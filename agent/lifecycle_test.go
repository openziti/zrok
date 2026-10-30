package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/openziti/zrok/v2/agent/agentGrpc"
	"github.com/openziti/zrok/v2/agent/proctree"
	"github.com/openziti/zrok/v2/environment/env_core"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type lifecycleRoot struct {
	env_core.Root
	dir string
}

func (*lifecycleRoot) IsEnabled() bool { return true }
func (r *lifecycleRoot) AgentRegistry() (string, error) {
	return filepath.Join(r.dir, "registry.json"), nil
}
func (r *lifecycleRoot) AgentSocket() (string, error) { return filepath.Join(r.dir, "agent.sock"), nil }
func (r *lifecycleRoot) AgentEnrollment() (string, error) {
	return filepath.Join(r.dir, "enrollment.json"), nil
}
func (r *lifecycleRoot) Metadata() *env_core.Metadata { return &env_core.Metadata{RootPath: r.dir} }

func newLifecycleAgent(t *testing.T) *Agent {
	t.Helper()
	// keep the socket path short even under macOS's long temporary directory; t.TempDir includes the test name.
	dir, err := os.MkdirTemp("", "agent-")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, os.RemoveAll(dir)) })
	a, err := NewAgent(DefaultConfig(), &lifecycleRoot{dir: dir})
	require.NoError(t, err)
	a.options.startup = 5 * time.Second
	a.options.shutdown = 5 * time.Second
	a.options.forcedExit = time.Second
	a.options.server = 50 * time.Millisecond
	a.deps.deleteShare = func(context.Context, string) error { return nil }
	a.deps.deleteAccess = func(context.Context, string, string) error { return nil }
	a.deps.startChild = func(tail proctree.TailFunction, _ ...string) (*proctree.Child, error) {
		return lifecycleChild(tail, "boot")
	}
	a.deps.createPublic = func(context.Context, *sdk.ShareRequest) (*sdk.Share, error) {
		return &sdk.Share{Token: "known-token", FrontendEndpoints: []string{"http://share.example"}}, nil
	}
	a.deps.deletePublic = func(ctx context.Context, token string) error { return a.deps.deleteShare(ctx, token) }
	// shared lifecycle fixtures can block either launch path at the same boundary.
	a.deps.startPublic = func(opts proctree.ControlledOptions, args ...string) (*proctree.Child, error) {
		return a.deps.startChild(opts.Stdout, args...)
	}
	t.Cleanup(func() { _ = a.Shutdown() })
	return a
}

func TestAgentLifecycleHelper(t *testing.T) {
	mode := ""
	token := "known-token"
	for index, arg := range os.Args {
		if arg == "--" && index+1 < len(os.Args) {
			mode = os.Args[index+1]
			if index+2 < len(os.Args) {
				token = os.Args[index+2]
			}
			break
		}
	}
	if mode == "" {
		return
	}
	if mode == "boot" {
		data, _ := json.Marshal(map[string]any{"msg": "boot", "token": token, "frontend_token": "known-frontend", "frontend_endpoints": []string{"http://share.example"}, "bind_address": "127.0.0.1:10000"})
		fmt.Println(string(data))
	}
	if mode == "eof" {
		_ = os.Stdout.Close()
	}
	_, _ = io.Copy(io.Discard, os.Stdin)
	os.Exit(0)
}

