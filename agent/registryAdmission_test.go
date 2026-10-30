package agent

import (
	"context"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/openziti/zrok/v2/agent/agentGrpc"
	"github.com/openziti/zrok/v2/agent/proctree"
	"github.com/stretchr/testify/require"
)

func admissionRequest(kind string) any {
	switch kind {
	case "public":
		return &SharePublicRequest{Target: "8080", BackendMode: "proxy", NameSelections: []NameSelection{{NamespaceToken: "public", Name: "app"}}}
	case "private":
		return &SharePrivateRequest{Target: "8080", BackendMode: "proxy", PrivateShareToken: "known-token"}
	default:
		return &AccessPrivateRequest{ShareToken: "shared"}
	}
}

func registryWithRequest(request any) *Registry {
	r := &Registry{V: RegistryV}
	switch req := request.(type) {
	case *SharePublicRequest:
		r.PublicShares = []*PublicShareRegistryEntry{{Request: req}}
	case *SharePrivateRequest:
		r.PrivateShares = []*PrivateShareRegistryEntry{{Request: req}}
	case *AccessPrivateRequest:
		r.PrivateAccesses = []*AccessRegistryEntry{{Request: req}}
	}
	return r
}

func admissionStart(ctx context.Context, a *Agent, request any, rpc bool) (string, error) {
	service := &agentGrpcImpl{agent: a}
	switch req := request.(type) {
	case *SharePublicRequest:
		if rpc {
			response, err := service.SharePublic(ctx, &agentGrpc.SharePublicRequest{Target: req.Target, BackendMode: req.BackendMode, NameSelections: []*agentGrpc.NameSelection{{NamespaceToken: "public", Name: "app"}}})
			if err != nil {
				return "", err
			}
			return response.Token, nil
		}
		token, _, err := a.SharePublic(req)
		return token, err
	case *SharePrivateRequest:
		if rpc {
			response, err := service.SharePrivate(ctx, &agentGrpc.SharePrivateRequest{Target: req.Target, BackendMode: req.BackendMode, PrivateShareToken: req.PrivateShareToken})
			if err != nil {
				return "", err
			}
			return response.Token, nil
		}
		return a.SharePrivate(req)
	case *AccessPrivateRequest:
		if rpc {
			response, err := service.AccessPrivate(ctx, &agentGrpc.AccessPrivateRequest{Token: req.ShareToken})
			if err != nil {
				return "", err
			}
			return response.FrontendToken, nil
		}
		return a.AccessPrivate(req)
	}
	panic("unsupported admission request")
}

func admissionRelease(ctx context.Context, a *Agent, kind, token string, rpc bool) error {
	if kind == "access" {
		if rpc {
			_, err := (&agentGrpcImpl{agent: a}).ReleaseAccess(ctx, &agentGrpc.ReleaseAccessRequest{FrontendToken: token})
			return err
		}
		return a.ReleaseAccess(token)
	}
	if rpc {
		_, err := (&agentGrpcImpl{agent: a}).ReleaseShare(ctx, &agentGrpc.ReleaseShareRequest{Token: token})
		return err
	}
	return a.ReleaseShare(token)
}

