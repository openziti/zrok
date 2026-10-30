package proctree

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/openziti/zrok/v2/cmd/zrok2/subordinate"
	"github.com/stretchr/testify/require"
)

func helperCommand(mode string) []string {
	return []string{os.Args[0], "-test.run=^TestProctreeHelper$", "--", mode}
}

func TestProctreeHelper(t *testing.T) {
	mode := ""
	for i, arg := range os.Args {
		if arg == "--" && i+1 < len(os.Args) {
			mode = os.Args[i+1]
			break
		}
	}
	if mode == "" {
		return
	}
	switch mode {
	case "eof":
		fmt.Fprintln(os.Stdout, `{"msg":"boot","token":"owned"}`)
		_, _ = io.Copy(io.Discard, os.Stdin)
	case "split":
		fmt.Fprint(os.Stdout, `{"msg":"bo`)
		var diagnostics sync.WaitGroup
		diagnostics.Add(1)
		go func() { defer diagnostics.Done(); fmt.Fprint(os.Stderr, "diagnostic fragment") }()
		fmt.Fprint(os.Stdout, "ot\",\"token\":\"owned\"}\n{\"msg\":\"access\"}\n")
		diagnostics.Wait()
		_, _ = io.Copy(io.Discard, os.Stdin)
	case "no-boot":
		fmt.Fprint(os.Stdout, `{"msg":"boot"`)
	case "terminal-error":
		fmt.Fprint(os.Stdout, "{\"msg\":\"boot\",\"token\":\"owned\"}\n{\"msg\":\"error\",\"error\":\"backend failed\"}\n")
		os.Exit(7)
	case "exit":
		os.Exit(23)
	case "ignore-eof":
		fmt.Fprintln(os.Stdout, `{"msg":"boot","token":"owned"}`)
		_, _ = io.Copy(io.Discard, os.Stdin)
		<-time.NewTimer(time.Hour).C
	case "hold-output":
		cmd := exec.Command(helperCommand("holder")[0], helperCommand("holder")[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			os.Exit(9)
		}
		fmt.Fprintln(os.Stdout, cmd.Process.Pid)
	case "holder":
		<-time.NewTimer(time.Hour).C
	default:
		os.Exit(8)
	}
	os.Exit(0)
}

func await(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(5 * time.Second):
		t.Fatal("child operation did not complete")
	}
}

func TestFailedSpawnAndActivation(t *testing.T) {
	for _, args := range [][]string{nil, {""}, {"/no-such-zrok-child"}} {
		c, err := StartChild(nil, args...)
		require.Error(t, err)
		require.Nil(t, c)
	}
	failure := errors.New("activation failed")
	var spawned *exec.Cmd
	c, err := startChild(false, ControlledOptions{DrainTimeout: 20 * time.Millisecond}, func(cmd *exec.Cmd) error { spawned = cmd; return failure }, helperCommand("ignore-eof")...)
	require.ErrorIs(t, err, failure)
	require.Nil(t, c)
	require.NotNil(t, spawned.ProcessState) // activation failure still joins the shared reaper.
}

func TestConcurrentWaitersShareExitResult(t *testing.T) {
	c, err := StartChild(nil, helperCommand("exit")...)
	require.NoError(t, err)
	results := make(chan error, 16)
	for range 16 {
		go func() { results <- WaitChild(c) }()
	}
	var first error
	for range 16 {
		select {
		case err := <-results:
			require.Error(t, err)
			if first == nil {
				first = err
			} else {
				require.Same(t, first, err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("waiter blocked")
		}
	}
	require.Same(t, first, WaitChild(c))
	require.Equal(t, 23, c.cmd.ProcessState.ExitCode())
	await(t, c.DrainWaitDone())
}

func TestControlledStreamsAndEOFStop(t *testing.T) {
	h := subordinate.NewMessageHandler()
	var token string
	h.BootHandler = func(_ string, msg subordinate.Message) { token, _ = msg["token"].(string) }
	var stderr bytes.Buffer
	c, err := StartControlledChild(ControlledOptions{Stdout: h.Tail, Stderr: func(data []byte) { stderr.Write(data) }, StdoutClosed: h.Close}, helperCommand("split")...)
	require.NoError(t, err)
	t.Cleanup(func() { _ = StopChild(c); _ = WaitChild(c) })
	await(t, h.BootComplete)
	require.Equal(t, "owned", token)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, StopControlledChild(ctx, c, 0, 0))
	require.NoError(t, c.RequestStop())
	require.NoError(t, WaitChild(c))
	await(t, c.DrainWaitDone())
	require.Equal(t, "diagnostic fragment", stderr.String())
}

func TestEOFBeforeBootAndTerminalError(t *testing.T) {
	for _, mode := range []string{"no-boot", "terminal-error"} {
		t.Run(mode, func(t *testing.T) {
			h := subordinate.NewMessageHandler()
			terminal := make(chan subordinate.Message, 1)
			h.MessageHandler = func(msg subordinate.Message) {
				if msg["msg"] == "error" {
					terminal <- msg
				}
			}
			c, err := StartControlledChild(ControlledOptions{Stdout: h.Tail, StdoutClosed: h.Close}, helperCommand(mode)...)
			require.NoError(t, err)
			await(t, c.Done())
			await(t, c.DrainWaitDone())
			await(t, h.BootComplete)
			if mode == "no-boot" {
				require.ErrorIs(t, h.Err(), io.ErrUnexpectedEOF)
			} else {
				require.Error(t, WaitChild(c))
				require.Equal(t, "backend failed", (<-terminal)["error"])
			}
		})
	}
}

func TestIgnoredEOFIsForceKilled(t *testing.T) {
	h := subordinate.NewMessageHandler()
	c, err := StartControlledChild(ControlledOptions{Stdout: h.Tail, StdoutClosed: h.Close}, helperCommand("ignore-eof")...)
	require.NoError(t, err)
	t.Cleanup(func() { _ = StopChild(c); _ = WaitChild(c) })
	await(t, h.BootComplete)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, StopControlledChild(ctx, c, 20*time.Millisecond, time.Second))
	require.Error(t, WaitChild(c))
	await(t, c.DrainWaitDone())
}

