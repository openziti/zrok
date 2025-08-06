package controller

import (
	"context"

	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/agent"
	"github.com/sirupsen/logrus"
)

type agentPingHandler struct{}

func newAgentPingHandler() *agentPingHandler {
	return &agentPingHandler{}
}

func (h *agentPingHandler) Handle(params agent.PingParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction for '%v': %v", principal.Email, err)
		return agent.NewPingInternalServerError()
	}
	defer trx.Rollback()

	env, err := str.FindEnvironmentForAccount(params.Body.EnvZID, int(principal.ID), trx)
	if err != nil {
		logrus.Errorf("error finding environment '%v' for '%v': %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewPingUnauthorized()
	}

	ae, err := str.FindAgentEnrollmentForEnvironment(env.Id, trx)
	if err != nil {
		logrus.Errorf("error finding agent enrollment for '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewPingBadGateway()
	}

	agentClient, agentConn, err := agentCtrl.NewClient(ae.Token)
	if err != nil {
		logrus.Errorf("error creating agent client for '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewPingInternalServerError()
	}
	defer agentConn.Close()

	resp, err := agentClient.Version(context.Background(), &agentGrpc.VersionRequest{})
	if err != nil {
		logrus.Errorf("error retrieving agent version for '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewPingBadGateway()
	}

	return agent.NewPingOK().WithPayload(&agent.PingOKBody{Version: resp.V})
}
