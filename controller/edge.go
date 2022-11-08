package controller

import (
	"context"
	"fmt"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/edge/rest_management_api_client/config"
	"github.com/openziti/edge/rest_management_api_client/edge_router_policy"
	identity_edge "github.com/openziti/edge/rest_management_api_client/identity"
	"github.com/openziti/edge/rest_management_api_client/service"
	"github.com/openziti/edge/rest_management_api_client/service_edge_router_policy"
	"github.com/openziti/edge/rest_management_api_client/service_policy"
	"github.com/sirupsen/logrus"
	"time"
)

func deleteServiceEdgeRouterPolicy(envZId, svcName string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	filter := fmt.Sprintf("name=\"%v\"", svcName)
	limit := int64(1)
	offset := int64(0)
	listReq := &service_edge_router_policy.ListServiceEdgeRouterPoliciesParams{
		Filter:  &filter,
		Limit:   &limit,
		Offset:  &offset,
		Context: context.Background(),
	}
	listReq.SetTimeout(30 * time.Second)
	listResp, err := edge.ServiceEdgeRouterPolicy.ListServiceEdgeRouterPolicies(listReq, nil)
	if err != nil {
		return err
	}
	if len(listResp.Payload.Data) == 1 {
		serpId := *(listResp.Payload.Data[0].ID)
		req := &service_edge_router_policy.DeleteServiceEdgeRouterPolicyParams{
			ID:      serpId,
			Context: context.Background(),
		}
		req.SetTimeout(30 * time.Second)
		_, err := edge.ServiceEdgeRouterPolicy.DeleteServiceEdgeRouterPolicy(req, nil)
		if err != nil {
			return err
		}
		logrus.Infof("deleted service edge router policy '%v' for environment '%v'", serpId, envZId)
	} else {
		logrus.Infof("did not find a service edge router policy")
	}
	return nil
}

func deleteServicePolicyBind(envZId, svcName string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	return deleteServicePolicy(envZId, fmt.Sprintf("name=\"%v-backend\"", svcName), edge)
}

func deleteServicePolicyDial(envZId, svcName string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	return deleteServicePolicy(envZId, fmt.Sprintf("name=\"%v-dial\"", svcName), edge)
}

func deleteServicePolicy(envZId, filter string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	limit := int64(1)
	offset := int64(0)
	listReq := &service_policy.ListServicePoliciesParams{
		Filter:  &filter,
		Limit:   &limit,
		Offset:  &offset,
		Context: context.Background(),
	}
	listReq.SetTimeout(30 * time.Second)
	listResp, err := edge.ServicePolicy.ListServicePolicies(listReq, nil)
	if err != nil {
		return err
	}
	if len(listResp.Payload.Data) == 1 {
		spId := *(listResp.Payload.Data[0].ID)
		req := &service_policy.DeleteServicePolicyParams{
			ID:      spId,
			Context: context.Background(),
		}
		req.SetTimeout(30 * time.Second)
		_, err := edge.ServicePolicy.DeleteServicePolicy(req, nil)
		if err != nil {
			return err
		}
		logrus.Infof("deleted service policy '%v' for environment '%v'", spId, envZId)
	} else {
		logrus.Infof("did not find a service policy")
	}
	return nil
}

func deleteConfig(envZId, svcName string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	filter := fmt.Sprintf("name=\"%v\"", svcName)
	limit := int64(0)
	offset := int64(0)
	listReq := &config.ListConfigsParams{
		Filter:  &filter,
		Limit:   &limit,
		Offset:  &offset,
		Context: context.Background(),
	}
	listReq.SetTimeout(30 * time.Second)
	listResp, err := edge.Config.ListConfigs(listReq, nil)
	if err != nil {
		return err
	}
	for _, cfg := range listResp.Payload.Data {
		deleteReq := &config.DeleteConfigParams{
			ID:      *cfg.ID,
			Context: context.Background(),
		}
		deleteReq.SetTimeout(30 * time.Second)
		_, err := edge.Config.DeleteConfig(deleteReq, nil)
		if err != nil {
			return err
		}
		logrus.Infof("deleted config '%v' for '%v'", *cfg.ID, envZId)
	}
	return nil
}

func deleteService(envZId, svcId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	req := &service.DeleteServiceParams{
		ID:      svcId,
		Context: context.Background(),
	}
	req.SetTimeout(30 * time.Second)
	_, err := edge.Service.DeleteService(req, nil)
	if err != nil {
		return err
	}
	logrus.Infof("deleted service '%v' for environment '%v'", svcId, envZId)
	return nil
}

func deleteEdgeRouterPolicy(envZId, id string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	filter := fmt.Sprintf("name=\"%v\"", id)
	limit := int64(0)
	offset := int64(0)
	listReq := &edge_router_policy.ListEdgeRouterPoliciesParams{
		Filter:  &filter,
		Limit:   &limit,
		Offset:  &offset,
		Context: context.Background(),
	}
	listReq.SetTimeout(30 * time.Second)
	listResp, err := edge.EdgeRouterPolicy.ListEdgeRouterPolicies(listReq, nil)
	if err != nil {
		return err
	}
	if len(listResp.Payload.Data) == 1 {
		erpId := *(listResp.Payload.Data[0].ID)
		req := &edge_router_policy.DeleteEdgeRouterPolicyParams{
			ID:      erpId,
			Context: context.Background(),
		}
		_, err := edge.EdgeRouterPolicy.DeleteEdgeRouterPolicy(req, nil)
		if err != nil {
			return err
		}
		logrus.Infof("deleted edge router policy '%v' for environment '%v'", erpId, envZId)
	} else {
		logrus.Infof("found '%d' edge router policies, expected 1", len(listResp.Payload.Data))
	}
	return nil
}

func deleteIdentity(id string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	req := &identity_edge.DeleteIdentityParams{
		ID:      id,
		Context: context.Background(),
	}
	req.SetTimeout(30 * time.Second)
	_, err := edge.Identity.DeleteIdentity(req, nil)
	if err != nil {
		return err
	}
	logrus.Infof("deleted environment identity '%v'", id)
	return nil
}
