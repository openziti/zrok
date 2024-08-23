package agent

import (
	agentGrpc "github.com/openziti/zrok/agent/grpc"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
)

type Agent struct {
	root     env_core.Root
	shares   map[string]*share
	accesses map[string]*access
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
	srv := grpc.NewServer()
	agentGrpc.RegisterAgentServer(srv, &agentGrpcImpl{})
	if err := srv.Serve(l); err != nil {
		return err
	}
	return nil
}
