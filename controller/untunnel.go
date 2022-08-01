package controller

import (
	"context"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/controller/store"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/tunnel"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/edge/rest_management_api_client/edge_router_policy"
	"github.com/openziti/edge/rest_management_api_client/service"
	"github.com/openziti/edge/rest_management_api_client/service_edge_router_policy"
	"github.com/openziti/edge/rest_management_api_client/service_policy"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

func untunnelHandler(params tunnel.UntunnelParams, principal *rest_model_zrok.Principal) middleware.Responder {
	logrus.Infof("untunneling for '%v' (%v)", principal.Username, principal.Token)

	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return tunnel.NewUntunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	defer func() { _ = tx.Rollback() }()

	edge, err := edgeClient()
	if err != nil {
		logrus.Error(err)
		return tunnel.NewUntunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	svcName := params.Body.Service
	svcId, err := findServiceId(svcName, edge)
	if err != nil {
		logrus.Error(err)
		return tunnel.NewUntunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	var ssvc *store.Service
	if svcs, err := str.FindServicesForAccount(int(principal.ID), tx); err == nil {
		for _, svc := range svcs {
			if svc.ZitiId == svcId {
				ssvc = svc
				break
			}
		}
		if ssvc == nil {
			err := errors.Errorf("service with id '%v' not found for '%v'", svcId, principal.Username)
			logrus.Error(err)
			return tunnel.NewUntunnelNotFound().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
		}
	} else {
		logrus.Errorf("error finding services for account '%v'", principal.Username)
		return tunnel.NewUntunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}

	if err := deleteEdgeRouterPolicy(svcName, edge); err != nil {
		logrus.Error(err)
		return tunnel.NewUntunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	if err := deleteServiceEdgeRouterPolicy(svcName, edge); err != nil {
		logrus.Error(err)
		return tunnel.NewUntunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	if err := deleteServicePolicyDial(svcName, edge); err != nil {
		logrus.Error(err)
		return tunnel.NewUntunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	if err := deleteServicePolicyBind(svcName, edge); err != nil {
		logrus.Error(err)
		return tunnel.NewUntunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	if err := deleteService(svcId, edge); err != nil {
		logrus.Error(err)
		return tunnel.NewUntunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}

	logrus.Infof("deallocated service '%v'", svcName)

	ssvc.Active = false
	if err := str.UpdateService(ssvc, tx); err != nil {
		logrus.Errorf("error deactivating service '%v': %v", svcId, err)
		return tunnel.NewUntunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing: %v", err)
		return tunnel.NewUntunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}

	return tunnel.NewUntunnelOK()
}

func findServiceId(svcName string, edge *rest_management_api_client.ZitiEdgeManagement) (string, error) {
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
		return "", err
	}
	if len(listResp.Payload.Data) == 1 {
		return *(listResp.Payload.Data[0].ID), nil
	}
	return "", errors.Errorf("service '%v' not found", svcName)
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

func deleteService(svcId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
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
	return nil
}