func TestBlockedLoggingCannotBlockExitOrDrainBound(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	defer close(release)
	c, err := StartControlledChild(ControlledOptions{Stdout: func([]byte) { close(entered); <-release }, DrainTimeout: 20 * time.Millisecond}, helperCommand("eof")...)
	require.NoError(t, err)
	t.Cleanup(func() { _ = StopChild(c); _ = WaitChild(c) })
	await(t, entered)
	require.NoError(t, c.RequestStop())
	await(t, c.Done())
	require.NoError(t, WaitChild(c))
	await(t, c.DrainWaitDone()) // callback is still blocked; this is the cutoff rather than callback completion.
}

func TestInheritedOutputWriterHasBoundedDrain(t *testing.T) {
	var mu sync.Mutex
	var output bytes.Buffer
	c, err := StartControlledChild(ControlledOptions{Stdout: func(data []byte) { mu.Lock(); defer mu.Unlock(); output.Write(data) }, DrainTimeout: 50 * time.Millisecond}, helperCommand("hold-output")...)
	require.NoError(t, err)
	await(t, c.Done())
	await(t, c.DrainWaitDone())
	require.NoError(t, WaitChild(c))
	mu.Lock()
	pidText := strings.TrimSpace(output.String())
	mu.Unlock()
	pid, err := strconv.Atoi(pidText)
	require.NoError(t, err)
	process, err := os.FindProcess(pid)
	require.NoError(t, err)
	require.NoError(t, process.Kill())
	_ = process.Release()
}

type finalRead struct{ done bool }

func (r *finalRead) Read(buf []byte) (int, error) {
	if r.done {
		return 0, io.EOF
	}
	r.done = true
	return copy(buf, "last bytes"), io.EOF
}
func (*finalRead) Close() error { return nil }

func TestReaderOwnsChunksAndDeliversFinalBytes(t *testing.T) {
	for _, input := range []io.ReadCloser{&finalRead{}, io.NopCloser(strings.NewReader(strings.Repeat("a", 64*1024) + "b"))} {
		output := make(chan []byte)
		go reader(input, output, make(chan struct{}))
		var chunks [][]byte
		for chunk := range output {
			chunks = append(chunks, chunk)
		}
		if len(chunks) == 1 {
			require.Equal(t, "last bytes", string(chunks[0]))
		} else {
			require.Equal(t, strings.Repeat("a", 64*1024), string(chunks[0]))
			require.Equal(t, "b", string(chunks[1]))
		}
	}
}

func TestUnconfirmedForcedExitIsFailure(t *testing.T) {
	input, parent, err := os.Pipe()
	require.NoError(t, err)
	defer input.Close()
	defer parent.Close()
	c := &Child{controlled: true, stdin: parent, exited: make(chan struct{})}
	killed := false
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = stopControlledChild(ctx, c, time.Millisecond, time.Millisecond, func(*Child) error { killed = true; return nil })
	require.True(t, killed)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.ErrorContains(t, err, "unconfirmed")
}

func TestCancelledStopStillForcesTermination(t *testing.T) {
	input, parent, err := os.Pipe()
	require.NoError(t, err)
	defer input.Close()
	defer parent.Close()
	c := &Child{controlled: true, stdin: parent, exited: make(chan struct{})}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	killed := false
	err = stopControlledChild(ctx, c, time.Second, time.Second, func(*Child) error { killed = true; return nil })
	require.True(t, killed)
	require.ErrorIs(t, err, context.Canceled)
}
