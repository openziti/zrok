package controller

import (
	"fmt"

	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/zrok/controller/automation"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/agent"
)

type agentEnrollHandler struct{}

func newAgentEnrollHandler() *agentEnrollHandler {
	return &agentEnrollHandler{}
}

func (h *agentEnrollHandler) Handle(params agent.EnrollParams, principal *rest_model_zrok.Principal) middleware.Responder {
	// start transaction early, if it fails, don't bother creating ziti resources
	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction for '%v': %v", principal.Email, err)
		return agent.NewEnrollInternalServerError()
	}
	defer trx.Rollback()

	env, err := str.FindEnvironmentForAccount(params.Body.EnvZID, int(principal.ID), trx)
	if err != nil {
		dl.Errorf("error finding environment '%v' for '%v': %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewEnrollUnauthorized()
	}

	if _, err := str.FindAgentEnrollmentForEnvironment(env.Id, trx); err == nil {
		dl.Errorf("environment '%v' (%v) is already enrolled!", params.Body.EnvZID, principal.Email)
		return agent.NewEnrollBadRequest()
	}

	token, err := CreateToken()
	if err != nil {
		dl.Errorf("error creating agent enrollment token for '%v': %v", principal.Email, err)
		return agent.NewEnrollInternalServerError()
	}
	dl.Infof("enrollment token: %v", token)

	ziti, err := automation.NewZitiAutomation(cfg.Ziti)
	if err != nil {
		dl.Errorf("error getting automation client for '%v': %v", principal.Email, err)
		return agent.NewEnrollInternalServerError()
	}

	// create service for agent remoting
	tags := automation.ZrokAgentRemoteTags(token, env.ZId).WithTag("zrokEnvZId", env.ZId)
	serviceOpts := &automation.ServiceOptions{
		BaseOptions: automation.BaseOptions{
			Name: token,
			Tags: tags,
		},
		EncryptionRequired: true,
	}
	zId, err := ziti.Services.Create(serviceOpts)
	if err != nil {
		dl.Errorf("error creating agent remoting service for '%v' (%v): %v", env.ZId, principal.Email, err)
		return agent.NewEnrollInternalServerError()
	}

	// create bind policy for the service
	bindPolicyName := env.ZId + "-" + token + "-bind"
	bindOpts := &automation.ServicePolicyOptions{
		BaseOptions: automation.BaseOptions{
			Name: bindPolicyName,
			Tags: automation.ZrokAgentRemoteTags(token, env.ZId),
		},
		IdentityRoles: []string{"@" + env.ZId},
		ServiceRoles:  []string{"@" + zId},
		PolicyType:    rest_model.DialBindBind,
		Semantic:      rest_model.SemanticAllOf,
	}
	if _, err := ziti.ServicePolicies.CreateBind(bindOpts); err != nil {
		dl.Errorf("error creating agent remoting bind policy for '%v' (%v): %v", env.ZId, principal.Email, err)
		return agent.NewEnrollInternalServerError()
	}

	// create dial policy for the service
	dialPolicyName := env.ZId + "-" + token + "-dial"
	dialOpts := &automation.ServicePolicyOptions{
		BaseOptions: automation.BaseOptions{
			Name: dialPolicyName,
			Tags: automation.ZrokAgentRemoteTags(token, env.ZId),
		},
		IdentityRoles: []string{"@" + cfg.AgentController.ZId},
		ServiceRoles:  []string{"@" + zId},
		PolicyType:    rest_model.DialBindDial,
		Semantic:      rest_model.SemanticAllOf,
	}
	if _, err := ziti.ServicePolicies.CreateDial(dialOpts); err != nil {
		dl.Errorf("error creating agent remoting dial policy for '%v' (%v): %v", env.ZId, principal.Email, err)
		return agent.NewEnrollInternalServerError()
	}

	// create service edge router policy
	serpOpts := &automation.ServiceEdgeRouterPolicyOptions{
		BaseOptions: automation.BaseOptions{
			Name: token,
			Tags: automation.ZrokAgentRemoteTags(token, env.ZId),
		},
		ServiceRoles:    []string{fmt.Sprintf("@%v", zId)},
		EdgeRouterRoles: []string{"#all"},
		Semantic:        rest_model.SemanticAllOf,
	}
	if _, err := ziti.ServiceEdgeRouterPolicies.Create(serpOpts); err != nil {
		dl.Errorf("error creating agent remoting serp for '%v' (%v): %v", env.ZId, principal.Email, err)
		return agent.NewEnrollInternalServerError()
	}

	if _, err := str.CreateAgentEnrollment(env.Id, token, trx); err != nil {
		dl.Errorf("error storing agent enrollment for '%v' (%v): %v", env.ZId, principal.Email, err)
		return agent.NewEnrollInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing agent enrollment record for '%v' (%v): %v", env.ZId, principal.Email, err)
		return agent.NewEnrollInternalServerError()
	}

	return agent.NewEnrollOK().WithPayload(&agent.EnrollOKBody{Token: token})
}
