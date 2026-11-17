package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/openziti/zrok/v2/agent/agentGrpc"
	"github.com/openziti/zrok/v2/agent/proctree"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/stretchr/testify/require"
)

// legacyOverview answers the account overview with the agent's environment and, when listed, the resource a
// cleanup is asking about.
func legacyOverview(kind string, listed bool, environment string) *rest_model_zrok.EnvironmentAndResources {
	resources := &rest_model_zrok.EnvironmentAndResources{Environment: &rest_model_zrok.Environment{ZID: environment}}
	if listed && kind == "access" {
		resources.Frontends = rest_model_zrok.Frontends{{FrontendToken: "known-frontend", ShareToken: "shared"}}
	}
	if listed && kind == "private share" {
		resources.Shares = rest_model_zrok.Shares{{ShareToken: "known"}}
	}
	return resources
}

func TestLegacyCleanupRequiresAuthenticatedConfirmedAbsence(t *testing.T) {
	for _, kind := range []string{"access", "private share"} {
		for _, outcome := range []string{"absent", "present", "other environment token", "missing environment", "overview unauthorized", "overview failure", "delete unauthorized", "generic 404", "delete failure", "successful delete"} {
			t.Run(kind+"/"+outcome, func(t *testing.T) {
				a := newLifecycleAgent(t)
				var overviews atomic.Int32
				deletePath := "/api/v2/unaccess"
				if kind == "private share" {
					deletePath = "/api/v2/unshare"
				}
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					if req.URL.Path == "/api/v2/clientVersionCheck" {
						// a 404 here is not the delete's typed answer; the client never gets as far as the delete.
						if outcome == "generic 404" {
							w.WriteHeader(404)
						}
						return
					}
					require.Equal(t, "original-account", req.Header.Get("X-TOKEN"))
					switch req.URL.Path {
					case deletePath:
						var body map[string]any
						require.NoError(t, json.NewDecoder(req.Body).Decode(&body))
						require.Equal(t, "original-environment", body["envZId"])
						if kind == "access" {
							require.Equal(t, "known-frontend", body["frontendToken"])
							require.Equal(t, "shared", body["shareToken"])
						} else {
							require.Equal(t, "known", body["shareToken"])
						}
						switch outcome {
						case "successful delete":
						case "delete unauthorized":
							w.WriteHeader(401)
						case "delete failure":
							w.WriteHeader(500)
						default:
							w.WriteHeader(404)
						}
					case "/api/v2/overview":
						overviews.Add(1)
						switch outcome {
						case "overview unauthorized":
							w.WriteHeader(401)
							return
						case "overview failure":
							w.WriteHeader(500)
							return
						}
						overview := &rest_model_zrok.Overview{Environments: []*rest_model_zrok.EnvironmentAndResources{legacyOverview(kind, outcome == "present", "original-environment")}}
						if outcome == "missing environment" {
							overview.Environments = nil
						}
						if outcome == "other environment token" {
							overview.Environments = append(overview.Environments, legacyOverview(kind, true, "other"))
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
				var err error
				if kind == "access" {
					err = a.deleteAccess(ctx, "shared", "known-frontend")
				} else {
					err = a.deletePrivateShare(ctx, "known")
				}
				if outcome == "absent" || outcome == "successful delete" {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
				}
				// only the controller's typed not-found is worth asking the overview about.
				switch outcome {
				case "delete unauthorized", "generic 404", "delete failure", "successful delete":
					require.Zero(t, overviews.Load())
				default:
					require.EqualValues(t, 1, overviews.Load())
				}
			})
		}
	}
}

// the controller removes every access to a share when the share is torn down, so an access is often gone before
// the agent deletes it. such an access must finish cleanup and be releasable; while the overview still lists it,
// the cleanup failure stays.
func TestAccessRemovedByControllerCanBeReleased(t *testing.T) {
	a := newLifecycleAgent(t)
	var listed atomic.Bool
	var deletes atomic.Int32
	firstDelete := make(chan struct{})
	listed.Store(true)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch req.URL.Path {
		case "/api/v2/clientVersionCheck":
		case "/api/v2/unaccess":
			if deletes.Add(1) == 1 {
				close(firstDelete)
			}
			w.WriteHeader(404)
		case "/api/v2/overview":
			overview := &rest_model_zrok.Overview{Environments: []*rest_model_zrok.EnvironmentAndResources{legacyOverview("access", listed.Load(), "original-environment")}}
			require.NoError(t, json.NewEncoder(w).Encode(overview))
		default:
			t.Errorf("unexpected path '%v'", req.URL.Path)
			w.WriteHeader(500)
		}
	}))
	defer server.Close()
	usePublicHTTPRoot(t, a, server.URL)
	a.deps.deleteAccess = a.deleteAccess
	a.deps.startChild = func(tail proctree.TailFunction, _ ...string) (*proctree.Child, error) {
		return lifecycleChild(tail, "boot")
	}
	_, err := a.AccessPrivate(&AccessPrivateRequest{ShareToken: "shared"})
	require.NoError(t, err)
	op := onlyOperation(t, a)
	a.mu.Lock()
	child := op.child
	a.mu.Unlock()

	// the worker dies; the controller answers not-found but the overview still lists the access: unresolved.
	require.NoError(t, proctree.StopChild(child))
	waitLifecycle(t, firstDelete)
	waitLifecycle(t, cleanupDone(a, op))
	a.mu.Lock()
	require.Equal(t, phaseFailed, op.phase)
	require.Equal(t, "known-frontend", op.resourceToken())
	a.mu.Unlock()
	require.Error(t, a.ReleaseAccess("known-frontend"))

	// once the overview agrees the access is gone, the same not-found finishes the cleanup.
	listed.Store(false)
	require.NoError(t, a.ReleaseAccess("known-frontend"))
	require.EqualValues(t, 3, deletes.Load())
	status, err := (&agentGrpcImpl{agent: a}).Status(context.Background(), &agentGrpc.StatusRequest{})
	require.NoError(t, err)
	require.Empty(t, status.Accesses)
	require.Empty(t, savedRegistry(t, a).PrivateAccesses)
}
