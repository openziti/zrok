package controller

import (
	"fmt"

	"github.com/go-openapi/runtime/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/controller/automation"
	"github.com/openziti/zrok/controller/store"
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

	ziti, err := automation.NewZitiAutomation(cfg)
	if err != nil {
		logrus.Errorf("error getting automation client for user '%v': %v", principal.Email, err)
		return environment.NewDisableInternalServerError()
	}

	if err := disableEnvironment(env, trx, ziti); err != nil {
		logrus.Errorf("error disabling environment for user '%v': %v", principal.Email, err)
		return environment.NewDisableInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		logrus.Errorf("error committing for user '%v': %v", principal.Email, err)
		return environment.NewDisableInternalServerError()
	}

	return environment.NewDisableOK()
}

func disableEnvironment(env *store.Environment, trx *sqlx.Tx, ziti *automation.ZitiAutomation) error {
	if err := removeSharesForEnvironment(env, trx, ziti); err != nil {
		return errors.Wrapf(err, "error removing shares for environment '%v'", env.ZId)
	}
	if err := removeFrontendsForEnvironment(env, trx, ziti); err != nil {
		return errors.Wrapf(err, "error removing frontends for environment '%v'", env.ZId)
	}
	if err := removeAgentRemoteForEnvironment(env, trx, ziti); err != nil {
		return errors.Wrapf(err, "error removing agent remote for '%v'", env.ZId)
	}

	// delete edge router policy for environment
	erpFilter := fmt.Sprintf("name=\"%v\"", env.ZId)
	if err := ziti.EdgeRouterPolicies.DeleteWithFilter(erpFilter); err != nil {
		return errors.Wrapf(err, "error deleting edge router policy for environment '%v'", env.ZId)
	}

	// delete identity for environment
	if err := ziti.Identities.Delete(env.ZId); err != nil {
		return errors.Wrapf(err, "error deleting identity for environment '%v'", env.ZId)
	}

	if err := removeEnvironmentFromStore(env, trx); err != nil {
		return errors.Wrapf(err, "error removing environment '%v' from store", env.ZId)
	}
	return nil
}

func removeSharesForEnvironment(env *store.Environment, trx *sqlx.Tx, ziti *automation.ZitiAutomation) error {
	shrs, err := str.FindSharesForEnvironment(env.Id, trx)
	if err != nil {
		return err
	}
	for _, shr := range shrs {
		shrToken := shr.Token
		logrus.Infof("garbage collecting share '%v' for environment '%v'", shrToken, env.ZId)

		// delete service edge router policies for share
		serpFilter := fmt.Sprintf("tags.zrokShareToken=\"%v\"", shrToken)
		if err := ziti.ServiceEdgeRouterPolicies.DeleteWithFilter(serpFilter); err != nil {
			logrus.Error(err)
		}

		// delete dial service policies for share
		dialFilter := fmt.Sprintf("tags.zrokShareToken=\"%v\" and type=1", shrToken)
		if err := ziti.ServicePolicies.DeleteWithFilter(dialFilter); err != nil {
			logrus.Error(err)
		}

		// delete bind service policies for share
		bindFilter := fmt.Sprintf("tags.zrokShareToken=\"%v\" and type=2", shrToken)
		if err := ziti.ServicePolicies.DeleteWithFilter(bindFilter); err != nil {
			logrus.Error(err)
		}

		// delete configs for share
		configFilter := fmt.Sprintf("tags.zrokShareToken=\"%v\"", shrToken)
		if err := ziti.Configs.DeleteWithFilter(configFilter); err != nil {
			logrus.Error(err)
		}

		// delete service
		if err := ziti.Services.Delete(shr.ZId); err != nil {
			logrus.Error(err)
		}

		logrus.Infof("removed share '%v' for environment '%v'", shr.Token, env.ZId)
	}
	return nil
}

func removeFrontendsForEnvironment(env *store.Environment, trx *sqlx.Tx, ziti *automation.ZitiAutomation) error {
	fes, err := str.FindFrontendsForEnvironment(env.Id, trx)
	if err != nil {
		return err
	}
	for _, fe := range fes {
		filter := fmt.Sprintf("tags.zrokFrontendToken=\"%v\" and type=1", fe.Token)
		if err := ziti.ServicePolicies.DeleteWithFilter(filter); err != nil {
			logrus.Errorf("error removing frontend access for '%v': %v", fe.Token, err)
		}
	}
	return nil
}

func removeAgentRemoteForEnvironment(env *store.Environment, trx *sqlx.Tx, ziti *automation.ZitiAutomation) error {
	enrolled, err := str.IsAgentEnrolledForEnvironment(env.Id, trx)
	if err != nil {
		return err
	}
	if enrolled {
		ae, err := str.FindAgentEnrollmentForEnvironment(env.Id, trx)
		if err != nil {
			return err
		}

		// delete service edge router policies for agent remote
		serpFilter := fmt.Sprintf("tags.zrokAgentRemote=\"%v\"", ae.Token)
		if err := ziti.ServiceEdgeRouterPolicies.DeleteWithFilter(serpFilter); err != nil {
			return err
		}

		// delete dial service policies for agent remote
		dialFilter := fmt.Sprintf("tags.zrokAgentRemote=\"%v\" and type=1", ae.Token)
		if err := ziti.ServicePolicies.DeleteWithFilter(dialFilter); err != nil {
			return err
		}

		// delete bind service policies for agent remote
		bindFilter := fmt.Sprintf("tags.zrokAgentRemote=\"%v\" and type=2", ae.Token)
		if err := ziti.ServicePolicies.DeleteWithFilter(bindFilter); err != nil {
			return err
		}

		// delete agent remote service by name
		serviceFilter := fmt.Sprintf("name=\"%v\"", ae.Token)
		if err := ziti.Services.DeleteWithFilter(serviceFilter); err != nil {
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