func lifecycleChild(tail proctree.TailFunction, mode string) (*proctree.Child, error) {
	return proctree.StartControlledChild(proctree.ControlledOptions{Stdout: tail}, os.Args[0], "-test.run=^TestAgentLifecycleHelper$", "--", mode)
}
func waitLifecycle(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("lifecycle operation did not complete")
	}
}
func lifecycleResult(t *testing.T, result <-chan error) error {
	t.Helper()
	select {
	case err := <-result:
		return err
	case <-time.After(10 * time.Second):
		t.Fatal("lifecycle result did not arrive")
		return nil
	}
}
func asyncResult(fn func() error) <-chan error {
	result := make(chan error, 1)
	go func() { result <- fn() }()
	return result
}
func savedRegistry(t *testing.T, a *Agent) *Registry {
	t.Helper()
	path, err := a.root.AgentRegistry()
	require.NoError(t, err)
	registry, err := LoadRegistry(path)
	require.NoError(t, err)
	return registry
}
func onlyOperation(t *testing.T, a *Agent) *operation {
	t.Helper()
	a.mu.Lock()
	defer a.mu.Unlock()
	require.Len(t, a.operations, 1)
	for _, op := range a.operations {
		return op
	}
	return nil
}
func cleanupDone(a *Agent, op *operation) <-chan struct{} {
	a.mu.Lock()
	defer a.mu.Unlock()
	return op.cleanup.done
}

func TestShutdownJoinsStartupAndRejectsAdmission(t *testing.T) {
	a := newLifecycleAgent(t)
	booted, finishLaunch := make(chan struct{}), make(chan struct{})
	a.deps.startChild = func(tail proctree.TailFunction, _ ...string) (*proctree.Child, error) {
		c, err := lifecycleChild(func(data []byte) { tail(data); close(booted) }, "boot")
		<-finishLaunch
		return c, err
	}
	var deletes atomic.Int32
	a.deps.deleteShare = func(context.Context, string) error { deletes.Add(1); return nil }
	started := asyncResult(func() error {
		_, err := a.SharePrivate(&SharePrivateRequest{Target: "8080", BackendMode: "proxy"})
		return err
	})
	waitLifecycle(t, booted)
	stopped := asyncResult(a.Shutdown)
	waitLifecycle(t, a.ctx.Done())
	select {
	case <-stopped:
		t.Fatal("shutdown returned while startup was registered")
	default:
	}
	_, err := a.SharePrivate(&SharePrivateRequest{Target: "new"})
	require.ErrorContains(t, err, "stopping")
	close(finishLaunch)
	require.Error(t, lifecycleResult(t, started))
	require.NoError(t, lifecycleResult(t, stopped))
	require.EqualValues(t, 1, deletes.Load())
	require.NoError(t, a.Shutdown())
}

func TestReleaseExitAndShutdownShareCleanupAttempt(t *testing.T) {
	a := newLifecycleAgent(t)
	deleting, finishDelete := make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	a.deps.deleteShare = func(context.Context, string) error { calls.Add(1); close(deleting); <-finishDelete; return nil }
	token, err := a.SharePrivate(&SharePrivateRequest{Target: "8080", BackendMode: "proxy"})
	require.NoError(t, err)
	released := asyncResult(func() error { return a.ReleaseShare(token) })
	waitLifecycle(t, deleting)
	stopped := asyncResult(a.Shutdown)
	waitLifecycle(t, a.ctx.Done())
	select {
	case <-stopped:
		t.Fatal("shutdown did not join deletion")
	default:
	}
	close(finishDelete)
	require.NoError(t, lifecycleResult(t, released))
	require.NoError(t, lifecycleResult(t, stopped))
	require.EqualValues(t, 1, calls.Load())
}

func TestOrdinaryStopPreservesPrivateTokenAndNamedIntent(t *testing.T) {
	a := newLifecycleAgent(t)
	var deletes atomic.Int32
	a.deps.deleteShare = func(context.Context, string) error { deletes.Add(1); return nil }
	_, err := a.SharePrivate(&SharePrivateRequest{Target: "8080", BackendMode: "proxy", PrivateShareToken: "known-token"})
	require.NoError(t, err)
	require.NoError(t, a.Shutdown())
	require.Zero(t, deletes.Load())
	registry := savedRegistry(t, a)
	require.Equal(t, RegistryV, registry.V)
	require.Len(t, registry.PrivateShares, 1)
}

