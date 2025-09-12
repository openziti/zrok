package dynamicProxy

import (
	"context"
	"net"
	"time"

	"github.com/michaelquigley/df"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/zrok/controller/dynamicProxyController"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/resolver"
)

type controllerClientConfig struct {
	IdentityPath string `df:"+required"`
	ServiceName  string `df:"+required"`
	Timeout      time.Duration
}

type controllerClient struct {
	cfg    *controllerClientConfig
	zCfg   *ziti.Config
	zCtx   ziti.Context
	conn   *grpc.ClientConn
	client dynamicProxyController.DynamicProxyControllerClient
	ctx    context.Context
	cancel context.CancelFunc
}

func buildControllerClient(app *df.Application[*config]) error {
	client, err := newControllerClient(app.Cfg.Controller)
	if err != nil {
		return err
	}
	df.Set(app.C, client)
	return nil
}

func newControllerClient(cfg *controllerClientConfig) (*controllerClient, error) {
	zCfg, err := ziti.NewConfigFromFile(cfg.IdentityPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load ziti config from '%s'", cfg.IdentityPath)
	}

	zCtx, err := ziti.NewContext(zCfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create ziti context")
	}

	ctx, cancel := context.WithCancel(context.Background())

	client := &controllerClient{
		cfg:    cfg,
		zCfg:   zCfg,
		zCtx:   zCtx,
		ctx:    ctx,
		cancel: cancel,
	}

	return client, nil
}

func (c *controllerClient) Start() error {
	opts := []grpc.DialOption{
		grpc.WithContextDialer(func(_ context.Context, addr string) (net.Conn, error) {
			logrus.Infof("dialing '%s'", addr)
			conn, err := c.zCtx.DialWithOptions(addr, &ziti.DialOptions{ConnectTimeout: c.cfg.Timeout})
			if err != nil {
				return nil, err
			}
			return conn, nil
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second, // send keepalive ping every 10 seconds
			Timeout:             5 * time.Second,  // wait 5 seconds for ping ack
			PermitWithoutStream: true,             // send pings even without active streams
		}),
	}
	resolver.SetDefaultScheme("passthrough")

	// create grpc connection using ziti service name
	conn, err := grpc.NewClient(c.cfg.ServiceName, opts...)
	if err != nil {
		return errors.Wrapf(err, "failed to connect to service '%s'", c.cfg.ServiceName)
	}

	c.conn = conn
	c.client = dynamicProxyController.NewDynamicProxyControllerClient(conn)

	logrus.Infof("grpc client connected to dynamic proxy controller service '%s'", c.cfg.ServiceName)
	return nil
}

func (c *controllerClient) Stop() error {
	c.cancel()
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			logrus.Warnf("error closing grpc connection: %v", err)
		}
		c.conn = nil
	}
	if c.zCtx != nil {
		c.zCtx.Close()
		c.zCtx = nil
	}
	logrus.Info("grpc client disconnected from dynamic proxy controller")
	return nil
}

// getFrontendMappings retrieves frontend mappings from the controller
func (c *controllerClient) getFrontendMappings(frontendToken, name string, version int64) ([]*dynamicProxyController.FrontendMapping, error) {
	if c.client == nil {
		return nil, errors.New("grpc client not connected")
	}

	ctx, cancel := context.WithTimeout(c.ctx, c.cfg.Timeout)
	defer cancel()

	req := &dynamicProxyController.FrontendMappingsRequest{
		FrontendToken: frontendToken,
		Name:          name,
		Version:       version,
	}

	resp, err := c.client.FrontendMappings(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get frontend mappings")
	}

	return resp.GetFrontendMappings(), nil
}

// getAllFrontendMappings is a convenience method to get all mappings for a frontend token
func (c *controllerClient) getAllFrontendMappings(frontendToken string, version int64) ([]*dynamicProxyController.FrontendMapping, error) {
	return c.getFrontendMappings(frontendToken, "", version)
}

// isConnected checks if the grpc connection is healthy
func (c *controllerClient) isConnected() bool {
	if c.conn == nil {
		return false
	}
	return c.conn.GetState().String() == "READY"
}
