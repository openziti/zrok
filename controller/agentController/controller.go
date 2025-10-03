package agentController

import (
	"context"
	"net"
	"time"

	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

type Controller struct {
	zCfg *ziti.Config
	zCtx ziti.Context
}

func NewAgentController(cfg *Config) (*Controller, error) {
	zCfg, err := ziti.NewConfigFromFile(cfg.IdentityPath)
	if err != nil {
		return nil, err
	}
	zCtx, err := ziti.NewContext(zCfg)
	if err != nil {
		return nil, err
	}
	return &Controller{zCfg: zCfg, zCtx: zCtx}, nil
}

func (ctrl *Controller) NewClient(serviceName string) (client agentGrpc.AgentClient, conn *grpc.ClientConn, err error) {
	opts := []grpc.DialOption{
		grpc.WithContextDialer(func(_ context.Context, addr string) (net.Conn, error) {
			conn, err := ctrl.zCtx.DialWithOptions(addr, &ziti.DialOptions{ConnectTimeout: 30 * time.Second})
			if err != nil {
				logrus.Warnf("initial dial failed; refreshing service '%v'", addr)
				if _, err := ctrl.zCtx.RefreshService(addr); err != nil {
					return nil, err
				}
				conn, err := ctrl.zCtx.DialWithOptions(addr, &ziti.DialOptions{ConnectTimeout: 30 * time.Second})
				if err != nil {
					return nil, err
				}
				return conn, nil
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
