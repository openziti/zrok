package agentClient

import (
	"context"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/build"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"net"
	"strings"
)

func NewClient(root env_core.Root) (client agentGrpc.AgentClient, conn *grpc.ClientConn, err error) {
	agentSocket, err := root.AgentSocket()
	if err != nil {
		return nil, nil, err
	}

	opts := []grpc.DialOption{
		grpc.WithContextDialer(func(_ context.Context, addr string) (net.Conn, error) {
			return net.Dial("unix", addr)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	resolver.SetDefaultScheme("passthrough")
	conn, err = grpc.NewClient(agentSocket, opts...)
	if err != nil {
		return nil, nil, err
	}

	return agentGrpc.NewAgentClient(conn), conn, nil
}

func IsAgentRunning(root env_core.Root) (bool, error) {
	client, conn, err := NewClient(root)
	if err != nil {
		return false, err
	}
	defer func() { _ = conn.Close() }()
	resp, err := client.Version(context.Background(), &agentGrpc.VersionRequest{})
	if err != nil {
		return false, nil
	}
	if !strings.HasPrefix(resp.GetV(), build.Series) {
		return false, errors.Errorf("agent reported version '%v'; we expected version '%v'", resp.GetV(), build.Series)
	}
	return true, nil
}
