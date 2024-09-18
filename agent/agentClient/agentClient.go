package agentClient

import (
	"context"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/tui"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"net"
)

func NewClient(root env_core.Root) (client agentGrpc.AgentClient, conn *grpc.ClientConn, err error) {
	agentSocket, err := root.AgentSocket()
	if err != nil {
		tui.Error("error getting agent socket", err)
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
		tui.Error("error connecting to agent socket", err)
	}

	return agentGrpc.NewAgentClient(conn), conn, nil
}
