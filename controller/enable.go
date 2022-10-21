package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/controller/store"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/identity"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/edge/rest_management_api_client/edge_router_policy"
	identity_edge "github.com/openziti/edge/rest_management_api_client/identity"
	rest_model_edge "github.com/openziti/edge/rest_model"
	sdk_config "github.com/openziti/sdk-golang/ziti/config"
	"github.com/openziti/sdk-golang/ziti/enroll"
	"github.com/sirupsen/logrus"
	"time"
)

type enableHandler struct {
}

func newEnableHandler() *enableHandler {
	return &enableHandler{}
}

func (self *enableHandler) Handle(params identity.EnableParams, principal *rest_model_zrok.Principal) middleware.Responder {
	// start transaction early; if it fails, don't bother creating ziti resources
	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return identity.NewEnableInternalServerError()
	}

	client, err := edgeClient()
	if err != nil {
		logrus.Errorf("error getting edge client: %v", err)
		return identity.NewEnableInternalServerError()
	}
	ident, err := self.createIdentity(principal.Email, client)
	if err != nil {
		logrus.Error(err)
		return identity.NewEnableInternalServerError()
	}
	cfg, err := self.enrollIdentity(ident.Payload.Data.ID, client)
	if err != nil {
		logrus.Error(err)
		return identity.NewEnableInternalServerError()
	}
	if err := self.createEdgeRouterPolicy(ident.Payload.Data.ID, client); err != nil {
		logrus.Error(err)
		return identity.NewEnableInternalServerError()
	}
	envId, err := str.CreateEnvironment(int(principal.ID), &store.Environment{
		Description: params.Body.Description,
		Host:        params.Body.Host,
		Address:     realRemoteAddress(params.HTTPRequest),
		ZId:         ident.Payload.Data.ID,
	}, tx)
	if err != nil {
		logrus.Errorf("error storing created identity: %v", err)
		_ = tx.Rollback()
		return identity.NewCreateAccountInternalServerError()
	}
	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing: %v", err)
		return identity.NewCreateAccountInternalServerError()
	}
	logrus.Infof("recorded identity '%v' with id '%v' for '%v'", ident.Payload.Data.ID, envId, principal.Email)

	resp := identity.NewEnableCreated().WithPayload(&rest_model_zrok.EnableResponse{
		Identity: ident.Payload.Data.ID,
	})

	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	err = enc.Encode(&cfg)
	if err != nil {
		panic(err)
	}
	resp.Payload.Cfg = out.String()

	return resp
}

func (self *enableHandler) createIdentity(email string, client *rest_management_api_client.ZitiEdgeManagement) (*identity_edge.CreateIdentityCreated, error) {
	iIsAdmin := false
	name, err := createToken()
	if err != nil {
		return nil, err
	}
	identityType := rest_model_edge.IdentityTypeUser
	tags := self.zrokTags()
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

func (_ *enableHandler) enrollIdentity(id string, client *rest_management_api_client.ZitiEdgeManagement) (*sdk_config.Config, error) {
	p := &identity_edge.DetailIdentityParams{
		Context: context.Background(),
		ID:      id,
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

func (self *enableHandler) createEdgeRouterPolicy(id string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	edgeRouterRoles := []string{"#all"}
	identityRoles := []string{fmt.Sprintf("@%v", id)}
	semantic := rest_model_edge.SemanticAllOf
	erp := &rest_model_edge.EdgeRouterPolicyCreate{
		EdgeRouterRoles: edgeRouterRoles,
		IdentityRoles:   identityRoles,
		Name:            &id,
		Semantic:        &semantic,
		Tags:            self.zrokTags(),
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

func (self *enableHandler) zrokTags() *rest_model_edge.Tags {
	return &rest_model_edge.Tags{
		SubTags: map[string]interface{}{
			"zrok": version,
		},
	}
}
