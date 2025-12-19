package agent

import (
	"context"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/agent/agentGrpc"
	"github.com/openziti/zrok/v2/agent/agentUi"
	"github.com/openziti/zrok/v2/agent/proctree"
	"github.com/openziti/zrok/v2/environment/env_core"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/openziti/zrok/v2/util"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Agent struct {
	cfg             *AgentConfig
	httpEndpoint    string
	root            env_core.Root
	agentSocket     string
	shares          map[string]*share
	addShare        chan *share
	rmShare         chan *share
	accesses        map[string]*access
	addAccess       chan *access
	rmAccess        chan *access
	retryManager    *retryManager
	retryCalc       *retryCalculator
	persistRegistry bool
}

func NewAgent(cfg *AgentConfig, root env_core.Root) (*Agent, error) {
	if !root.IsEnabled() {
		return nil, errors.Errorf("unable to load environment; did you 'zrok enable'?")
	}
	a := &Agent{
		cfg:       cfg,
		root:      root,
		shares:    make(map[string]*share),
		addShare:  make(chan *share),
		rmShare:   make(chan *share),
		accesses:  make(map[string]*access),
		addAccess: make(chan *access),
		rmAccess:  make(chan *access),
		retryCalc: newRetryCalculator(cfg),
	}
	a.retryManager = newRetryManager(a)
	return a, nil
}

func (a *Agent) Run() error {
	dl.Infof("started")

	if err := proctree.Init("zrok Agent"); err != nil {
		return err
	}

	agentSocket, err := a.root.AgentSocket()
	if err != nil {
		return err
	}
	l, err := net.Listen("unix", agentSocket)
	if err != nil {
		return err
	}
	a.agentSocket = agentSocket

	go a.retryManager.run()

	go a.manager()
	if a.cfg.ConsoleEnabled {
		go a.gateway()
	}
	failure := make(chan error, 1)
	go a.remoteAgent(failure)
	go func() {
		err := <-failure
		if a.cfg.RequireRemoting {
			panic(errors.Errorf("remote agent requires remoting: %v", err))
		}
	}()

	a.persistRegistry = false
	if err := a.ReloadRegistry(); err != nil {
		dl.Errorf("error reloading registry '%v'", err)
	}
	a.persistRegistry = true

	srv := grpc.NewServer()
	agentGrpc.RegisterAgentServer(srv, &agentGrpcImpl{agent: a})
	if err := srv.Serve(l); err != nil {
		return err
	}

	return nil
}

func (a *Agent) Shutdown() {
	dl.Infof("stopping")

	a.persistRegistry = false

	if err := os.Remove(a.agentSocket); err != nil {
		dl.Warnf("unable to remove agent socket: %v", err)
	}
	for _, shr := range a.shares {
		dl.Debugf("stopping share '%v'", shr.token)
		a.rmShare <- shr
	}
	for _, acc := range a.accesses {
		dl.Debugf("stopping access '%v'", acc.token)
		a.rmAccess <- acc
	}

	a.retryManager.stop()
}

func (a *Agent) Config() *AgentConfig {
	return a.cfg
}

