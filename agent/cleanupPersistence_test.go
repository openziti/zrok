package agent

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/openziti/zrok/v2/agent/agentGrpc"
	"github.com/openziti/zrok/v2/agent/proctree"
	"github.com/stretchr/testify/require"
)

func TestReleaseDuringCleanupPersistence(t *testing.T) {
	for _, outcome := range []string{"success", "delete failure", "save failure"} {
		t.Run(outcome, func(t *testing.T) {
			a := newLifecycleAgent(t)
			initial, saving, finish := make(chan struct{}), make(chan struct{}), make(chan struct{})
			var unblock sync.Once
			defer unblock.Do(func() { close(finish) })
			deleteFailure := errors.New("controller cleanup failed")
			saveFailure := errors.New("cleanup persistence failed")
			var writes, launches, deletes atomic.Int32
			a.deps.startChild = func(tail proctree.TailFunction, _ ...string) (*proctree.Child, error) {
				launches.Add(1)
				return lifecycleChild(tail, "boot")
			}
			a.deps.deleteAccess = func(context.Context, string, string) error {
				deletes.Add(1)
				if outcome == "delete failure" {
					return deleteFailure
				}
				return nil
			}
			a.deps.saveRegistry = func(r *Registry, path string) error {
				n := writes.Add(1)
				if n == 2 {
					close(saving)
					<-finish
					if outcome == "save failure" {
						return saveFailure
					}
				}
				err := r.Save(path)
				if n == 1 {
					close(initial)
				}
				return err
			}
			_, err := a.AccessPrivate(&AccessPrivateRequest{ShareToken: "shared"})
			require.NoError(t, err)
			waitLifecycle(t, initial)
			op := onlyOperation(t, a)
			a.mu.Lock()
			child := op.child
			a.mu.Unlock()
			require.NoError(t, proctree.StopChild(child))
			waitLifecycle(t, saving)

			intentRecorded := make(chan struct{})
			var notified sync.Once
			a.mu.Lock()
			id := op.failureID
			attempt := op.cleanup
			cancel := op.start.cancel
			// release cancels startup after recording intent under the lifecycle mutex.
			op.start.cancel = func() { cancel(); notified.Do(func() { close(intentRecorded) }) }
			a.mu.Unlock()
			require.NotEmpty(t, id)
			released := asyncResult(func() error { return a.ReleaseAccess(id) })
			waitLifecycle(t, intentRecorded)
			a.mu.Lock()
			releaseIntent, restartIntent, joinedAttempt := op.released, op.desired, op.cleanup
			a.mu.Unlock()
			require.True(t, releaseIntent)
			require.False(t, restartIntent)
			require.Same(t, attempt, joinedAttempt)
			select {
			case <-attempt.done:
				t.Fatal("cleanup completed before its persistence barrier opened")
			default:
			}
			a.retryDue(time.Now().Add(time.Hour))
			require.EqualValues(t, 1, launches.Load())
			unblock.Do(func() { close(finish) })

			err = lifecycleResult(t, released)
			switch outcome {
			case "success":
				require.NoError(t, err)
				require.NoError(t, attempt.err)
			case "delete failure":
				require.ErrorIs(t, err, deleteFailure)
				require.ErrorIs(t, attempt.err, deleteFailure)
			case "save failure":
				require.ErrorIs(t, err, saveFailure)
				require.ErrorIs(t, attempt.err, saveFailure)
			}
			status, err := (&agentGrpcImpl{agent: a}).Status(context.Background(), &agentGrpc.StatusRequest{})
			require.NoError(t, err)
			require.Empty(t, savedRegistry(t, a).PrivateAccesses)
			a.mu.Lock()
			currentAttempt, registered := op.cleanup, a.operations[op.id]
			token, phase, remainingChild := op.resourceToken(), op.phase, op.child
			a.mu.Unlock()
			require.Same(t, attempt, currentAttempt)
			if outcome == "delete failure" {
				require.Same(t, op, registered)
				require.Equal(t, "known-frontend", token)
				require.Equal(t, phaseFailed, phase)
			} else {
				require.Nil(t, registered)
				require.Nil(t, remainingChild)
				require.Empty(t, token)
				require.Equal(t, phaseIdle, phase)
			}
			if outcome == "delete failure" {
				require.Len(t, status.Accesses, 1)
				require.Equal(t, "known-frontend", status.Accesses[0].FrontendToken)
				require.Equal(t, "failed", status.Accesses[0].Status)
				require.Contains(t, status.Accesses[0].Failure.LastError, "pending cleanup")
			} else {
				require.Empty(t, status.Accesses, "released operation remains visible after cleanup completed")
				a.retryDue(time.Now().Add(time.Hour))
				require.EqualValues(t, 1, launches.Load())
			}
			require.EqualValues(t, 1, deletes.Load())
		})
	}
}
