package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/openziti/edge-api/rest_management_api_client"
	edge_service "github.com/openziti/edge-api/rest_management_api_client/service"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/environment"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type disableHandler struct{}

func newDisableHandler() *disableHandler {
	return &disableHandler{}
}

func (h *disableHandler) Handle(params environment.DisableParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction for user '%v': %v", principal.Email, err)
		return environment.NewDisableInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	env, err := str.FindEnvironmentForAccount(params.Body.Identity, int(principal.ID), trx)
	if err != nil {
		logrus.Errorf("identity check failed for user '%v': %v", principal.Email, err)
		return environment.NewDisableUnauthorized()
	}

	edge, err := zrokEdgeSdk.Client(cfg.Ziti)
	if err != nil {
		logrus.Errorf("error getting edge client for user '%v': %v", principal.Email, err)
		return environment.NewDisableInternalServerError()
	}

	if err := disableEnvironment(env, trx, edge); err != nil {
		logrus.Errorf("error disabling environment for user '%v': %v", principal.Email, err)
		return environment.NewDisableInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		logrus.Errorf("error committing for user '%v': %v", principal.Email, err)
		return environment.NewDisableInternalServerError()
	}

	return environment.NewDisableOK()
}

func disableEnvironment(env *store.Environment, trx *sqlx.Tx, edge *rest_management_api_client.ZitiEdgeManagement) error {
	if err := removeSharesForEnvironment(env, trx, edge); err != nil {
		return errors.Wrapf(err, "error removing shares for environment '%v'", env.ZId)
	}
	if err := removeFrontendsForEnvironment(env, trx, edge); err != nil {
		return errors.Wrapf(err, "error removing frontends for environment '%v'", env.ZId)
	}
	if err := removeAgentRemoteForEnvironment(env, trx, edge); err != nil {
		return errors.Wrapf(err, "error removing agent remote for '%v'", env.ZId)
	}
	if err := zrokEdgeSdk.DeleteEdgeRouterPolicy(env.ZId, edge); err != nil {
		return errors.Wrapf(err, "error deleting edge router policy for environment '%v'", env.ZId)
	}
	if err := zrokEdgeSdk.DeleteIdentity(env.ZId, edge); err != nil {
		return errors.Wrapf(err, "error deleting identity for environment '%v'", env.ZId)
	}
	if err := removeEnvironmentFromStore(env, trx); err != nil {
		return errors.Wrapf(err, "error removing environment '%v' from store", env.ZId)
	}
	return nil
}

func removeSharesForEnvironment(env *store.Environment, trx *sqlx.Tx, edge *rest_management_api_client.ZitiEdgeManagement) error {
	shrs, err := str.FindSharesForEnvironment(env.Id, trx)
	if err != nil {
		return err
	}
	for _, shr := range shrs {
		shrToken := shr.Token
		logrus.Infof("garbage collecting share '%v' for environment '%v'", shrToken, env.ZId)
		if err := zrokEdgeSdk.DeleteServiceEdgeRouterPolicyForShare(env.ZId, shrToken, edge); err != nil {
			logrus.Error(err)
		}
		if err := zrokEdgeSdk.DeleteServicePoliciesDialForShare(env.ZId, shrToken, edge); err != nil {
			logrus.Error(err)
		}
		if err := zrokEdgeSdk.DeleteServicePoliciesBindForShare(env.ZId, shrToken, edge); err != nil {
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

func removeFrontendsForEnvironment(env *store.Environment, trx *sqlx.Tx, edge *rest_management_api_client.ZitiEdgeManagement) error {
	fes, err := str.FindFrontendsForEnvironment(env.Id, trx)
	if err != nil {
		return err
	}
	for _, fe := range fes {
		if err := zrokEdgeSdk.DeleteServicePolicies(env.ZId, fmt.Sprintf("tags.zrokFrontendToken=\"%v\" and type=1", fe.Token), edge); err != nil {
			logrus.Errorf("error removing frontend access for '%v': %v", fe.Token, err)
		}
	}
	return nil
}

func removeAgentRemoteForEnvironment(env *store.Environment, trx *sqlx.Tx, edge *rest_management_api_client.ZitiEdgeManagement) error {
	enrolled, err := str.IsAgentEnrolledForEnvironment(env.Id, trx)
	if err != nil {
		return err
	}
	if enrolled {
		ae, err := str.FindAgentEnrollmentForEnvironment(env.Id, trx)
		if err != nil {
			return err
		}
		if err := zrokEdgeSdk.DeleteServiceEdgeRouterPolicyForAgentRemote(env.ZId, ae.Token, edge); err != nil {
			return err
		}
		if err := zrokEdgeSdk.DeleteServicePoliciesDialForAgentRemote(env.ZId, ae.Token, edge); err != nil {
			return err
		}
		if err := zrokEdgeSdk.DeleteServicePoliciesBindForAgentRemote(env.ZId, ae.Token, edge); err != nil {
			return err
		}
		filter := fmt.Sprintf("name=\"%v\"", ae.Token)
		limit := int64(1)
		offset := int64(0)
		listReq := &edge_service.ListServicesParams{
			Filter:  &filter,
			Limit:   &limit,
			Offset:  &offset,
			Context: context.Background(),
		}
		listReq.SetTimeout(30 * time.Second)
		listResp, err := edge.Service.ListServices(listReq, nil)
		if err != nil {
			return err
		}
		aeZId := ""
		if len(listResp.Payload.Data) > 0 {
			aeZId = *(listResp.Payload.Data[0].ID)
		} else {
			return errors.New("no agent remoting identity found")
		}
		if err := zrokEdgeSdk.DeleteService(env.ZId, aeZId, edge); err != nil {
			return err
		}
		if err := str.DeleteAgentEnrollment(ae.Id, trx); err != nil {
			return err
		}
	}
	return nil
}

func removeEnvironmentFromStore(env *store.Environment, trx *sqlx.Tx) error {
	shrs, err := str.FindSharesForEnvironment(env.Id, trx)
	if err != nil {
		return errors.Wrapf(err, "error finding shares for environment '%d'", env.Id)
	}
	for _, shr := range shrs {
		if err := str.DeleteShare(shr.Id, trx); err != nil {
			return errors.Wrapf(err, "error deleting share '%d' for environment '%d'", shr.Id, env.Id)
		}
	}
	fes, err := str.FindFrontendsForEnvironment(env.Id, trx)
	if err != nil {
		return errors.Wrapf(err, "error finding frontends for environment '%d'", env.Id)
	}
	for _, fe := range fes {
		if err := str.DeleteFrontend(fe.Id, trx); err != nil {
			return errors.Wrapf(err, "error deleting frontend '%d' for environment '%d'", fe.Id, env.Id)
		}
	}
	if err := str.DeleteEnvironment(env.Id, trx); err != nil {
		return errors.Wrapf(err, "error deleting environment '%d'", env.Id)
	}
	return nil
}
