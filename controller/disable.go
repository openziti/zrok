package controller

import (
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/environment"
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
	if err := h.removeSharesForEnvironment(envId, tx, edge); err != nil {
		logrus.Errorf("error removing shares for environment: %v", err)
		return environment.NewDisableInternalServerError()
	}
	if err := h.removeFrontendsForEnvironment(envId, tx, edge); err != nil {
		logrus.Errorf("error removing frontends for environment: %v", err)
		return environment.NewDisableInternalServerError()
	}
	if err := h.removeEnvironment(envId, tx); err != nil {
		logrus.Errorf("error removing environment: %v", err)
		return environment.NewDisableInternalServerError()
	}
	if err := zrokEdgeSdk.DeleteEdgeRouterPolicy(env.ZId, edge); err != nil {
		logrus.Errorf("error deleting edge router policy: %v", err)
		return environment.NewDisableInternalServerError()
	}
	if err := zrokEdgeSdk.DeleteIdentity(params.Body.Identity, edge); err != nil {
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

func (h *disableHandler) removeSharesForEnvironment(envId int, tx *sqlx.Tx, edge *rest_management_api_client.ZitiEdgeManagement) error {
	env, err := str.GetEnvironment(envId, tx)
	if err != nil {
		return err
	}
	shrs, err := str.FindSharesForEnvironment(envId, tx)
	if err != nil {
		return err
	}
	for _, shr := range shrs {
		shrToken := shr.Token
		logrus.Infof("garbage collecting share '%v' for environment '%v'", shrToken, env.ZId)
		if err := zrokEdgeSdk.DeleteServiceEdgeRouterPolicy(env.ZId, shrToken, edge); err != nil {
			logrus.Error(err)
		}
		if err := zrokEdgeSdk.DeleteServicePolicyDial(env.ZId, shrToken, edge); err != nil {
			logrus.Error(err)
		}
		if err := zrokEdgeSdk.DeleteServicePolicyBind(env.ZId, shrToken, edge); err != nil {
			logrus.Error(err)
		}
		if err := zrokEdgeSdk.DeleteConfig(env.ZId, shrToken, edge); err != nil {
			logrus.Error(err)
		}
		if err := zrokEdgeSdk.DeleteService(env.ZId, shr.ZId, edge); err != nil {
			logrus.Error(err)
		}
		logrus.Infof("removed share '%v' for environment '%v'", shr.Token, env.ZId)
	}
	return nil
}

func (h *disableHandler) removeFrontendsForEnvironment(envId int, tx *sqlx.Tx, edge *rest_management_api_client.ZitiEdgeManagement) error {
	env, err := str.GetEnvironment(envId, tx)
	if err != nil {
		return err
	}
	fes, err := str.FindFrontendsForEnvironment(envId, tx)
	if err != nil {
		return err
	}
	for _, fe := range fes {
		if err := zrokEdgeSdk.DeleteServicePolicy(env.ZId, fmt.Sprintf("tags.zrokFrontendToken=\"%v\" and type=1", fe.Token), edge); err != nil {
			logrus.Errorf("error removing frontend access for '%v': %v", fe.Token, err)
		}
	}
	return nil
}

func (h *disableHandler) removeEnvironment(envId int, tx *sqlx.Tx) error {
	shrs, err := str.FindSharesForEnvironment(envId, tx)
	if err != nil {
		return errors.Wrapf(err, "error finding shares for environment '%d'", envId)
	}
	for _, shr := range shrs {
		if err := str.DeleteShare(shr.Id, tx); err != nil {
			return errors.Wrapf(err, "error deleting share '%d' for environment '%d'", shr.Id, envId)
		}
	}
	fes, err := str.FindFrontendsForEnvironment(envId, tx)
	if err != nil {
		return errors.Wrapf(err, "error finding frontends for environment '%d'", envId)
	}
	for _, fe := range fes {
		if err := str.DeleteFrontend(fe.Id, tx); err != nil {
			return errors.Wrapf(err, "error deleting frontend '%d' for environment '%d'", fe.Id, envId)
		}
	}
	if err := str.DeleteEnvironment(envId, tx); err != nil {
		return errors.Wrapf(err, "error deleting environment '%d'", envId)
	}
	return nil
}
