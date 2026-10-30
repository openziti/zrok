//go:build agent_smoke && windows

package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/kolesnikovae/go-winjob"
	"golang.org/x/sys/windows"
)

const (
	// localPlatform is the identity format this build writes and may act on
	localPlatform = windowsPlatform
	// stopEvent is the graceful-stop event the harness delivers; stopScope is who receives it. go's runtime routes
	// ctrl-break to os.Interrupt. the agent starts its workers without a new process group, so they share its group
	// and receive the event too.
	stopEvent = "CTRL_BREAK_EVENT"
	stopScope = "the agent's console process group, which its workers share"

	recordedAccess = windows.PROCESS_QUERY_LIMITED_INFORMATION | windows.PROCESS_TERMINATE | windows.SYNCHRONIZE
)

var (
	kernel32                  = windows.NewLazySystemDLL("kernel32.dll")
	procAllocConsole          = kernel32.NewProc("AllocConsole")
	procGetConsoleProcessList = kernel32.NewProc("GetConsoleProcessList")

	consoleOnce   sync.Once
	consoleOrigin string
	consoleErr    error

	// inspectProcess and terminateProcess act on a held handle to a recorded process; tests replace them to make
	// inspection fail or termination go unconfirmed.
	inspectProcess   = processIdentity
	terminateProcess = func(h windows.Handle) error { return windows.TerminateProcess(h, 1) }
)

// processTree is the job object that contains a child and everything it starts, including workers that outlive
// it. the job is killed when its last handle closes, so nothing the harness started survives the harness. the
// agent nests its own worker job inside it.
type processTree struct {
	job *winjob.JobObject
}

// ensureConsole gives the harness a console that its children inherit, which console control events require.
// a runner step may have none, so one is allocated when absent. events are only ever sent to a child's own
// process group, never to group 0, which would reach the harness and the runner's shell.
func ensureConsole() error {
	consoleOnce.Do(func() {
		var pids [1]uint32
		if n, _, _ := procGetConsoleProcessList.Call(uintptr(unsafe.Pointer(&pids[0])), 1); n != 0 {
			consoleOrigin = "inherited console"
			return
		}
		if ok, _, err := procAllocConsole.Call(); ok == 0 {
			consoleErr = fmt.Errorf("unable to allocate a console for console control events: %w", err)
			return
		}
		consoleOrigin = "allocated console"
	})
	return consoleErr
}

// stopDescription names the event, who receives it and where the console came from, for the report.
func stopDescription() string {
	return fmt.Sprintf("%v to %v (%v)", stopEvent, stopScope, consoleOrigin)
}

// startChild starts the child suspended in a new console process group, places it in its own job object, then
// resumes it, so it runs no code outside the job.
func startChild(binary string, args, env []string, out *boundedLog) (*child, error) {
	if err := ensureConsole(); err != nil {
		return nil, err
	}
	job, err := winjob.Create("", winjob.LimitKillOnJobClose)
	if err != nil {
		return nil, fmt.Errorf("unable to create a job object: %w", err)
	}
	cmd := exec.Command(binary, args...)
	cmd.Env = env
	cmd.Stdout = out
	cmd.Stderr = out
	cmd.SysProcAttr = &windows.SysProcAttr{CreationFlags: windows.CREATE_NEW_PROCESS_GROUP | windows.CREATE_SUSPENDED}
	cmd.WaitDelay = ioWaitDelay
	if err := cmd.Start(); err != nil {
		_ = job.Close()
		return nil, err
	}
	c := reap(cmd, processTree{job: job})
	if err := job.Assign(cmd.Process); err != nil {
		// still suspended and outside the job, so it has run nothing and can be terminated directly
		_ = cmd.Process.Kill()
		<-c.done
		_ = job.Close()
		return nil, fmt.Errorf("unable to assign pid %d to its job object: %w", c.pid(), err)
	}
	if err := winjob.ResumeProcess(cmd.Process.Pid); err != nil {
		_ = job.Terminate()
		<-c.done
		_ = job.Close()
		return nil, fmt.Errorf("unable to resume pid %d: %w", c.pid(), err)
	}
	return c, nil
}

