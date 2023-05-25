package controller

import (
	"bytes"
	"encoding/json"

	"github.com/go-openapi/runtime/middleware"
	rest_model_edge "github.com/openziti/edge-api/rest_model"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
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

	edge, err := zrokEdgeSdk.Client(cfg.Ziti)
	if err != nil {
		logrus.Errorf("error getting edge client: %v", err)
		return admin.NewCreateIdentityInternalServerError()
	}

	idc, err := zrokEdgeSdk.CreateIdentity(name, rest_model_edge.IdentityTypeService, nil, edge)
	if err != nil {
		logrus.Errorf("error creating identity: %v", err)
		return admin.NewCreateIdentityInternalServerError()
	}

	zId := idc.Payload.Data.ID
	idCfg, err := zrokEdgeSdk.EnrollIdentity(zId, edge)
	if err != nil {
		logrus.Errorf("error enrolling identity: %v", err)
		return admin.NewCreateIdentityInternalServerError()
	}

	if err := zrokEdgeSdk.CreateEdgeRouterPolicy(name, zId, edge); err != nil {
		logrus.Errorf("error creating edge router policy for identity: %v", err)
		return admin.NewCreateIdentityInternalServerError()
	}

	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	err = enc.Encode(&idCfg)
	if err != nil {
		logrus.Errorf("error encoding identity config: %v", err)
		return admin.NewCreateFrontendInternalServerError()
	}

	return admin.NewCreateIdentityCreated().WithPayload(&admin.CreateIdentityCreatedBody{
		Identity: zId,
		Cfg:      out.String(),
	})
}
