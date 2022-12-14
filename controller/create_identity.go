package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/controller/zrok_edge_sdk"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/admin"
	"github.com/openziti/edge/rest_management_api_client/service"
	rest_model_edge "github.com/openziti/edge/rest_model"
	"github.com/sirupsen/logrus"
	"time"
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
	idCfg, err := enrollIdentity(zId, edge)
	if err != nil {
		logrus.Errorf("error enrolling identity: %v", err)
		return admin.NewCreateIdentityInternalServerError()
	}

	if err := createEdgeRouterPolicy(name, zId, edge); err != nil {
		logrus.Errorf("error creating edge router policy for identity: %v", err)
		return admin.NewCreateIdentityInternalServerError()
	}

	filter := fmt.Sprintf("name=\"%v\" and tags.zrok != null", cfg.Metrics.ServiceName)
	limit := int64(0)
	offset := int64(0)
	listSvcReq := &service.ListServicesParams{
		Filter: &filter,
		Limit:  &limit,
		Offset: &offset,
	}
	listSvcReq.SetTimeout(30 * time.Second)
	listSvcResp, err := edge.Service.ListServices(listSvcReq, nil)
	if err != nil {
		logrus.Errorf("error listing metrics service: %v", err)
		return admin.NewCreateIdentityInternalServerError()
	}
	if len(listSvcResp.Payload.Data) != 1 {
		logrus.Errorf("could not find metrics service")
		return admin.NewCreateIdentityInternalServerError()
	}
	svcZId := *listSvcResp.Payload.Data[0].ID

	spName := fmt.Sprintf("%v-%v-dial", name, cfg.Metrics.ServiceName)
	if err := zrok_edge_sdk.CreateNamedDialServicePolicy(spName, svcZId, zId, edge); err != nil {
		logrus.Errorf("error creating named dial service policy '%v': %v", spName, err)
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