// interrupt delivers CTRL_BREAK_EVENT to the child's console process group, whose id is the child's pid.
func (c *child) interrupt() error {
	return windows.GenerateConsoleCtrlEvent(windows.CTRL_BREAK_EVENT, uint32(c.pid()))
}

// kill terminates the child's job, which holds the child and every process it started, then requires the
// reaper and an empty job. it reports what it terminated, or nothing when the child exited and left nothing.
func (c *child) kill(bound time.Duration) (string, error) {
	active, err := c.tree.active()
	if err != nil {
		return "", fmt.Errorf("unable to inspect the job object of pid %d: %w", c.pid(), err)
	}
	if c.exited() && active == 0 {
		return "", nil
	}
	if err := c.tree.job.Terminate(); err != nil {
		return "", fmt.Errorf("unable to terminate the job object of pid %d: %w", c.pid(), err)
	}
	deadline := time.Now().Add(bound)
	select {
	case <-c.done:
	case <-time.After(bound):
		return "", fmt.Errorf("process '%d' was not reaped within %v of its job being terminated", c.pid(), bound)
	}
	for {
		remaining, err := c.tree.active()
		if err == nil && remaining == 0 {
			return fmt.Sprintf("pid %d with %d processes in its job", c.pid(), active), nil
		}
		if time.Now().After(deadline) {
			if err != nil {
				return "", fmt.Errorf("termination of the job object of pid %d not confirmed within %v: %w", c.pid(), bound, err)
			}
			return "", fmt.Errorf("%d processes remain in the job object of pid %d %v after it was terminated", remaining, c.pid(), bound)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func (t processTree) active() (uint32, error) {
	counters, err := t.job.Counters()
	if err != nil {
		return 0, err
	}
	return counters.ActiveProcesses, nil
}

// identify records a started process's identity: its creation time and executable path, read from a handle.
func identify(_ context.Context, pid int, binary string, _ []string) processRecord {
	record := processRecord{PID: pid, Platform: localPlatform, Command: canonicalPath(binary)}
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return record
	}
	defer func() { _ = windows.CloseHandle(h) }()
	if started, _, err := processIdentity(h); err == nil {
		record.Started = started
	}
	return record
}

// stopRecorded resolves a process recorded by an earlier invocation through a handle held for the whole
// decision, so a pid reused after the check is never acted on. it terminates the process and its descendants
// only while the handle shows the recorded creation time and executable, and reports termination only once
// every handle is signalled.
func stopRecorded(ctx context.Context, record processRecord, bound, _ time.Duration) (recordOutcome, error) {
	h, err := windows.OpenProcess(recordedAccess, false, uint32(record.PID))
	if errors.Is(err, windows.ERROR_INVALID_PARAMETER) {
		return recordNotRunning, nil
	}
	if err != nil {
		return 0, &inspectError{err}
	}
	defer func() { _ = windows.CloseHandle(h) }()
	started, command, err := inspectProcess(h)
	if err != nil {
		return 0, &inspectError{err}
	}
	if signalled(h) {
		return recordNotRunning, nil
	}
	if record.Started == "" || started != record.Started || !strings.EqualFold(command, canonicalPath(record.Command)) {
		return recordReplaced, nil
	}
	descendants, err := descendantHandles(record.PID, started)
	if err != nil {
		return 0, &inspectError{err}
	}
	defer func() {
		for _, d := range descendants {
			_ = windows.CloseHandle(d)
		}
	}()
	handles := append([]windows.Handle{h}, descendants...)
	for _, handle := range handles {
		_ = terminateProcess(handle)
	}
	if err := awaitHandles(ctx, handles, bound); err != nil {
		return 0, err
	}
	return recordTerminated, nil
}

// processIdentity reads a process's creation time, as 100ns intervals since 1601, and executable path.
func processIdentity(h windows.Handle) (started, command string, err error) {
	var creation, exit, kernel, user windows.Filetime
	if err := windows.GetProcessTimes(h, &creation, &exit, &kernel, &user); err != nil {
		return "", "", err
	}
	buf := make([]uint16, windows.MAX_LONG_PATH)
	size := uint32(len(buf))
	if err := windows.QueryFullProcessImageName(h, 0, &buf[0], &size); err != nil {
		return "", "", err
	}
	stamp := uint64(creation.HighDateTime)<<32 | uint64(creation.LowDateTime)
	return strconv.FormatUint(stamp, 10), canonicalPath(windows.UTF16ToString(buf[:size])), nil
}

// descendantHandles opens every live descendant of a process. a pid counts as a child only when it was created
// after its recorded parent, which rules out a parent pid that has since been reused.
func descendantHandles(root int, rootStarted string) ([]windows.Handle, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer func() { _ = windows.CloseHandle(snapshot) }()
	children := make(map[uint32][]uint32)
	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	for err = windows.Process32First(snapshot, &entry); err == nil; err = windows.Process32Next(snapshot, &entry) {
		children[entry.ParentProcessID] = append(children[entry.ParentProcessID], entry.ProcessID)
	}
	if !errors.Is(err, windows.ERROR_NO_MORE_FILES) {
		return nil, err
	}

	var handles []windows.Handle
	type parent struct {
		pid     uint32
		started uint64
	}
	rootStamp, _ := strconv.ParseUint(rootStarted, 10, 64)
	queue := []parent{{uint32(root), rootStamp}}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		for _, pid := range children[p.pid] {
			h, err := windows.OpenProcess(recordedAccess, false, pid)
			if err != nil {
				continue
			}
			started, _, err := processIdentity(h)
			stamp, _ := strconv.ParseUint(started, 10, 64)
			if err != nil || stamp < p.started || signalled(h) {
				_ = windows.CloseHandle(h)
				continue
			}
			handles = append(handles, h)
			queue = append(queue, parent{pid, stamp})
		}
	}
	return handles, nil
}

// signalled reports whether a process handle shows the process has exited.
func signalled(h windows.Handle) bool {
	event, err := windows.WaitForSingleObject(h, 0)
	return err == nil && event == windows.WAIT_OBJECT_0
}

// awaitHandles requires every process to exit within bound.
func awaitHandles(ctx context.Context, handles []windows.Handle, bound time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, bound)
	defer cancel()
	deadline, _ := ctx.Deadline()
	for _, h := range handles {
		wait := max(time.Until(deadline), 0)
		event, err := windows.WaitForSingleObject(h, uint32(wait.Milliseconds()))
		if err != nil {
			return fmt.Errorf("termination not confirmed within %v of being killed: %w", bound, err)
		}
		if event != windows.WAIT_OBJECT_0 {
			return fmt.Errorf("not confirmed terminated within %v of being killed", bound)
		}
	}
	return nil
}

// canonicalPath expands short path components, so a path reported by the system compares equal to the one the
// harness was given.
func canonicalPath(path string) string {
	path = filepath.Clean(path)
	in, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return path
	}
	buf := make([]uint16, windows.MAX_LONG_PATH)
	n, err := windows.GetLongPathName(in, &buf[0], uint32(len(buf)))
	if err != nil || n == 0 || int(n) > len(buf) {
		return path
	}
	return windows.UTF16ToString(buf[:n])
}

// executableFile reports whether a binary under test can be executed: a regular .exe file. windows has no
// execute bit; the extension is what CreateProcess requires.
func executableFile(path string, info fs.FileInfo) bool {
	return info.Mode().IsRegular() && strings.EqualFold(filepath.Ext(path), ".exe")
}
