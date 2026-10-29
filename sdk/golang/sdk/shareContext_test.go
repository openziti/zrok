package sdk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-openapi/runtime"
	"github.com/openziti/zrok/v2/environment/env_core"
	"github.com/openziti/zrok/v2/environment/env_v0_4"
	"github.com/openziti/zrok/v2/rest_client_zrok"
	"github.com/stretchr/testify/require"
)

type shareHTTPRoot struct{ env_core.Root }

func (r *shareHTTPRoot) IsEnabled() bool { return true }
func (r *shareHTTPRoot) Environment() *env_core.Environment {
	return &env_core.Environment{AccountToken: "account", ZitiIdentity: "environment"}
}

type contextShareHTTPRoot struct{ *shareHTTPRoot }

func (r *contextShareHTTPRoot) ClientContext(ctx context.Context) (*rest_client_zrok.Zrok, error) {
	return env_core.ClientContext(ctx, r.Root)
}

func TestShareContextIncludesClientVersionCheck(t *testing.T) {
	for _, operation := range []string{"create", "delete"} {
		t.Run(operation, func(t *testing.T) {
			var laterRequests atomic.Int32
			release := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.URL.Path == "/api/v2/clientVersionCheck" {
					select {
					case <-req.Context().Done():
					case <-release:
					}
					return
				}
				laterRequests.Add(1)
			}))
			defer server.Close()
			defer close(release)
			t.Setenv("ZROK2_API_ENDPOINT", server.URL)
			base, err := env_v0_4.Default()
			require.NoError(t, err)
			root := &contextShareHTTPRoot{&shareHTTPRoot{Root: base}}
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			started := time.Now()
			if operation == "create" {
				_, err = CreateShareContext(ctx, root, &ShareRequest{ShareMode: PublicShareMode})
			} else {
				err = DeleteShareContext(ctx, root, &Share{Token: "owned"})
			}
			require.ErrorIs(t, err, context.DeadlineExceeded)
			require.Less(t, time.Since(started), time.Second)
			require.Zero(t, laterRequests.Load())
		})
	}
}

func TestShareContextSurvivesSuccessfulInitialization(t *testing.T) {
	for _, operation := range []string{"create", "delete"} {
		t.Run(operation, func(t *testing.T) {
			version, finishVersion, dispatched, release := make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.URL.Path == "/api/v2/clientVersionCheck" {
					close(version)
					<-finishVersion
					w.WriteHeader(http.StatusOK)
					return
				}
				close(dispatched)
				select {
				case <-req.Context().Done():
				case <-release:
				}
			}))
			defer server.Close()
			defer close(release)
			t.Setenv("ZROK2_API_ENDPOINT", server.URL)
			base, err := env_v0_4.Default()
			require.NoError(t, err)
			root := &contextShareHTTPRoot{&shareHTTPRoot{Root: base}}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			result := make(chan error, 1)
			go func() {
				if operation == "create" {
					_, err := CreateShareContext(ctx, root, &ShareRequest{ShareMode: PublicShareMode})
					result <- err
				} else {
					result <- DeleteShareContext(ctx, root, &Share{Token: "owned"})
				}
			}()
			<-version
			close(finishVersion)
			<-dispatched
			cancel()
			select {
			case err := <-result:
				require.ErrorIs(t, err, context.Canceled)
			case <-time.After(time.Second):
				t.Fatal("operation discarded the initialization context")
			}
		})
	}
}

func TestShareContextRequiresCapabilityAndLegacyRemainsCompatible(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		if req.URL.Path == "/api/v2/share" {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"shareToken":"created"}`))
		}
	}))
	defer server.Close()
	t.Setenv("ZROK2_API_ENDPOINT", server.URL)
	base, err := env_v0_4.Default()
	require.NoError(t, err)
	root := &shareHTTPRoot{Root: base}
	_, err = CreateShareContext(context.Background(), root, &ShareRequest{ShareMode: PublicShareMode})
	require.ErrorContains(t, err, "does not support context-aware")
	require.ErrorContains(t, DeleteShareContext(context.Background(), root, &Share{Token: "owned"}), "does not support context-aware")
	require.Zero(t, requests.Load())
	created, err := CreateShare(root, &ShareRequest{ShareMode: PublicShareMode})
	require.NoError(t, err)
	require.Equal(t, "created", created.Token)
	require.NoError(t, DeleteShare(root, created))
	require.EqualValues(t, 4, requests.Load())
}

type deadlineShareRoot struct {
	*contextShareHTTPRoot
	dispatched chan time.Time
}

func (r *deadlineShareRoot) ClientContext(ctx context.Context) (*rest_client_zrok.Zrok, error) {
	client, err := r.contextShareHTTPRoot.ClientContext(ctx)
	if err == nil {
		client.SetTransport(&deadlineShareTransport{ClientTransport: client.Transport, dispatched: r.dispatched})
	}
	return client, err
}

type deadlineShareTransport struct {
	runtime.ClientTransport
	dispatched chan time.Time
}

func (t *deadlineShareTransport) Submit(operation *runtime.ClientOperation) (any, error) {
	var deadline time.Time
	if operation.Context != nil {
		deadline, _ = operation.Context.Deadline()
	}
	t.dispatched <- deadline
	return t.ClientTransport.Submit(operation)
}

func TestShareDeadlineIsUnchangedAfterDelayedVersionCheck(t *testing.T) {
	for _, operation := range []string{"create", "delete"} {
		t.Run(operation, func(t *testing.T) {
			versionEntered, allowVersion, operationEntered, release := make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				if request.URL.Path == "/api/v2/clientVersionCheck" {
					close(versionEntered)
					select {
					case <-allowVersion:
					case <-request.Context().Done():
					}
					return
				}
				close(operationEntered)
				select {
				case <-request.Context().Done():
				case <-release:
				}
			}))
			defer server.Close()
			defer close(release)
			t.Setenv("ZROK2_API_ENDPOINT", server.URL)
			base, err := env_v0_4.Default()
			require.NoError(t, err)
			root := &deadlineShareRoot{contextShareHTTPRoot: &contextShareHTTPRoot{&shareHTTPRoot{Root: base}}, dispatched: make(chan time.Time, 1)}
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()
			originalDeadline, _ := ctx.Deadline()
			result := make(chan error, 1)
			go func() {
				if operation == "create" {
					_, err := CreateShareContext(ctx, root, &ShareRequest{ShareMode: PublicShareMode})
					result <- err
				} else {
					result <- DeleteShareContext(ctx, root, &Share{Token: "owned"})
				}
			}()
			select {
			case <-versionEntered:
			case <-ctx.Done():
				t.Fatal("version check was not dispatched")
			}
			select {
			case <-root.dispatched:
				t.Fatal("operation preceded version approval")
			default:
			}
			close(allowVersion)
			select {
			case deadline := <-root.dispatched:
				require.Equal(t, originalDeadline, deadline, "initialization reset the operation budget")
			case <-ctx.Done():
				t.Fatal("operation was not dispatched after version approval")
			}
			select {
			case <-operationEntered:
			case <-ctx.Done():
				t.Fatal("generated HTTP transport never reached the operation")
			}
			select {
			case err := <-result:
				require.ErrorIs(t, err, context.DeadlineExceeded)
			case <-time.After(time.Second):
				t.Fatal("stalled operation exceeded the original deadline")
			}
		})
	}
}
