package agent

import (
	"context"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/build"
)

type agentGrpcImpl struct {
	agentGrpc.UnimplementedAgentServer
}

func (s *agentGrpcImpl) Version(ctx context.Context, req *agentGrpc.VersionRequest) (*agentGrpc.VersionReply, error) {
	v := build.String()
	return &agentGrpc.VersionReply{V: &v}, nil
}