func (a *Agent) ReloadRegistry() error {
	registryPath, err := a.root.AgentRegistry()
	if err != nil {
		return err
	}
	registry, err := LoadRegistry(registryPath)
	if err != nil {
		return err
	}

	dl.Infof("loaded %d private accesses", len(registry.PrivateAccesses))
	registryModified := false
	for _, access := range registry.PrivateAccesses {
		if feToken, err := a.AccessPrivate(access.Request); err == nil {
			dl.Infof("restarted private access '%v' -> '%v'", access.Request.ShareToken, feToken)
			if access.Failure != nil {
				access.Failure = nil
				registryModified = true
			}
		} else {
			dl.Warnf("failed to restart private access '%v': %v (will retry)", access.Request.ShareToken, err)
			if access.Failure != nil {
				access.Failure.Count++
				access.Failure.LastError = err.Error()
			} else {
				access.Failure = &FailureEntry{
					Count:     1,
					LastError: err.Error(),
				}
			}

			// calculate next retry with exponential backoff
			access.Failure.NextRetry = a.retryCalc.nextRetryTime(access.Failure.Count)
			registryModified = true

			a.retryManager.addFailedAccess(access)

			dl.Infof("next retry for private access '%v' scheduled for %v", access.Request.ShareToken, access.Failure.NextRetry.Format(time.RFC3339))
		}
	}

	dl.Infof("loaded %d public shares", len(registry.PublicShares))
	for _, share := range registry.PublicShares {
		if token, frontends, err := a.SharePublic(share.Request); err == nil {
			dl.Infof("restarted public share '%v' -> token='%v', frontends='%v'", share.Request.Target, token, frontends)
			if share.Failure != nil {
				share.Failure = nil
				registryModified = true
			}
		} else {
			dl.Warnf("failed to restart public share '%v': %v (will retry)", share.Request.Target, err)
			if share.Failure != nil {
				share.Failure.Count++
				share.Failure.LastError = err.Error()
			} else {
				share.Failure = &FailureEntry{
					Count:     1,
					LastError: err.Error(),
				}
			}

			// calculate next retry with exponential backoff
			share.Failure.NextRetry = a.retryCalc.nextRetryTime(share.Failure.Count)
			registryModified = true

			a.retryManager.addFailedPublicShare(share)

			dl.Infof("next retry for public share '%v' scheduled for %v", share.Request.Target, share.Failure.NextRetry.Format(time.RFC3339))
		}
	}

	dl.Infof("loaded %d private shares", len(registry.PrivateShares))
	for _, share := range registry.PrivateShares {
		if token, err := a.SharePrivate(share.Request); err == nil {
			dl.Infof("restarted private share '%v' -> token='%v'", share.Request.Target, token)
			if share.Failure != nil {
				share.Failure = nil
				registryModified = true
			}
		} else {
			dl.Warnf("failed to restart private share '%v': %v (will retry)", share.Request.Target, err)
			if share.Failure != nil {
				share.Failure.Count++
				share.Failure.LastError = err.Error()
			} else {
				share.Failure = &FailureEntry{
					Count:     1,
					LastError: err.Error(),
				}
			}

			// calculate next retry with exponential backoff
			share.Failure.NextRetry = a.retryCalc.nextRetryTime(share.Failure.Count)
			registryModified = true

			a.retryManager.addFailedPrivateShare(share)

			dl.Infof("next retry for private share '%v' scheduled for %v", share.Request.Target, share.Failure.NextRetry.Format(time.RFC3339))
		}
	}

	// save updated registry with retry state
	if registryModified {
		if err := registry.Save(registryPath); err != nil {
			dl.Errorf("error saving updated registry: %v", err)
		}
	}

	dl.Infof("reload complete")
	return nil
}

func (a *Agent) SaveRegistry() error {
	r := &Registry{}
	// save private accesses
	for _, acc := range a.accesses {
		if acc.request != nil {
			entry := &AccessRegistryEntry{
				Request: acc.request,
			}
			r.PrivateAccesses = append(r.PrivateAccesses, entry)
		}
	}

	// save named shares
	for _, shr := range a.shares {
		if req, ok := shr.request.(*SharePublicRequest); ok {
			// only save public] shares with at least one registered name (not just namespace)
			if req.hasReservedName() {
				entry := &PublicShareRegistryEntry{
					Request: req,
				}
				r.PublicShares = append(r.PublicShares, entry)
			}
		} else if req, ok := shr.request.(*SharePrivateRequest); ok {
			// only save private shares with a specified share token
			if req.hasReservedToken() {
				entry := &PrivateShareRegistryEntry{
					Request: req,
				}
				r.PrivateShares = append(r.PrivateShares, entry)
			}
		}
	}

	// failures
	failedAccesses, failedPublicShares, failedPrivateShares := a.retryManager.failures()
	for _, failedAccess := range failedAccesses {
		r.PrivateAccesses = append(r.PrivateAccesses, failedAccess)
	}
	for _, failedPublicShare := range failedPublicShares {
		r.PublicShares = append(r.PublicShares, failedPublicShare)
	}
	for _, failedPrivateShare := range failedPrivateShares {
		r.PrivateShares = append(r.PrivateShares, failedPrivateShare)
	}

	registryPath, err := a.root.AgentRegistry()
	if err != nil {
		return err
	}
	if err := r.Save(registryPath); err != nil {
		return err
	}
	return nil
}

