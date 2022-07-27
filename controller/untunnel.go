package controller

import (
	"context"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/tunnel"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/edge/rest_management_api_client/edge_router_policy"
	"github.com/openziti/edge/rest_management_api_client/service"
	"github.com/openziti/edge/rest_management_api_client/service_edge_router_policy"
	"github.com/openziti/edge/rest_management_api_client/service_policy"
	"github.com/sirupsen/logrus"
	"time"
)

func untunnelHandler(params tunnel.UntunnelParams) middleware.Responder {
	edge, err := edgeClient()
	if err != nil {
		logrus.Error(err)
		return tunnel.NewUntunnelInternalServerError()
	}
	svcName := params.Body.Service
	if err := deleteEdgeRouterPolicy(svcName, edge); err != nil {
		logrus.Error(err)
		return tunnel.NewUntunnelInternalServerError()
	}
	if err := deleteServiceEdgeRouterPolicy(svcName, edge); err != nil {
		logrus.Error(err)
		return tunnel.NewUntunnelInternalServerError()
	}
	if err := deleteServicePolicyDial(svcName, edge); err != nil {
		logrus.Error(err)
		return tunnel.NewUntunnelInternalServerError()
	}
	if err := deleteServicePolicyBind(svcName, edge); err != nil {
		logrus.Error(err)
		return tunnel.NewUntunnelInternalServerError()
	}
	if err := deleteService(svcName, edge); err != nil {
		logrus.Error(err)
		return tunnel.NewUntunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}

	logrus.Infof("deallocated service '%v'", svcName)

	return tunnel.NewUntunnelOK()
}

func deleteEdgeRouterPolicy(svcName string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	filter := fmt.Sprintf("name=\"%v\"", svcName)
	limit := int64(1)
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
		req.SetTimeout(30 * time.Second)
		_, err := edge.EdgeRouterPolicy.DeleteEdgeRouterPolicy(req, nil)
		if err != nil {
			return err
		}
		logrus.Infof("deleted edge router policy '%v'", erpId)
	} else {
		logrus.Infof("did not find an edge router policy")
	}
	return nil
}

func deleteServiceEdgeRouterPolicy(svcName string, edge *rest_management_api_client.ZitiEdgeManagement) error {
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
		logrus.Infof("deleted service edge router policy '%v'", serpId)
	} else {
		logrus.Infof("did not find a service edge router policy")
	}
	return nil
}

func deleteServicePolicyBind(svcName string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	return deleteServicePolicy(fmt.Sprintf("name=\"%v-bind\"", svcName), edge)
}

func deleteServicePolicyDial(svcName string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	return deleteServicePolicy(fmt.Sprintf("name=\"%v-dial\"", svcName), edge)
}

func deleteServicePolicy(filter string, edge *rest_management_api_client.ZitiEdgeManagement) error {
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
		logrus.Infof("deleted service policy '%v'", spId)
	} else {
		logrus.Infof("did not find a service policy")
	}
	return nil
}

func deleteService(svcName string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	filter := fmt.Sprintf("name=\"%v\"", svcName)
	limit := int64(1)
	offset := int64(0)
	listReq := &service.ListServicesParams{
		Filter:  &filter,
		Limit:   &limit,
		Offset:  &offset,
		Context: context.Background(),
	}
	listReq.SetTimeout(30 * time.Second)
	listResp, err := edge.Service.ListServices(listReq, nil)
	if err != nil {
		return err
	}
	if len(listResp.Payload.Data) == 1 {
		svcId := *(listResp.Payload.Data[0].ID)
		req := &service.DeleteServiceParams{
			ID:      svcId,
			Context: context.Background(),
		}
		req.SetTimeout(30 * time.Second)
		_, err := edge.Service.DeleteService(req, nil)
		if err != nil {
			return err
		}
		logrus.Infof("deleted service '%v'", svcId)
	} else {
		logrus.Infof("did not find a service")
	}
	return nil
}
