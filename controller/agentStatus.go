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

type agentStatusHandler struct {
	cfg *config.Config
}

func newAgentStatusHandler(cfg *config.Config) *agentStatusHandler {
	return &agentStatusHandler{cfg: cfg}
}

func (h *agentStatusHandler) Handle(params agent.AgentStatusParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if h.cfg.AgentController != nil {
		acli, aconn, err := agentController.NewAgentClient(params.Body.EnvZID, h.cfg.AgentController)
		if err != nil {
			logrus.Errorf("error creating agent client for '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
			return agent.NewAgentStatusInternalServerError()
		}
		defer aconn.Close()

		resp, err := acli.Version(context.Background(), &agentGrpc.VersionRequest{})
		if err != nil {
			logrus.Errorf("error retrieving agent version for '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
			return agent.NewAgentStatusInternalServerError()
		}

		return agent.NewAgentStatusOK().WithPayload(&agent.AgentStatusOKBody{Version: resp.V})
	}
	return agent.NewAgentStatusUnauthorized()
}
