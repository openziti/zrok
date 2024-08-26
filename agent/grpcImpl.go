package agent

import (
	"context"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/build"
	"github.com/sirupsen/logrus"
)

type agentGrpcImpl struct {
	agentGrpc.UnimplementedAgentServer
}

func (s *agentGrpcImpl) Version(_ context.Context, _ *agentGrpc.VersionRequest) (*agentGrpc.VersionReply, error) {
	v := build.String()
	logrus.Infof("responding to version inquiry with '%v'", v)
	return &agentGrpc.VersionReply{V: v}, nil
}