func (a *Agent) remoteAgent(failure chan error) {
	enrollmentPath, err := a.root.AgentEnrollment()
	if err != nil {
		dl.Errorf("unable to get agent enrollment path: %v", err)
		if failure != nil {
			failure <- err
		}
		return
	}

	enrollment, err := LoadEnrollment(enrollmentPath)
	if err != nil {
		dl.Warnf("unable to load agent enrollment: %v", err)
		if failure != nil {
			failure <- err
		}
		return
	}

	dl.Infof("listening for remote commands at '%v'", enrollment.Token)

	l, err := sdk.NewListener(enrollment.Token, a.root)
	if err != nil {
		dl.Errorf("error listening for remote agent: %v", err)
		if failure != nil {
			failure <- err
		}
		return
	}

	srv := grpc.NewServer()
	agentGrpc.RegisterAgentServer(srv, &agentGrpcImpl{agent: a})
	if err := srv.Serve(l); err != nil {
		dl.Errorf("error serving: %v", err)
		if failure != nil {
			failure <- err
		}
		return
	}
}

func (a *Agent) gateway() {
	dl.Info("started")
	defer dl.Warn("exited")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	endpoint := "unix:" + a.agentSocket
	dl.Debugf("endpoint: '%v'", endpoint)
	if err := agentGrpc.RegisterAgentHandlerFromEndpoint(ctx, mux, "unix:"+a.agentSocket, opts); err != nil {
		dl.Fatalf("unable to register gateway: %v", err)
	}

	listener, err := util.AutoListener("tcp", a.cfg.ConsoleAddress, a.cfg.ConsoleStartPort, a.cfg.ConsoleEndPort)
	if err != nil {
		dl.Fatalf("unable to create a listener: %v", err)
	}
	a.httpEndpoint = listener.Addr().String()

	if err := http.Serve(listener, agentUi.Middleware(mux)); err != nil {
		dl.Error(err)
	}
}

