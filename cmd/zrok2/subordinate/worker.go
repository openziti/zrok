package subordinate

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// WorkerContext registers both terminal signals and parent-pipe EOF before the caller constructs a backend.
// input must be a closeable pipe: cleanup closes the control reader and joins the watcher when the backend returns first.
func WorkerContext(parent context.Context, input io.ReadCloser) (context.Context, func(), error) {
	control, err := workerInput(input)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to prepare worker control input: %w", err)
	}
	ctx, stop := signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = io.Copy(io.Discard, control)
		stop()
	}()
	var once sync.Once
	return ctx, func() { once.Do(func() { stop(); _ = control.Close(); <-done }) }, nil
}
