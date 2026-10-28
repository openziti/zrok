package subordinate

import (
	"context"
	"io"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWorkerEOFAndCleanup(t *testing.T) {
	input, parent := io.Pipe()
	ctx, stop, err := WorkerContext(context.Background(), input)
	require.NoError(t, err)
	select {
	case <-ctx.Done():
		t.Fatal("worker stopped before EOF")
	default:
	}
	require.NoError(t, parent.Close())
	select {
	case <-ctx.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("worker did not stop on EOF")
	}
	stop()
	input, parent = io.Pipe()
	_, stop, err = WorkerContext(context.Background(), input)
	require.NoError(t, err)
	stop() // a backend failing first also joins the otherwise blocked stdin watcher.
	_ = parent.Close()
}

func TestWorkerSignalHelper(t *testing.T) {
	if os.Getenv("ZROK_WORKER_SIGNAL_HELPER") != "1" {
		return
	}
	ctx, stop, err := WorkerContext(context.Background(), os.Stdin)
	require.NoError(t, err)
	defer stop()
	_, _ = os.Stdout.Write([]byte("ready\n"))
	<-ctx.Done()
}

func TestWorkerTerminalSignal(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows Process.Signal cannot send terminal SIGINT; EOF control is covered by portable helper tests")
	}
	for _, signal := range []os.Signal{os.Interrupt, syscall.SIGTERM} {
		t.Run(signal.String(), func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestWorkerSignalHelper$")
			cmd.Env = append(os.Environ(), "ZROK_WORKER_SIGNAL_HELPER=1")
			input, err := cmd.StdinPipe()
			require.NoError(t, err)
			output, err := cmd.StdoutPipe()
			require.NoError(t, err)
			require.NoError(t, cmd.Start())
			t.Cleanup(func() { _ = input.Close(); _ = cmd.Process.Kill() })
			ready := make([]byte, len("ready\n"))
			_, err = io.ReadFull(output, ready)
			require.NoError(t, err)
			require.Equal(t, "ready\n", string(ready))
			require.NoError(t, cmd.Process.Signal(signal))
			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			select {
			case err := <-done:
				require.NoError(t, err)
			case <-time.After(5 * time.Second):
				t.Fatal("worker ignored terminal signal")
			}

		})
	}
}
