package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/environment"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type disableHandler struct {
}

func newDisableHandler() *disableHandler {
	return &disableHandler{}
}

func (h *disableHandler) Handle(params environment.DisableParams, principal *rest_model_zrok.Principal) middleware.Responder {
	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return environment.NewDisableInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()
	envId, err := h.checkZitiIdentity(params.Body.Identity, principal, tx)
	if err != nil {
		logrus.Errorf("identity check failed: %v", err)
		return environment.NewDisableUnauthorized()
	}
	env, err := str.GetEnvironment(envId, tx)
	if err != nil {
		logrus.Errorf("error getting environment: %v", err)
		return environment.NewDisableInternalServerError()
	}
	edge, err := edgeClient()
	if err != nil {
		logrus.Errorf("error getting edge client: %v", err)
		return environment.NewDisableInternalServerError()
	}
	if err := h.removeServicesForEnvironment(envId, tx, edge); err != nil {
		logrus.Errorf("error removing services for environment: %v", err)
		return environment.NewDisableInternalServerError()
	}
	if err := h.removeEnvironment(envId, tx); err != nil {
		logrus.Errorf("error removing environment: %v", err)
		return environment.NewDisableInternalServerError()
	}
	if err := deleteEdgeRouterPolicy(env.ZId, edge); err != nil {
		logrus.Errorf("error deleting edge router policy: %v", err)
		return environment.NewDisableInternalServerError()
	}
	if err := deleteIdentity(params.Body.Identity, edge); err != nil {
		logrus.Errorf("error deleting identity: %v", err)
		return environment.NewDisableInternalServerError()
	}
	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing: %v", err)
	}
	return environment.NewDisableOK()
}

func (h *disableHandler) checkZitiIdentity(id string, principal *rest_model_zrok.Principal, tx *sqlx.Tx) (int, error) {
	envs, err := str.FindEnvironmentsForAccount(int(principal.ID), tx)
	if err != nil {
		return -1, err
	}
	for _, env := range envs {
		if env.ZId == id {
			return env.Id, nil
		}
	}
	return -1, errors.Errorf("no such environment '%v'", id)
}

func (h *disableHandler) removeServicesForEnvironment(envId int, tx *sqlx.Tx, edge *rest_management_api_client.ZitiEdgeManagement) error {
	env, err := str.GetEnvironment(envId, tx)
	if err != nil {
		return err
	}
	svcs, err := str.FindServicesForEnvironment(envId, tx)
	if err != nil {
		return err
	}
	for _, svc := range svcs {
		svcToken := svc.Token
		logrus.Infof("garbage collecting service '%v' for environment '%v'", svcToken, env.ZId)
		if err := deleteServiceEdgeRouterPolicy(env.ZId, svcToken, edge); err != nil {
			logrus.Error(err)
		}
		if err := deleteServicePolicyDial(env.ZId, svcToken, edge); err != nil {
			logrus.Error(err)
		}
		if err := deleteServicePolicyBind(env.ZId, svcToken, edge); err != nil {
			logrus.Error(err)
		}
		if err := deleteConfig(env.ZId, svcToken, edge); err != nil {
			logrus.Error(err)
		}
		if err := deleteService(env.ZId, svc.ZId, edge); err != nil {
			logrus.Error(err)
		}
		logrus.Infof("removed service '%v' for environment '%v'", svc.Token, env.ZId)
	}
	return nil
}

func (h *disableHandler) removeEnvironment(envId int, tx *sqlx.Tx) error {
	svcs, err := str.FindServicesForEnvironment(envId, tx)
	if err != nil {
		return errors.Wrapf(err, "error finding services for environment '%d'", envId)
	}
	for _, svc := range svcs {
		if err := str.DeleteService(svc.Id, tx); err != nil {
			return errors.Wrapf(err, "error deleting service '%d' for environment '%d'", svc.Id, envId)
		}
	}
	if err := str.DeleteEnvironment(envId, tx); err != nil {
		return errors.Wrapf(err, "error deleting environment '%d'", envId)
	}
	return nil
}
