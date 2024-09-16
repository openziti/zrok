package agent

import (
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/agent/proctree"
	"github.com/openziti/zrok/environment/env_core"
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
		logrus.Infof("stopping share '%v'", shr.token)
		a.outShares <- shr
	}
	for _, acc := range a.accesses {
		logrus.Infof("stopping access '%v'", acc.token)
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
				delete(a.shares, outShare.token)
			} else {
				logrus.Debug("skipping unidentified (orphaned) share removal")
			}
		}
	}
}
