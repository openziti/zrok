package controller

import (
	"context"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/controller/store"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/tunnel"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/edge/rest_management_api_client/config"
	"github.com/openziti/edge/rest_management_api_client/service"
	"github.com/openziti/edge/rest_management_api_client/service_edge_router_policy"
	"github.com/openziti/edge/rest_management_api_client/service_policy"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

type untunnelHandler struct {
	cfg *Config
}

func newUntunnelHandler(cfg *Config) *untunnelHandler {
	return &untunnelHandler{cfg: cfg}
}

func (self *untunnelHandler) Handle(params tunnel.UntunnelParams, principal *rest_model_zrok.Principal) middleware.Responder {
	logrus.Infof("untunneling for '%v' (%v)", principal.Email, principal.Token)

	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return tunnel.NewUntunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	defer func() { _ = tx.Rollback() }()

	edge, err := edgeClient(self.cfg.Ziti)
	if err != nil {
		logrus.Error(err)
		return tunnel.NewUntunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	svcName := params.Body.Service
	svcId, err := self.findServiceId(svcName, edge)
	if err != nil {
		logrus.Error(err)
		return tunnel.NewUntunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	var senv *store.Environment
	if envs, err := str.FindEnvironmentsForAccount(int(principal.ID), tx); err == nil {
		for _, env := range envs {
			if env.ZitiIdentityId == params.Body.ZitiIdentityID {
				senv = env
				break
			}
		}
		if senv == nil {
			err := errors.Errorf("environment with id '%v' not found for '%v", params.Body.ZitiIdentityID, principal.Email)
			logrus.Error(err)
			return tunnel.NewUntunnelNotFound().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
		}
	} else {
		logrus.Errorf("error finding environments for account '%v': %v", principal.Email, err)
		return tunnel.NewUntunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}

	var ssvc *store.Service
	if svcs, err := str.FindServicesForEnvironment(senv.Id, tx); err == nil {
		for _, svc := range svcs {
			if svc.ZitiServiceId == svcId {
				ssvc = svc
				break
			}
		}
		if ssvc == nil {
			err := errors.Errorf("service with id '%v' not found for '%v'", svcId, principal.Email)
			logrus.Error(err)
			return tunnel.NewUntunnelNotFound().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
		}
	} else {
		logrus.Errorf("error finding services for account '%v': %v", principal.Email, err)
		return tunnel.NewUntunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}

	if err := self.deleteServiceEdgeRouterPolicy(svcName, edge); err != nil {
		logrus.Error(err)
		return tunnel.NewUntunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	if err := self.deleteServicePolicyDial(svcName, edge); err != nil {
		logrus.Error(err)
		return tunnel.NewUntunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	if err := self.deleteServicePolicyBind(svcName, edge); err != nil {
		logrus.Error(err)
		return tunnel.NewUntunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	if err := self.deleteConfig(svcName, edge); err != nil {
		logrus.Error(err)
		return tunnel.NewTunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	if err := self.deleteService(svcId, edge); err != nil {
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

func (_ *untunnelHandler) findServiceId(svcName string, edge *rest_management_api_client.ZitiEdgeManagement) (string, error) {
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

func (_ *untunnelHandler) deleteServiceEdgeRouterPolicy(svcName string, edge *rest_management_api_client.ZitiEdgeManagement) error {
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

func (self *untunnelHandler) deleteServicePolicyBind(svcName string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	return self.deleteServicePolicy(fmt.Sprintf("name=\"%v-backend\"", svcName), edge)
}

func (self *untunnelHandler) deleteServicePolicyDial(svcName string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	return self.deleteServicePolicy(fmt.Sprintf("name=\"%v-dial\"", svcName), edge)
}

func (_ *untunnelHandler) deleteServicePolicy(filter string, edge *rest_management_api_client.ZitiEdgeManagement) error {
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

func (_ *untunnelHandler) deleteConfig(svcName string, edge *rest_management_api_client.ZitiEdgeManagement) error {
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
		logrus.Infof("deleted config '%v'", *cfg.ID)
	}
	return nil
}

func (_ *untunnelHandler) deleteService(svcId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
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
