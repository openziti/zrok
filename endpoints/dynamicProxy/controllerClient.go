package dynamicProxy

import (
	"context"
	"net"
	"time"

	"github.com/michaelquigley/df/da"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/zrok/controller/dynamicProxyController"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

type controllerClientConfig struct {
	IdentityPath string `dd:"+required"`
	ServiceName  string `dd:"+required"`
	Timeout      time.Duration
}

type controllerClient struct {
	cfg    *controllerClientConfig
	zCfg   *ziti.Config
	zCtx   ziti.Context
	ctx    context.Context
	cancel context.CancelFunc
}

func buildControllerClient(app *da.Application[*config]) error {
	client, err := newControllerClient(app.Cfg.Controller)
	if err != nil {
		return err
	}
	da.Set(app.C, client)
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
	dl.Infof("dynamic proxy controller client started for service '%s'", c.cfg.ServiceName)
	return nil
}

func (c *controllerClient) Stop() error {
	c.cancel()
	if c.zCtx != nil {
		c.zCtx.Close()
		c.zCtx = nil
	}
	dl.Info("dynamic proxy controller client stopped")
	return nil
}

// dialAndCall creates a connection, makes an RPC call, and closes the connection
func (c *controllerClient) dialAndCall(call func(client dynamicProxyController.DynamicProxyControllerClient, ctx context.Context) error) error {
	opts := []grpc.DialOption{
		grpc.WithContextDialer(func(_ context.Context, addr string) (net.Conn, error) {
			conn, err := c.zCtx.DialWithOptions(addr, &ziti.DialOptions{ConnectTimeout: c.cfg.Timeout})
			if err != nil {
				return nil, err
			}
			return conn, nil
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	resolver.SetDefaultScheme("passthrough")

	// create grpc connection using ziti service name
	conn, err := grpc.NewClient(c.cfg.ServiceName, opts...)
	if err != nil {
		return errors.Wrapf(err, "failed to connect to service '%s'", c.cfg.ServiceName)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			dl.Warnf("error closing grpc connection: %v", err)
		}
	}()

	client := dynamicProxyController.NewDynamicProxyControllerClient(conn)
	ctx, cancel := context.WithTimeout(c.ctx, c.cfg.Timeout)
	defer cancel()

	return call(client, ctx)
}

// getFrontendMappings retrieves frontend mappings from the controller
func (c *controllerClient) getFrontendMappings(frontendToken, name string, id int64) ([]*dynamicProxyController.FrontendMapping, error) {
	var mappings []*dynamicProxyController.FrontendMapping
	err := c.dialAndCall(func(client dynamicProxyController.DynamicProxyControllerClient, ctx context.Context) error {
		req := &dynamicProxyController.FrontendMappingsRequest{
			Id:            id,
			FrontendToken: frontendToken,
			Name:          name,
		}

		resp, err := client.FrontendMappings(ctx, req)
		if err != nil {
			return errors.Wrap(err, "failed to get frontend mappings")
		}

		mappings = resp.GetFrontendMappings()
		return nil
	})

	if err != nil {
		return nil, err
	}
	return mappings, nil
}

// getAllFrontendMappings is a convenience method to get all mappings for a frontend token
func (c *controllerClient) getAllFrontendMappings(frontendToken string, version int64) ([]*dynamicProxyController.FrontendMapping, error) {
	return c.getFrontendMappings(frontendToken, "", version)
}
