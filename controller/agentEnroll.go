package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/agent"
	"github.com/sirupsen/logrus"
)

type agentEnrollHandler struct{}

func newAgentEnrollHandler() *agentEnrollHandler {
	return &agentEnrollHandler{}
}

func (h *agentEnrollHandler) Handle(params agent.EnrollParams, principal *rest_model_zrok.Principal) middleware.Responder {
	// start transaction early, if it fails, don't bother creating ziti resources
	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction for '%v': %v", principal.Email, err)
		return agent.NewEnrollInternalServerError()
	}
	defer trx.Rollback()

	env, err := str.FindEnvironmentForAccount(params.Body.EnvZID, int(principal.ID), trx)
	if err != nil {
		logrus.Errorf("error finding environment '%v' for '%v': %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewEnrollUnauthorized()
	}

	if _, err := str.FindAgentEnrollmentForEnvironment(env.Id, trx); err == nil {
		logrus.Errorf("environment '%v' (%v) is already enrolled!", params.Body.EnvZID, principal.Email)
		return agent.NewEnrollBadRequest()
	}

	client, err := zrokEdgeSdk.Client(cfg.Ziti)
	if err != nil {
		logrus.Errorf("error getting ziti client for '%v': %v", principal.Email, err)
		return agent.NewEnrollInternalServerError()
	}

	token, err := CreateToken()
	if err != nil {
		logrus.Errorf("error creating agent enrollment token for '%v': %v", principal.Email, err)
		return agent.NewEnrollInternalServerError()
	}
	logrus.Infof("enrollment token: %v", token)

	zId, err := zrokEdgeSdk.CreateService(token, nil, map[string]interface{}{"zrokEnvZId": env.ZId}, client)
	if err != nil {
		logrus.Errorf("error creating agent remoting service for '%v' (%v): %v", env.ZId, principal.Email, err)
		return agent.NewEnrollInternalServerError()
	}

	if err := zrokEdgeSdk.CreateServicePolicyBind(env.ZId+"-"+token+"-bind", zId, env.ZId, zrokEdgeSdk.ZrokAgentRemoteTags(token, env.ZId).SubTags, client); err != nil {
		logrus.Errorf("error creating agent remoting bind policy for '%v' (%v): %v", env.ZId, principal.Email, err.Error())
		return agent.NewEnrollInternalServerError()
	}

	if err := zrokEdgeSdk.CreateServicePolicyDial(env.ZId+"-"+token+"-dial", zId, []string{cfg.AgentController.ZId}, zrokEdgeSdk.ZrokAgentRemoteTags(token, env.ZId).SubTags, client); err != nil {
		logrus.Errorf("error creating agent remoting dial policy for '%v' (%v): %v", env.ZId, principal.Email, err.Error())
		return agent.NewEnrollInternalServerError()
	}

	if err := zrokEdgeSdk.CreateAgentRemoteServiceEdgeRouterPolicy(env.ZId, token, zId, client); err != nil {
		logrus.Errorf("error creating agent remoting serp for '%v' (%v): %v", env.ZId, principal.Email, err)
		return agent.NewEnrollInternalServerError()
	}

	if _, err := str.CreateAgentEnrollment(env.Id, token, trx); err != nil {
		logrus.Errorf("error storing agent enrollment for '%v' (%v): %v", env.ZId, principal.Email, err)
		return agent.NewEnrollInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		logrus.Errorf("error committing agent enrollment record for '%v' (%v): %v", env.ZId, principal.Email, err)
		return agent.NewEnrollInternalServerError()
	}

	return agent.NewEnrollOK().WithPayload(&agent.EnrollOKBody{Token: token})
}
