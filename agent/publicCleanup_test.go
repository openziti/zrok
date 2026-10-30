package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/openziti/zrok/v2/environment/env_core"
	"github.com/openziti/zrok/v2/environment/env_v0_4"
	"github.com/openziti/zrok/v2/rest_client_zrok"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/stretchr/testify/require"
)

type publicHTTPRoot struct {
	*lifecycleRoot
	client *env_v0_4.Root
}

func (r *publicHTTPRoot) Environment() *env_core.Environment {
	return &env_core.Environment{AccountToken: "original-account", ZitiIdentity: "original-environment"}
}
func (r *publicHTTPRoot) ClientContext(ctx context.Context) (*rest_client_zrok.Zrok, error) {
	return r.client.ClientContext(ctx)
}

func usePublicHTTPRoot(t *testing.T, a *Agent, endpoint string) {
	t.Helper()
	t.Setenv("ZROK2_API_ENDPOINT", endpoint)
	client, err := env_v0_4.Default()
	require.NoError(t, err)
	a.root = &publicHTTPRoot{lifecycleRoot: a.root.(*lifecycleRoot), client: client}
}

func TestPublicCleanupRequiresAuthenticatedConfirmedAbsence(t *testing.T) {
	for _, outcome := range []string{"absent", "present", "other environment token", "missing environment", "overview unauthorized", "overview failure", "unshare unauthorized", "generic 404", "successful delete"} {
		t.Run(outcome, func(t *testing.T) {
			a := newLifecycleAgent(t)
			var overviews atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if req.URL.Path == "/api/v2/clientVersionCheck" {
					if outcome == "generic 404" {
						w.WriteHeader(404)
					}
					return
				}
				require.Equal(t, "original-account", req.Header.Get("X-TOKEN"))
				switch req.URL.Path {
				case "/api/v2/unshare":
					var body map[string]any
					require.NoError(t, json.NewDecoder(req.Body).Decode(&body))
					require.Equal(t, "original-environment", body["envZId"])
					if outcome == "successful delete" {
						return
					}
					if outcome == "unshare unauthorized" {
						w.WriteHeader(401)
					} else {
						w.WriteHeader(404)
					}
					if outcome == "generic 404" {
						_, _ = w.Write([]byte(`{"error":"untyped proxy response"}`))
					} else {
						_, _ = w.Write([]byte(`"not found"`))
					}
				case "/api/v2/overview":
					overviews.Add(1)
					if outcome == "overview unauthorized" {
						w.WriteHeader(401)
						return
					}
					if outcome == "overview failure" {
						w.WriteHeader(500)
						return
					}
					resources := &rest_model_zrok.EnvironmentAndResources{Environment: &rest_model_zrok.Environment{ZID: "original-environment"}}
					overview := &rest_model_zrok.Overview{Environments: []*rest_model_zrok.EnvironmentAndResources{resources}}
					if outcome == "present" {
						resources.Shares = rest_model_zrok.Shares{{ShareToken: "known"}}
					}
					if outcome == "missing environment" {
						overview.Environments = nil
					}
					if outcome == "other environment token" {
						overview.Environments = append(overview.Environments, &rest_model_zrok.EnvironmentAndResources{Environment: &rest_model_zrok.Environment{ZID: "other"}, Shares: rest_model_zrok.Shares{{ShareToken: "known"}}})
					}
					require.NoError(t, json.NewEncoder(w).Encode(overview))
				default:
					t.Errorf("unexpected path '%v'", req.URL.Path)
					w.WriteHeader(500)
				}
			}))
			defer server.Close()
			usePublicHTTPRoot(t, a, server.URL)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			err := a.deletePublicShare(ctx, "known")
			if outcome == "absent" || outcome == "successful delete" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
			if outcome == "unshare unauthorized" || outcome == "generic 404" || outcome == "successful delete" {
				require.Zero(t, overviews.Load())
			} else {
				require.EqualValues(t, 1, overviews.Load())
			}
		})
	}
}

func TestPublicOverviewInitializationConsumesRemainingDeadline(t *testing.T) {
	a := newLifecycleAgent(t)
	var versions, overviews atomic.Int32
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch req.URL.Path {
		case "/api/v2/clientVersionCheck":
			if versions.Add(1) == 2 {
				select {
				case <-req.Context().Done():
				case <-release:
				}
			}
		case "/api/v2/unshare":
			w.WriteHeader(404)
			_, _ = w.Write([]byte(`"missing"`))
		case "/api/v2/overview":
			overviews.Add(1)
		}
	}))
	defer server.Close()
	defer close(release)
	usePublicHTTPRoot(t, a, server.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	started := time.Now()
	err := a.deletePublicShare(ctx, "known")
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Less(t, time.Since(started), time.Second)
	require.EqualValues(t, 2, versions.Load())
	require.Zero(t, overviews.Load())
}

func TestPublicCleanupInitializationFailureReleasesAttempt(t *testing.T) {
	a := newLifecycleAgent(t)
	var calls atomic.Int32
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if calls.Add(1) == 1 {
			select {
			case <-req.Context().Done():
			case <-release:
			}
			return
		}
		w.WriteHeader(200)
	}))
	defer server.Close()
	defer close(release)
	usePublicHTTPRoot(t, a, server.URL)
	a.options.cleanup = 100 * time.Millisecond
	a.deps.deletePublic = a.deletePublicShare
	token, _, err := a.SharePublic(publicRequest())
	require.NoError(t, err)
	op := onlyOperation(t, a)
	require.ErrorIs(t, a.ReleaseShare(token), context.DeadlineExceeded)
	a.mu.Lock()
	require.Equal(t, token, op.resourceToken())
	require.Equal(t, phaseFailed, op.phase)
	first := op.cleanup
	a.mu.Unlock()
	require.NoError(t, a.ReleaseShare(token))
	a.mu.Lock()
	require.NotSame(t, first, op.cleanup)
	a.mu.Unlock()
}
