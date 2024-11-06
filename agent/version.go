package agent

import (
	"context"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/build"
	"github.com/sirupsen/logrus"
)

func (i *agentGrpcImpl) Version(_ context.Context, _ *agentGrpc.VersionRequest) (*agentGrpc.VersionResponse, error) {
	v := build.String()
	logrus.Debugf("responding to version inquiry with '%v'", v)
	return &agentGrpc.VersionResponse{
		V:               v,
		ConsoleEndpoint: i.agent.httpEndpoint,
	}, nil
}