func TestExplicitReleaseDeletesSuppliedPrivateToken(t *testing.T) {
	a := newLifecycleAgent(t)
	var deletes atomic.Int32
	a.deps.deleteShare = func(context.Context, string) error { deletes.Add(1); return nil }
	token, err := a.SharePrivate(&SharePrivateRequest{Target: "8080", BackendMode: "proxy", PrivateShareToken: "known-token"})
	require.NoError(t, err)
	require.NoError(t, a.ReleaseShare(token))
	require.EqualValues(t, 1, deletes.Load())
	require.Empty(t, savedRegistry(t, a).PrivateShares)
}

func TestCallerContextDetachesAfterSuccessfulStartup(t *testing.T) {
	a := newLifecycleAgent(t)
	ctx, cancel := context.WithCancel(context.Background())
	token, err := a.sharePrivate(ctx, &SharePrivateRequest{Target: "8080", BackendMode: "proxy"})
	require.NoError(t, err)
	cancel()
	op := onlyOperation(t, a)
	a.mu.Lock()
	require.Equal(t, phaseActive, op.phase)
	child := op.child
	a.mu.Unlock()
	select {
	case <-child.Done():
		t.Fatal("completed RPC cancelled an active worker")
	default:
	}
	require.NoError(t, a.ReleaseShare(token))
}

func TestLegacyStartupWithoutBootIsBounded(t *testing.T) {
	a := newLifecycleAgent(t)
	a.options.startup = 20 * time.Millisecond
	a.deps.startChild = func(tail proctree.TailFunction, _ ...string) (*proctree.Child, error) {
		return lifecycleChild(tail, "no-boot")
	}
	_, err := a.SharePrivate(&SharePrivateRequest{Target: "8080", BackendMode: "proxy"})
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Empty(t, savedRegistry(t, a).PrivateShares)
}

func TestShutdownDuringRegistryReplayPreservesUnattemptedRequests(t *testing.T) {
	a := newLifecycleAgent(t)
	registry := &Registry{V: RegistryV}
	for _, name := range []string{"one", "two", "three", "four"} {
		registry.PublicShares = append(registry.PublicShares, &PublicShareRegistryEntry{Request: &SharePublicRequest{BackendMode: "proxy", Target: "8080", NameSelections: []NameSelection{{NamespaceToken: "public", Name: name}}}})
	}
	path, _ := a.root.AgentRegistry()
	require.NoError(t, registry.Save(path))
	entered, finishLaunch := make(chan struct{}), make(chan struct{})
	var launches atomic.Int32
	a.deps.startChild = func(tail proctree.TailFunction, _ ...string) (*proctree.Child, error) {
		launches.Add(1)
		c, err := lifecycleChild(func(data []byte) { tail(data); close(entered) }, "boot")
		<-finishLaunch
		return c, err
	}
	replay := asyncResult(a.RestoreRegistryOnce)
	waitLifecycle(t, entered)
	a.mu.Lock()
	var releaseID string
	for _, op := range a.operations {
		if op.phase == phaseQueued {
			releaseID = op.id
			break
		}
	}
	a.mu.Unlock()
	require.NotEmpty(t, releaseID)
	require.NoError(t, a.ReleaseShare(releaseID))
	stopped := asyncResult(a.Shutdown)
	waitLifecycle(t, a.ctx.Done())
	close(finishLaunch)
	require.NoError(t, lifecycleResult(t, replay))
	require.NoError(t, lifecycleResult(t, stopped))
	require.Len(t, savedRegistry(t, a).PublicShares, 3)
	require.EqualValues(t, 1, launches.Load())
}

