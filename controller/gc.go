package controller

import (
	"context"
	"github.com/openziti-test-kitchen/zrok/controller/store"
	"github.com/openziti/edge/rest_management_api_client/service"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

func GC(cfg *Config) error {
	if v, err := store.Open(cfg.Store); err == nil {
		str = v
	} else {
		return errors.Wrap(err, "error opening store")
	}
	defer func() {
		if err := str.Close(); err != nil {
			logrus.Errorf("error closing store: %v", err)
		}
	}()
	if err := gcServices(cfg, str); err != nil {
		return errors.Wrap(err, "error garbage collecting services")
	}
	return nil
}

func gcServices(cfg *Config, str *store.Store) error {
	edge, err := edgeClient(cfg.Ziti)
	if err != nil {
		return err
	}
	filter := "tags.zrok != null"
	limit := int64(0)
	offset := int64(0)
	listReq := &service.ListServicesParams{
		Filter:  &filter,
		Limit:   &limit,
		Offset:  &offset,
		Context: context.Background(),
	}
	listReq.SetTimeout(30 * time.Second)
	if listResp, err := edge.Service.ListServices(listReq, nil); err == nil {
		for _, svc := range listResp.Payload.Data {
			logrus.Infof("found svcId='%v', name='%v'", *svc.ID, *svc.Name)
		}
	}
	return nil
}