func TestCallerAdmissionWaitsForRegistryImport(t *testing.T) {
	for _, kind := range []string{"private", "public", "access"} {
		for _, rpc := range []bool{false, true} {
			for _, release := range []bool{false, true} {
				name := kind + "/direct"
				if rpc {
					name = kind + "/grpc"
				}
				if release {
					name += "/release"
				} else {
					name += "/start"
				}
				t.Run(name, func(t *testing.T) {
					a := newLifecycleAgent(t)
					req := admissionRequest(kind)
					path, _ := a.root.AgentRegistry()
					require.NoError(t, registryWithRequest(req).Save(path))
					entered, finish := make(chan struct{}), make(chan struct{})
					var unblock sync.Once
					defer unblock.Do(func() { close(finish) })
					a.deps.loadRegistry = func(path string) (*Registry, error) { close(entered); <-finish; return LoadRegistry(path) }
					var launches, deletes atomic.Int32
					a.deps.startChild = func(tail proctree.TailFunction, _ ...string) (*proctree.Child, error) {
						launches.Add(1)
						return lifecycleChild(tail, "boot")
					}
					a.deps.deleteShare = func(context.Context, string) error { deletes.Add(1); return nil }
					a.deps.deleteAccess = func(context.Context, string, string) error { deletes.Add(1); return nil }
					token := "known-token"
					if kind == "access" {
						token = "known-frontend"
					}
					var existing *operation
					if release {
						// reproduce an already registered caller while its first lazy registry import is still pending.
						existing = &operation{id: operationID(), request: cloneRequest(req), phase: phaseActive, desired: true}
						if kind == "access" {
							existing.acc = &access{token: "shared", frontendToken: token}
						} else {
							existing.shr = &share{token: token}
						}
						a.mu.Lock()
						a.operations[existing.id] = existing
						a.mu.Unlock()
					}
					activity := asyncResult(func() error {
						if release {
							return admissionRelease(context.Background(), a, kind, token, rpc)
						}
						var err error
						token, err = admissionStart(context.Background(), a, req, rpc)
						return err
					})
					waitLifecycle(t, entered)
					require.Zero(t, launches.Load())
					require.Zero(t, deletes.Load())
					a.mu.Lock()
					if release {
						require.False(t, existing.released)
						require.True(t, existing.desired)
					} else {
						require.Empty(t, a.operations)
					}
					a.mu.Unlock()
					select {
					case <-activity:
						t.Fatal("caller was admitted before import completed")
					default:
					}
					unblock.Do(func() { close(finish) })
					require.NoError(t, lifecycleResult(t, activity))
					if !release {
						require.EqualValues(t, 1, launches.Load())
						onlyOperation(t, a) // the caller uses the matching queued intent instead of duplicating it.
						require.NoError(t, admissionRelease(context.Background(), a, kind, token, rpc))
					}
					saved := savedRegistry(t, a)
					require.Empty(t, saved.PrivateShares)
					require.Empty(t, saved.PublicShares)
					require.Empty(t, saved.PrivateAccesses)
					require.NoError(t, a.RestoreRegistryOnce())
					saved = savedRegistry(t, a)
					require.Empty(t, saved.PrivateShares)
					require.Empty(t, saved.PublicShares)
					require.Empty(t, saved.PrivateAccesses)
				})
			}
		}
	}
}

func TestImportFailureRejectsCallerActivityWithoutSideEffects(t *testing.T) {
	for _, kind := range []string{"private", "public", "access"} {
		for _, rpc := range []bool{false, true} {
			for _, release := range []bool{false, true} {
				name := kind + "/direct"
				if rpc {
					name = kind + "/grpc"
				}
				if release {
					name += "/release"
				}
				t.Run(name, func(t *testing.T) {
					a := newLifecycleAgent(t)
					path, _ := a.root.AgentRegistry()
					raw := []byte("unread registry")
					require.NoError(t, os.WriteFile(path, raw, 0600))
					failure := errors.New("registry import failed")
					a.deps.loadRegistry = func(string) (*Registry, error) { return nil, failure }
					var launches, deletes, saves atomic.Int32
					a.deps.startChild = func(proctree.TailFunction, ...string) (*proctree.Child, error) {
						launches.Add(1)
						return nil, errors.New("unexpected launch")
					}
					a.deps.deleteShare = func(context.Context, string) error { deletes.Add(1); return nil }
					a.deps.deleteAccess = func(context.Context, string, string) error { deletes.Add(1); return nil }
					a.deps.saveRegistry = func(*Registry, string) error { saves.Add(1); return nil }
					var err error
					if release {
						err = admissionRelease(context.Background(), a, kind, "known", rpc)
					} else {
						_, err = admissionStart(context.Background(), a, admissionRequest(kind), rpc)
					}
					require.ErrorIs(t, err, failure)
					require.Zero(t, launches.Load())
					require.Zero(t, deletes.Load())
					require.Zero(t, saves.Load())
					a.mu.Lock()
					require.Empty(t, a.operations)
					a.mu.Unlock()
					require.ErrorIs(t, a.Shutdown(), failure)
					unchanged, err := os.ReadFile(path)
					require.NoError(t, err)
					require.Equal(t, raw, unchanged)
					require.Zero(t, saves.Load())
				})
			}
		}
	}
}

