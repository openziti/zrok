package controller

import (
	"context"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/identity"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/edge/rest_management_api_client/edge_router_policy"
	identity_edge "github.com/openziti/edge/rest_management_api_client/identity"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

type disableHandler struct {
	cfg *Config
}

func newDisableHandler(cfg *Config) *disableHandler {
	return &disableHandler{cfg: cfg}
}

func (self *disableHandler) Handle(params identity.DisableParams, principal *rest_model_zrok.Principal) middleware.Responder {
	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return identity.NewDisableInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	defer func() { _ = tx.Rollback() }()
	envId, err := self.checkZitiIdentity(params.Body.Identity, principal, tx)
	if err != nil {
		logrus.Errorf("identity check failed: %v", err)
		return identity.NewDisableUnauthorized()
	}
	if err := self.removeEnvironment(envId, tx); err == nil {
		logrus.Errorf("error removing environment: %v", err)
		return identity.NewDisableInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	edge, err := edgeClient(self.cfg.Ziti)
	if err != nil {
		logrus.Errorf("error getting edge client: %v", err)
		return identity.NewDisableInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	if err := self.deleteEdgeRouterPolicy(params.Body.Identity, edge); err != nil {
		logrus.Errorf("error deleting edge router policy: %v", err)
		return identity.NewDisableInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	if err := self.deleteIdentity(params.Body.Identity, edge); err != nil {
		logrus.Errorf("error deleting identity: %v", err)
		return identity.NewDisableInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing: %v", err)
	}
	return identity.NewDisableOK()
}

func (self *disableHandler) checkZitiIdentity(id string, principal *rest_model_zrok.Principal, tx *sqlx.Tx) (int, error) {
	envs, err := str.FindEnvironmentsForAccount(int(principal.ID), tx)
	if err != nil {
		return -1, err
	}
	for _, env := range envs {
		if env.ZitiIdentityId == id {
			return env.Id, nil
		}
	}
	return -1, errors.Errorf("no such environment '%v'", id)
}

func (self *disableHandler) deleteEdgeRouterPolicy(id string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	filter := fmt.Sprintf("name=\"zrok-%v\"", id)
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
		logrus.Infof("deleted edge router policy '%v'", erpId)
	} else {
		logrus.Infof("found '%d' edge router policies, expected 1", len(listResp.Payload.Data))
	}
	return nil
}

func (self *disableHandler) deleteIdentity(id string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	req := &identity_edge.DeleteIdentityParams{
		ID:      id,
		Context: context.Background(),
	}
	req.SetTimeout(30 * time.Second)
	_, err := edge.Identity.DeleteIdentity(req, nil)
	if err != nil {
		return err
	}
	logrus.Infof("deleted identity '%v'", id)
	return nil
}

func (self *disableHandler) removeEnvironment(envId int, tx *sqlx.Tx) error {
	if err := str.DeleteEnvironment(envId, tx); err != nil {
		return errors.Wrapf(err, "error deleting environment '%d'", envId)
	}
	return nil
}