func TestShutdownDuringRegistryImport(t *testing.T) {
	for _, broken := range []bool{false, true} {
		t.Run(fmt.Sprint(broken), func(t *testing.T) {
			a := newLifecycleAgent(t)
			entered, finishLoad := make(chan struct{}), make(chan struct{})
			var saves atomic.Int32
			originalSave := a.deps.saveRegistry
			a.deps.saveRegistry = func(r *Registry, path string) error { saves.Add(1); return originalSave(r, path) }
			a.deps.loadRegistry = func(string) (*Registry, error) {
				close(entered)
				<-finishLoad
				if broken {
					return nil, errors.New("unreadable registry")
				}
				return &Registry{V: RegistryV, PrivateShares: []*PrivateShareRegistryEntry{{Request: &SharePrivateRequest{Target: "first", PrivateShareToken: "first"}}, {Request: &SharePrivateRequest{Target: "second", PrivateShareToken: "second"}}}}, nil
			}
			replay := asyncResult(a.RestoreRegistryOnce)
			waitLifecycle(t, entered)
			stopped := asyncResult(a.Shutdown)
			waitLifecycle(t, a.ctx.Done())
			require.Zero(t, saves.Load())
			close(finishLoad)
			if broken {
				require.ErrorContains(t, lifecycleResult(t, replay), "unreadable")
				require.ErrorContains(t, lifecycleResult(t, stopped), "unreadable")
				require.Zero(t, saves.Load())
			} else {
				require.NoError(t, lifecycleResult(t, replay))
				require.NoError(t, lifecycleResult(t, stopped))
				require.Len(t, savedRegistry(t, a).PrivateShares, 2)
			}
		})
	}
}

func TestReleaseCancelsInFlightRetryByFailureID(t *testing.T) {
	a := newLifecycleAgent(t)
	path, _ := a.root.AgentRegistry()
	require.NoError(t, (&Registry{V: RegistryV, PublicShares: []*PublicShareRegistryEntry{{Request: &SharePublicRequest{BackendMode: "proxy", Target: "8080", NameSelections: []NameSelection{{NamespaceToken: "public", Name: "app"}}}}}}).Save(path))
	entered, finishLaunch := make(chan struct{}), make(chan struct{})
	var launches, deletes atomic.Int32
	a.deps.startChild = func(tail proctree.TailFunction, _ ...string) (*proctree.Child, error) {
		if launches.Add(1) == 1 {
			return nil, errors.New("initial startup failed")
		}
		c, err := lifecycleChild(func(data []byte) { tail(data); close(entered) }, "boot")
		<-finishLaunch
		return c, err
	}
	a.deps.deleteShare = func(context.Context, string) error { deletes.Add(1); return nil }
	require.NoError(t, a.RestoreRegistryOnce())
	op := onlyOperation(t, a)
	a.mu.Lock()
	id := op.failureID
	a.mu.Unlock()
	require.NotEmpty(t, id)
	a.retryDue(time.Now().Add(time.Hour))
	waitLifecycle(t, entered)
	a.retryDue(time.Now().Add(time.Hour))
	a.mu.Lock()
	cancelled := op.start.ctx.Done()
	a.mu.Unlock()
	released := asyncResult(func() error { return a.ReleaseShare(id) })
	waitLifecycle(t, cancelled)
	close(finishLaunch)
	require.NoError(t, lifecycleResult(t, released))
	require.EqualValues(t, 2, launches.Load())
	// both the failed launch's resource and the cancelled retry's resource are agent-owned.
	require.EqualValues(t, 2, deletes.Load())
	a.retryDue(time.Now().Add(time.Hour))
	require.EqualValues(t, 2, launches.Load())
	require.Empty(t, savedRegistry(t, a).PublicShares)
}

func TestRegistryWriteCannotOverwriteNewerRelease(t *testing.T) {
	a := newLifecycleAgent(t)
	initialSaved, oldSnapshot, finishOld, deleted := make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{})
	var writes atomic.Int32
	a.deps.saveRegistry = func(r *Registry, path string) error {
		n := writes.Add(1)
		if n == 2 {
			close(oldSnapshot)
			<-finishOld
		}
		err := r.Save(path)
		if n == 1 {
			close(initialSaved)
		}
		return err
	}
	a.deps.deleteShare = func(context.Context, string) error { close(deleted); return nil }
	token, err := a.SharePrivate(&SharePrivateRequest{Target: "8080", BackendMode: "proxy", PrivateShareToken: "known-token"})
	require.NoError(t, err)
	waitLifecycle(t, initialSaved)
	saved := asyncResult(a.SaveRegistry)
	waitLifecycle(t, oldSnapshot)
	released := asyncResult(func() error { return a.ReleaseShare(token) })
	waitLifecycle(t, deleted)
	close(finishOld)
	require.NoError(t, lifecycleResult(t, saved))
	require.NoError(t, lifecycleResult(t, released))
	require.Empty(t, savedRegistry(t, a).PrivateShares)
}