func TestImportWaitHonorsCallerCancellationAndShutdown(t *testing.T) {
	for _, shutdown := range []bool{false, true} {
		for _, release := range []bool{false, true} {
			name := "cancel/start"
			if shutdown {
				name = "shutdown/start"
			}
			if release {
				name += "-release"
			}
			t.Run(name, func(t *testing.T) {
				a := newLifecycleAgent(t)
				req := admissionRequest("private")
				entered, finish := make(chan struct{}), make(chan struct{})
				var unblock sync.Once
				defer unblock.Do(func() { close(finish) })
				a.deps.loadRegistry = func(string) (*Registry, error) { close(entered); <-finish; return registryWithRequest(req), nil }
				var launches atomic.Int32
				a.deps.startChild = func(proctree.TailFunction, ...string) (*proctree.Child, error) {
					launches.Add(1)
					return nil, errors.New("unexpected launch")
				}
				var existing *operation
				if release {
					existing = &operation{id: operationID(), request: cloneRequest(req), phase: phaseQueued, desired: true}
					a.mu.Lock()
					a.operations[existing.id] = existing
					a.mu.Unlock()
				}
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				activity := asyncResult(func() error {
					if release {
						return admissionRelease(ctx, a, "private", existing.id, !shutdown)
					}
					_, err := admissionStart(ctx, a, req, !shutdown)
					return err
				})
				waitLifecycle(t, entered)
				var stopped <-chan error
				if shutdown {
					stopped = asyncResult(a.Shutdown)
					waitLifecycle(t, a.ctx.Done())
				} else {
					cancel()
				}
				err := lifecycleResult(t, activity)
				if shutdown {
					require.ErrorContains(t, err, "stopping")
				} else {
					require.ErrorIs(t, err, context.Canceled)
				}
				require.Zero(t, launches.Load())
				a.mu.Lock()
				if release {
					require.False(t, existing.released)
					require.True(t, existing.desired)
				} else {
					require.Empty(t, a.operations)
				}
				a.mu.Unlock()
				unblock.Do(func() { close(finish) })
				if shutdown {
					require.NoError(t, lifecycleResult(t, stopped))
				}
			})
		}
	}
}

func TestCallerAdmissionDoesNotWaitForWorkerReplay(t *testing.T) {
	a := newLifecycleAgent(t)
	path, _ := a.root.AgentRegistry()
	loaded := &SharePrivateRequest{Target: "8080", BackendMode: "proxy", PrivateShareToken: "loaded-token"}
	require.NoError(t, registryWithRequest(loaded).Save(path))
	replayEntered, finishReplay := make(chan struct{}), make(chan struct{})
	var unblock sync.Once
	defer unblock.Do(func() { close(finishReplay) })
	var launches atomic.Int32
	a.deps.startChild = func(tail proctree.TailFunction, _ ...string) (*proctree.Child, error) {
		if launches.Add(1) == 1 {
			close(replayEntered)
			<-finishReplay
		}
		return lifecycleChild(tail, "boot")
	}
	replay := asyncResult(a.RestoreRegistryOnce)
	waitLifecycle(t, replayEntered)
	caller := &SharePrivateRequest{Target: "9090", BackendMode: "proxy", PrivateShareToken: "caller-token"}
	started := asyncResult(func() error {
		token, err := a.SharePrivate(caller)
		if err != nil {
			return err
		}
		return a.ReleaseShare(token)
	})
	require.NoError(t, lifecycleResult(t, started))
	require.EqualValues(t, 2, launches.Load())
	require.Len(t, savedRegistry(t, a).PrivateShares, 1)
	unblock.Do(func() { close(finishReplay) })
	require.NoError(t, lifecycleResult(t, replay))
}
