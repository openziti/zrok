package controller

import (
	"bytes"
	"encoding/json"
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/admin"
	rest_model_edge "github.com/openziti/edge/rest_model"
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

	edge, err := edgeClient()
	if err != nil {
		logrus.Errorf("error getting edge client: %v", err)
		return admin.NewCreateIdentityInternalServerError()
	}

	idc, err := createIdentity(name, rest_model_edge.IdentityTypeService, nil, edge)
	if err != nil {
		logrus.Errorf("error creating identity: %v", err)
		return admin.NewCreateIdentityInternalServerError()
	}

	zId := idc.Payload.Data.ID
	cfg, err := enrollIdentity(zId, edge)
	if err != nil {
		logrus.Errorf("error enrolling identity: %v", err)
		return admin.NewCreateIdentityInternalServerError()
	}

	if err := createEdgeRouterPolicy(name, zId, edge); err != nil {
		logrus.Errorf("error creating edge router policy for identity: %v", err)
		return admin.NewCreateIdentityInternalServerError()
	}

	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	err = enc.Encode(&cfg)
	if err != nil {
		logrus.Errorf("error encoding identity config: %v", err)
		return admin.NewCreateFrontendInternalServerError()
	}

	return admin.NewCreateIdentityCreated().WithPayload(&admin.CreateIdentityCreatedBody{
		Identity: zId,
		Cfg:      out.String(),
	})
}
