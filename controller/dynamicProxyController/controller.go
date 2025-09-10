package dynamicProxyController

import (
	"context"
	"time"

	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/dynamicProxyModel"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Controller struct {
	UnimplementedDynamicProxyControllerServer
	str       *store.Store
	publisher *AmqpPublisher
	zCfg      *ziti.Config
	zCtx      ziti.Context
}

func NewController(cfg *Config, str *store.Store) (*Controller, error) {
	publisher, err := NewAmqpPublisher(cfg.AmqpPublisher)
	if err != nil {
		return nil, err
	}

	zCfg, err := ziti.NewConfigFromFile(cfg.IdentityPath)
	if err != nil {
		return nil, err
	}
	zCtx, err := ziti.NewContext(zCfg)
	if err != nil {
		return nil, err
	}
	srv := grpc.NewServer()
	ctrl := &Controller{
		str:       str,
		publisher: publisher,
		zCfg:      zCfg,
		zCtx:      zCtx,
	}
	RegisterDynamicProxyControllerServer(srv, ctrl)
	l, err := zCtx.Listen(cfg.ServiceName)
	if err != nil {
		return nil, err
	}
	go func() {
		if err := srv.Serve(l); err != nil {
			logrus.Errorf("error serving dynamic proxy controller: %v", err)
			return
		}
	}()
	logrus.Infof("started dynamic proxy controller server")

	return ctrl, nil
}

func (c *Controller) BindFrontendMapping(frontendToken, name, shareToken string) error {
	trx, err := c.str.Begin()
	if err != nil {
		return err
	}
	defer trx.Rollback()

	// use current timestamp as version to ensure it's always increasing
	version := time.Now().UnixNano()

	// create new frontend mapping
	fm := &store.FrontendMapping{
		FrontendToken: frontendToken,
		Name:          name,
		Version:       version,
		ShareToken:    shareToken,
	}

	if err := c.str.CreateFrontendMapping(fm, trx); err != nil {
		return err
	}

	if err := trx.Commit(); err != nil {
		return err
	}

	// broadcast the mapping update via AMQP
	mapping := dynamicProxyModel.Mapping{
		Operation:  dynamicProxyModel.OperationBind,
		Name:       name,
		Version:    version,
		ShareToken: shareToken,
	}
	return c.sendMappingUpdate(frontendToken, mapping)
}

func (c *Controller) UnbindFrontendMapping(frontendToken, name string) error {
	trx, err := c.str.Begin()
	if err != nil {
		return err
	}
	defer trx.Rollback()

	if err := c.str.DeleteFrontendMappingsByFrontendTokenAndName(frontendToken, name, trx); err != nil {
		return err
	}

	if err := trx.Commit(); err != nil {
		return err
	}

	// broadcast the mapping update via AMQP
	mapping := dynamicProxyModel.Mapping{
		Operation: dynamicProxyModel.OperationUnbind,
		Name:      name,
		Version:   time.Now().UnixNano(),
	}
	return c.sendMappingUpdate(frontendToken, mapping)
}

func (c *Controller) FrontendMappings(_ context.Context, req *FrontendMappingsRequest) (*FrontendMappingsResponse, error) {
	trx, err := c.str.Begin()
	if err != nil {
		return nil, err
	}
	defer trx.Rollback()

	var mappings []*store.FrontendMapping
	if req.GetName() == "" {
		mappings, err = c.str.FindFrontendMappingsByFrontendTokenWithVersionOrHigher(req.GetFrontendToken(), req.GetVersion(), trx)
	} else {
		mappings, err = c.str.FindFrontendMappingsWithVersionOrHigher(req.GetFrontendToken(), req.GetName(), req.GetVersion(), trx)
	}
	if err != nil {
		return nil, err
	}

	out := make([]*FrontendMapping, len(mappings))
	for i, storeMapping := range mappings {
		out[i] = &FrontendMapping{
			Name:       storeMapping.Name,
			Version:    storeMapping.Version,
			ShareToken: storeMapping.ShareToken,
		}
	}

	return &FrontendMappingsResponse{FrontendMappings: out}, nil
}

func (c *Controller) sendMappingUpdate(frontendToken string, m dynamicProxyModel.Mapping) error {
	if err := c.publisher.Publish(context.Background(), frontendToken, m); err != nil {
		return err
	}
	logrus.Infof("sent mapping update '%+v' -> '%s'", m, frontendToken)
	return nil
}
