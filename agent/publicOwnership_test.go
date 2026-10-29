package agent

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/openziti/zrok/v2/agent/agentGrpc"
	"github.com/openziti/zrok/v2/agent/proctree"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/stretchr/testify/require"
)

func publicRequest() *SharePublicRequest {
	return &SharePublicRequest{BackendMode: "proxy", Target: "8080", NameSelections: []NameSelection{{NamespaceToken: "team", Name: "app"}}}
}

func TestPublicAgentOwnsResourceAndExecutionArguments(t *testing.T) {
	a := newLifecycleAgent(t)
	req := publicRequest()
	req.BasicAuth, req.AccessGrants, req.Closed, req.Insecure = []string{"user:secret"}, []string{"account"}, true, true
	var creates, deletes atomic.Int32
	a.deps.createPublic = func(ctx context.Context, prepared *sdk.ShareRequest) (*sdk.Share, error) {
		creates.Add(1)
		_, bounded := ctx.Deadline()
		require.True(t, bounded)
		require.Equal(t, "http://127.0.0.1:8080", prepared.Target)
		require.Equal(t, []string{"user:secret"}, prepared.BasicAuth)
		require.Equal(t, []string{"account"}, prepared.AccessGrants)
		require.Equal(t, sdk.ClosedPermissionMode, prepared.PermissionMode)
		return &sdk.Share{Token: "known-token", FrontendEndpoints: []string{"https://parent.example"}}, nil
	}
	a.deps.startPublic = func(opts proctree.ControlledOptions, args ...string) (*proctree.Child, error) {
		require.Equal(t, []string{os.Args[0], "share", "public", "--subordinate", "--share-token", "known-token", "-b", "proxy", "--insecure", "http://127.0.0.1:8080"}, args)
		return proctree.StartControlledChild(opts, os.Args[0], "-test.run=^TestAgentLifecycleHelper$", "--", "boot")
	}
	var child *proctree.Child
	a.deps.deletePublic = func(ctx context.Context, token string) error {
		deletes.Add(1)
		require.Equal(t, "known-token", token)
		require.NoError(t, ctx.Err())
		select {
		case <-child.Done():
		default:
			t.Error("deleted before worker exit")
		}
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	token, endpoints, err := a.sharePublic(ctx, req)
	require.NoError(t, err)
	require.Equal(t, "known-token", token)
	require.Equal(t, []string{"https://parent.example"}, endpoints, "child endpoints must not replace the controller response")
	op := onlyOperation(t, a)
	a.mu.Lock()
	child = op.child
	a.mu.Unlock()
	cancel()
	require.NoError(t, a.Shutdown())
	require.EqualValues(t, 1, creates.Load())
	require.EqualValues(t, 1, deletes.Load())
	require.Len(t, savedRegistry(t, a).PublicShares, 1)
}

func TestPublicCreateSettlesAfterCancellationOrShutdown(t *testing.T) {
	for _, shutdown := range []bool{false, true} {
		t.Run(map[bool]string{false: "caller", true: "shutdown"}[shutdown], func(t *testing.T) {
			a := newLifecycleAgent(t)
			creating, finishCreate, deleting, finishDelete := make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{})
			a.deps.createPublic = func(ctx context.Context, _ *sdk.ShareRequest) (*sdk.Share, error) {
				close(creating)
				<-finishCreate
				require.NoError(t, ctx.Err(), "caller cancellation must not discard a late create response")
				return &sdk.Share{Token: "late"}, nil
			}
			a.deps.startPublic = func(proctree.ControlledOptions, ...string) (*proctree.Child, error) {
				t.Error("launched after stop intent")
				return nil, errors.New("unexpected launch")
			}
			a.deps.deletePublic = func(ctx context.Context, token string) error {
				require.Equal(t, "late", token)
				require.NoError(t, ctx.Err())
				close(deleting)
				<-finishDelete
				return nil
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			result := asyncResult(func() error { _, _, err := a.sharePublic(ctx, publicRequest()); return err })
			waitLifecycle(t, creating)
			var stopped <-chan error
			if shutdown {
				stopped = asyncResult(a.Shutdown)
				waitLifecycle(t, a.ctx.Done())
			} else {
				cancel()
			}
			close(finishCreate)
			waitLifecycle(t, deleting)
			select {
			case <-result:
				t.Fatal("startup returned before late-token cleanup")
			default:
			}
			if shutdown {
				select {
				case <-stopped:
					t.Fatal("shutdown returned during delete")
				default:
				}
			}
			close(finishDelete)
			require.Error(t, lifecycleResult(t, result))
			if shutdown {
				require.NoError(t, lifecycleResult(t, stopped))
			}
			require.Empty(t, savedRegistry(t, a).PublicShares, "failed caller start gained restart intent")
		})
	}
}

func TestPublicCleanupFailureRetainsTokenAndBlocksRecreation(t *testing.T) {
	a := newLifecycleAgent(t)
	var creates, deletes atomic.Int32
	a.deps.createPublic = func(context.Context, *sdk.ShareRequest) (*sdk.Share, error) {
		creates.Add(1)
		return &sdk.Share{Token: "known-token"}, nil
	}
	failure := errors.New("controller unavailable")
	a.deps.deletePublic = func(context.Context, string) error { deletes.Add(1); return failure }
	token, _, err := a.SharePublic(publicRequest())
	require.NoError(t, err)
	op := onlyOperation(t, a)
	a.mu.Lock()
	child := op.child
	a.mu.Unlock()
	require.NoError(t, proctree.StopChild(child))
	// the monitor claims cleanup after observing the reaped worker.
	a.mu.Lock()
	done := a.idle
	a.mu.Unlock()
	waitLifecycle(t, done)
	a.mu.Lock()
	require.Equal(t, phaseFailed, op.phase)
	a.mu.Unlock()
	status, err := (&agentGrpcImpl{agent: a}).Status(context.Background(), &agentGrpc.StatusRequest{})
	require.NoError(t, err)
	require.Equal(t, token, status.Shares[0].Token)
	require.Equal(t, "failed", status.Shares[0].Status)
	require.Contains(t, status.Shares[0].Failure.LastError, "pending cleanup")
	a.retryDue(time.Now().Add(time.Hour))
	waitLifecycle(t, cleanupDone(a, op))
	require.EqualValues(t, 1, creates.Load())
	require.ErrorIs(t, a.ReleaseShare(token), failure)
	require.Empty(t, savedRegistry(t, a).PublicShares)
	a.deps.deletePublic = func(context.Context, string) error { deletes.Add(1); return nil }
	require.NoError(t, a.ReleaseShare(token))
	a.retryDue(time.Now().Add(time.Hour))
	require.EqualValues(t, 1, creates.Load())
	require.EqualValues(t, 4, deletes.Load())
}

func TestPublicFailedStartupReportsKnownCleanupToken(t *testing.T) {
	a := newLifecycleAgent(t)
	launchErr, deleteErr := errors.New("spawn failed"), errors.New("delete failed")
	a.deps.startPublic = func(proctree.ControlledOptions, ...string) (*proctree.Child, error) { return nil, launchErr }
	a.deps.deletePublic = func(context.Context, string) error { return deleteErr }
	_, _, err := a.SharePublic(publicRequest())
	require.ErrorIs(t, err, launchErr)
	require.ErrorIs(t, err, deleteErr)
	require.ErrorContains(t, err, "known-token")
	require.Empty(t, savedRegistry(t, a).PublicShares)
	require.ErrorContains(t, a.Shutdown(), "known-token")
}

func TestPublicCreateUsesFreshDefaultWithoutMutatingIntent(t *testing.T) {
	t.Setenv("ZROK2_DEFAULT_NAMESPACE", "")
	a := newLifecycleAgent(t)
	path := filepath.Join(a.root.Metadata().RootPath, "config.json")
	var selections []sdk.NameSelection
	a.deps.createPublic = func(_ context.Context, req *sdk.ShareRequest) (*sdk.Share, error) {
		selections = append(selections, req.NameSelections...)
		return &sdk.Share{Token: "known-token", FrontendEndpoints: []string{"https://public.example"}}, nil
	}
	for _, namespace := range []string{"team-a", "team-b"} {
		require.NoError(t, os.WriteFile(path, []byte(`{"default_namespace":"`+namespace+`"}`), 0600))
		req := &agentGrpc.SharePublicRequest{BackendMode: "proxy", Target: "8080"}
		result, err := (&agentGrpcImpl{agent: a}).SharePublic(context.Background(), req)
		require.NoError(t, err)
		require.Equal(t, []string{"https://public.example"}, result.FrontendEndpoints)
		require.Nil(t, req.NameSelections)
		require.NoError(t, a.ReleaseShare(result.Token))
	}
	require.Equal(t, []sdk.NameSelection{{NamespaceToken: "team-a"}, {NamespaceToken: "team-b"}}, selections)
}

func TestPublicWorkerStartupFailuresCleanKnownToken(t *testing.T) {
	for _, mode := range []string{"eof", "mismatched boot", "no-boot"} {
		t.Run(mode, func(t *testing.T) {
			a := newLifecycleAgent(t)
			a.options.startup = 100 * time.Millisecond
			a.deps.startPublic = func(opts proctree.ControlledOptions, _ ...string) (*proctree.Child, error) {
				if mode == "mismatched boot" {
					return proctree.StartControlledChild(opts, os.Args[0], "-test.run=^TestAgentLifecycleHelper$", "--", "boot", "wrong-token")
				}
				return proctree.StartControlledChild(opts, os.Args[0], "-test.run=^TestAgentLifecycleHelper$", "--", mode)
			}
			var deletes atomic.Int32
			a.deps.deletePublic = func(_ context.Context, token string) error {
				require.Equal(t, "known-token", token)
				deletes.Add(1)
				return nil
			}
			_, _, err := a.SharePublic(publicRequest())
			require.Error(t, err)
			require.EqualValues(t, 1, deletes.Load())
			require.Empty(t, savedRegistry(t, a).PublicShares)
		})
	}
}

func TestPublicRetryCreatesReplacementOnlyAfterDeletion(t *testing.T) {
	a := newLifecycleAgent(t)
	var creates atomic.Int32
	a.deps.createPublic = func(context.Context, *sdk.ShareRequest) (*sdk.Share, error) {
		return &sdk.Share{Token: "token-" + strconv.Itoa(int(creates.Add(1)))}, nil
	}
	a.deps.startPublic = func(opts proctree.ControlledOptions, args ...string) (*proctree.Child, error) {
		return proctree.StartControlledChild(opts, os.Args[0], "-test.run=^TestAgentLifecycleHelper$", "--", "boot", args[5])
	}
	deleting, finish := make(chan struct{}), make(chan struct{})
	a.deps.deletePublic = func(_ context.Context, token string) error {
		if token == "token-1" {
			close(deleting)
			<-finish
		}
		return nil
	}
	token, _, err := a.SharePublic(publicRequest())
	require.NoError(t, err)
	require.Equal(t, "token-1", token)
	op := onlyOperation(t, a)
	a.mu.Lock()
	child := op.child
	a.mu.Unlock()
	require.NoError(t, proctree.StopChild(child))
	waitLifecycle(t, deleting)
	a.retryDue(time.Now().Add(time.Hour))
	require.EqualValues(t, 1, creates.Load())
	close(finish)
	waitLifecycle(t, cleanupDone(a, op))
	a.retryDue(time.Now().Add(time.Hour))
	a.mu.Lock()
	attempt := op.start
	a.mu.Unlock()
	waitLifecycle(t, attempt.done)
	require.NoError(t, attempt.err)
	require.Equal(t, "token-2", attempt.shr.token)
	require.EqualValues(t, 2, creates.Load())
	require.NoError(t, a.Shutdown())
	require.Len(t, savedRegistry(t, a).PublicShares, 1)
}
