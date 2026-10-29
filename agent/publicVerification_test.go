package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	"github.com/openziti/zrok/v2/environment/env_v0_4"
	restshare "github.com/openziti/zrok/v2/rest_client_zrok/share"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/stretchr/testify/require"
)

func TestPublicDefaultChangeDuringCreateKeepsPreparedSelectionAndCredentials(t *testing.T) {
	t.Setenv("ZROK2_DEFAULT_NAMESPACE", "")
	firstCreate, finishFirst := make(chan *rest_model_zrok.ShareRequest, 1), make(chan struct{})
	var unblock sync.Once
	defer unblock.Do(func() { close(finishFirst) })
	var creates, deletes atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if req.URL.Path == "/api/v2/clientVersionCheck" {
			return
		}
		require.Equal(t, "original-account", req.Header.Get("X-TOKEN"))
		switch req.URL.Path {
		case "/api/v2/share":
			body := &rest_model_zrok.ShareRequest{}
			require.NoError(t, json.NewDecoder(req.Body).Decode(body))
			require.Equal(t, "original-environment", body.EnvZID)
			n := creates.Add(1)
			if n == 1 {
				firstCreate <- body
				select {
				case <-finishFirst:
				case <-req.Context().Done():
					return
				}
			} else if n == 2 {
				require.Equal(t, "team-b", body.NameSelections[0].NamespaceToken)
			} else {
				require.Equal(t, "cli-team", body.NameSelections[0].NamespaceToken)
				require.Equal(t, "app", body.NameSelections[0].Name)
			}
			w.WriteHeader(http.StatusCreated)
			require.NoError(t, json.NewEncoder(w).Encode(&rest_model_zrok.ShareResponse{ShareToken: fmt.Sprintf("token-%v", n), FrontendProxyEndpoints: []string{"https://public.example"}}))
		case "/api/v2/unshare":
			body := &restshare.UnshareBody{}
			require.NoError(t, json.NewDecoder(req.Body).Decode(body))
			require.Equal(t, "original-environment", body.EnvZID)
			deletes.Add(1)
		default:
			t.Errorf("unexpected request '%v'", req.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	t.Cleanup(server.Close)
	previous, err := env_v0_4.Default()
	require.NoError(t, err)
	env_v0_4.SetRootDirName(t.TempDir())
	t.Cleanup(func() { env_v0_4.SetRootDirName(previous.Metadata().RootPath) })
	root, err := env_v0_4.Default()
	require.NoError(t, err)
	require.NoError(t, root.SetEnvironment(&env_core.Environment{AccountToken: "original-account", ZitiIdentity: "original-environment", ApiEndpoint: server.URL}))
	config := filepath.Join(root.Metadata().RootPath, "config.json")
	writeDefault := func(namespace string) {
		t.Helper()
		// the refresh must not adopt a changed endpoint or credentials from this file.
		require.NoError(t, os.WriteFile(config, []byte(`{"default_namespace":"`+namespace+`","api_endpoint":"http://unselected.invalid"}`), 0600))
	}
	writeDefault("team-a")
	a := newLifecycleAgent(t)
	a.root = root
	a.deps.createPublic = func(ctx context.Context, req *sdk.ShareRequest) (*sdk.Share, error) {
		return sdk.CreateShareContext(ctx, root, req)
	}
	a.deps.deletePublic = a.deletePublicShare
	a.deps.startPublic = func(opts proctree.ControlledOptions, args ...string) (*proctree.Child, error) {
		return proctree.StartControlledChild(opts, os.Args[0], "-test.run=^TestAgentLifecycleHelper$", "--", "boot", args[5])
	}
	service := &agentGrpcImpl{agent: a}
	firstRequest := &agentGrpc.SharePublicRequest{BackendMode: "proxy", Target: "8080"}
	var firstResponse *agentGrpc.SharePublicResponse
	first := asyncResult(func() error {
		var err error
		firstResponse, err = service.SharePublic(context.Background(), firstRequest)
		return err
	})
	var prepared *rest_model_zrok.ShareRequest
	select {
	case prepared = <-firstCreate:
	case <-time.After(5 * time.Second):
		t.Fatal("first create did not reach the controller")
	}
	writeDefault("team-b")
	require.NoError(t, os.WriteFile(filepath.Join(root.Metadata().RootPath, "environment.json"), []byte(fmt.Sprintf(`{"zrok_token":"changed-account","ziti_identity":"changed-environment","api_endpoint":%q}`, server.URL)), 0600))
	secondRequest := &agentGrpc.SharePublicRequest{BackendMode: "proxy", Target: "8080", NameSelections: []*agentGrpc.NameSelection{}}
	second, err := service.SharePublic(context.Background(), secondRequest)
	require.NoError(t, err)
	require.Equal(t, []string{"https://public.example"}, second.FrontendEndpoints)
	require.Equal(t, "team-a", prepared.NameSelections[0].NamespaceToken)
	unblock.Do(func() { close(finishFirst) })
	require.NoError(t, lifecycleResult(t, first))
	require.Nil(t, firstRequest.NameSelections)
	require.Empty(t, secondRequest.NameSelections)
	// explicit CLI selections remain authoritative even with unreadable current configuration.
	require.NoError(t, os.WriteFile(config, []byte("{"), 0600))
	explicit, err := service.SharePublic(context.Background(), &agentGrpc.SharePublicRequest{BackendMode: "proxy", Target: "8080", NameSelections: []*agentGrpc.NameSelection{{NamespaceToken: "cli-team", Name: "app"}}})
	require.NoError(t, err)
	a.mu.Lock()
	selections := map[string]string{}
	for _, op := range a.operations {
		selections[op.resourceToken()] = op.shr.nameSelections[0].NamespaceToken
	}
	a.mu.Unlock()
	require.Equal(t, map[string]string{firstResponse.Token: "team-a", second.Token: "team-b", explicit.Token: "cli-team"}, selections)
	require.Same(t, root, a.root)
	for _, token := range []string{firstResponse.Token, second.Token, explicit.Token} {
		require.NoError(t, a.ReleaseShare(token))
	}
	require.EqualValues(t, 3, creates.Load())
	require.EqualValues(t, 3, deletes.Load())
}

func TestPublicUISelectionsAndPreparationFailuresBeforeCreate(t *testing.T) {
	for _, test := range []struct {
		name       string
		config     string
		override   string
		unreadable bool
		selections []*agentGrpc.NameSelection
		want       []sdk.NameSelection
		fail       bool
	}{
		{name: "nil UI list", config: `{"default_namespace":"team"}`, want: []sdk.NameSelection{{NamespaceToken: "team"}}},
		{name: "empty UI list", config: `{"default_namespace":"team"}`, selections: []*agentGrpc.NameSelection{}, want: []sdk.NameSelection{{NamespaceToken: "team"}}},
		{name: "process override", config: `{"default_namespace":"changed"}`, override: "process-team", want: []sdk.NameSelection{{NamespaceToken: "process-team"}}},
		{name: "missing config", want: []sdk.NameSelection{{NamespaceToken: "public"}}},
		{name: "unset config", config: `{}`, want: []sdk.NameSelection{{NamespaceToken: "public"}}},
		{name: "malformed config", config: `{`, fail: true},
		{name: "unreadable config", unreadable: true, fail: true},
		{name: "explicit named", config: `{`, selections: []*agentGrpc.NameSelection{{NamespaceToken: "cli", Name: "app"}}, want: []sdk.NameSelection{{NamespaceToken: "cli", Name: "app"}}},
		{name: "explicit namespace", config: `{"default_namespace":"different"}`, selections: []*agentGrpc.NameSelection{{NamespaceToken: "explicit"}}, want: []sdk.NameSelection{{NamespaceToken: "explicit"}}},
		{name: "multiple selections", selections: []*agentGrpc.NameSelection{{NamespaceToken: "one", Name: "app"}, {NamespaceToken: "two"}}, want: []sdk.NameSelection{{NamespaceToken: "one", Name: "app"}, {NamespaceToken: "two"}}},
		{name: "invalid supplied selection", override: "fallback", selections: []*agentGrpc.NameSelection{{Name: "app"}}, fail: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("ZROK2_DEFAULT_NAMESPACE", test.override)
			a := newLifecycleAgent(t)
			path := filepath.Join(a.root.Metadata().RootPath, "config.json")
			if test.unreadable {
				require.NoError(t, os.Mkdir(path, 0700))
			} else if test.config != "" {
				require.NoError(t, os.WriteFile(path, []byte(test.config), 0600))
			}
			var creates atomic.Int32
			a.deps.createPublic = func(_ context.Context, req *sdk.ShareRequest) (*sdk.Share, error) {
				creates.Add(1)
				require.Equal(t, test.want, req.NameSelections)
				return &sdk.Share{Token: "known-token", FrontendEndpoints: []string{"https://public.example"}}, nil
			}
			request := &agentGrpc.SharePublicRequest{BackendMode: "proxy", Target: "8080", NameSelections: test.selections}
			response, err := (&agentGrpcImpl{agent: a}).SharePublic(context.Background(), request)
			if test.fail {
				require.Error(t, err)
				require.Zero(t, creates.Load())
				return
			}
			require.NoError(t, err)
			require.Equal(t, []string{"https://public.example"}, response.FrontendEndpoints)
			require.Equal(t, test.selections, request.NameSelections)
			require.EqualValues(t, 1, creates.Load())
			require.NoError(t, a.ReleaseShare(response.Token))
		})
	}
}

func TestPublicExitReleaseAndShutdownJoinBlockedCleanup(t *testing.T) {
	a := newLifecycleAgent(t)
	deleting, finishDelete, removed := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var unblock, notify sync.Once
	defer unblock.Do(func() { close(finishDelete) })
	a.deps.saveRegistry = func(registry *Registry, path string) error {
		err := registry.Save(path)
		if len(registry.PublicShares) == 0 {
			notify.Do(func() { close(removed) })
		}
		return err
	}
	var creates, deletes atomic.Int32
	a.deps.createPublic = func(context.Context, *sdk.ShareRequest) (*sdk.Share, error) {
		creates.Add(1)
		return &sdk.Share{Token: "known-token"}, nil
	}
	a.deps.deletePublic = func(context.Context, string) error { deletes.Add(1); close(deleting); <-finishDelete; return nil }
	token, _, err := a.SharePublic(publicRequest())
	require.NoError(t, err)
	op := onlyOperation(t, a)
	a.mu.Lock()
	child := op.child
	a.mu.Unlock()
	require.NoError(t, proctree.StopChild(child))
	waitLifecycle(t, deleting)
	released := asyncResult(func() error { return a.ReleaseShare(token) })
	waitLifecycle(t, removed)
	stopped := asyncResult(a.Shutdown)
	waitLifecycle(t, a.ctx.Done())
	a.retryDue(time.Now().Add(time.Hour))
	require.EqualValues(t, 1, creates.Load())
	for _, result := range []<-chan error{released, stopped} {
		select {
		case <-result:
			t.Fatal("cleanup caller returned before deletion completed")
		default:
		}
	}
	unblock.Do(func() { close(finishDelete) })
	require.NoError(t, lifecycleResult(t, released))
	require.NoError(t, lifecycleResult(t, stopped))
	require.EqualValues(t, 1, deletes.Load())
	require.Empty(t, savedRegistry(t, a).PublicShares)
}

func TestPublicLostDeleteResponseIsReconciledOnRetry(t *testing.T) {
	a := newLifecycleAgent(t)
	var deletes, overviews atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch req.URL.Path {
		case "/api/v2/clientVersionCheck":
			return
		case "/api/v2/unshare":
			if deletes.Add(1) == 1 {
				connection, _, err := w.(http.Hijacker).Hijack()
				require.NoError(t, err)
				_ = connection.Close() // deletion committed, but no response reached the agent.
				return
			}
			w.WriteHeader(http.StatusNotFound)
		case "/api/v2/overview":
			overviews.Add(1)
			require.NoError(t, json.NewEncoder(w).Encode(&rest_model_zrok.Overview{Environments: []*rest_model_zrok.EnvironmentAndResources{{Environment: &rest_model_zrok.Environment{ZID: "original-environment"}}}}))
		default:
			t.Errorf("unexpected request '%v'", req.URL.Path)
		}
	}))
	defer server.Close()
	usePublicHTTPRoot(t, a, server.URL)
	a.deps.deletePublic = a.deletePublicShare
	token, _, err := a.SharePublic(publicRequest())
	require.NoError(t, err)
	err = a.ReleaseShare(token)
	require.Error(t, err)
	require.ErrorContains(t, err, "pending cleanup")
	require.Zero(t, overviews.Load())
	op := onlyOperation(t, a)
	a.mu.Lock()
	retained := op.resourceToken()
	a.mu.Unlock()
	require.Equal(t, token, retained)
	require.NoError(t, a.ReleaseShare(token))
	require.EqualValues(t, 2, deletes.Load())
	require.EqualValues(t, 1, overviews.Load())
}

func TestPublicShutdownMakesFinalAttemptForFailedCleanup(t *testing.T) {
	a := newLifecycleAgent(t)
	var deletes atomic.Int32
	failure := errors.New("controller unavailable")
	a.deps.deletePublic = func(context.Context, string) error { deletes.Add(1); return failure }
	token, _, err := a.SharePublic(publicRequest())
	require.NoError(t, err)
	require.ErrorIs(t, a.ReleaseShare(token), failure)
	require.EqualValues(t, 1, deletes.Load())
	err = a.Shutdown()
	require.ErrorIs(t, err, failure)
	require.ErrorContains(t, err, token)
	require.EqualValues(t, 2, deletes.Load())
	require.Same(t, err, a.Shutdown())
	require.EqualValues(t, 2, deletes.Load())
}
