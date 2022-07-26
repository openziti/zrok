package controller

import (
	"context"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/tunnel"
	"github.com/openziti/edge/rest_management_api_client/edge_router_policy"
	"github.com/openziti/edge/rest_management_api_client/service"
	"github.com/openziti/edge/rest_management_api_client/service_edge_router_policy"
	"github.com/openziti/edge/rest_management_api_client/service_policy"
	"github.com/openziti/edge/rest_model"
	"github.com/sirupsen/logrus"
	"time"
)

func tunnelHandler(params tunnel.TunnelParams) middleware.Responder {
	edge, err := edgeClient()
	if err != nil {
		logrus.Error(err)
		return middleware.Error(500, err.Error())
	}

	serviceId, err := randomId()
	if err != nil {
		logrus.Error(err)
		return middleware.Error(500, err.Error())
	}
	logrus.Infof("using service '%v'", serviceId)

	semantic := rest_model.SemanticAllOf

	// Service
	svcConfigs := make([]string, 0)
	svcEnc := true
	svc := &rest_model.ServiceCreate{
		Configs:            svcConfigs,
		EncryptionRequired: &svcEnc,
		Name:               &serviceId,
	}
	svcParams := &service.CreateServiceParams{
		Service: svc,
		Context: context.Background(),
	}
	svcParams.SetTimeout(30 * time.Second)
	svcResp, err := edge.Service.CreateService(svcParams, nil)
	if err != nil {
		logrus.Error(err)
		return middleware.Error(500, err.Error())
	}
	logrus.Infof("created service '%v'", serviceId)

	// Service Policy (Bind)
	svcpIdRoles := []string{fmt.Sprintf("@%v", params.Body.Identity)}
	svcpName := fmt.Sprintf("%v-bind", serviceId)
	svcpPcRoles := []string{}
	svcpSvcRoles := []string{fmt.Sprintf("@%v", svcResp.Payload.Data.ID)}
	svcpDialBind := rest_model.DialBindBind
	svcp := &rest_model.ServicePolicyCreate{
		IdentityRoles:     svcpIdRoles,
		Name:              &svcpName,
		PostureCheckRoles: svcpPcRoles,
		Semantic:          &semantic,
		ServiceRoles:      svcpSvcRoles,
		Type:              &svcpDialBind,
	}
	svcpParams := &service_policy.CreateServicePolicyParams{
		Policy:  svcp,
		Context: context.Background(),
	}
	svcpParams.SetTimeout(30 * time.Second)
	_, err = edge.ServicePolicy.CreateServicePolicy(svcpParams, nil)
	if err != nil {
		logrus.Error(err)
		return middleware.Error(500, err.Error())
	}
	logrus.Infof("created service policy '%v' (bind)", serviceId)

	// Service Policy (Dial)
	svcpIdRoles = []string{"@PyB606.S."} // @proxy
	svcpName = fmt.Sprintf("%v-dial", serviceId)
	svcpDialBind = rest_model.DialBindDial
	svcp = &rest_model.ServicePolicyCreate{
		IdentityRoles:     svcpIdRoles,
		Name:              &svcpName,
		PostureCheckRoles: svcpPcRoles,
		Semantic:          &semantic,
		ServiceRoles:      svcpSvcRoles,
		Type:              &svcpDialBind,
	}
	svcpParams = &service_policy.CreateServicePolicyParams{
		Policy:  svcp,
		Context: context.Background(),
	}
	svcpParams.SetTimeout(30 * time.Second)
	_, err = edge.ServicePolicy.CreateServicePolicy(svcpParams, nil)
	if err != nil {
		logrus.Error(err)
		return middleware.Error(500, err.Error())
	}
	logrus.Infof("created service policy '%v' (dial)", serviceId)

	// Service Edge Router Policy
	serpErRoles := []string{"@tDnhG8jkG9"} // @linux-edge-router
	serpSvcRoles := []string{fmt.Sprintf("@%v", svcResp.Payload.Data.ID)}
	serp := &rest_model.ServiceEdgeRouterPolicyCreate{
		EdgeRouterRoles: serpErRoles,
		Name:            &serviceId,
		Semantic:        &semantic,
		ServiceRoles:    serpSvcRoles,
	}
	serpParams := &service_edge_router_policy.CreateServiceEdgeRouterPolicyParams{
		Policy:  serp,
		Context: context.Background(),
	}
	serpParams.SetTimeout(30 * time.Second)
	_, err = edge.ServiceEdgeRouterPolicy.CreateServiceEdgeRouterPolicy(serpParams, nil)
	if err != nil {
		logrus.Error(err)
		return middleware.Error(500, err.Error())
	}
	logrus.Infof("created service edge router policy '%v'", serviceId)

	// Edge Router Policy
	erpErRoles := []string{"@tDnhG8jkG9"} // @linux-edge-router
	erpIdRoles := []string{fmt.Sprintf("@%v", params.Body.Identity)}
	erp := &rest_model.EdgeRouterPolicyCreate{
		EdgeRouterRoles: erpErRoles,
		IdentityRoles:   erpIdRoles,
		Name:            &serviceId,
		Semantic:        &semantic,
	}
	erpParams := &edge_router_policy.CreateEdgeRouterPolicyParams{
		Policy:  erp,
		Context: context.Background(),
	}
	erpParams.SetTimeout(30 * time.Second)
	_, err = edge.EdgeRouterPolicy.CreateEdgeRouterPolicy(erpParams, nil)
	if err != nil {
		logrus.Error(err)
		return middleware.Error(500, err.Error())
	}
	logrus.Infof("created edge router policy '%v'", serviceId)

	resp := tunnel.NewTunnelCreated().WithPayload(&rest_model_zrok.TunnelResponse{
		Service: serviceId,
	})
	return resp
}
