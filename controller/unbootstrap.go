package controller

import (
	"context"
	"fmt"

	"github.com/openziti/edge-api/rest_management_api_client"
	apiConfig "github.com/openziti/edge-api/rest_management_api_client/config"
	"github.com/openziti/edge-api/rest_management_api_client/edge_router_policy"
	"github.com/openziti/edge-api/rest_management_api_client/identity"
	"github.com/openziti/edge-api/rest_management_api_client/service"
	"github.com/openziti/edge-api/rest_management_api_client/service_edge_router_policy"
	"github.com/openziti/edge-api/rest_management_api_client/service_policy"
	"github.com/openziti/zrok/controller/config"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/sirupsen/logrus"
)

func Unbootstrap(cfg *config.Config) error {
	edge, err := zrokEdgeSdk.Client(cfg.Ziti)
	if err != nil {
		return err
	}
	if err := unbootstrapServiceEdgeRouterPolicies(edge); err != nil {
		logrus.Errorf("error unbootstrapping service edge router policies: %v", err)
	}
	if err := unbootstrapServicePolicies(edge); err != nil {
		logrus.Errorf("error unbootstrapping service policies: %v", err)
	}
	if err := unbootstrapConfigs(edge); err != nil {
		logrus.Errorf("error unbootrapping configs: %v", err)
	}
	if err := unbootstrapServices(edge); err != nil {
		logrus.Errorf("error unbootstrapping services: %v", err)
	}
	if err := unbootstrapEdgeRouterPolicies(edge); err != nil {
		logrus.Errorf("error unbootstrapping edge router policies: %v", err)
	}
	if err := unbootstrapIdentities(edge); err != nil {
		logrus.Errorf("error unbootstrapping identities: %v", err)
	}
	if err := unbootstrapConfigType(edge); err != nil {
		logrus.Errorf("error unbootstrapping config type: %v", err)
	}
	return nil
}

func unbootstrapServiceEdgeRouterPolicies(edge *rest_management_api_client.ZitiEdgeManagement) error {
	for {
		filter := "tags.zrok != null"
		limit := int64(100)
		offset := int64(0)
		listReq := &service_edge_router_policy.ListServiceEdgeRouterPoliciesParams{
			Filter:  &filter,
			Limit:   &limit,
			Offset:  &offset,
			Context: context.Background(),
		}
		listResp, err := edge.ServiceEdgeRouterPolicy.ListServiceEdgeRouterPolicies(listReq, nil)
		if err != nil {
			return err
		}
		if len(listResp.Payload.Data) < 1 {
			break
		}
		for _, serp := range listResp.Payload.Data {
			delReq := &service_edge_router_policy.DeleteServiceEdgeRouterPolicyParams{
				ID:      *serp.ID,
				Context: context.Background(),
			}
			_, err := edge.ServiceEdgeRouterPolicy.DeleteServiceEdgeRouterPolicy(delReq, nil)
			if err == nil {
				logrus.Infof("deleted service edge router policy '%v'", *serp.ID)
			} else {
				return err
			}
		}
	}
	return nil
}

func unbootstrapServicePolicies(edge *rest_management_api_client.ZitiEdgeManagement) error {
	for {
		filter := "tags.zrok != null"
		limit := int64(100)
		offset := int64(0)
		listReq := &service_policy.ListServicePoliciesParams{
			Filter:  &filter,
			Limit:   &limit,
			Offset:  &offset,
			Context: context.Background(),
		}
		listResp, err := edge.ServicePolicy.ListServicePolicies(listReq, nil)
		if err != nil {
			return err
		}
		if len(listResp.Payload.Data) < 1 {
			break
		}
		for _, sp := range listResp.Payload.Data {
			delReq := &service_policy.DeleteServicePolicyParams{
				ID:      *sp.ID,
				Context: context.Background(),
			}
			_, err := edge.ServicePolicy.DeleteServicePolicy(delReq, nil)
			if err == nil {
				logrus.Infof("deleted service policy '%v'", *sp.ID)
			} else {
				return err
			}
		}
	}
	return nil
}

func unbootstrapServices(edge *rest_management_api_client.ZitiEdgeManagement) error {
	for {
		filter := "tags.zrok != null"
		limit := int64(100)
		offset := int64(0)
		listReq := &service.ListServicesParams{
			Filter:  &filter,
			Limit:   &limit,
			Offset:  &offset,
			Context: context.Background(),
		}
		listResp, err := edge.Service.ListServices(listReq, nil)
		if err != nil {
			return err
		}
		if len(listResp.Payload.Data) < 1 {
			break
		}
		for _, svc := range listResp.Payload.Data {
			delReq := &service.DeleteServiceParams{
				ID:      *svc.ID,
				Context: context.Background(),
			}
			_, err := edge.Service.DeleteService(delReq, nil)
			if err == nil {
				logrus.Infof("deleted service '%v' (%v)", *svc.ID, *svc.Name)
			} else {
				return err
			}
		}
	}
	return nil
}

