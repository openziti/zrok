package agent

import (
	"context"
	"github.com/openziti/zrok/agent/grpc"
	"github.com/prometheus/common/version"
)

type grpcServer struct {
}

func (s *grpcServer) Version(ctx context.Context, req *grpc.VersionRequest) (*grpc.VersionReply, error) {
	return &grpc.VersionReply{V: &version.Version}, nil
}