func (a *Agent) manager() {
	dl.Info("started")
	defer dl.Warn("exited")

	for {
		select {
		case inShare := <-a.addShare:
			dl.Infof("adding new share '%v'", inShare.token)
			a.shares[inShare.token] = inShare

			if a.persistRegistry {
				if err := a.SaveRegistry(); err != nil {
					dl.Errorf("unable to persist registry: %v", err)
				}
			}

		case outShare := <-a.rmShare:
			if shr, found := a.shares[outShare.token]; found {
				dl.Infof("removing share '%v'", shr.token)
				if err := proctree.StopChild(shr.process); err != nil {
					dl.Errorf("error stopping share '%v': %v", shr.token, err)
				}
				if err := proctree.WaitChild(shr.process); err != nil {
					dl.Errorf("error joining share '%v': %v", shr.token, err)
				}
				if err := a.deleteShare(shr.token); err != nil {
					dl.Errorf("error deleting share '%v': %v", shr.token, err)
				}
				delete(a.shares, shr.token)

				// submit the share for retry if it exited abnormally
				if outShare.processExited && !outShare.releaseRequested {
					if reqPub, ok := outShare.request.(*SharePublicRequest); ok {
						if reqPub.hasReservedName() {
							share := &PublicShareRegistryEntry{
								Request: reqPub,
								Failure: &FailureEntry{
									Count: 1,
								},
							}
							if outShare.lastError != nil {
								share.Failure.LastError = outShare.lastError.Error()
							}
							// calculate next retry with exponential backoff
							share.Failure.NextRetry = a.retryCalc.nextRetryTime(share.Failure.Count)
							a.retryManager.addFailedPublicShare(share)
						}
					} else if reqPriv, ok := outShare.request.(*SharePrivateRequest); ok {
						if reqPriv.hasReservedToken() {
							share := &PrivateShareRegistryEntry{
								Request: reqPriv,
								Failure: &FailureEntry{
									Count: 1,
								},
							}
							if outShare.lastError != nil {
								share.Failure.LastError = outShare.lastError.Error()
							}
							// calculate next retry with exponential backoff
							share.Failure.NextRetry = a.retryCalc.nextRetryTime(share.Failure.Count)
							a.retryManager.addFailedPrivateShare(share)
						}
					}
				}

				if a.persistRegistry {
					if err := a.SaveRegistry(); err != nil {
						dl.Errorf("unable to persist registry: %v", err)
					}
				}

			} else {
				dl.Debug("skipping unidentified (orphaned) share removal")
			}

		case inAccess := <-a.addAccess:
			dl.Infof("adding new access '%v'", inAccess.frontendToken)
			a.accesses[inAccess.frontendToken] = inAccess

			if a.persistRegistry {
				if err := a.SaveRegistry(); err != nil {
					dl.Errorf("unable to persist registry: %v", err)
				}
			} else {
				dl.Warn("no persist registry?")
			}

		case outAccess := <-a.rmAccess:
			if acc, found := a.accesses[outAccess.frontendToken]; found {
				dl.Infof("removing access '%v'", acc.frontendToken)
				if err := proctree.StopChild(acc.process); err != nil {
					dl.Errorf("error stopping access '%v': %v", acc.frontendToken, err)
				}
				if err := proctree.WaitChild(acc.process); err != nil {
					dl.Errorf("error joining access '%v': %v", acc.frontendToken, err)
				}
				if err := a.deleteAccess(acc.token, acc.frontendToken); err != nil {
					dl.Errorf("error deleting access '%v': %v", acc.frontendToken, err)
				}
				delete(a.accesses, acc.frontendToken)

				// submit the access for retry if it exited abnormally
				if outAccess.processExited && !outAccess.releaseRequested {
					access := &AccessRegistryEntry{
						Request: outAccess.request,
						Failure: &FailureEntry{
							Count: 1,
						},
					}
					if outAccess.lastError != nil {
						access.Failure.LastError = outAccess.lastError.Error()
					}
					// calculate next retry with exponential backoff
					access.Failure.NextRetry = a.retryCalc.nextRetryTime(access.Failure.Count)
					a.retryManager.addFailedAccess(access)
				}

				if a.persistRegistry {
					if err := a.SaveRegistry(); err != nil {
						dl.Errorf("unable to persist registry: %v", err)
					}
				}

			} else {
				dl.Debug("skipping unidentified (orphaned) access removal")
			}
		}
	}
}

func (a *Agent) deleteShare(token string) error {
	dl.Debugf("deleting share '%v'", token)
	if err := sdk.DeleteShare(a.root, &sdk.Share{Token: token}); err != nil {
		return err
	}
	return nil
}

func (a *Agent) deleteAccess(token, frontendToken string) error {
	dl.Debugf("deleting access '%v'", frontendToken)
	if err := sdk.DeleteAccess(a.root, &sdk.Access{Token: frontendToken, ShareToken: token}); err != nil {
		return err
	}
	return nil
}

type agentGrpcImpl struct {
	agentGrpc.UnimplementedAgentServer
	agent *Agent
}
