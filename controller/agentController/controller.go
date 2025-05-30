package agentController

import (
	"context"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/zrok/agent/agentGrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"net"
	"time"
)

func NewAgentClient(serviceName string, cfg *Config) (client agentGrpc.AgentClient, conn *grpc.ClientConn, err error) {
	opts := []grpc.DialOption{
		grpc.WithContextDialer(func(_ context.Context, addr string) (net.Conn, error) {
			zcfg, err := ziti.NewConfigFromFile(cfg.IdentityPath)
			if err != nil {
				return nil, err
			}
			zctx, err := ziti.NewContext(zcfg)
			if err != nil {
				return nil, err
			}
			conn, err := zctx.DialWithOptions(addr, &ziti.DialOptions{ConnectTimeout: 30 * time.Second})
			if err != nil {
				return nil, err
			}
			return conn, nil
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	resolver.SetDefaultScheme("passthrough")
	conn, err = grpc.NewClient(serviceName, opts...)
	if err != nil {
		return nil, nil, err
	}
	return agentGrpc.NewAgentClient(conn), conn, nil
}
