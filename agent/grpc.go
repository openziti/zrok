package agent

import (
	"context"
	"github.com/openziti/zrok/agent/grpc"
	"github.com/openziti/zrok/build"
	_ "google.golang.org/grpc"
)

type agentGrpcImpl struct {
	grpc.UnimplementedAgentServer
}

func (s *agentGrpcImpl) Version(ctx context.Context, req *grpc.VersionRequest) (*grpc.VersionReply, error) {
	v := build.String()
	return &grpc.VersionReply{V: &v}, nil
}
