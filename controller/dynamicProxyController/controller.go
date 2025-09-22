package dynamicProxyController

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/zrok/controller/store"
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

func (c *Controller) BindFrontendMapping(frontendToken, name, shareToken string, trx *sqlx.Tx) error {
	// use current timestamp as version to ensure it's always increasing
	version := time.Now().UnixNano()

	// create new frontend mapping
	fm := &store.FrontendMapping{
		FrontendToken: frontendToken,
		Name:          name,
		ShareToken:    shareToken,
	}

	if err := c.str.CreateFrontendMapping(fm, trx); err != nil {
		return err
	}

	// broadcast the mapping update via AMQP
	mapping := Mapping{
		Operation:  OperationBind,
		Name:       name,
		Version:    version,
		ShareToken: shareToken,
	}
	return c.sendMappingUpdate(frontendToken, mapping)
}

func (c *Controller) UnbindFrontendMapping(frontendToken, name string, trx *sqlx.Tx) error {
	if err := c.str.DeleteFrontendMappingsByFrontendTokenAndName(frontendToken, name, trx); err != nil {
		return err
	}

	// broadcast the mapping update via AMQP
	mapping := Mapping{
		Operation: OperationUnbind,
		Name:      name,
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
		mappings, err = c.str.FindFrontendMappingsByFrontendTokenWithIdOrHigher(req.GetFrontendToken(), req.GetId(), trx)
	} else {
		mappings, err = c.str.FindFrontendMappingsWithIdOrHigher(req.GetFrontendToken(), req.GetName(), req.GetId(), trx)
	}
	if err != nil {
		return nil, err
	}

	out := make([]*FrontendMapping, len(mappings))
	for i, storeMapping := range mappings {
		out[i] = &FrontendMapping{
			Id:         storeMapping.Id,
			Name:       storeMapping.Name,
			ShareToken: storeMapping.ShareToken,
		}
	}

	return &FrontendMappingsResponse{FrontendMappings: out}, nil
}

func (c *Controller) sendMappingUpdate(frontendToken string, m Mapping) error {
	if err := c.publisher.Publish(context.Background(), frontendToken, m); err != nil {
		return err
	}
	logrus.Infof("sent mapping update '%+v' -> '%s'", m, frontendToken)
	return nil
}
