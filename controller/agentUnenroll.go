package controller

import (
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/controller/automation"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/agent"
	"github.com/sirupsen/logrus"
)

type agentUnenrollHandler struct{}

func newAgentUnenrollHandler() *agentUnenrollHandler {
	return &agentUnenrollHandler{}
}

func (h *agentUnenrollHandler) Handle(params agent.UnenrollParams, principal *rest_model_zrok.Principal) middleware.Responder {
	// start transaction early, if it fails, don't bother creating ziti resources
	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction for '%v': %v", principal.Email, err)
		return agent.NewUnenrollInternalServerError()
	}
	defer trx.Rollback()

	env, err := str.FindEnvironmentForAccount(params.Body.EnvZID, int(principal.ID), trx)
	if err != nil {
		logrus.Errorf("error finding environment '%v' for '%v': %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewUnenrollUnauthorized()
	}

	ae, err := str.FindAgentEnrollmentForEnvironment(env.Id, trx)
	if err != nil {
		logrus.Errorf("error finding agent enrollment for '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewUnenrollBadRequest()
	}

	ziti, err := automation.NewZitiAutomation(cfg.Ziti)
	if err != nil {
		logrus.Errorf("error getting automation client for '%v': %v", principal.Email, err)
		return agent.NewUnenrollInternalServerError()
	}

	// delete service edge router policies for agent remote
	serpFilter := fmt.Sprintf("tags.zrokAgentRemote=\"%v\"", ae.Token)
	if err := ziti.ServiceEdgeRouterPolicies.DeleteWithFilter(serpFilter); err != nil {
		logrus.Errorf("error removing agent remote serp for '%v' (%v): %v", env.ZId, principal.Email, err)
		return agent.NewUnenrollInternalServerError()
	}

	// delete dial service policies for agent remote
	dialFilter := fmt.Sprintf("tags.zrokAgentRemote=\"%v\" and type=1", ae.Token)
	if err := ziti.ServicePolicies.DeleteWithFilter(dialFilter); err != nil {
		logrus.Errorf("error removing agent remote dial service policy for '%v' (%v): %v", env.ZId, principal.Email, err)
		return agent.NewUnenrollInternalServerError()
	}

	// delete bind service policies for agent remote
	bindFilter := fmt.Sprintf("tags.zrokAgentRemote=\"%v\" and type=2", ae.Token)
	if err := ziti.ServicePolicies.DeleteWithFilter(bindFilter); err != nil {
		logrus.Errorf("error removing agent remote bind service policy for '%v' (%v): %v", env.ZId, principal.Email, err)
		return agent.NewUnenrollInternalServerError()
	}

	// find and delete the agent remote service
	serviceFilter := fmt.Sprintf("name=\"%v\"", ae.Token)
	if err := ziti.Services.DeleteWithFilter(serviceFilter); err != nil {
		logrus.Errorf("error removing agent remote service for '%v' (%v): %v", env.ZId, principal.Email, err)
		return agent.NewUnenrollInternalServerError()
	}

	if err := str.DeleteAgentEnrollment(ae.Id, trx); err != nil {
		logrus.Errorf("error deleting agent enrollment for '%v' (%v): %v", env.ZId, principal.Email, err)
		return agent.NewUnenrollInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		logrus.Errorf("error committing agent unenrollment for '%v' (%v): %v", env.ZId, principal.Email, err)
		return agent.NewUnenrollInternalServerError()
	}

	return agent.NewUnenrollOK()
}
