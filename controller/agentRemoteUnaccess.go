package controller

import (
	"context"

	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/agent"
	"github.com/sirupsen/logrus"
)

type agentRemoteUnaccessHandler struct{}

func newAgentRemoteUnaccessHandler() *agentRemoteUnaccessHandler {
	return &agentRemoteUnaccessHandler{}
}

func (h *agentRemoteUnaccessHandler) Handle(params agent.RemoteUnaccessParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction for '%v': %v", principal.Email, err)
		return agent.NewRemoteUnshareInternalServerError()
	}
	defer trx.Rollback()

	env, err := str.FindEnvironmentForAccount(params.Body.EnvZID, int(principal.ID), trx)
	if err != nil {
		logrus.Errorf("error finding environment '%v' for '%v': %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewRemoteUnshareUnauthorized()
	}

	ae, err := str.FindAgentEnrollmentForEnvironment(env.Id, trx)
	if err != nil {
		logrus.Errorf("error finding agent enrollment for environment '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewRemoteUnshareBadGateway()
	}
	_ = trx.Rollback() // ...or will block unshare trx on sqlite

	agentClient, agentConn, err := agentCtrl.NewClient(ae.Token)
	if err != nil {
		logrus.Errorf("error creating agent client for '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewRemoteUnshareInternalServerError()
	}
	defer agentConn.Close()

	req := &agentGrpc.ReleaseAccessRequest{FrontendToken: params.Body.FrontendToken}
	_, err = agentClient.ReleaseAccess(context.Background(), req)
	if err != nil {
		logrus.Errorf("error releasing access '%v' for '%v' (%v): %v", params.Body.FrontendToken, params.Body.EnvZID, principal.Email, err)
		return agent.NewRemoteUnaccessBadGateway()
	}

	return agent.NewRemoteUnaccessOK()
}