func TestLateRemoteListenerCannotRegisterAfterShutdown(t *testing.T) {
	a := newLifecycleAgent(t)
	path, _ := a.root.AgentEnrollment()
	require.NoError(t, NewEnrollment("remote").Save(path))
	entered, finish := make(chan struct{}), make(chan struct{})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	a.deps.remoteListener = func(string) (net.Listener, error) { close(entered); <-finish; return listener, nil }
	a.mu.Lock()
	a.beginWorkLocked()
	a.mu.Unlock()
	go a.remoteAgent(make(chan error, 1))
	waitLifecycle(t, entered)
	stopped := asyncResult(a.Shutdown)
	waitLifecycle(t, a.ctx.Done())
	close(finish)
	require.NoError(t, lifecycleResult(t, stopped))
	a.mu.Lock()
	require.Empty(t, a.grpcServers)
	a.mu.Unlock()
	_, err = listener.Accept()
	require.Error(t, err)
}

func TestStatusSnapshotsAndShutdownAreRaceFree(t *testing.T) {
	a := newLifecycleAgent(t)
	req := &AccessPrivateRequest{ShareToken: "share", ResponseHeaders: []string{"header"}}
	_, err := a.AccessPrivate(req)
	require.NoError(t, err)
	req.ResponseHeaders[0] = "caller change"
	rpc := &agentGrpcImpl{agent: a}
	status, err := rpc.Status(context.Background(), &agentGrpc.StatusRequest{})
	require.NoError(t, err)
	require.Equal(t, "header", status.Accesses[0].ResponseHeaders[0])
	status.Accesses[0].ResponseHeaders[0] = "response change"
	var workers sync.WaitGroup
	for range 8 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for range 20 {
				_, _ = rpc.Status(context.Background(), &agentGrpc.StatusRequest{})
				_, _ = rpc.Version(context.Background(), &agentGrpc.VersionRequest{})
			}
		}()
	}
	require.NoError(t, a.Shutdown())
	workers.Wait()
}

func TestReleaseQueuedFailureRemovesStatusAndIntent(t *testing.T) {
	a := newLifecycleAgent(t)
	path, _ := a.root.AgentRegistry()
	require.NoError(t, (&Registry{V: RegistryV, PrivateShares: []*PrivateShareRegistryEntry{{Request: &SharePrivateRequest{Target: "8080", PrivateShareToken: "reserved"}}}}).Save(path))
	a.deps.startChild = func(proctree.TailFunction, ...string) (*proctree.Child, error) {
		return nil, errors.New("spawn failed")
	}
	require.NoError(t, a.RestoreRegistryOnce())
	op := onlyOperation(t, a)
	a.mu.Lock()
	id := op.failureID
	a.mu.Unlock()
	require.NoError(t, a.ReleaseShare(id))
	status, err := (&agentGrpcImpl{agent: a}).Status(context.Background(), &agentGrpc.StatusRequest{})
	require.NoError(t, err)
	require.Empty(t, status.Shares)
	require.Empty(t, savedRegistry(t, a).PrivateShares)
}

