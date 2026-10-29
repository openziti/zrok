package sdk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

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
