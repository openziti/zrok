package controller

import (
	"context"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/agent"
)

type agentRemoteStatusHandler struct{}

func newAgentRemoteStatusHandler() *agentRemoteStatusHandler {
	return &agentRemoteStatusHandler{}
}

func (h *agentRemoteStatusHandler) Handle(params agent.RemoteStatusParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction for '%v': %v", principal.Email, err)
		return agent.NewRemoteStatusInternalServerError()
	}
	defer trx.Rollback()

	env, err := str.FindEnvironmentForAccount(params.Body.EnvZID, int(principal.ID), trx)
	if err != nil {
		dl.Errorf("error finding environment '%v' for '%v' (%v)", params.Body.EnvZID, principal.Email, err)
		return agent.NewRemoteStatusUnauthorized()
	}

	ae, err := str.FindAgentEnrollmentForEnvironment(env.Id, trx)
	if err != nil {
		dl.Errorf("error finding agent enrollment for environment '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewRemoteStatusBadGateway()
	}

	agentClient, agentConn, err := agentCtrl.NewClient(ae.Token)
	if err != nil {
		dl.Errorf("error creating agent client for '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewRemoteStatusInternalServerError()
	}
	defer agentConn.Close()

	resp, err := agentClient.Status(context.Background(), &agentGrpc.StatusRequest{})
	if err != nil {
		dl.Errorf("error retrieving remote agent status for '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewRemoteStatusBadGateway()
	}

	out := &agent.RemoteStatusOKBody{}
	for _, share := range resp.Shares {
		shareItem := &agent.RemoteStatusOKBodySharesItems0{
			BackendEndpoint:   share.BackendEndpoint,
			BackendMode:       share.BackendMode,
			FrontendEndpoints: share.FrontendEndpoint,
			Open:              !share.Closed,
			ShareMode:         share.ShareMode,
			Status:            share.Status,
			Token:             share.Token,
		}
		if share.Failure != nil {
			shareItem.Failure = &agent.RemoteStatusOKBodySharesItems0Failure{
				ID:        share.Failure.Id,
				Count:     int64(share.Failure.Count),
				LastError: share.Failure.LastError,
			}
			if share.Failure.NextRetry != nil {
				shareItem.Failure.NextRetry = share.Failure.NextRetry.AsTime().Format(time.RFC3339)
			}
		}
		out.Shares = append(out.Shares, shareItem)
	}
	for _, access := range resp.Accesses {
		accessItem := &agent.RemoteStatusOKBodyAccessesItems0{
			BindAddress:     access.BindAddress,
			FrontendToken:   access.FrontendToken,
			ResponseHeaders: access.ResponseHeaders,
			Token:           access.Token,
			Status:          access.Status,
		}
		if access.Failure != nil {
			accessItem.Failure = &agent.RemoteStatusOKBodyAccessesItems0Failure{
				ID:        access.Failure.Id,
				Count:     int64(access.Failure.Count),
				LastError: access.Failure.LastError,
			}
			if access.Failure.NextRetry != nil {
				accessItem.Failure.NextRetry = access.Failure.NextRetry.AsTime().Format(time.RFC3339)
			}
		}
		out.Accesses = append(out.Accesses, accessItem)
	}

	return agent.NewRemoteStatusOK().WithPayload(out)
}
