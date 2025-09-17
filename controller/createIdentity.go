package controller

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/go-openapi/runtime/middleware"
	rest_model_edge "github.com/openziti/edge-api/rest_model"
	"github.com/openziti/zrok/controller/automation"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
	"github.com/sirupsen/logrus"
)

type createIdentityHandler struct{}

func newCreateIdentityHandler() *createIdentityHandler {
	return &createIdentityHandler{}
}

func (h *createIdentityHandler) Handle(params admin.CreateIdentityParams, principal *rest_model_zrok.Principal) middleware.Responder {
	name := params.Body.Name

	if !principal.Admin {
		logrus.Errorf("invalid admin principal")
		return admin.NewCreateIdentityUnauthorized()
	}

	ziti, err := automation.NewZitiAutomation(cfg.Ziti)
	if err != nil {
		logrus.Errorf("error getting automation client: %v", err)
		return admin.NewCreateIdentityInternalServerError()
	}

	// create identity
	identityOpts := &automation.IdentityOptions{
		BaseOptions: automation.BaseOptions{
			Name: name,
			Tags: automation.ZrokTags(),
		},
		Type:    rest_model_edge.IdentityTypeService,
		IsAdmin: false,
	}
	zId, err := ziti.Identities.Create(identityOpts)
	if err != nil {
		logrus.Errorf("error creating identity: %v", err)
		return admin.NewCreateIdentityInternalServerError()
	}

	// enroll identity
	idCfg, err := ziti.Identities.Enroll(zId)
	if err != nil {
		logrus.Errorf("error enrolling identity: %v", err)
		return admin.NewCreateIdentityInternalServerError()
	}

	// create edge router policy for the identity
	erpOpts := &automation.EdgeRouterPolicyOptions{
		BaseOptions: automation.BaseOptions{
			Name: zId,
			Tags: automation.ZrokTags(),
		},
		IdentityRoles:   []string{fmt.Sprintf("@%v", zId)},
		EdgeRouterRoles: []string{"#all"},
		Semantic:        rest_model_edge.SemanticAllOf,
	}
	if _, err := ziti.EdgeRouterPolicies.Create(erpOpts); err != nil {
		logrus.Errorf("error creating edge router policy for identity: %v", err)
		return admin.NewCreateIdentityInternalServerError()
	}

	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	err = enc.Encode(&idCfg)
	if err != nil {
		logrus.Errorf("error encoding identity config: %v", err)
		return admin.NewCreateIdentityInternalServerError()
	}

	return admin.NewCreateIdentityCreated().WithPayload(&admin.CreateIdentityCreatedBody{
		Identity: zId,
		Cfg:      out.String(),
	})
}
