package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/openziti/zrok/v2/endpoints"
	"github.com/openziti/zrok/v2/endpoints/drive"
	"github.com/openziti/zrok/v2/endpoints/proxy"
	"github.com/openziti/zrok/v2/environment/env_core"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
)

// publicShareLifecycle leaves backend execution to run, whose return joins backend stop before deletion.
func publicShareLifecycle(root env_core.Root, req *sdk.ShareRequest,
	create func(env_core.Root, *sdk.ShareRequest) (*sdk.Share, error),
	remove func(env_core.Root, *sdk.Share) error, run func(*sdk.Share) error) error {
	shr, err := create(root, req)
	if err != nil {
		return err
	}
	runErr := run(shr)
	deleteErr := remove(root, shr)
	if deleteErr != nil {
		deleteErr = fmt.Errorf("error deleting share '%v': %w", shr.Token, deleteErr)
	}
	return errors.Join(runErr, deleteErr)
}

// publicBackendRunner manages only backend startup and stop, never controller resources.
type publicBackendRunner struct {
	backend runningBackend
	failed  chan error
}

func (r *publicBackendRunner) start(ctx context.Context, mode string, construct func() (shareBackend, error)) error {
	r.failed = make(chan error, 1)
	be, err := construct()
	if err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return err
	}
	// serialize Caddy's inline startup with stop so a stopped backend cannot be reactivated.
	r.backend.mu.Lock()
	if r.backend.stopping || ctx.Err() != nil {
		r.backend.mu.Unlock()
		if err := be.Stop(); err != nil {
			return fmt.Errorf("error stopping late backend: %w", err)
		}
		return nil
	}
	r.backend.be = be
	if startsInline(mode) {
		err := be.Run()
		r.backend.mu.Unlock()
		if err != nil {
			return fmt.Errorf("unable to start '%v' backend: %w", mode, err)
		}
	} else {
		go func() {
			err := be.Run()
			if !r.backend.isStopping() {
				if err == nil {
					err = fmt.Errorf("backend stopped unexpectedly")
				}
				r.failed <- fmt.Errorf("error running '%v' backend: %w", mode, err)
			}
		}()
		r.backend.mu.Unlock()
	}
	return nil
}

func (r *publicBackendRunner) stop() { r.backend.stop() }

func newPublicBackend(root env_core.Root, req *sdk.ShareRequest, shr *sdk.Share, insecure bool, requests chan *endpoints.Request) (shareBackend, error) {
	zif, err := root.ZitiIdentityNamed(root.EnvironmentIdentityName())
	if err != nil {
		return nil, fmt.Errorf("unable to access ziti identity file: %w", err)
	}
	superNetwork, _ := root.SuperNetwork()
	switch req.BackendMode {
	case sdk.ProxyBackendMode:
		return proxy.NewBackend(&proxy.BackendConfig{IdentityPath: zif, EndpointAddress: req.Target, ShrToken: shr.Token, Insecure: insecure, Requests: requests, SuperNetwork: superNetwork})
	case sdk.WebBackendMode:
		return proxy.NewCaddyWebBackend(&proxy.CaddyWebBackendConfig{IdentityPath: zif, WebRoot: req.Target, ShrToken: shr.Token, Requests: requests})
	case sdk.CaddyBackendMode:
		return proxy.NewCaddyfileBackend(&proxy.CaddyfileBackendConfig{CaddyfilePath: req.Target, Shr: shr, Requests: requests})
	case sdk.DriveBackendMode:
		return drive.NewBackend(&drive.BackendConfig{IdentityPath: zif, DriveRoot: req.Target, ShrToken: shr.Token, Requests: requests, SuperNetwork: superNetwork})
	default:
		return nil, fmt.Errorf("invalid backend mode '%v'", req.BackendMode)
	}
}
