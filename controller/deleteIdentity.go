package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/edge-api/rest_management_api_client"
	"github.com/openziti/edge-api/rest_management_api_client/edge_router_policy"
	"github.com/openziti/edge-api/rest_management_api_client/identity"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
	"github.com/sirupsen/logrus"
)

type deleteIdentityHandler struct{}

func newDeleteIdentityHandler() *deleteIdentityHandler {
	return &deleteIdentityHandler{}
}

func (h *deleteIdentityHandler) Handle(params admin.DeleteIdentityParams, principal *rest_model_zrok.Principal) middleware.Responder {
	identityZId := params.Body.ZID

	if !principal.Admin {
		logrus.Errorf("invalid admin principal")
		return admin.NewDeleteIdentityUnauthorized()
	}

	edge, err := zrokEdgeSdk.Client(cfg.Ziti)
	if err != nil {
		logrus.Errorf("error getting edge client: %v", err)
		return admin.NewDeleteIdentityInternalServerError()
	}

	if err := h.deleteEdgeRouterPolicy(identityZId, edge); err != nil {
		logrus.Warnf("unable to delete edge router policy: %v", err)
	}

	if err := h.deleteIdentity(identityZId, edge); err != nil {
		logrus.Errorf("error deleting identity '%v': %v", identityZId, err)
		return admin.NewDeleteIdentityInternalServerError()
	}

	return admin.NewDeleteIdentityOK()
}

func (h *deleteIdentityHandler) getIdentityNameFromZId(identityZId string, edge *rest_management_api_client.ZitiEdgeManagement) (string, error) {
	filter := fmt.Sprintf("id=\"%v\"", identityZId)
	limit := int64(0)
	offset := int64(0)
	listReq := identity.NewListIdentitiesParams()
	listReq.Filter = &filter
	listReq.Limit = &limit
	listReq.Offset = &offset
	listReq.SetTimeout(30 * time.Second)
	listResp, err := edge.Identity.ListIdentities(listReq, nil)
	if err != nil {
		return "", err
	}
	if len(listResp.Payload.Data) != 1 {
		return "", fmt.Errorf("expected 1 identity, found %v", len(listResp.Payload.Data))
	}
	return *(listResp.Payload.Data[0].Name), nil
}

func (h *deleteIdentityHandler) deleteEdgeRouterPolicy(identityZid string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	identityName, err := h.getIdentityNameFromZId(identityZid, edge)
	if err != nil {
		return err
	}
	filter := fmt.Sprintf("name=\"%v\"", identityName)
	limit := int64(0)
	offset := int64(0)
	listReq := edge_router_policy.NewListEdgeRouterPoliciesParams()
	listReq.Filter = &filter
	listReq.Limit = &limit
	listReq.Offset = &offset
	listReq.SetTimeout(30 * time.Second)
	listResp, err := edge.EdgeRouterPolicy.ListEdgeRouterPolicies(listReq, nil)
	if err != nil {
		return err
	}
	if len(listResp.Payload.Data) != 1 {
		return fmt.Errorf("expected 1 edge router policy, found %v", len(listResp.Payload.Data))
	}
	erpZId := *(listResp.Payload.Data[0].ID)
	deleteReq := edge_router_policy.NewDeleteEdgeRouterPolicyParams()
	deleteReq.ID = erpZId
	deleteReq.SetTimeout(30 * time.Second)
	if _, err := edge.EdgeRouterPolicy.DeleteEdgeRouterPolicy(deleteReq, nil); err != nil {
		return err
	}
	logrus.Infof("deleted edge router policy '%v'", erpZId)
	return nil
}

func (h *deleteIdentityHandler) deleteIdentity(identityZId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	req := &identity.DeleteIdentityParams{
		ID:      identityZId,
		Context: context.Background(),
	}
	req.SetTimeout(30 * time.Second)
	if _, err := edge.Identity.DeleteIdentity(req, nil); err != nil {
		return err
	}
	return nil
}
