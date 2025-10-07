package controller

import (
	"context"

	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/agent"
	"github.com/sirupsen/logrus"
)

type agentRemoteShareHandler struct{}

func newAgentRemoteShareHandler() *agentRemoteShareHandler {
	return &agentRemoteShareHandler{}
}

func (h *agentRemoteShareHandler) Handle(params agent.RemoteShareParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction for '%v': %v", principal.Email, err)
		return agent.NewRemoteShareInternalServerError()
	}
	defer trx.Rollback()

	env, err := str.FindEnvironmentForAccount(params.Body.EnvZID, int(principal.ID), trx)
	if err != nil {
		logrus.Errorf("error finding environment '%v' for '%v': %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewRemoteShareUnauthorized()
	}
	logrus.Infof("found environment '%v' for '%v'", params.Body.EnvZID, principal.Email)

	ae, err := str.FindAgentEnrollmentForEnvironment(env.Id, trx)
	if err != nil {
		logrus.Errorf("error finding agent enrollment for environment '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewRemoteShareBadGateway()
	}
	logrus.Infof("found agent enrollment '%v' for environment '%v' for '%v'", ae.Token, params.Body.EnvZID, principal.Email)

	agentClient, agentConn, err := agentCtrl.NewClient(ae.Token)
	if err != nil {
		logrus.Errorf("error creating agent client for '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewRemoteShareInternalServerError()
	}
	defer agentConn.Close()
	logrus.Infof("created agentCtrl client for environment '%v' for '%v'", params.Body.EnvZID, principal.Email)

	out := &agent.RemoteShareOKBody{}
	switch params.Body.ShareMode {
	case "public":
		token, frontendEndpoints, err := h.publicShare(params, agentClient)
		if err != nil {
			logrus.Errorf("error creating public remote agent share for '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
			return agent.NewRemoteShareBadGateway()
		}
		out.Token = token
		out.FrontendEndpoints = frontendEndpoints

	case "private":
		token, err := h.privateShare(params, agentClient)
		if err != nil {
			logrus.Errorf("error creating private remote agent share for '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
			return agent.NewRemoteShareBadGateway()
		}
		out.Token = token

	case "reserved":
		token, err := h.reservedShare(params, agentClient)
		if err != nil {
			logrus.Errorf("error creating reserved remote agent share for '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
			return agent.NewRemoteShareBadGateway()
		}
		out.Token = token
	}

	return agent.NewRemoteShareOK().WithPayload(out)
}

func (h *agentRemoteShareHandler) publicShare(params agent.RemoteShareParams, client agentGrpc.AgentClient) (token string, frontendEndpoints []string, err error) {
	req := &agentGrpc.SharePublicRequest{
		Target:                    params.Body.Target,
		BasicAuth:                 params.Body.BasicAuth,
		FrontendSelection:         params.Body.FrontendSelection,
		BackendMode:               params.Body.BackendMode,
		Insecure:                  params.Body.Insecure,
		OauthProvider:             params.Body.OauthProvider,
		OauthEmailAddressPatterns: params.Body.OauthEmailAddressPatterns,
		OauthCheckInterval:        params.Body.OauthCheckInterval,
		Closed:                    !params.Body.Open,
		AccessGrants:              params.Body.AccessGrants,
	}
	resp, err := client.SharePublic(context.Background(), req)
	if err != nil {
		return "", nil, err
	}
	logrus.Infof("got token '%v'", resp.Token)
	return resp.Token, resp.FrontendEndpoints, nil
}

func (h *agentRemoteShareHandler) privateShare(params agent.RemoteShareParams, client agentGrpc.AgentClient) (token string, err error) {
	req := &agentGrpc.SharePrivateRequest{
		Target:       params.Body.Target,
		BackendMode:  params.Body.BackendMode,
		Insecure:     params.Body.Insecure,
		Closed:       !params.Body.Open,
		AccessGrants: params.Body.AccessGrants,
	}
	resp, err := client.SharePrivate(context.Background(), req)
	if err != nil {
		return "", err
	}
	logrus.Infof("got token '%v'", resp.Token)
	return resp.Token, nil
}

func (h *agentRemoteShareHandler) reservedShare(params agent.RemoteShareParams, client agentGrpc.AgentClient) (token string, err error) {
	logrus.Infof("requesting reserved share '%v'", params.Body.Token)
	req := &agentGrpc.ShareReservedRequest{
		Token:            params.Body.Token,
		OverrideEndpoint: params.Body.Target,
		Insecure:         params.Body.Insecure,
	}
	resp, err := client.ShareReserved(context.Background(), req)
	if err != nil {
		logrus.Errorf("reserved share failed for '%v': %v", params.Body.Token, err)
		return "", err
	}
	logrus.Infof("got token '%v'", resp.Token)
	return resp.Token, nil
}