func TestFailedCleanupPrecedesRecreation(t *testing.T) {
	a := newLifecycleAgent(t)
	firstDelete, secondDelete, finishSecond := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var launches, deletes atomic.Int32
	a.deps.startChild = func(tail proctree.TailFunction, _ ...string) (*proctree.Child, error) {
		launches.Add(1)
		return lifecycleChild(tail, "boot")
	}
	a.deps.deleteAccess = func(context.Context, string, string) error {
		switch deletes.Add(1) {
		case 1:
			close(firstDelete)
			return errors.New("controller unavailable")
		case 2:
			close(secondDelete)
			<-finishSecond
		}
		return nil
	}
	_, err := a.AccessPrivate(&AccessPrivateRequest{ShareToken: "shared"})
	require.NoError(t, err)
	op := onlyOperation(t, a)
	a.mu.Lock()
	child := op.child
	a.mu.Unlock()
	require.NoError(t, proctree.StopChild(child))
	waitLifecycle(t, firstDelete)
	waitLifecycle(t, cleanupDone(a, op))
	a.mu.Lock()
	require.Equal(t, phaseFailed, op.phase)
	require.Equal(t, "known-frontend", op.resourceToken())
	a.mu.Unlock()
	a.retryDue(time.Now().Add(time.Hour))
	waitLifecycle(t, secondDelete)
	a.retryDue(time.Now().Add(time.Hour))
	require.EqualValues(t, 1, launches.Load())
	close(finishSecond)
	waitLifecycle(t, cleanupDone(a, op))
	a.mu.Lock()
	require.Equal(t, phaseRetry, op.phase)
	a.mu.Unlock()
	a.retryDue(time.Now().Add(time.Hour))
	a.mu.Lock()
	attempt := op.start
	a.mu.Unlock()
	waitLifecycle(t, attempt.done)
	require.EqualValues(t, 2, launches.Load())
}

func TestReleaseReportsCleanupAndRegistryFailures(t *testing.T) {
	a := newLifecycleAgent(t)
	persisted := make(chan struct{})
	var writes atomic.Int32
	diskFailure := errors.New("registry disk failure")
	a.deps.saveRegistry = func(r *Registry, path string) error {
		if writes.Add(1) == 1 {
			err := r.Save(path)
			close(persisted)
			return err
		}
		return diskFailure
	}
	deleteFailure := errors.New("controller delete failure")
	a.deps.deleteShare = func(context.Context, string) error { return deleteFailure }
	token, err := a.SharePrivate(&SharePrivateRequest{Target: "8080", PrivateShareToken: "reserved"})
	require.NoError(t, err)
	waitLifecycle(t, persisted)
	err = a.ReleaseShare(token)
	require.ErrorIs(t, err, deleteFailure)
	require.ErrorIs(t, err, diskFailure)
	op := onlyOperation(t, a)
	a.mu.Lock()
	require.False(t, op.desired)
	require.Equal(t, token, op.resourceToken())
	a.mu.Unlock()
	require.Error(t, a.Shutdown())
}

func TestShutdownDeadlineReportsUnresolvedToken(t *testing.T) {
	a := newLifecycleAgent(t)
	a.options.shutdown = 30 * time.Millisecond
	deleting, finishDelete := make(chan struct{}), make(chan struct{})
	a.deps.deleteShare = func(context.Context, string) error { close(deleting); <-finishDelete; return nil }
	token, err := a.SharePrivate(&SharePrivateRequest{Target: "8080"})
	require.NoError(t, err)
	stopped := asyncResult(a.Shutdown)
	waitLifecycle(t, deleting)
	err = lifecycleResult(t, stopped)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.ErrorContains(t, err, token)
	op := onlyOperation(t, a)
	close(finishDelete)
	waitLifecycle(t, cleanupDone(a, op))
	require.Same(t, err, a.Shutdown())
}

