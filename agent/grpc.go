package agent

import (
	"context"
	"github.com/openziti/zrok/agent/grpc"
	"github.com/prometheus/common/version"
	_ "google.golang.org/grpc"
)

type agentGrpcImpl struct {
	grpc.UnimplementedAgentServer
}

func (s *agentGrpcImpl) Version(ctx context.Context, req *grpc.VersionRequest) (*grpc.VersionReply, error) {
	return &grpc.VersionReply{V: &version.Version}, nil
}
