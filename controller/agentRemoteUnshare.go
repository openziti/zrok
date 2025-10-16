package controller

import (
	"context"

	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/agent"
)

type agentRemoteUnshareHandler struct{}

func newAgentRemoteUnshareHandler() *agentRemoteUnshareHandler {
	return &agentRemoteUnshareHandler{}
}

func (h *agentRemoteUnshareHandler) Handle(params agent.RemoteUnshareParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction for '%v': %v", principal.Email, err)
		return agent.NewRemoteUnshareInternalServerError()
	}
	defer trx.Rollback()

	env, err := str.FindEnvironmentForAccount(params.Body.EnvZID, int(principal.ID), trx)
	if err != nil {
		dl.Errorf("error finding environment '%v' for '%v': %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewRemoteUnshareUnauthorized()
	}

	ae, err := str.FindAgentEnrollmentForEnvironment(env.Id, trx)
	if err != nil {
		dl.Errorf("error finding agent enrollment for environment '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewRemoteUnshareBadGateway()
	}
	_ = trx.Rollback() // ...or will block unshare trx on sqlite

	agentClient, agentConn, err := agentCtrl.NewClient(ae.Token)
	if err != nil {
		dl.Errorf("error creating agent client for '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewRemoteUnshareInternalServerError()
	}
	defer agentConn.Close()

	req := &agentGrpc.ReleaseShareRequest{Token: params.Body.Token}
	_, err = agentClient.ReleaseShare(context.Background(), req)
	if err != nil {
		dl.Errorf("error releasing share '%v' for '%v' (%v): %v", params.Body.Token, params.Body.EnvZID, principal.Email, err)
		return agent.NewRemoteUnshareBadGateway()
	}

	return agent.NewRemoteUnshareOK()
}
