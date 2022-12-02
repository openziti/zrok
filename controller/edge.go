package controller

import (
	"context"
	"fmt"
	"github.com/openziti-test-kitchen/zrok/build"
	"github.com/openziti-test-kitchen/zrok/model"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/edge/rest_management_api_client/config"
	"github.com/openziti/edge/rest_management_api_client/edge_router_policy"
	identity_edge "github.com/openziti/edge/rest_management_api_client/identity"
	"github.com/openziti/edge/rest_management_api_client/service"
	edge_service "github.com/openziti/edge/rest_management_api_client/service"
	"github.com/openziti/edge/rest_management_api_client/service_edge_router_policy"
	"github.com/openziti/edge/rest_management_api_client/service_policy"
	"github.com/openziti/edge/rest_model"
	rest_model_edge "github.com/openziti/edge/rest_model"
	sdk_config "github.com/openziti/sdk-golang/ziti/config"
	"github.com/openziti/sdk-golang/ziti/enroll"
	"github.com/sirupsen/logrus"
	"time"
)

func createServiceEdgeRouterPolicy(envZId, svcToken, svcZId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	edgeRouterRoles := []string{"#all"}
	semantic := rest_model.SemanticAllOf
	serviceRoles := []string{fmt.Sprintf("@%v", svcZId)}
	serp := &rest_model.ServiceEdgeRouterPolicyCreate{
		EdgeRouterRoles: edgeRouterRoles,
		Name:            &svcToken,
		Semantic:        &semantic,
		ServiceRoles:    serviceRoles,
		Tags:            zrokServiceTags(svcToken),
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
	logrus.Infof("created service edge router policy '%v' for service '%v' for environment '%v'", resp.Payload.Data.ID, svcZId, envZId)
	return nil
}

func deleteServiceEdgeRouterPolicy(envZId, svcToken string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	filter := fmt.Sprintf("tags.zrokServiceToken=\"%v\"", svcToken)
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

func createServicePolicyBind(envZId, svcToken, svcZId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	semantic := rest_model.SemanticAllOf
	identityRoles := []string{fmt.Sprintf("@%v", envZId)}
	name := fmt.Sprintf("%v-backend", svcToken)
	var postureCheckRoles []string
	serviceRoles := []string{fmt.Sprintf("@%v", svcZId)}
	dialBind := rest_model.DialBindBind
	svcp := &rest_model.ServicePolicyCreate{
		IdentityRoles:     identityRoles,
		Name:              &name,
		PostureCheckRoles: postureCheckRoles,
		Semantic:          &semantic,
		ServiceRoles:      serviceRoles,
		Type:              &dialBind,
		Tags:              zrokServiceTags(svcToken),
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
	logrus.Infof("created bind service policy '%v' for service '%v' for environment '%v'", resp.Payload.Data.ID, svcZId, envZId)
	return nil
}

func deleteServicePolicyBind(envZId, svcToken string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	// type=2 == "Bind"
	return deleteServicePolicy(envZId, fmt.Sprintf("tags.zrokServiceToken=\"%v\" and type=2", svcToken), edge)
}

func createServicePolicyDial(envZId, svcToken, svcZId string, edge *rest_management_api_client.ZitiEdgeManagement, tags ...*rest_model.Tags) error {
	allTags := zrokServiceTags(svcToken)
	for _, t := range tags {
		for k, v := range t.SubTags {
			allTags.SubTags[k] = v
		}
	}

	var identityRoles []string
	for _, proxyIdentity := range cfg.Proxy.Identities {
		identityRoles = append(identityRoles, "@"+proxyIdentity)
		logrus.Infof("added proxy identity role '%v'", proxyIdentity)
	}
	name := fmt.Sprintf("%v-dial", svcToken)
	var postureCheckRoles []string
	semantic := rest_model.SemanticAllOf
	serviceRoles := []string{fmt.Sprintf("@%v", svcZId)}
	dialBind := rest_model.DialBindDial
	svcp := &rest_model.ServicePolicyCreate{
		IdentityRoles:     identityRoles,
		Name:              &name,
		PostureCheckRoles: postureCheckRoles,
		Semantic:          &semantic,
		ServiceRoles:      serviceRoles,
		Type:              &dialBind,
		Tags:              allTags,
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
	logrus.Infof("created dial service policy '%v' for service '%v' for environment '%v'", resp.Payload.Data.ID, svcZId, envZId)
	return nil
}

func deleteServicePolicyDial(envZId, svcToken string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	// type=1 == "Dial"
	return deleteServicePolicy(envZId, fmt.Sprintf("tags.zrokServiceToken=\"%v\" and type=1", svcToken), edge)
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

func createConfig(envZId, svcToken string, authSchemeStr string, authUsers []*model.AuthUser, edge *rest_management_api_client.ZitiEdgeManagement) (cfgID string, err error) {
	authScheme, err := model.ParseAuthScheme(authSchemeStr)
	if err != nil {
		return "", err
	}
	cfg := &model.ProxyConfig{
		AuthScheme: authScheme,
	}
	if cfg.AuthScheme == model.Basic {
		cfg.BasicAuth = &model.BasicAuth{}
		for _, authUser := range authUsers {
			cfg.BasicAuth.Users = append(cfg.BasicAuth.Users, &model.AuthUser{Username: authUser.Username, Password: authUser.Password})
		}
	}
	cfgCrt := &rest_model.ConfigCreate{
		ConfigTypeID: &zrokProxyConfigId,
		Data:         cfg,
		Name:         &svcToken,
		Tags:         zrokServiceTags(svcToken),
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
	logrus.Infof("created config '%v' for environment '%v'", cfgResp.Payload.Data.ID, envZId)
	return cfgResp.Payload.Data.ID, nil
}

func deleteConfig(envZId, svcToken string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	filter := fmt.Sprintf("tags.zrokServiceToken=\"%v\"", svcToken)
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

func createService(envZId, svcToken, cfgId string, edge *rest_management_api_client.ZitiEdgeManagement) (serviceId string, err error) {
	configs := []string{cfgId}
	encryptionRequired := true
	svc := &rest_model.ServiceCreate{
		Configs:            configs,
		EncryptionRequired: &encryptionRequired,
		Name:               &svcToken,
		Tags:               zrokServiceTags(svcToken),
	}
	req := &edge_service.CreateServiceParams{
		Service: svc,
		Context: context.Background(),
	}
	req.SetTimeout(30 * time.Second)
	resp, err := edge.Service.CreateService(req, nil)
	if err != nil {
		return "", err
	}
	logrus.Infof("created zrok service named '%v' (with ziti id '%v') for environment '%v'", svcToken, resp.Payload.Data.ID, envZId)
	return resp.Payload.Data.ID, nil
}

func deleteService(envZId, svcZId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	req := &service.DeleteServiceParams{
		ID:      svcZId,
		Context: context.Background(),
	}
	req.SetTimeout(30 * time.Second)
	_, err := edge.Service.DeleteService(req, nil)
	if err != nil {
		return err
	}
	logrus.Infof("deleted service '%v' for environment '%v'", svcZId, envZId)
	return nil
}

func createEdgeRouterPolicy(zId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	edgeRouterRoles := []string{"#all"}
	identityRoles := []string{fmt.Sprintf("@%v", zId)}
	semantic := rest_model_edge.SemanticAllOf
	erp := &rest_model_edge.EdgeRouterPolicyCreate{
		EdgeRouterRoles: edgeRouterRoles,
		IdentityRoles:   identityRoles,
		Name:            &zId,
		Semantic:        &semantic,
		Tags:            zrokTags(),
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
	logrus.Infof("created edge router policy '%v' for ziti identity '%v'", resp.Payload.Data.ID, zId)
	return nil
}

func deleteEdgeRouterPolicy(envZId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	filter := fmt.Sprintf("name=\"%v\"", envZId)
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

func createIdentity(email string, client *rest_management_api_client.ZitiEdgeManagement) (*identity_edge.CreateIdentityCreated, error) {
	iIsAdmin := false
	name, err := createToken()
	if err != nil {
		return nil, err
	}
	identityType := rest_model_edge.IdentityTypeUser
	tags := zrokTags()
	tags.SubTags["zrokEmail"] = email
	i := &rest_model_edge.IdentityCreate{
		Enrollment:          &rest_model_edge.IdentityCreateEnrollment{Ott: true},
		IsAdmin:             &iIsAdmin,
		Name:                &name,
		RoleAttributes:      nil,
		ServiceHostingCosts: nil,
		Tags:                tags,
		Type:                &identityType,
	}
	req := identity_edge.NewCreateIdentityParams()
	req.Identity = i
	resp, err := client.Identity.CreateIdentity(req, nil)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func enrollIdentity(zId string, client *rest_management_api_client.ZitiEdgeManagement) (*sdk_config.Config, error) {
	p := &identity_edge.DetailIdentityParams{
		Context: context.Background(),
		ID:      zId,
	}
	p.SetTimeout(30 * time.Second)
	resp, err := client.Identity.DetailIdentity(p, nil)
	if err != nil {
		return nil, err
	}
	tkn, _, err := enroll.ParseToken(resp.GetPayload().Data.Enrollment.Ott.JWT)
	if err != nil {
		return nil, err
	}
	flags := enroll.EnrollmentFlags{
		Token:  tkn,
		KeyAlg: "RSA",
	}
	conf, err := enroll.Enroll(flags)
	if err != nil {
		return nil, err
	}
	return conf, nil
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

func zrokTags() *rest_model.Tags {
	return &rest_model.Tags{
		SubTags: map[string]interface{}{
			"zrok": build.String(),
		},
	}
}

func zrokServiceTags(svcToken string) *rest_model.Tags {
	tags := zrokTags()
	tags.SubTags["zrokServiceToken"] = svcToken
	return tags
}
