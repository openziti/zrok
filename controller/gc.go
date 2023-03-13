package controller

import (
	"context"
	"fmt"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/edge/rest_management_api_client/config"
	"github.com/openziti/edge/rest_management_api_client/service"
	"github.com/openziti/edge/rest_management_api_client/service_edge_router_policy"
	"github.com/openziti/edge/rest_management_api_client/service_policy"
	zrok_config "github.com/openziti/zrok/controller/config"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

func GC(inCfg *zrok_config.Config) error {
	cfg = inCfg
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
	edge, err := zrokEdgeSdk.Client(cfg.Ziti)
	if err != nil {
		return err
	}
	tx, err := str.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	sshrs, err := str.FindAllShares(tx)
	if err != nil {
		return err
	}
	liveMap := make(map[string]struct{})
	for _, sshr := range sshrs {
		liveMap[sshr.Token] = struct{}{}
	}
	if err := gcServices(edge, liveMap); err != nil {
		return errors.Wrap(err, "error garbage collecting services")
	}
	if err := gcServiceEdgeRouterPolicies(edge, liveMap); err != nil {
		return errors.Wrap(err, "error garbage collecting service edge router policies")
	}
	if err := gcServicePolicies(edge, liveMap); err != nil {
		return errors.Wrap(err, "error garbage collecting service policies")
	}
	if err := gcConfigs(edge, liveMap); err != nil {
		return errors.Wrap(err, "error garbage collecting configs")
	}
	return nil
}

func gcServices(edge *rest_management_api_client.ZitiEdgeManagement, liveMap map[string]struct{}) error {
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
				if err := zrokEdgeSdk.DeleteServiceEdgeRouterPolicy("gc", *svc.Name, edge); err != nil {
					logrus.Errorf("error garbage collecting service edge router policy: %v", err)
				}
				if err := zrokEdgeSdk.DeleteServicePolicyDial("gc", *svc.Name, edge); err != nil {
					logrus.Errorf("error garbage collecting service dial policy: %v", err)
				}
				if err := zrokEdgeSdk.DeleteServicePolicyBind("gc", *svc.Name, edge); err != nil {
					logrus.Errorf("error garbage collecting service bind policy: %v", err)
				}
				if err := zrokEdgeSdk.DeleteConfig("gc", *svc.Name, edge); err != nil {
					logrus.Errorf("error garbage collecting config: %v", err)
				}
				if err := zrokEdgeSdk.DeleteService("gc", *svc.ID, edge); err != nil {
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

func gcServiceEdgeRouterPolicies(edge *rest_management_api_client.ZitiEdgeManagement, liveMap map[string]struct{}) error {
	listReq := &service_edge_router_policy.ListServiceEdgeRouterPoliciesParams{
		Filter:  &filter,
		Limit:   &limit,
		Offset:  &offset,
		Context: context.Background(),
	}
	listReq.SetTimeout(30 * time.Second)
	if listResp, err := edge.ServiceEdgeRouterPolicy.ListServiceEdgeRouterPolicies(listReq, nil); err == nil {
		for _, serp := range listResp.Payload.Data {
			if _, found := liveMap[*serp.Name]; !found {
				logrus.Infof("garbage collecting, svcId='%v'", *serp.Name)
				if err := zrokEdgeSdk.DeleteServiceEdgeRouterPolicy("gc", *serp.Name, edge); err != nil {
					logrus.Errorf("error garbage collecting service edge router policy: %v", err)
				}
			} else {
				logrus.Infof("remaining live, svcId='%v'", *serp.Name)
			}
		}
	} else {
		return errors.Wrap(err, "error listing service edge router policies")
	}
	return nil
}

func gcServicePolicies(edge *rest_management_api_client.ZitiEdgeManagement, liveMap map[string]struct{}) error {
	listReq := &service_policy.ListServicePoliciesParams{
		Filter:  &filter,
		Limit:   &limit,
		Offset:  &offset,
		Context: context.Background(),
	}
	listReq.SetTimeout(30 * time.Second)
	if listResp, err := edge.ServicePolicy.ListServicePolicies(listReq, nil); err == nil {
		for _, sp := range listResp.Payload.Data {
			spName := strings.Split(*sp.Name, "-")[0]
			if _, found := liveMap[spName]; !found {
				logrus.Infof("garbage collecting, svcId='%v'", spName)
				deleteFilter := fmt.Sprintf("id=\"%v\"", *sp.ID)
				if err := zrokEdgeSdk.DeleteServicePolicy("gc", deleteFilter, edge); err != nil {
					logrus.Errorf("error garbage collecting service policy: %v", err)
				}
			} else {
				logrus.Infof("remaining live, svcId='%v'", spName)
			}
		}
	} else {
		return errors.Wrap(err, "error listing service policies")
	}
	return nil
}

func gcConfigs(edge *rest_management_api_client.ZitiEdgeManagement, liveMap map[string]struct{}) error {
	listReq := &config.ListConfigsParams{
		Filter:  &filter,
		Limit:   &limit,
		Offset:  &offset,
		Context: context.Background(),
	}
	listReq.SetTimeout(30 * time.Second)
	if listResp, err := edge.Config.ListConfigs(listReq, nil); err == nil {
		for _, c := range listResp.Payload.Data {
			if _, found := liveMap[*c.Name]; !found {
				if err := zrokEdgeSdk.DeleteConfig("gc", *c.Name, edge); err != nil {
					logrus.Errorf("error garbage collecting config: %v", err)
				}
			} else {
				logrus.Infof("remaining live, svcId='%v'", *c.Name)
			}
		}
	} else {
		return errors.Wrap(err, "error listing configs")
	}
	return nil
}

var filter = "tags.zrok != null"
var limit = int64(0)
var offset = int64(0)
