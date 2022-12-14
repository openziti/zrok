package controller

import (
	"context"
	"fmt"
	"github.com/openziti-test-kitchen/zrok/controller/edge_ctrl"
	"github.com/openziti-test-kitchen/zrok/model"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/edge/rest_management_api_client/config"
	"github.com/openziti/edge/rest_management_api_client/edge_router_policy"
	identity_edge "github.com/openziti/edge/rest_management_api_client/identity"
	"github.com/openziti/edge/rest_management_api_client/service_edge_router_policy"
	"github.com/openziti/edge/rest_management_api_client/service_policy"
	"github.com/openziti/edge/rest_model"
	rest_model_edge "github.com/openziti/edge/rest_model"
	sdk_config "github.com/openziti/sdk-golang/ziti/config"
	"github.com/openziti/sdk-golang/ziti/enroll"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

func createShareServiceEdgeRouterPolicy(envZId, svcToken, svcZId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	serpZId, err := createServiceEdgeRouterPolicy(svcToken, svcZId, edge_ctrl.ZrokServiceTags(svcToken).SubTags, edge)
	if err != nil {
		return err
	}
	logrus.Infof("created service edge router policy '%v' for service '%v' for environment '%v'", serpZId, svcZId, envZId)
	return nil
}

func createServiceEdgeRouterPolicy(name, svcZId string, moreTags map[string]interface{}, edge *rest_management_api_client.ZitiEdgeManagement) (string, error) {
	edgeRouterRoles := []string{"#all"}
	semantic := rest_model.SemanticAllOf
	serviceRoles := []string{fmt.Sprintf("@%v", svcZId)}
	tags := edge_ctrl.ZrokTags()
	for k, v := range moreTags {
		tags.SubTags[k] = v
	}
	serp := &rest_model.ServiceEdgeRouterPolicyCreate{
		EdgeRouterRoles: edgeRouterRoles,
		Name:            &name,
		Semantic:        &semantic,
		ServiceRoles:    serviceRoles,
		Tags:            tags,
	}
	serpParams := &service_edge_router_policy.CreateServiceEdgeRouterPolicyParams{
		Policy:  serp,
		Context: context.Background(),
	}
	serpParams.SetTimeout(30 * time.Second)
	resp, err := edge.ServiceEdgeRouterPolicy.CreateServiceEdgeRouterPolicy(serpParams, nil)
	if err != nil {
		return "", errors.Wrapf(err, "error creating serp '%v' for service '%v'", name, svcZId)
	}
	return resp.Payload.Data.ID, nil
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
		Tags:              edge_ctrl.ZrokServiceTags(svcToken),
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

func createNamedBindServicePolicy(name, svcZId, idZId string, edge *rest_management_api_client.ZitiEdgeManagement, tags ...*rest_model.Tags) error {
	allTags := &rest_model_edge.Tags{SubTags: make(rest_model_edge.SubTags)}
	for _, t := range tags {
		for k, v := range t.SubTags {
			allTags.SubTags[k] = v
		}
	}
	identityRoles := []string{"@" + idZId}
	var postureCheckRoles []string
	semantic := rest_model.SemanticAllOf
	serviceRoles := []string{"@" + svcZId}
	dialBind := rest_model.DialBindBind
	sp := &rest_model.ServicePolicyCreate{
		IdentityRoles:     identityRoles,
		Name:              &name,
		PostureCheckRoles: postureCheckRoles,
		Semantic:          &semantic,
		ServiceRoles:      serviceRoles,
		Type:              &dialBind,
		Tags:              allTags,
	}
	req := &service_policy.CreateServicePolicyParams{
		Policy:  sp,
		Context: context.Background(),
	}
	req.SetTimeout(30 * time.Second)
	_, err := edge.ServicePolicy.CreateServicePolicy(req, nil)
	if err != nil {
		return err
	}
	return nil
}

func deleteServicePolicyBind(envZId, svcToken string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	// type=2 == "Bind"
	return deleteServicePolicy(envZId, fmt.Sprintf("tags.zrokServiceToken=\"%v\" and type=2", svcToken), edge)
}

func createServicePolicyDial(envZId, svcToken, svcZId string, dialZIds []string, edge *rest_management_api_client.ZitiEdgeManagement, tags ...*rest_model.Tags) error {
	allTags := edge_ctrl.ZrokServiceTags(svcToken)
	for _, t := range tags {
		for k, v := range t.SubTags {
			allTags.SubTags[k] = v
		}
	}

	var identityRoles []string
	for _, proxyIdentity := range dialZIds {
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

func createNamedDialServicePolicy(name, svcZId, idZId string, edge *rest_management_api_client.ZitiEdgeManagement, tags ...*rest_model.Tags) error {
	allTags := &rest_model_edge.Tags{SubTags: make(rest_model_edge.SubTags)}
	for _, t := range tags {
		for k, v := range t.SubTags {
			allTags.SubTags[k] = v
		}
	}
	identityRoles := []string{"@" + idZId}
	var postureCheckRoles []string
	semantic := rest_model.SemanticAllOf
	serviceRoles := []string{"@" + svcZId}
	dialBind := rest_model.DialBindDial
	sp := &rest_model.ServicePolicyCreate{
		IdentityRoles:     identityRoles,
		Name:              &name,
		PostureCheckRoles: postureCheckRoles,
		Semantic:          &semantic,
		ServiceRoles:      serviceRoles,
		Type:              &dialBind,
		Tags:              allTags,
	}
	req := &service_policy.CreateServicePolicyParams{
		Policy:  sp,
		Context: context.Background(),
	}
	req.SetTimeout(30 * time.Second)
	_, err := edge.ServicePolicy.CreateServicePolicy(req, nil)
	if err != nil {
		return err
	}
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
		Tags:         edge_ctrl.ZrokServiceTags(svcToken),
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

func createShareService(envZId, svcToken, cfgId string, edge *rest_management_api_client.ZitiEdgeManagement) (svcZId string, err error) {
	configs := []string{cfgId}
	tags := edge_ctrl.ZrokServiceTags(svcToken)
	svcZId, err = edge_ctrl.CreateService(svcToken, configs, tags.SubTags, edge)
	if err != nil {
		return "", errors.Wrapf(err, "error creating service '%v'", svcToken)
	}
	logrus.Infof("created zrok service named '%v' (with ziti id '%v') for environment '%v'", svcToken, svcZId, envZId)
	return svcZId, nil
}

func createEdgeRouterPolicy(name, zId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	edgeRouterRoles := []string{"#all"}
	identityRoles := []string{fmt.Sprintf("@%v", zId)}
	semantic := rest_model_edge.SemanticAllOf
	erp := &rest_model_edge.EdgeRouterPolicyCreate{
		EdgeRouterRoles: edgeRouterRoles,
		IdentityRoles:   identityRoles,
		Name:            &name,
		Semantic:        &semantic,
		Tags:            edge_ctrl.ZrokTags(),
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

func createEnvironmentIdentity(accountEmail string, client *rest_management_api_client.ZitiEdgeManagement) (*identity_edge.CreateIdentityCreated, error) {
	name, err := createToken()
	if err != nil {
		return nil, err
	}
	identityType := rest_model_edge.IdentityTypeUser
	moreTags := map[string]interface{}{"zrokEmail": accountEmail}
	return createIdentity(name, identityType, moreTags, client)
}

func createIdentity(name string, identityType rest_model_edge.IdentityType, moreTags map[string]interface{}, client *rest_management_api_client.ZitiEdgeManagement) (*identity_edge.CreateIdentityCreated, error) {
	isAdmin := false
	tags := edge_ctrl.ZrokTags()
	for k, v := range moreTags {
		tags.SubTags[k] = v
	}
	req := identity_edge.NewCreateIdentityParams()
	req.Identity = &rest_model_edge.IdentityCreate{
		Enrollment:          &rest_model_edge.IdentityCreateEnrollment{Ott: true},
		IsAdmin:             &isAdmin,
		Name:                &name,
		RoleAttributes:      nil,
		ServiceHostingCosts: nil,
		Tags:                tags,
		Type:                &identityType,
	}
	req.SetTimeout(30 * time.Second)
	resp, err := client.Identity.CreateIdentity(req, nil)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func getIdentity(zId string, client *rest_management_api_client.ZitiEdgeManagement) (*identity_edge.ListIdentitiesOK, error) {
	filter := fmt.Sprintf("id=\"%v\"", zId)
	limit := int64(0)
	offset := int64(0)
	req := &identity_edge.ListIdentitiesParams{
		Filter:  &filter,
		Limit:   &limit,
		Offset:  &offset,
		Context: context.Background(),
	}
	req.SetTimeout(30 * time.Second)
	resp, err := client.Identity.ListIdentities(req, nil)
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
