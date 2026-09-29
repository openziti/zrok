package dynamicProxyController

import (
	"context"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/zrok/v2/controller/store"
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
			dl.Errorf("error serving dynamic proxy controller: %v", err)
			return
		}
	}()
	dl.Infof("started dynamic proxy controller server")

	return ctrl, nil
}

func (c *Controller) FrontendMappings(_ context.Context, req *FrontendMappingsRequest) (*FrontendMappingsResponse, error) {
	trx, err := c.str.Begin()
	if err != nil {
		return nil, err
	}
	defer trx.Rollback()

	var mappings []*store.FrontendMapping
	if req.GetName() == "" {
		mappings, err = c.str.FindFrontendMappingsByFrontendTokenWithHigherId(req.GetFrontendToken(), req.GetId(), trx)
	} else {
		mappings, err = c.str.FindFrontendMappingsWithHigherId(req.GetFrontendToken(), req.GetName(), req.GetId(), trx)
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

// Publish sends a mapping update to the frontend's queue. the caller has already committed the row
// the update describes; the store is not touched here.
func (c *Controller) Publish(ctx context.Context, frontendToken string, m Mapping) error {
	if err := c.publisher.Publish(ctx, frontendToken, m); err != nil {
		return err
	}
	dl.Infof("sent mapping update '%+v' -> '%s'", m, frontendToken)
	return nil
}
