//go:build agent_smoke

package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"sync"
	"time"
)

const (
	posixPlatform   = "posix"
	windowsPlatform = "windows"
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

// child is a subprocess started in its own interrupt scope: a process group on POSIX, a console process group
// inside a job object on Windows. its single reaper goroutine closes done; everything else waits on done.
type child struct {
	cmd     *exec.Cmd
	done    chan struct{}
	waitErr error
	tree    processTree
}

// reap starts the single goroutine that owns cmd.Wait.
func reap(cmd *exec.Cmd, tree processTree) *child {
	c := &child{cmd: cmd, done: make(chan struct{}), tree: tree}
	go func() {
		c.waitErr = cmd.Wait()
		close(c.done)
	}()
	return c
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

// recordPlatform names the platform whose process identity a record holds. records written before the field
// existed are POSIX.
func recordPlatform(record processRecord) string {
	if record.Platform == "" {
		return posixPlatform
	}
	return record.Platform
}

// inspectError marks a failure to read a recorded process's identity, as distinct from a failure to confirm its
// termination.
type inspectError struct {
	err error
}

func (e *inspectError) Error() string { return e.err.Error() }
func (e *inspectError) Unwrap() error { return e.err }

// recordOutcome is how a recorded process from an earlier invocation was resolved.
type recordOutcome int

const (
	recordNotRunning recordOutcome = iota
	recordReplaced
	recordTerminated
)