func TestRunShutdownClosesExistingGrpcClients(t *testing.T) {
	a := newLifecycleAgent(t)
	a.cfg.ConsoleEnabled = false
	imported := make(chan struct{})
	a.deps.loadRegistry = func(string) (*Registry, error) { close(imported); return &Registry{V: RegistryV}, nil }
	run := asyncResult(a.Run)
	waitLifecycle(t, imported)
	socket, _ := a.root.AgentSocket()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	connection, err := grpc.DialContext(ctx, "unix:"+socket, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	require.NoError(t, err)
	defer connection.Close()
	client := agentGrpc.NewAgentClient(connection)
	_, err = client.Version(ctx, &agentGrpc.VersionRequest{})
	require.NoError(t, err)
	require.NoError(t, a.Shutdown())
	require.NoError(t, lifecycleResult(t, run))
	_, err = client.Version(ctx, &agentGrpc.VersionRequest{})
	require.Error(t, err)
	_, err = os.Stat(socket)
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestRequiredRemotingFailureInterruptsRegistryReplay(t *testing.T) {
	a := newLifecycleAgent(t)
	a.cfg.ConsoleEnabled = false
	a.cfg.RequireRemoting = true
	enrollment, _ := a.root.AgentEnrollment()
	require.NoError(t, NewEnrollment("remote").Save(enrollment))
	path, _ := a.root.AgentRegistry()
	require.NoError(t, (&Registry{V: RegistryV, PrivateShares: []*PrivateShareRegistryEntry{{Request: &SharePrivateRequest{Target: "8080", PrivateShareToken: "reserved"}}}}).Save(path))
	entered, finishLaunch := make(chan struct{}), make(chan struct{})
	failure := errors.New("required remoting failed")
	a.deps.remoteListener = func(string) (net.Listener, error) { <-entered; return nil, failure }
	a.deps.startChild = func(tail proctree.TailFunction, _ ...string) (*proctree.Child, error) {
		c, err := lifecycleChild(func(data []byte) { tail(data); close(entered) }, "boot")
		<-finishLaunch
		return c, err
	}
	run := asyncResult(a.Run)
	waitLifecycle(t, entered)
	waitLifecycle(t, a.ctx.Done())
	close(finishLaunch)
	require.ErrorIs(t, lifecycleResult(t, run), failure)
	require.Len(t, savedRegistry(t, a).PrivateShares, 1)
}

func TestHealthcheckDoesNotHoldLifecycleLockDuringHTTP(t *testing.T) {
	entered, finish := make(chan struct{}), make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { close(entered); <-finish; w.WriteHeader(http.StatusOK) }))
	defer server.Close()
	defer func() {
		select {
		case <-finish:
		default:
			close(finish)
		}
	}()
	a := newLifecycleAgent(t)
	token, err := a.SharePrivate(&SharePrivateRequest{Target: server.URL, BackendMode: "proxy"})
	require.NoError(t, err)
	health := asyncResult(func() error { return a.ShareHttpHealthcheck(token, "/", "GET", 200, 5000) })
	waitLifecycle(t, entered)
	require.NoError(t, a.Shutdown())
	// the independent HTTP request is still in progress after its resource has been drained.
	select {
	case <-health:
		t.Fatal("health request returned before the server responded")
	default:
	}
	close(finish)
	require.NoError(t, lifecycleResult(t, health))
}

func TestEmptyReleaseIdentifierCannotSelectAnActiveOperation(t *testing.T) {
	a := newLifecycleAgent(t)
	_, err := a.SharePrivate(&SharePrivateRequest{Target: "8080"})
	require.NoError(t, err)
	require.Error(t, a.ReleaseShare(""))
	op := onlyOperation(t, a)
	a.mu.Lock()
	require.Equal(t, phaseActive, op.phase)
	require.False(t, op.released)
	a.mu.Unlock()
}

func TestCancelledCallerCannotLaunchAnOperation(t *testing.T) {
	a := newLifecycleAgent(t)
	var launches atomic.Int32
	a.deps.startChild = func(proctree.TailFunction, ...string) (*proctree.Child, error) {
		launches.Add(1)
		return nil, errors.New("unexpected launch")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := a.sharePrivate(ctx, &SharePrivateRequest{Target: "8080"})
	require.ErrorIs(t, err, context.Canceled)
	require.Zero(t, launches.Load())
	a.mu.Lock()
	require.Empty(t, a.operations)
	a.mu.Unlock()
}
