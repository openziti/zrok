package controller

import (
	"context"
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/controller/agentController"
	"github.com/openziti/zrok/controller/config"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/agent"
	"github.com/sirupsen/logrus"
)

type agentPingHandler struct {
	cfg *config.Config
}

func newAgentPingHandler(cfg *config.Config) *agentPingHandler {
	return &agentPingHandler{cfg: cfg}
}

func (h *agentPingHandler) Handle(params agent.PingParams, principal *rest_model_zrok.Principal) middleware.Responder {
	acli, aconn, err := agentController.NewAgentClient(params.Body.EnvZID, h.cfg.AgentController)
	if err != nil {
		logrus.Errorf("error creating agent client for '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewPingInternalServerError()
	}
	defer aconn.Close()

	resp, err := acli.Version(context.Background(), &agentGrpc.VersionRequest{})
	if err != nil {
		logrus.Errorf("error retrieving agent version for '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewPingBadGateway()
	}

	return agent.NewPingOK().WithPayload(&agent.PingOKBody{Version: resp.V})
}
