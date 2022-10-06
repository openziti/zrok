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
	tx, err := str.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	dbSvcs, err := str.GetAllServices(tx)
	if err != nil {
		return err
	}
	liveMap := make(map[string]struct{})
	for _, dbSvc := range dbSvcs {
		liveMap[dbSvc.ZrokServiceId] = struct{}{}
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
			if _, found := liveMap[*svc.Name]; !found {
				logrus.Infof("garbage collecting, zitiSvcId='%v', zrokSvcId='%v'", *svc.ID, *svc.Name)
				if err := deleteServiceEdgeRouterPolicy(*svc.Name, edge); err != nil {
					logrus.Errorf("error garbage collecting service edge router policy: %v", err)
				}
				if err := deleteServicePolicyDial(*svc.Name, edge); err != nil {
					logrus.Errorf("error garbage collecting service dial policy: %v", err)
				}
				if err := deleteServicePolicyBind(*svc.Name, edge); err != nil {
					logrus.Errorf("error garbage collecting service bind policy: %v", err)
				}
				if err := deleteConfig(*svc.Name, edge); err != nil {
					logrus.Errorf("error garbage collecting config: %v", err)
				}
				if err := deleteService(*svc.ID, edge); err != nil {
					logrus.Errorf("error garbage collecting service: %v", err)
				}
			} else {
				logrus.Infof("remaining live, zitiSvcId='%v', zrokSvcId='%v'", *svc.ID, *svc.Name)
			}
		}
	} else {
		return errors.Wrap(err, "error listing services")
	}
	return nil
}
