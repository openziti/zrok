package proctree

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/michaelquigley/df/dl"
)

const (
	DefaultDrainTimeout = time.Second
	DefaultGracefulStop = 5 * time.Second
	DefaultForcedExit   = 5 * time.Second
)

type TailFunction func(data []byte)

// ControlledOptions separates protocol stdout from diagnostic stderr.
type ControlledOptions struct {
	Stdout       TailFunction
	Stderr       TailFunction
	StdoutClosed func()
	DrainTimeout time.Duration
}

type Child struct {
	cmd           *exec.Cmd
	controlled    bool
	stdin         *os.File
	stopOnce      sync.Once
	stopErr       error
	exited        chan struct{}
	drainWaitDone chan struct{}
	exitErr       error
}

// Done closes as soon as the process has been reaped, independently of output callbacks.
func (c *Child) Done() <-chan struct{} { return c.exited }

// DrainWaitDone closes when the bounded drain wait ends, including timeout with a callback still running.
func (c *Child) DrainWaitDone() <-chan struct{} { return c.drainWaitDone }

// StartChild retains the merged-output, uncontrolled launch used by legacy commands.
func StartChild(tail TailFunction, args ...string) (*Child, error) {
	return startChild(false, ControlledOptions{Stdout: tail, Stderr: tail}, finishStart, args...)
}

// StartControlledChild gives a public worker a parent-owned stdin lifetime and separate output streams.
func StartControlledChild(opts ControlledOptions, args ...string) (*Child, error) {
	return startChild(true, opts, finishStart, args...)
}

func startChild(controlled bool, opts ControlledOptions, activate func(*exec.Cmd) error, args ...string) (*Child, error) {
	if len(args) == 0 || args[0] == "" {
		return nil, errors.New("missing child executable")
	}
	if opts.DrainTimeout < 0 {
		return nil, errors.New("output drain timeout must not be negative")
	}
	if opts.DrainTimeout == 0 {
		opts.DrainTimeout = DefaultDrainTimeout
	}
	dl.Infof("executing '%v'", strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:]...)
	if err := configureChild(cmd, controlled); err != nil {
		return nil, err
	}
	var ends []*os.File
	closeEnds := func() {
		for _, end := range ends {
			_ = end.Close()
		}
	}
	stdout, outWrite, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	ends = append(ends, stdout, outWrite)
	stderr, errWrite, err := os.Pipe()
	if err != nil {
		closeEnds()
		return nil, err
	}
	ends = append(ends, stderr, errWrite)
	cmd.Stdout = outWrite
	cmd.Stderr = errWrite
	c := &Child{cmd: cmd, controlled: controlled, exited: make(chan struct{}), drainWaitDone: make(chan struct{})}
	var inRead *os.File
	if controlled {
		inRead, c.stdin, err = os.Pipe()
		if err != nil {
			closeEnds()
			return nil, err
		}
		ends = append(ends, inRead, c.stdin)
		cmd.Stdin = inRead
	}
	if err := cmd.Start(); err != nil {
		closeEnds()
		return nil, err
	}
	// these explicit pipe writers are outside exec.Cmd's copying machinery: Wait cannot close their readers.
	_ = outWrite.Close()
	_ = errWrite.Close()
	if inRead != nil {
		_ = inRead.Close()
	}
	cancel := make(chan struct{})
	outStream, errStream := make(chan []byte), make(chan []byte)
	drained := make(chan struct{})
	go reader(stdout, outStream, cancel)
	go reader(stderr, errStream, cancel)
	go func() { defer close(drained); combine(outStream, errStream, cancel, opts) }()
	// there is exactly one Cmd.Wait, even when platform activation fails after spawning.
	go func() {
		c.exitErr = cmd.Wait()
		close(c.exited)
		if controlled {
			_ = c.RequestStop()
		}
		timer := time.NewTimer(opts.DrainTimeout)
		defer timer.Stop()
		select {
		case <-drained:
		case <-timer.C:
		}
		close(cancel)
		_ = stdout.Close()
		_ = stderr.Close()
		close(c.drainWaitDone)
	}()
	if err := activate(cmd); err != nil {
		_ = StopChild(c)
		_ = WaitChild(c)
		<-c.DrainWaitDone()
		return nil, err
	}
	return c, nil
}

// RequestStop requests graceful shutdown by closing stdin; it does not claim that the worker has exited.
func (c *Child) RequestStop() error {
	if !c.controlled {
		return errors.New("child does not support stdin stop control")
	}
	c.stopOnce.Do(func() { c.stopErr = c.stdin.Close() })
	return c.stopErr
}

func WaitChild(c *Child) error { <-c.exited; return c.exitErr }

// StopChild is the independent force-kill operation used by legacy and controlled children.
func StopChild(c *Child) error {
	select {
	case <-c.exited:
		return nil
	default:
	}
	err := killChild(c)
	if errors.Is(err, os.ErrProcessDone) {
		return nil
	}
	return err
}

// StopControlledChild waits for graceful exit, then forces termination and confirms exit within a second bound.
func StopControlledChild(ctx context.Context, c *Child, graceful, forced time.Duration) error {
	return stopControlledChild(ctx, c, graceful, forced, StopChild)
}

func stopControlledChild(ctx context.Context, c *Child, graceful, forced time.Duration, kill func(*Child) error) error {
	if graceful == 0 {
		graceful = DefaultGracefulStop
	}
	if forced == 0 {
		forced = DefaultForcedExit
	}
	if graceful < 0 || forced < 0 {
		return errors.New("child stop timeouts must not be negative")
	}
	if err := c.RequestStop(); err != nil {
		return err
	}
	if waitExit(ctx, c, graceful) == nil {
		return nil
	}
	if err := kill(c); err != nil {
		return fmt.Errorf("error killing worker: %w", err)
	}
	if err := waitExit(ctx, c, forced); err != nil {
		return fmt.Errorf("worker exit is unconfirmed: %w", err)
	}
	return nil
}

func waitExit(ctx context.Context, c *Child, bound time.Duration) error {
	// an already confirmed exit wins over a cancelled wait.
	select {
	case <-c.exited:
		return nil
	default:
	}
	timer := time.NewTimer(bound)
	defer timer.Stop()
	select {
	case <-c.exited:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return context.DeadlineExceeded
	}
}

func combine(stdout, stderr <-chan []byte, cancel <-chan struct{}, opts ControlledOptions) {
	for stdout != nil || stderr != nil {
		select {
		case <-cancel:
			return
		case data, ok := <-stdout:
			if !ok {
				stdout = nil
				if opts.StdoutClosed != nil {
					opts.StdoutClosed()
				}
				continue
			}
			if opts.Stdout != nil {
				opts.Stdout(data)
			}
		case data, ok := <-stderr:
			if !ok {
				stderr = nil
				continue
			}
			if opts.Stderr != nil {
				opts.Stderr(data)
			}
		}
	}
}

func reader(r io.ReadCloser, out chan<- []byte, cancel <-chan struct{}) {
	defer close(out)
	defer r.Close()
	buf := make([]byte, 64*1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			chunk := append([]byte(nil), buf[:n]...)
			select {
			case out <- chunk:
			case <-cancel:
				return
			}
		}
		if err != nil {
			if !errors.Is(err, io.EOF) && !errors.Is(err, os.ErrClosed) {
				dl.Warnf("error reading child output: %v", err)
			}
			return
		}
	}
}
