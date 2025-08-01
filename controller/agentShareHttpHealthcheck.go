package controller

import (
	"context"

	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/controller/agentController"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/agent"
	"github.com/sirupsen/logrus"
)

type agentShareHttpHealthcheckHandler struct{}

func newAgentShareHttpHealthcheckHandler() *agentShareHttpHealthcheckHandler {
	return &agentShareHttpHealthcheckHandler{}
}

func (h *agentShareHttpHealthcheckHandler) Handle(params agent.ShareHTTPHealthcheckParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction for '%v': %v", principal.Email, err)
		return agent.NewShareHTTPHealthcheckInternalServerError()
	}
	defer trx.Rollback()

	env, err := str.FindEnvironmentForAccount(params.Body.EnvZID, int(principal.ID), trx)
	if err != nil {
		logrus.Errorf("error finding environment '%v' for '%v': %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewShareHTTPHealthcheckUnauthorized()
	}

	ae, err := str.FindAgentEnrollmentForEnvironment(env.Id, trx)
	if err != nil {
		logrus.Errorf("error finding agent enrollment for environment '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewShareHTTPHealthcheckBadGateway()
	}
	_ = trx.Rollback() // ...or will block share trx on sqlite

	acli, aconn, err := agentController.NewAgentClient(ae.Token, cfg.AgentController)
	if err != nil {
		logrus.Errorf("error creating agent client for '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewShareHTTPHealthcheckInternalServerError()
	}
	defer aconn.Close()

	req := &agentGrpc.ShareHttpHealthcheckRequest{
		Token:                params.Body.ShareToken,
		HttpVerb:             params.Body.HTTPVerb,
		Endpoint:             params.Body.Endpoint,
		ExpectedHttpResponse: uint32(params.Body.ExpectedHTTPResponse),
		TimeoutMs:            uint64(params.Body.TimeoutMs),
	}
	resp, err := acli.ShareHttpHealthcheck(context.Background(), req)
	if err != nil {
		logrus.Infof("error invoking remoted share '%v' http healthcheck for '%v': %v", params.Body.ShareToken, params.Body.EnvZID, err)
		return agent.NewShareHTTPHealthcheckBadGateway()
	}

	return agent.NewShareHTTPHealthcheckOK().WithPayload(&agent.ShareHTTPHealthcheckOKBody{
		Healthy: resp.Healthy,
		Error:   resp.Error,
	})
}
