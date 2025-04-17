package controller

import (
	"context"
	"fmt"
	"github.com/openziti/edge-api/rest_management_api_client"
	apiConfig "github.com/openziti/edge-api/rest_management_api_client/config"
	"github.com/openziti/edge-api/rest_management_api_client/edge_router_policy"
	"github.com/openziti/edge-api/rest_management_api_client/identity"
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
	if err := unbootstrapEdgeRouterPolicies(edge); err != nil {
		logrus.Errorf("error unbootstrapping edge router policies: %v", err)
	}
	if err := unbootstrapIdentities(edge); err != nil {
		logrus.Errorf("error unbootstrapping identities: %v", err)
	}
	if err := unbootstrapConfigs(edge); err != nil {
		logrus.Errorf("error unbootrapping configs: %v", err)
	}
	if err := unbootstrapConfigType(edge); err != nil {
		logrus.Errorf("error unbootstrapping config type: %v", err)
	}
	return nil
}

func unbootstrapServiceEdgeRouterPolicies(edge *rest_management_api_client.ZitiEdgeManagement) error {
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
	for _, serp := range listResp.Payload.Data {
		logrus.Infof("found service edge router policy: %v", *serp.ID)
	}
	return nil
}

func unbootstrapServicePolicies(edge *rest_management_api_client.ZitiEdgeManagement) error {
	filter := "tags.zrok != null"
	limit := int64(100)
	offset := int64(0)
	req := &service_policy.ListServicePoliciesParams{
		Filter:  &filter,
		Limit:   &limit,
		Offset:  &offset,
		Context: context.Background(),
	}
	resp, err := edge.ServicePolicy.ListServicePolicies(req, nil)
	if err != nil {
		return err
	}
	for _, sp := range resp.Payload.Data {
		logrus.Infof("found service policy: %v", *sp.ID)
	}
	return nil
}

func unbootstrapEdgeRouterPolicies(edge *rest_management_api_client.ZitiEdgeManagement) error {
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
	for _, erp := range listResp.Payload.Data {
		logrus.Infof("found edge router policy: %v", *erp.ID)
	}
	return nil
}

func unbootstrapIdentities(edge *rest_management_api_client.ZitiEdgeManagement) error {
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
	for _, identity := range listResp.Payload.Data {
		logrus.Infof("found identity: %v", *identity.ID)
	}
	return nil
}

func unbootstrapConfigs(edge *rest_management_api_client.ZitiEdgeManagement) error {
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
	for _, listCfg := range listResp.Payload.Data {
		logrus.Infof("found config: %v", *listCfg.ID)
	}
	return nil
}

func unbootstrapConfigType(edge *rest_management_api_client.ZitiEdgeManagement) error {
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
	for _, listCfgType := range listResp.Payload.Data {
		logrus.Infof("found config type: %v (%v)", *listCfgType.ID, *listCfgType.Name)
	}
	return nil
}
