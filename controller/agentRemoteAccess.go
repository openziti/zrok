package controller

import (
	"context"

	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/agent"
)

type agentRemoteAccessHandler struct{}

func newAgentRemoteAccessHandler() *agentRemoteAccessHandler {
	return &agentRemoteAccessHandler{}
}

func (h *agentRemoteAccessHandler) Handle(params agent.RemoteAccessParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction for '%v': %v", principal.Email, err)
		return agent.NewRemoteAccessInternalServerError()
	}
	defer trx.Rollback()

	env, err := str.FindEnvironmentForAccount(params.Body.EnvZID, int(principal.ID), trx)
	if err != nil {
		dl.Errorf("error finding environment for '%v' (%v): %v", params.Body.EnvZID, principal.ID, err)
		return agent.NewRemoteAccessUnauthorized()
	}

	ae, err := str.FindAgentEnrollmentForEnvironment(env.Id, trx)
	if err != nil {
		dl.Errorf("error finding agent enrollment for environment '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewRemoteAccessBadGateway()
	}
	_ = trx.Rollback() // ...or will block the access trx on sqlite

	agentClient, agentConn, err := agentCtrl.NewClient(ae.Token)
	if err != nil {
		dl.Errorf("error creating agent client for '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewRemoteAccessInternalServerError()
	}
	defer agentConn.Close()

	req := &agentGrpc.AccessPrivateRequest{
		Token:           params.Body.Token,
		BindAddress:     params.Body.BindAddress,
		AutoMode:        params.Body.AutoMode,
		AutoAddress:     params.Body.AutoAddress,
		AutoStartPort:   uint32(params.Body.AutoStartPort),
		AutoEndPort:     uint32(params.Body.AutoEndPort),
		ResponseHeaders: params.Body.ResponseHeaders,
	}
	resp, err := agentClient.AccessPrivate(context.Background(), req)
	if err != nil {
		dl.Errorf("error creating remote agent private access for '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewRemoteAccessBadGateway()
	}

	return agent.NewRemoteAccessOK().WithPayload(&agent.RemoteAccessOKBody{FrontendToken: resp.FrontendToken})
}
