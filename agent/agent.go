package agent

import (
	"github.com/openziti/zrok/agent/agentGrpc"
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
	accesses    map[string]*access
}

func NewAgent(root env_core.Root) (*Agent, error) {
	if !root.IsEnabled() {
		return nil, errors.Errorf("unable to load environment; did you 'zrok enable'?")
	}
	return &Agent{
		root:     root,
		shares:   make(map[string]*share),
		accesses: make(map[string]*access),
	}, nil
}

func (a *Agent) Run() error {
	logrus.Infof("started")

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
	if err := os.Remove(a.agentSocket); err != nil {
		logrus.Warnf("unable to remove agent socket: %v", err)
	}
}
