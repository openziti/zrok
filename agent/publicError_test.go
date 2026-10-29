package agent

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-openapi/runtime"
	"github.com/openziti/zrok/v2/agent/agentGrpc"
	"github.com/openziti/zrok/v2/agent/proctree"
	"github.com/openziti/zrok/v2/rest_client_zrok/metadata"
	restshare "github.com/openziti/zrok/v2/rest_client_zrok/share"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func publicErrorClient(t *testing.T, a *Agent) agentGrpc.AgentClient {
	t.Helper()
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	agentGrpc.RegisterAgentServer(server, &agentGrpcImpl{agent: a})
	go func() { _ = server.Serve(listener) }()
	connection, err := grpc.NewClient("passthrough:///agent", grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }))
	require.NoError(t, err)
	t.Cleanup(func() { _ = connection.Close(); server.Stop(); _ = listener.Close() })
	return agentGrpc.NewAgentClient(connection)
}

func TestPublicCreateControllerAnswerSurvivesRPC(t *testing.T) {
	for _, test := range []struct {
		name string
		err  error
		want agentGrpc.ControllerResponse
	}{
		{"conflict", &restshare.ShareConflict{Payload: rest_model_zrok.ErrorMessage("name 'app' is held by share 'existing'")}, agentGrpc.ControllerResponse{Status: 409, Payload: "name 'app' is held by share 'existing'"}},
		{"unauthorized", restshare.NewShareUnauthorized(), agentGrpc.ControllerResponse{Status: 401}},
		{"rate limited", &runtime.APIError{OperationName: "share", Code: 429}, agentGrpc.ControllerResponse{Status: 429}},
		{"server error", &restshare.ShareInternalServerError{Payload: rest_model_zrok.ErrorMessage("controller failure")}, agentGrpc.ControllerResponse{Status: 500, Payload: "controller failure"}},
		{"busy", &restshare.ShareServiceUnavailable{RetryAfter: 5}, agentGrpc.ControllerResponse{Status: 503, RetryAfter: 5}},
	} {
		t.Run(test.name, func(t *testing.T) {
			a := newLifecycleAgent(t)
			a.deps.createPublic = func(context.Context, *sdk.ShareRequest) (*sdk.Share, error) {
				return nil, errors.Join(errors.New("create failed"), test.err)
			}
			a.deps.startPublic = func(proctree.ControlledOptions, ...string) (*proctree.Child, error) {
				t.Error("refused create launched a worker")
				return nil, errors.New("unexpected worker")
			}
			client := publicErrorClient(t, a)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, err := client.SharePublic(ctx, &agentGrpc.SharePublicRequest{BackendMode: "proxy", Target: "8080", NameSelections: []*agentGrpc.NameSelection{{NamespaceToken: "team", Name: "app"}}})
			require.Error(t, err)
			answer, ok := agentGrpc.ControllerResponseFromError(err)
			require.True(t, ok)
			require.Equal(t, test.want, answer)
			require.Contains(t, err.Error(), "create failed")
		})
	}
}

func TestPublicCleanupControllerErrorDoesNotBecomeCreateRefusal(t *testing.T) {
	a := newLifecycleAgent(t)
	a.deps.startPublic = func(proctree.ControlledOptions, ...string) (*proctree.Child, error) {
		return nil, errors.New("worker spawn failed")
	}
	a.deps.deletePublic = func(context.Context, string) error { return restshare.NewUnshareUnauthorized() }
	client := publicErrorClient(t, a)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := client.SharePublic(ctx, &agentGrpc.SharePublicRequest{BackendMode: "proxy", Target: "8080", NameSelections: []*agentGrpc.NameSelection{{NamespaceToken: "team", Name: "app"}}})
	require.Error(t, err)
	_, controller := agentGrpc.ControllerResponseFromError(err)
	require.False(t, controller)
	require.Contains(t, err.Error(), "worker spawn failed")
	require.Contains(t, err.Error(), "pending cleanup for 'known-token'")
}

func TestPublicVersionCheckRejectionPreservesControllerMessage(t *testing.T) {
	const message = "client upgrade required; minimum supported version is 'v2.0.9'"
	var versionChecks, laterRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if req.URL.Path == "/api/v2/clientVersionCheck" {
			versionChecks.Add(1)
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`"` + message + `"`))
			return
		}
		laterRequests.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	a := newLifecycleAgent(t)
	usePublicHTTPRoot(t, a, server.URL)
	a.deps.createPublic = func(ctx context.Context, request *sdk.ShareRequest) (*sdk.Share, error) {
		return sdk.CreateShareContext(ctx, a.root, request)
	}
	a.deps.startPublic = func(proctree.ControlledOptions, ...string) (*proctree.Child, error) {
		t.Error("version rejection launched a worker")
		return nil, errors.New("unexpected worker")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// the environment wrapper already preserves the generated error and its original payload.
	_, err := sdk.CreateShareContext(ctx, a.root, &sdk.ShareRequest{ShareMode: sdk.PublicShareMode})
	var rejection *metadata.ClientVersionCheckBadRequest
	require.ErrorAs(t, err, &rejection)
	require.Equal(t, message, rejection.Payload)
	client := publicErrorClient(t, a)
	_, err = client.SharePublic(ctx, &agentGrpc.SharePublicRequest{BackendMode: "proxy", Target: "8080", NameSelections: []*agentGrpc.NameSelection{{NamespaceToken: "team", Name: "app"}}})
	require.Error(t, err)
	answer, ok := agentGrpc.ControllerResponseFromError(err)
	require.True(t, ok)
	require.Equal(t, agentGrpc.ControllerResponse{Status: http.StatusBadRequest, Payload: message}, answer)
	require.EqualValues(t, 2, versionChecks.Load())
	require.Zero(t, laterRequests.Load())
}
