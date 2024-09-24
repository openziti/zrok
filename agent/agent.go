package agent

import (
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/agent/proctree"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"os"
)

type Agent struct {
	root        env_core.Root
	agentSocket string
	shares      map[string]*share
	inShares    chan *share
	outShares   chan *share
	accesses    map[string]*access
	inAccesses  chan *access
	outAccesses chan *access
}

func NewAgent(root env_core.Root) (*Agent, error) {
	if !root.IsEnabled() {
		return nil, errors.Errorf("unable to load environment; did you 'zrok enable'?")
	}
	return &Agent{
		root:        root,
		shares:      make(map[string]*share),
		inShares:    make(chan *share),
		outShares:   make(chan *share),
		accesses:    make(map[string]*access),
		inAccesses:  make(chan *access),
		outAccesses: make(chan *access),
	}, nil
}

func (a *Agent) Run() error {
	logrus.Infof("started")

	if err := proctree.Init("zrok Agent"); err != nil {
		return err
	}
	go a.manager()

	agentSocket, err := a.root.AgentSocket()
	if err != nil {
		return err
	}
	l, err := net.Listen("unix", agentSocket)
	if err != nil {
		return err
	}
	a.agentSocket = agentSocket

	srv := grpc.NewServer()
	agentGrpc.RegisterAgentServer(srv, &agentGrpcImpl{a: a})
	if err := srv.Serve(l); err != nil {
		return err
	}

	return nil
}

func (a *Agent) Shutdown() {
	logrus.Infof("stopping")

	if err := os.Remove(a.agentSocket); err != nil {
		logrus.Warnf("unable to remove agent socket: %v", err)
	}
	for _, shr := range a.shares {
		logrus.Debugf("stopping share '%v'", shr.token)
		a.outShares <- shr
	}
	for _, acc := range a.accesses {
		logrus.Debugf("stopping access '%v'", acc.token)
		a.outAccesses <- acc
	}
}

func (a *Agent) manager() {
	logrus.Info("started")
	defer logrus.Warn("exited")

	for {
		select {
		case inShare := <-a.inShares:
			logrus.Infof("adding new share '%v'", inShare.token)
			a.shares[inShare.token] = inShare

		case outShare := <-a.outShares:
			if outShare.token != "" {
				logrus.Infof("removing share '%v'", outShare.token)
				if err := proctree.StopChild(outShare.process); err != nil {
					logrus.Errorf("error stopping share '%v': %v", outShare.token, err)
				}
				if err := proctree.WaitChild(outShare.process); err != nil {
					logrus.Errorf("error joining share '%v': %v", outShare.token, err)
				}
				if !outShare.reserved {
					if err := a.deleteShare(outShare.token); err != nil {
						logrus.Errorf("error deleting share '%v': %v", outShare.token, err)
					}
				}
				delete(a.shares, outShare.token)
			} else {
				logrus.Debug("skipping unidentified (orphaned) share removal")
			}

		case inAccess := <-a.inAccesses:
			logrus.Infof("adding new access '%v'", inAccess.frontendToken)
			a.accesses[inAccess.frontendToken] = inAccess

		case outAccess := <-a.outAccesses:
			if outAccess.frontendToken != "" {
				logrus.Infof("removing access '%v'", outAccess.frontendToken)
				if err := proctree.StopChild(outAccess.process); err != nil {
					logrus.Errorf("error stopping access '%v': %v", outAccess.frontendToken, err)
				}
				if err := proctree.WaitChild(outAccess.process); err != nil {
					logrus.Errorf("error joining access '%v': %v", outAccess.frontendToken, err)
				}
				if err := a.deleteAccess(outAccess.token, outAccess.frontendToken); err != nil {
					logrus.Errorf("error deleting access '%v': %v", outAccess.frontendToken, err)
				}
				delete(a.accesses, outAccess.frontendToken)
			} else {
				logrus.Debug("skipping unidentified (orphaned) access removal")
			}
		}
	}
}

func (a *Agent) deleteShare(token string) error {
	if err := sdk.DeleteShare(a.root, &sdk.Share{Token: token}); err != nil {
		return err
	}
	return nil
}

func (a *Agent) deleteAccess(token, frontendToken string) error {
	if err := sdk.DeleteAccess(a.root, &sdk.Access{Token: frontendToken, ShareToken: token}); err != nil {
		return err
	}
	return nil
}

type agentGrpcImpl struct {
	agentGrpc.UnimplementedAgentServer
	a *Agent
}
