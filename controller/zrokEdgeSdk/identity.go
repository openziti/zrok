package zrokEdgeSdk

import (
	"context"
	"fmt"
	"time"

	"github.com/openziti/edge-api/rest_management_api_client"
	"github.com/openziti/edge-api/rest_management_api_client/identity"
	rest_model_edge "github.com/openziti/edge-api/rest_model"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/enroll"
	"github.com/sirupsen/logrus"
)

func CreateEnvironmentIdentity(uniqueToken, accountEmail, envDescription string, edge *rest_management_api_client.ZitiEdgeManagement) (*identity.CreateIdentityCreated, error) {
	identityType := rest_model_edge.IdentityTypeUser
	moreTags := map[string]interface{}{"zrokEmail": accountEmail}
	return CreateIdentity(accountEmail+"-"+uniqueToken+"-"+envDescription, identityType, moreTags, edge)
}

func CreateIdentity(name string, identityType rest_model_edge.IdentityType, addlTags map[string]interface{}, edge *rest_management_api_client.ZitiEdgeManagement) (*identity.CreateIdentityCreated, error) {
	isAdmin := false
	tags := ZrokTags()
	for k, v := range addlTags {
		tags.SubTags[k] = v
	}
	req := identity.NewCreateIdentityParams()
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
	resp, err := edge.Identity.CreateIdentity(req, nil)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func GetIdentityByZId(zId string, edge *rest_management_api_client.ZitiEdgeManagement) (*identity.ListIdentitiesOK, error) {
	filter := fmt.Sprintf("id=\"%v\"", zId)
	limit := int64(0)
	offset := int64(0)
	req := &identity.ListIdentitiesParams{
		Filter:  &filter,
		Limit:   &limit,
		Offset:  &offset,
		Context: context.Background(),
	}
	req.SetTimeout(30 * time.Second)
	resp, err := edge.Identity.ListIdentities(req, nil)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func EnrollIdentity(zId string, edge *rest_management_api_client.ZitiEdgeManagement) (*ziti.Config, error) {
	p := &identity.DetailIdentityParams{
		Context: context.Background(),
		ID:      zId,
	}
	p.SetTimeout(30 * time.Second)
	resp, err := edge.Identity.DetailIdentity(p, nil)
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
	logrus.Infof("enrolled ziti identity '%v'", zId)
	return conf, nil
}

func DeleteIdentity(zId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	req := &identity.DeleteIdentityParams{
		ID:      zId,
		Context: context.Background(),
	}
	req.SetTimeout(30 * time.Second)
	_, err := edge.Identity.DeleteIdentity(req, nil)
	if err != nil {
		return err
	}
	logrus.Infof("deleted ziti identity '%v'", zId)
	return nil
}
