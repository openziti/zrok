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
	"github.com/openziti/edge/rest_model"
	"github.com/sirupsen/logrus"
	"time"
)

func tunnelHandler(params tunnel.TunnelParams, principal *rest_model_zrok.Principal) middleware.Responder {
	logrus.Infof("tunneling for '%v' (%v)", principal.Username, principal.Token)

	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return tunnel.NewTunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}

	edge, err := edgeClient()
	if err != nil {
		logrus.Error(err)
		return tunnel.NewTunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	svcName, err := randomId()
	if err != nil {
		logrus.Error(err)
		return tunnel.NewTunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	svcId, err := createService(svcName, edge)
	if err != nil {
		logrus.Error(err)
		return tunnel.NewTunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	envId := params.Body.Identity
	if err := createServicePolicyBind(svcName, svcId, envId, edge); err != nil {
		logrus.Error(err)
		return tunnel.NewTunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	if err := createServicePolicyDial(svcName, svcId, edge); err != nil {
		logrus.Error(err)
		return tunnel.NewTunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	if err := createServiceEdgeRouterPolicy(svcName, svcId, edge); err != nil {
		logrus.Error(err)
		return tunnel.NewTunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	if err := createEdgeRouterPolicy(svcName, envId, edge); err != nil {
		logrus.Error(err)
		return tunnel.NewTunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}

	logrus.Infof("allocated service '%v'", svcName)

	sid, err := str.CreateService(int(principal.ID), &store.Service{ZitiId: svcId, Endpoint: params.Body.Endpoint}, tx)
	if err != nil {
		logrus.Errorf("error creating service record: %v", err)
		_ = tx.Rollback()
		return tunnel.NewUntunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing service record: %v", err)
		return tunnel.NewTunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	logrus.Infof("recorded service '%v' with id '%v' for '%v'", svcId, sid, principal.Username)

	return tunnel.NewTunnelCreated().WithPayload(&rest_model_zrok.TunnelResponse{
		Service: svcName,
	})
}

func createService(name string, edge *rest_management_api_client.ZitiEdgeManagement) (serviceId string, err error) {
	configs := make([]string, 0)
	encryptionRequired := true
	svc := &rest_model.ServiceCreate{
		Configs:            configs,
		EncryptionRequired: &encryptionRequired,
		Name:               &name,
	}
	req := &service.CreateServiceParams{
		Service: svc,
		Context: context.Background(),
	}
	req.SetTimeout(30 * time.Second)
	resp, err := edge.Service.CreateService(req, nil)
	if err != nil {
		return "", err
	}
	logrus.Infof("created service '%v'", resp.Payload.Data.ID)
	return resp.Payload.Data.ID, nil
}

func createServicePolicyBind(svcName, svcId, envId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	semantic := rest_model.SemanticAllOf
	identityRoles := []string{fmt.Sprintf("@%v", envId)}
	name := fmt.Sprintf("%v-bind", svcName)
	postureCheckRoles := []string{}
	serviceRoles := []string{fmt.Sprintf("@%v", svcId)}
	dialBind := rest_model.DialBindBind
	svcp := &rest_model.ServicePolicyCreate{
		IdentityRoles:     identityRoles,
		Name:              &name,
		PostureCheckRoles: postureCheckRoles,
		Semantic:          &semantic,
		ServiceRoles:      serviceRoles,
		Type:              &dialBind,
	}
	req := &service_policy.CreateServicePolicyParams{
		Policy:  svcp,
		Context: context.Background(),
	}
	req.SetTimeout(30 * time.Second)
	resp, err := edge.ServicePolicy.CreateServicePolicy(req, nil)
	if err != nil {
		return err
	}
	logrus.Infof("created service policy '%v'", resp.Payload.Data.ID)
	return nil
}

func createServicePolicyDial(svcName, svcId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	identityRoles := []string{"@PyB606.S."} // @proxy
	name := fmt.Sprintf("%v-dial", svcName)
	postureCheckRoles := []string{}
	semantic := rest_model.SemanticAllOf
	serviceRoles := []string{fmt.Sprintf("@%v", svcId)}
	dialBind := rest_model.DialBindDial
	svcp := &rest_model.ServicePolicyCreate{
		IdentityRoles:     identityRoles,
		Name:              &name,
		PostureCheckRoles: postureCheckRoles,
		Semantic:          &semantic,
		ServiceRoles:      serviceRoles,
		Type:              &dialBind,
	}
	req := &service_policy.CreateServicePolicyParams{
		Policy:  svcp,
		Context: context.Background(),
	}
	req.SetTimeout(30 * time.Second)
	resp, err := edge.ServicePolicy.CreateServicePolicy(req, nil)
	if err != nil {
		return err
	}
	logrus.Infof("created service policy '%v'", resp.Payload.Data.ID)
	return nil
}

func createServiceEdgeRouterPolicy(svcName, svcId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	edgeRouterRoles := []string{"@tDnhG8jkG9"} // @linux-edge-router
	semantic := rest_model.SemanticAllOf
	serviceRoles := []string{fmt.Sprintf("@%v", svcId)}
	serp := &rest_model.ServiceEdgeRouterPolicyCreate{
		EdgeRouterRoles: edgeRouterRoles,
		Name:            &svcName,
		Semantic:        &semantic,
		ServiceRoles:    serviceRoles,
	}
	serpParams := &service_edge_router_policy.CreateServiceEdgeRouterPolicyParams{
		Policy:  serp,
		Context: context.Background(),
	}
	serpParams.SetTimeout(30 * time.Second)
	resp, err := edge.ServiceEdgeRouterPolicy.CreateServiceEdgeRouterPolicy(serpParams, nil)
	if err != nil {
		return err
	}
	logrus.Infof("created service edge router policy '%v'", resp.Payload.Data.ID)
	return nil
}

func createEdgeRouterPolicy(svcName, envId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	edgeRouterRoles := []string{"@tDnhG8jkG9"} // @linux-edge-router
	identityRoles := []string{fmt.Sprintf("@%v", envId)}
	semantic := rest_model.SemanticAllOf
	erp := &rest_model.EdgeRouterPolicyCreate{
		EdgeRouterRoles: edgeRouterRoles,
		IdentityRoles:   identityRoles,
		Name:            &svcName,
		Semantic:        &semantic,
	}
	req := &edge_router_policy.CreateEdgeRouterPolicyParams{
		Policy:  erp,
		Context: context.Background(),
	}
	req.SetTimeout(30 * time.Second)
	resp, err := edge.EdgeRouterPolicy.CreateEdgeRouterPolicy(req, nil)
	if err != nil {
		return err
	}
	logrus.Infof("created edge router policy '%v'", resp.Payload.Data.ID)
	return nil
}
