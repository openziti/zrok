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

type agentRemoteStatusHandler struct{}

func newAgentRemoteStatusHandler() *agentRemoteStatusHandler {
	return &agentRemoteStatusHandler{}
}

func (h *agentRemoteStatusHandler) Handle(params agent.RemoteStatusParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction for '%v': %v", principal.Email, err)
		return agent.NewRemoteStatusInternalServerError()
	}
	defer trx.Rollback()

	env, err := str.FindEnvironmentForAccount(params.Body.EnvZID, int(principal.ID), trx)
	if err != nil {
		logrus.Errorf("error finding environment '%v' for '%v' (%v)", params.Body.EnvZID, principal.Email, err)
		return agent.NewRemoteStatusUnauthorized()
	}

	ae, err := str.FindAgentEnrollmentForEnvironment(env.Id, trx)
	if err != nil {
		logrus.Errorf("error finding agent enrollment for environment '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewRemoteStatusBadGateway()
	}

	acli, aconn, err := agentController.NewAgentClient(ae.Token, cfg.AgentController)
	if err != nil {
		logrus.Errorf("error creating agent client for '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewRemoteStatusInternalServerError()
	}
	defer aconn.Close()

	resp, err := acli.Status(context.Background(), &agentGrpc.StatusRequest{})
	if err != nil {
		logrus.Errorf("error retrieving remote agent status for '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewRemoteStatusBadGateway()
	}

	out := &agent.RemoteStatusOKBody{}
	for _, share := range resp.Shares {
		out.Shares = append(out.Shares, &agent.RemoteStatusOKBodySharesItems0{
			BackendEndpoint:   share.BackendEndpoint,
			BackendMode:       share.BackendMode,
			FrontendEndpoints: share.FrontendEndpoint,
			Open:              !share.Closed,
			Reserved:          share.Reserved,
			ShareMode:         share.ShareMode,
			Status:            share.Status,
			Token:             share.Token,
		})
	}
	for _, access := range resp.Accesses {
		out.Accesses = append(out.Accesses, &agent.RemoteStatusOKBodyAccessesItems0{
			BindAddress:     access.BindAddress,
			FrontendToken:   access.FrontendToken,
			ResponseHeaders: access.ResponseHeaders,
			Token:           access.Token,
		})
	}

	return agent.NewRemoteStatusOK().WithPayload(out)
}
