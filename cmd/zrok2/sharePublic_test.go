package main

import (
	"context"
	"errors"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/openziti/zrok/v2/environment/env_core"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/stretchr/testify/require"
)

func TestPublicCLIRequestPreparation(t *testing.T) {
	cmd := &sharePublicCommand{backendMode: "proxy", nameSelections: []string{"team:app", "other"}, basicAuth: []string{"u:p"}, accessGrants: []string{"account"}}
	out, err := cmd.prepareRequest([]string{"443"})
	require.NoError(t, err)
	require.Equal(t, &sdk.ShareRequest{BackendMode: sdk.ProxyBackendMode, ShareMode: sdk.PublicShareMode, Target: "https://127.0.0.1:443", NameSelections: []sdk.NameSelection{{NamespaceToken: "team", Name: "app"}, {NamespaceToken: "other"}}, BasicAuth: []string{"u:p"}, AccessGrants: []string{"account"}, PermissionMode: sdk.ClosedPermissionMode}, out)
	for _, args := range [][]string{nil, {"a", "b"}} {
		_, err := cmd.prepareRequest(args)
		require.Error(t, err)
	}
	cmd.basicAuth = nil
	cmd.oauthProvider = "provider"
	cmd.oauthEmailDomains = []string{"*@example.com"}
	cmd.oauthRefreshInterval = 3 * time.Hour
	cmd.open = true
	out, err = cmd.prepareRequest([]string{"8080"})
	require.NoError(t, err)
	require.Equal(t, 3*time.Hour, out.OauthRefreshInterval)
	require.Equal(t, sdk.OpenPermissionMode, out.PermissionMode)
	for _, mode := range []string{"web", "drive", "caddy"} {
		cmd.backendMode = mode
		out, err = cmd.prepareRequest([]string{"relative"})
		require.NoError(t, err)
		want, err := filepath.Abs("relative")
		require.NoError(t, err)
		require.Equal(t, want, out.Target)
	}
}

func TestPublicCLIRejectsEmptyNamespaces(t *testing.T) {
	for _, selections := range [][]string{{""}, {":app"}, {"team", ""}} {
		cmd := &sharePublicCommand{backendMode: "proxy", nameSelections: selections}
		out, err := cmd.prepareRequest([]string{"8080"})
		require.ErrorContains(t, err, "namespace")
		require.Nil(t, out)
		require.Equal(t, selections, cmd.nameSelections)
	}
}

type publicTestBackend struct {
	runs  atomic.Int32
	stops atomic.Int32
	run   func() error
	stop  func() error
}

func (b *publicTestBackend) Run() error {
	b.runs.Add(1)
	if b.run != nil {
		return b.run()
	}
	return nil
}
func (b *publicTestBackend) Stop() error {
	b.stops.Add(1)
	if b.stop != nil {
		return b.stop()
	}
	return nil
}

func receivePublicResult(t *testing.T, ch <-chan error) error {
	t.Helper()
	select {
	case err := <-ch:
		return err
	case <-time.After(5 * time.Second):
		t.Fatal("public backend did not complete")
		return nil
	}
}

func TestPublicLateBackendInstallation(t *testing.T) {
	for _, mode := range []string{"proxy", "web", "caddy", "drive"} {
		t.Run(mode, func(t *testing.T) {
			r := &publicBackendRunner{}
			entered, release := make(chan struct{}), make(chan struct{})
			be := &publicTestBackend{}
			started := make(chan error, 1)
			go func() {
				started <- r.start(context.Background(), mode, func() (shareBackend, error) { close(entered); <-release; return be, nil })
			}()
			<-entered
			r.stop()
			r.stop()
			close(release)
			require.NoError(t, receivePublicResult(t, started))
			require.EqualValues(t, 0, be.runs.Load())
			require.EqualValues(t, 1, be.stops.Load())
		})
	}
}

func TestPublicBackendFailuresAndStop(t *testing.T) {
	failure := errors.New("backend failed")
	for _, mode := range []string{"web", "caddy"} {
		r := &publicBackendRunner{}
		be := &publicTestBackend{run: func() error { return failure }}
		require.ErrorIs(t, r.start(context.Background(), mode, func() (shareBackend, error) { return be, nil }), failure)
		r.stop()
		require.EqualValues(t, 1, be.stops.Load())
	}
	r := &publicBackendRunner{}
	be := &publicTestBackend{run: func() error { return failure }}
	require.NoError(t, r.start(context.Background(), "proxy", func() (shareBackend, error) { return be, nil }))
	require.ErrorIs(t, receivePublicResult(t, r.failed), failure)
	r.stop()
	r = &publicBackendRunner{}
	exited := make(chan struct{})
	stopped := make(chan struct{})
	be = &publicTestBackend{run: func() error { defer close(exited); <-stopped; return failure }, stop: func() error { close(stopped); return nil }}
	require.NoError(t, r.start(context.Background(), "drive", func() (shareBackend, error) { return be, nil }))
	r.stop()
	<-exited
	// join the observer's isStopping call before checking failure reporting.
	r.backend.mu.Lock()
	r.backend.mu.Unlock()
	select {
	case err := <-r.failed:
		t.Fatalf("stop reported as failure: %v", err)
	default:
	}
}

func TestStandalonePublicLifecycle(t *testing.T) {
	failure := errors.New("startup failed")
	for _, runErr := range []error{nil, failure} {
		var events []string
		shr := &sdk.Share{Token: "created"}
		err := publicShareLifecycle(nil, &sdk.ShareRequest{}, func(env_core.Root, *sdk.ShareRequest) (*sdk.Share, error) {
			events = append(events, "create")
			return shr, nil
		}, func(_ env_core.Root, got *sdk.Share) error {
			require.Same(t, shr, got)
			events = append(events, "delete")
			return nil
		}, func(got *sdk.Share) error {
			require.Same(t, shr, got)
			events = append(events, "run", "stop")
			return runErr
		})
		require.ErrorIs(t, err, runErr)
		require.Equal(t, []string{"create", "run", "stop", "delete"}, events)
	}
	deletes := 0
	err := publicShareLifecycle(nil, nil, func(env_core.Root, *sdk.ShareRequest) (*sdk.Share, error) { return nil, failure }, func(env_core.Root, *sdk.Share) error { deletes++; return nil }, func(*sdk.Share) error { t.Fatal("ran without a share"); return nil })
	require.ErrorIs(t, err, failure)
	require.Zero(t, deletes)
	deleteErr := errors.New("delete failed")
	err = publicShareLifecycle(nil, nil, func(env_core.Root, *sdk.ShareRequest) (*sdk.Share, error) { return &sdk.Share{Token: "known"}, nil }, func(env_core.Root, *sdk.Share) error { return deleteErr }, func(*sdk.Share) error { return failure })
	require.ErrorIs(t, err, failure)
	require.ErrorIs(t, err, deleteErr)
	require.ErrorContains(t, err, "'known'")
}
