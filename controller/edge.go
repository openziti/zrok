package controller

import (
	"context"
	"fmt"
	"github.com/openziti-test-kitchen/zrok/controller/zrok_edge_sdk"
	"github.com/openziti/edge/rest_management_api_client"
	identity_edge "github.com/openziti/edge/rest_management_api_client/identity"
	rest_model_edge "github.com/openziti/edge/rest_model"
	sdk_config "github.com/openziti/sdk-golang/ziti/config"
	"github.com/openziti/sdk-golang/ziti/enroll"
	"github.com/sirupsen/logrus"
	"time"
)

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
	tags := zrok_edge_sdk.ZrokTags()
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
