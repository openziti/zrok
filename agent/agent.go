package agent

import (
	"context"
	"net"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/agent/agentUi"
	"github.com/openziti/zrok/agent/proctree"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/openziti/zrok/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	persistRegistry bool
}

func NewAgent(cfg *AgentConfig, root env_core.Root) (*Agent, error) {
	if !root.IsEnabled() {
		return nil, errors.Errorf("unable to load environment; did you 'zrok enable'?")
	}
	return &Agent{
		cfg:       cfg,
		root:      root,
		shares:    make(map[string]*share),
		addShare:  make(chan *share),
		rmShare:   make(chan *share),
		accesses:  make(map[string]*access),
		addAccess: make(chan *access),
		rmAccess:  make(chan *access),
	}, nil
}

func (a *Agent) Run() error {
	logrus.Infof("started")

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

	go a.manager()
	if a.cfg.ConsoleEnabled {
		go a.gateway()
	}
	go a.remoteAgent()

	a.persistRegistry = false
	if err := a.ReloadRegistry(); err != nil {
		logrus.Errorf("error reloading registry '%v'", err)
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
	logrus.Infof("stopping")

	a.persistRegistry = false
	if err := os.Remove(a.agentSocket); err != nil {
		logrus.Warnf("unable to remove agent socket: %v", err)
	}
	for _, shr := range a.shares {
		logrus.Debugf("stopping share '%v'", shr.token)
		a.rmShare <- shr
	}
	for _, acc := range a.accesses {
		logrus.Debugf("stopping access '%v'", acc.token)
		a.rmAccess <- acc
	}
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
	logrus.Infof("loaded %d reserved shares, %d accesses", len(registry.ReservedShares), len(registry.PrivateAccesses))
	for _, req := range registry.ReservedShares {
		if resp, err := a.ShareReserved(req); err == nil {
			logrus.Infof("restarted reserved share '%v' -> '%v'", req, resp)
		} else {
			logrus.Errorf("error restarting reserved share '%v': %v", req, err)
		}
	}
	for _, req := range registry.PrivateAccesses {
		if resp, err := a.AccessPrivate(req); err == nil {
			logrus.Infof("restarted private access '%v' -> '%v'", req, resp)
		} else {
			logrus.Errorf("error restarting private access '%v': %v", req, err)
		}
	}
	logrus.Infof("reload complete")
	return nil
}

func (a *Agent) SaveRegistry() error {
	r := &Registry{}
	for _, shr := range a.shares {
		if shr.request != nil {
			switch shr.request.(type) {
			case *ShareReservedRequest:
				logrus.Infof("persisting reserved share '%v'", shr.token)
				r.ReservedShares = append(r.ReservedShares, shr.request.(*ShareReservedRequest))
			}
		}
	}
	for _, acc := range a.accesses {
		if acc.request != nil {
			r.PrivateAccesses = append(r.PrivateAccesses, acc.request)
		}
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

func (a *Agent) remoteAgent() {
	enrollmentPath, err := a.root.AgentEnrollment()
	if err != nil {
		logrus.Errorf("unable to get agent enrollment path: %v", err)
		return
	}

	enrollment, err := LoadEnrollment(enrollmentPath)
	if err != nil {
		logrus.Warnf("unable to load agent enrollment: %v", err)
		return
	}

	logrus.Infof("listening for remote commands at '%v'", enrollment.Token)

	l, err := sdk.NewListener(enrollment.Token, a.root)
	if err != nil {
		logrus.Errorf("error listening for remote agent: %v", err)
		return
	}

	srv := grpc.NewServer()
	agentGrpc.RegisterAgentServer(srv, &agentGrpcImpl{agent: a})
	if err := srv.Serve(l); err != nil {
		logrus.Errorf("error serving: %v", err)
		return
	}
}

func (a *Agent) gateway() {
	logrus.Info("started")
	defer logrus.Warn("exited")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	endpoint := "unix:" + a.agentSocket
	logrus.Debugf("endpoint: '%v'", endpoint)
	if err := agentGrpc.RegisterAgentHandlerFromEndpoint(ctx, mux, "unix:"+a.agentSocket, opts); err != nil {
		logrus.Fatalf("unable to register gateway: %v", err)
	}

	listener, err := util.AutoListener("tcp", a.cfg.ConsoleAddress, a.cfg.ConsoleStartPort, a.cfg.ConsoleEndPort)
	if err != nil {
		logrus.Fatalf("unable to create a listener: %v", err)
	}
	a.httpEndpoint = listener.Addr().String()

	if err := http.Serve(listener, agentUi.Middleware(mux)); err != nil {
		logrus.Error(err)
	}
}

func (a *Agent) manager() {
	logrus.Info("started")
	defer logrus.Warn("exited")

	for {
		select {
		case inShare := <-a.addShare:
			logrus.Infof("adding new share '%v'", inShare.token)
			a.shares[inShare.token] = inShare

			if a.persistRegistry {
				if err := a.SaveRegistry(); err != nil {
					logrus.Errorf("unable to persist registry: %v", err)
				}
			}

		case outShare := <-a.rmShare:
			if shr, found := a.shares[outShare.token]; found {
				logrus.Infof("removing share '%v'", shr.token)
				if err := proctree.StopChild(shr.process); err != nil {
					logrus.Errorf("error stopping share '%v': %v", shr.token, err)
				}
				if err := proctree.WaitChild(shr.process); err != nil {
					logrus.Errorf("error joining share '%v': %v", shr.token, err)
				}
				if !shr.reserved {
					if err := a.deleteShare(shr.token); err != nil {
						logrus.Errorf("error deleting share '%v': %v", shr.token, err)
					}
				}
				delete(a.shares, shr.token)

				if a.persistRegistry {
					if err := a.SaveRegistry(); err != nil {
						logrus.Errorf("unable to persist registry: %v", err)
					}
				}

			} else {
				logrus.Debug("skipping unidentified (orphaned) share removal")
			}

		case inAccess := <-a.addAccess:
			logrus.Infof("adding new access '%v'", inAccess.frontendToken)
			a.accesses[inAccess.frontendToken] = inAccess

			if a.persistRegistry {
				if err := a.SaveRegistry(); err != nil {
					logrus.Errorf("unable to persist registry: %v", err)
				}
			}

		case outAccess := <-a.rmAccess:
			if acc, found := a.accesses[outAccess.frontendToken]; found {
				logrus.Infof("removing access '%v'", acc.frontendToken)
				if err := proctree.StopChild(acc.process); err != nil {
					logrus.Errorf("error stopping access '%v': %v", acc.frontendToken, err)
				}
				if err := proctree.WaitChild(acc.process); err != nil {
					logrus.Errorf("error joining access '%v': %v", acc.frontendToken, err)
				}
				if err := a.deleteAccess(acc.token, acc.frontendToken); err != nil {
					logrus.Errorf("error deleting access '%v': %v", acc.frontendToken, err)
				}
				delete(a.accesses, acc.frontendToken)

				if a.persistRegistry {
					if err := a.SaveRegistry(); err != nil {
						logrus.Errorf("unable to persist registry: %v", err)
					}
				}

			} else {
				logrus.Debug("skipping unidentified (orphaned) access removal")
			}
		}
	}
}

func (a *Agent) deleteShare(token string) error {
	logrus.Debugf("deleting share '%v'", token)
	if err := sdk.DeleteShare(a.root, &sdk.Share{Token: token}); err != nil {
		return err
	}
	return nil
}

func (a *Agent) deleteAccess(token, frontendToken string) error {
	logrus.Debugf("deleting access '%v'", frontendToken)
	if err := sdk.DeleteAccess(a.root, &sdk.Access{Token: frontendToken, ShareToken: token}); err != nil {
		return err
	}
	return nil
}

type agentGrpcImpl struct {
	agentGrpc.UnimplementedAgentServer
	agent *Agent
}
