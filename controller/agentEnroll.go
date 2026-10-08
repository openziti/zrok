package controller

import (
	"fmt"

	"github.com/go-openapi/runtime/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/zrok/v2/controller/automation"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/agent"
)

type agentEnrollHandler struct {
	// commit is replaceable so a test can fail the commit after the ziti objects exist.
	commit func(*sqlx.Tx) error
}

func newAgentEnrollHandler() *agentEnrollHandler {
	return &agentEnrollHandler{commit: (*sqlx.Tx).Commit}
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
	dl.Infof("created enrollment token for '%v'", principal.Email)

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

	// every ziti object created from here on is deleted by id unless the enrollment record commits
	compensation := newZitiCompensation(compensatingAgentEnrollment, token)
	compensation.add(zitiService, zId)
	committed := false
	defer func() {
		if !committed {
			compensation.run(ziti)
		}
	}()

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
	bindZId, err := ziti.ServicePolicies.CreateBind(bindOpts)
	if err != nil {
		dl.Errorf("error creating agent remoting bind policy for '%v' (%v): %v", env.ZId, principal.Email, err)
		return agent.NewEnrollInternalServerError()
	}
	compensation.add(zitiServicePolicy, bindZId)

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
	dialZId, err := ziti.ServicePolicies.CreateDial(dialOpts)
	if err != nil {
		dl.Errorf("error creating agent remoting dial policy for '%v' (%v): %v", env.ZId, principal.Email, err)
		return agent.NewEnrollInternalServerError()
	}
	compensation.add(zitiServicePolicy, dialZId)

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
	serpZId, err := ziti.ServiceEdgeRouterPolicies.Create(serpOpts)
	if err != nil {
		dl.Errorf("error creating agent remoting serp for '%v' (%v): %v", env.ZId, principal.Email, err)
		return agent.NewEnrollInternalServerError()
	}
	compensation.add(zitiServiceEdgeRouterPolicy, serpZId)

	if _, err := str.CreateAgentEnrollment(env.Id, token, trx); err != nil {
		dl.Errorf("error storing agent enrollment for '%v' (%v): %v", env.ZId, principal.Email, err)
		return agent.NewEnrollInternalServerError()
	}

	if err := h.commit(trx); err != nil {
		dl.Errorf("error committing agent enrollment record for '%v' (%v): %v", env.ZId, principal.Email, err)
		return agent.NewEnrollInternalServerError()
	}
	committed = true

	return agent.NewEnrollOK().WithPayload(&agent.EnrollOKBody{Token: token})
}