func unbootstrapEdgeRouterPolicies(edge *rest_management_api_client.ZitiEdgeManagement) error {
	for {
		filter := "tags.zrok != null"
		limit := int64(100)
		offset := int64(0)
		listReq := &edge_router_policy.ListEdgeRouterPoliciesParams{
			Filter:  &filter,
			Limit:   &limit,
			Offset:  &offset,
			Context: context.Background(),
		}
		listResp, err := edge.EdgeRouterPolicy.ListEdgeRouterPolicies(listReq, nil)
		if err != nil {
			return err
		}
		if len(listResp.Payload.Data) < 1 {
			break
		}
		for _, erp := range listResp.Payload.Data {
			delReq := &edge_router_policy.DeleteEdgeRouterPolicyParams{
				ID:      *erp.ID,
				Context: context.Background(),
			}
			_, err := edge.EdgeRouterPolicy.DeleteEdgeRouterPolicy(delReq, nil)
			if err == nil {
				logrus.Infof("deleted edge router policy '%v'", *erp.ID)
			} else {
				return err
			}
		}
	}
	return nil
}

func unbootstrapIdentities(edge *rest_management_api_client.ZitiEdgeManagement) error {
	for {
		filter := "tags.zrok != null"
		limit := int64(100)
		offset := int64(0)
		listReq := &identity.ListIdentitiesParams{
			Filter:  &filter,
			Limit:   &limit,
			Offset:  &offset,
			Context: context.Background(),
		}
		listResp, err := edge.Identity.ListIdentities(listReq, nil)
		if err != nil {
			return err
		}
		if len(listResp.Payload.Data) < 1 {
			break
		}
		for _, i := range listResp.Payload.Data {
			delReq := &identity.DeleteIdentityParams{
				ID:      *i.ID,
				Context: context.Background(),
			}
			_, err := edge.Identity.DeleteIdentity(delReq, nil)
			if err == nil {
				logrus.Infof("deleted identity '%v' (%v)", *i.ID, *i.Name)
			} else {
				return err
			}
		}
	}
	return nil
}

func unbootstrapConfigs(edge *rest_management_api_client.ZitiEdgeManagement) error {
	for {
		filter := "tags.zrok != null"
		limit := int64(100)
		offset := int64(0)
		listReq := &apiConfig.ListConfigsParams{
			Filter:  &filter,
			Limit:   &limit,
			Offset:  &offset,
			Context: context.Background(),
		}
		listResp, err := edge.Config.ListConfigs(listReq, nil)
		if err != nil {
			return err
		}
		if len(listResp.Payload.Data) < 1 {
			break
		}
		for _, listCfg := range listResp.Payload.Data {
			delReq := &apiConfig.DeleteConfigParams{
				ID:      *listCfg.ID,
				Context: context.Background(),
			}
			_, err := edge.Config.DeleteConfig(delReq, nil)
			if err == nil {
				logrus.Infof("deleted config '%v'", *listCfg.ID)
			} else {
				return nil
			}
		}
	}
	return nil
}

func unbootstrapConfigType(edge *rest_management_api_client.ZitiEdgeManagement) error {
	for {
		filter := fmt.Sprintf("name = \"%v\"", sdk.ZrokProxyConfig)
		limit := int64(100)
		offset := int64(0)
		listReq := &apiConfig.ListConfigTypesParams{
			Filter:  &filter,
			Limit:   &limit,
			Offset:  &offset,
			Context: context.Background(),
		}
		listResp, err := edge.Config.ListConfigTypes(listReq, nil)
		if err != nil {
			return err
		}
		if len(listResp.Payload.Data) < 1 {
			break
		}
		for _, listCfgType := range listResp.Payload.Data {
			delReq := &apiConfig.DeleteConfigTypeParams{
				ID:      *listCfgType.ID,
				Context: context.Background(),
			}
			_, err := edge.Config.DeleteConfigType(delReq, nil)
			if err == nil {
				logrus.Infof("deleted config type '%v' (%v)", *listCfgType.ID, *listCfgType.Name)
			} else {
				return err
			}
		}
	}
	return nil
}
