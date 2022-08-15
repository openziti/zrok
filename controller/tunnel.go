package controller

import (
	"context"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/controller/store"
	"github.com/openziti-test-kitchen/zrok/model"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/tunnel"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/edge/rest_management_api_client/config"
	"github.com/openziti/edge/rest_management_api_client/edge_router_policy"
	"github.com/openziti/edge/rest_management_api_client/service"
	"github.com/openziti/edge/rest_management_api_client/service_edge_router_policy"
	"github.com/openziti/edge/rest_management_api_client/service_policy"
	"github.com/openziti/edge/rest_model"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

type tunnelHandler struct {
	cfg *Config
}

func newTunnelHandler(cfg *Config) *tunnelHandler {
	return &tunnelHandler{cfg: cfg}
}

func (self *tunnelHandler) Handle(params tunnel.TunnelParams, principal *rest_model_zrok.Principal) middleware.Responder {
	logrus.Infof("tunneling for '%v' (%v)", principal.Username, principal.Token)

	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return tunnel.NewTunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	defer func() { _ = tx.Rollback() }()

	envId := params.Body.ZitiIdentityID
	if envs, err := str.FindEnvironmentsForAccount(int(principal.ID), tx); err == nil {
		found := false
		for _, env := range envs {
			if env.ZitiIdentityId == envId {
				logrus.Infof("found identity '%v' for user '%v'", envId, principal.Username)
				found = true
				break
			}
		}
		if !found {
			logrus.Errorf("environment '%v' not found for user '%v'", envId, principal.Username)
			return tunnel.NewTunnelUnauthorized().WithPayload("bad environment identity")
		}
	} else {
		logrus.Errorf("error finding environments for account '%v'", principal.Username)
		return tunnel.NewTunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}

	edge, err := edgeClient(self.cfg.Ziti)
	if err != nil {
		logrus.Error(err)
		return tunnel.NewTunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	svcName, err := randomId()
	if err != nil {
		logrus.Error(err)
		return tunnel.NewTunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	cfgId, err := self.createConfig(edge)
	if err != nil {
		logrus.Error(err)
		return tunnel.NewTunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	svcId, err := self.createService(svcName, cfgId, edge)
	if err != nil {
		logrus.Error(err)
		return tunnel.NewTunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	if err := self.createServicePolicyBind(svcName, svcId, envId, edge); err != nil {
		logrus.Error(err)
		return tunnel.NewTunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	if err := self.createServicePolicyDial(svcName, svcId, edge); err != nil {
		logrus.Error(err)
		return tunnel.NewTunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	if err := self.createServiceEdgeRouterPolicy(svcName, svcId, edge); err != nil {
		logrus.Error(err)
		return tunnel.NewTunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	if err := self.createEdgeRouterPolicy(svcName, envId, edge); err != nil {
		logrus.Error(err)
		return tunnel.NewTunnelInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}

	logrus.Infof("allocated service '%v'", svcName)

	sid, err := str.CreateService(int(principal.ID), &store.Service{
		ZitiServiceId: svcId,
		Endpoint:      params.Body.Endpoint,
	}, tx)
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
		ProxyEndpoint: self.proxyUrl(svcName),
		Service:       svcName,
	})
}

func (self *tunnelHandler) createConfig(edge *rest_management_api_client.ZitiEdgeManagement) (cfgID string, err error) {
	cfg := &model.ZrokAuth{Hello: "World"}
	name := "zrok.auth.v1"
	cfgCrt := &rest_model.ConfigCreate{
		ConfigTypeID: &zrokAuthV1Id,
		Data:         cfg,
		Name:         &name,
	}
	cfgReq := &config.CreateConfigParams{
		Config:  cfgCrt,
		Context: context.Background(),
	}
	cfgReq.SetTimeout(30 * time.Second)
	cfgResp, err := edge.Config.CreateConfig(cfgReq, nil)
	if err != nil {
		return "", err
	}
	logrus.Infof("created config '%v'", cfgResp.Payload.Data.ID)
	return cfgResp.Payload.Data.ID, nil
}

func (self *tunnelHandler) createService(name, cfgId string, edge *rest_management_api_client.ZitiEdgeManagement) (serviceId string, err error) {
	configs := []string{cfgId}
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

func (self *tunnelHandler) createServicePolicyBind(svcName, svcId, envId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
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

func (self *tunnelHandler) createServicePolicyDial(svcName, svcId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	var identityRoles []string
	for _, proxyIdentity := range self.cfg.Proxy.Identities {
		identityRoles = append(identityRoles, "@"+proxyIdentity)
		logrus.Infof("added proxy identity role '%v'", proxyIdentity)
	}
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

func (self *tunnelHandler) createServiceEdgeRouterPolicy(svcName, svcId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	edgeRouterRoles := []string{"#all"}
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

func (self *tunnelHandler) createEdgeRouterPolicy(svcName, envId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	edgeRouterRoles := []string{"#all"}
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

func (self *tunnelHandler) proxyUrl(svcName string) string {
	return strings.Replace(self.cfg.Proxy.UrlTemplate, "{svcName}", svcName, -1)
}
