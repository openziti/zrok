//go:build agent_smoke && unix

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	logHeadLimit = 64 << 10
	logTailLimit = 192 << 10
	ioWaitDelay  = 2 * time.Second
)

// boundedLog retains the head and tail of a child's combined output in memory. nothing reaches the harness's own
// output; the report writes it out after redaction.
type boundedLog struct {
	mu      sync.Mutex
	head    []byte
	tail    []byte
	dropped int64
}

func (l *boundedLog) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	n := len(p)
	if room := logHeadLimit - len(l.head); room > 0 {
		take := min(room, len(p))
		l.head = append(l.head, p[:take]...)
		p = p[take:]
	}
	if len(p) > 0 {
		l.tail = append(l.tail, p...)
		if over := len(l.tail) - logTailLimit; over > logTailLimit {
			l.dropped += int64(over)
			l.tail = append([]byte(nil), l.tail[over:]...)
		}
	}
	return n, nil
}

func (l *boundedLog) contents() []byte {
	l.mu.Lock()
	defer l.mu.Unlock()
	tail := l.tail
	dropped := l.dropped
	if over := len(tail) - logTailLimit; over > 0 {
		dropped += int64(over)
		tail = tail[over:]
	}
	var out bytes.Buffer
	out.Write(l.head)
	if dropped > 0 {
		_, _ = fmt.Fprintf(&out, "\n[... %d bytes omitted ...]\n", dropped)
	}
	out.Write(tail)
	return out.Bytes()
}

// child is a subprocess in its own process group. its single reaper goroutine closes done; everything else waits
// on done.
type child struct {
	cmd     *exec.Cmd
	done    chan struct{}
	waitErr error
}

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
	c := &child{cmd: cmd, done: make(chan struct{})}
	go func() {
		c.waitErr = cmd.Wait()
		close(c.done)
	}()
	return c, nil
}

func (c *child) pid() int {
	return c.cmd.Process.Pid
}

func (c *child) exited() bool {
	select {
	case <-c.done:
		return true
	default:
		return false
	}
}

// exitCode is valid once done is closed; a signal-terminated process reports -1.
func (c *child) exitCode() int {
	if c.cmd.ProcessState == nil {
		return -1
	}
	return c.cmd.ProcessState.ExitCode()
}

func (c *child) interrupt() error {
	return syscall.Kill(-c.pid(), syscall.SIGINT)
}

// kill force-terminates the child's process group and the groups of its direct children, which the agent places
// in their own groups, then waits for the reaper.
func (c *child) kill(bound time.Duration) error {
	if c.exited() {
		return nil
	}
	descendants := childPIDs(c.pid())
	_ = syscall.Kill(-c.pid(), syscall.SIGKILL)
	for _, pid := range descendants {
		killGroupOrProcess(pid)
	}
	select {
	case <-c.done:
		return nil
	case <-time.After(bound):
		return fmt.Errorf("process '%d' was not reaped within %v of being killed", c.pid(), bound)
	}
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
