package agent

import (
	"context"
	"errors"
	"fmt"
)

func (a *Agent) release(ctx context.Context, id string, isAccess bool) error {
	if id == "" {
		return errors.New("missing operation token or failure id")
	}
	if err := a.waitForImport(ctx); err != nil {
		return err
	}
	a.mu.Lock()
	if a.stopping {
		a.mu.Unlock()
		return errors.New("agent is stopping")
	}
	if err := ctx.Err(); err != nil {
		a.mu.Unlock()
		return err
	}
	var selected *operation
	for _, op := range a.operations {
		_, access := op.request.(*AccessPrivateRequest)
		if access == isAccess && (op.id == id || op.failureID == id || op.resourceToken() == id && id != "") {
			selected = op
			break
		}
	}
	if selected == nil {
		a.mu.Unlock()
		return fmt.Errorf("agent has no operation with token or failure id '%v'", id)
	}
	a.beginWorkLocked()
	selected.released = true
	selected.desired = false
	selected.stop = true
	if selected.start != nil {
		selected.start.cancel()
	}
	attempt := a.claimCleanupLocked(selected)
	a.mu.Unlock()
	defer a.endWork()
	saveErr := a.SaveRegistry()
	select {
	case <-attempt.done:
		return errors.Join(saveErr, attempt.err)
	case <-ctx.Done():
		return errors.Join(saveErr, ctx.Err())
	}
}
