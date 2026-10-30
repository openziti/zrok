//go:build agent_smoke && unix

package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	// localPlatform is the identity format this build writes and may act on
	localPlatform = posixPlatform
	// stopEvent is the graceful-stop event the harness delivers; stopScope is who receives it. the agent starts its
	// workers in their own process groups, so they do not receive it.
	stopEvent = "SIGINT"
	stopScope = "the agent's process group"
)

// stopDescription names the event and who receives it, for the report.
func stopDescription() string {
	return stopEvent + " to " + stopScope
}

// processTree is empty on POSIX: the child's process group is its pid.
type processTree struct{}

func startChild(binary string, args, env []string, out *boundedLog) (*child, error) {
	cmd := exec.Command(binary, args...)
	cmd.Env = env
	cmd.Stdout = out
	cmd.Stderr = out
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.WaitDelay = ioWaitDelay
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return reap(cmd, processTree{}), nil
}

// interrupt delivers SIGINT to the child's process group, as a terminal's Ctrl-C does.
func (c *child) interrupt() error {
	return syscall.Kill(-c.pid(), syscall.SIGINT)
}

// kill force-terminates the child's process group and the groups of its direct children, which the agent places
// in their own groups, then waits for the reaper. it reports what it terminated, or nothing when the child had
// already exited.
func (c *child) kill(bound time.Duration) (string, error) {
	if c.exited() {
		return "", nil
	}
	descendants := childPIDs(c.pid())
	_ = syscall.Kill(-c.pid(), syscall.SIGKILL)
	for _, pid := range descendants {
		killGroupOrProcess(pid)
	}
	select {
	case <-c.done:
		return fmt.Sprintf("pid %d", c.pid()), nil
	case <-time.After(bound):
		return "", fmt.Errorf("process '%d' was not reaped within %v of being killed", c.pid(), bound)
	}
}

// identify records a started process's identity as ps reports it: start time and the command line it was given.
func identify(ctx context.Context, pid int, binary string, args []string) processRecord {
	record := processRecord{PID: pid, Platform: localPlatform, Command: strings.Join(append([]string{binary}, args...), " ")}
	if started, _, alive, err := processIdentity(ctx, pid); err == nil && alive {
		record.Started = started
	}
	return record
}

// stopRecorded resolves a process recorded by an earlier invocation. it is signalled only while ps still reports
// the recorded start time and command line, and is reported terminated only once confirmed gone.
func stopRecorded(ctx context.Context, record processRecord, bound, poll time.Duration) (recordOutcome, error) {
	started, command, alive, err := processIdentity(ctx, record.PID)
	switch {
	case err != nil:
		return 0, &inspectError{err}
	case !alive:
		return recordNotRunning, nil
	case record.Started == "" || started != record.Started || command != record.Command:
		return recordReplaced, nil
	}
	descendants := childPIDs(record.PID)
	_ = syscall.Kill(-record.PID, syscall.SIGKILL)
	for _, pid := range descendants {
		killGroupOrProcess(pid)
	}
	if err := awaitGone(ctx, record, bound, poll); err != nil {
		return 0, err
	}
	return recordTerminated, nil
}

// executableFile reports whether a binary under test can be executed: a regular file with an execute bit.
func executableFile(_ string, info fs.FileInfo) bool {
	return info.Mode().IsRegular() && info.Mode().Perm()&0o111 != 0
}

func killGroupOrProcess(pid int) {
	if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil {
		_ = syscall.Kill(pid, syscall.SIGKILL)
	}
}

func childPIDs(parent int) []int {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, _ := exec.CommandContext(ctx, "pgrep", "-P", strconv.Itoa(parent)).Output()
	var pids []int
	for _, field := range strings.Fields(string(out)) {
		if pid, err := strconv.Atoi(field); err == nil {
			pids = append(pids, pid)
		}
	}
	return pids
}

// processIdentity reports a live pid's start time and command line as ps prints them, or alive false.
func processIdentity(ctx context.Context, pid int) (started, command string, alive bool, err error) {
	if started, alive, err = psField(ctx, pid, "lstart"); err != nil || !alive {
		return "", "", alive, err
	}
	if command, alive, err = psField(ctx, pid, "args"); err != nil || !alive {
		return "", "", alive, err
	}
	return started, command, true, nil
}

func psField(ctx context.Context, pid int, field string) (string, bool, error) {
	cmd := exec.CommandContext(ctx, "ps", "-o", field+"=", "-p", strconv.Itoa(pid))
	cmd.Env = append(os.Environ(), "LC_ALL=C")
	out, err := cmd.Output()
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() == 1 {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return strings.TrimSpace(string(out)), true, nil
}

// processGone reports whether a recorded process has terminated: its pid is free, names another process, or is
// a zombie awaiting its parent.
func processGone(ctx context.Context, record processRecord) (bool, error) {
	started, command, alive, err := processIdentity(ctx, record.PID)
	if err != nil {
		return false, err
	}
	if !alive || started != record.Started || command != record.Command {
		return true, nil
	}
	stat, alive, err := psField(ctx, record.PID, "stat")
	if err != nil {
		return false, err
	}
	return !alive || strings.HasPrefix(stat, "Z"), nil
}

// awaitGone polls until a killed recorded process is confirmed gone, within bound. inspection failures are
// retried until the bound expires.
func awaitGone(ctx context.Context, record processRecord, bound, poll time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, bound)
	defer cancel()
	for {
		gone, err := processGone(ctx, record)
		if err == nil && gone {
			return nil
		}
		select {
		case <-ctx.Done():
			if err != nil {
				return fmt.Errorf("termination not confirmed within %v of being killed: %w", bound, err)
			}
			return fmt.Errorf("not confirmed terminated within %v of being killed", bound)
		case <-time.After(poll):
		}
	}
}
