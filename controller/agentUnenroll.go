package controller

import (
	"context"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/edge-api/rest_management_api_client"
	edge_service "github.com/openziti/edge-api/rest_management_api_client/service"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/agent"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

type agentUnenrollHandler struct{}

func newAgentUnenrollHandler() *agentUnenrollHandler {
	return &agentUnenrollHandler{}
}

func (h *agentUnenrollHandler) Handle(params agent.UnenrollParams, principal *rest_model_zrok.Principal) middleware.Responder {
	// start transaction early, if it fails, don't bother creating ziti resources
	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction for '%v': %v", principal.Email, err)
		return agent.NewUnenrollInternalServerError()
	}
	defer trx.Rollback()

	env, err := str.FindEnvironmentForAccount(params.Body.EnvZID, int(principal.ID), trx)
	if err != nil {
		logrus.Errorf("error finding environment '%v' for '%v': %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewUnenrollUnauthorized()
	}

	ae, err := str.FindAgentEnrollmentForEnvironment(env.Id, trx)
	if err != nil {
		logrus.Errorf("error finding agent enrollment for '%v' (%v): %v", params.Body.EnvZID, principal.Email, err)
		return agent.NewUnenrollBadRequest()
	}

	client, err := zrokEdgeSdk.Client(cfg.Ziti)
	if err != nil {
		logrus.Errorf("error getting ziti client for '%v': %v", principal.Email, err)
		return agent.NewUnenrollInternalServerError()
	}

	if err := zrokEdgeSdk.DeleteServiceEdgeRouterPolicyForAgentRemote(env.ZId, ae.Token, client); err != nil {
		logrus.Errorf("error removing agent remote serp for '%v' (%v): %v", env.ZId, principal.Email, err)
		return agent.NewUnenrollInternalServerError()
	}

	if err := zrokEdgeSdk.DeleteServicePoliciesDialForAgentRemote(env.ZId, ae.Token, client); err != nil {
		logrus.Errorf("error removing agent remote dial service policy for '%v' (%v): %v", env.ZId, principal.Email, err)
		return agent.NewUnenrollInternalServerError()
	}

	if err := zrokEdgeSdk.DeleteServicePoliciesBindForAgentRemote(env.ZId, ae.Token, client); err != nil {
		logrus.Errorf("error removing agent remote bind service policy for '%v' (%v): %v", env.ZId, principal.Email, err)
		return agent.NewUnenrollInternalServerError()
	}

	aeZId, err := h.findAgentRemoteZId(ae.Token, client)
	if err != nil {
		logrus.Errorf("error finding zId for agent remote service for '%v' (%v): %v", env.ZId, principal.Email, err)
		return agent.NewUnenrollInternalServerError()
	}

	if err := zrokEdgeSdk.DeleteService(env.ZId, aeZId, client); err != nil {
		logrus.Errorf("error removing agent remote service for '%v' (%v): %v", env.ZId, principal.Email, err)
		return agent.NewUnenrollInternalServerError()
	}

	if err := str.DeleteAgentEnrollment(ae.Id, trx); err != nil {
		logrus.Errorf("error deleting agent enrollment for '%v' (%v): %v", env.ZId, principal.Email, err)
		return agent.NewUnenrollInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		logrus.Errorf("error committing agent unenrollment for '%v' (%v): %v", env.ZId, principal.Email, err)
		return agent.NewUnenrollInternalServerError()
	}

	return agent.NewUnenrollOK()
}

func (h *agentUnenrollHandler) findAgentRemoteZId(enrollmentToken string, edge *rest_management_api_client.ZitiEdgeManagement) (string, error) {
	filter := fmt.Sprintf("name=\"%v\"", enrollmentToken)
	limit := int64(1)
	offset := int64(0)
	listReq := &edge_service.ListServicesParams{
		Filter:  &filter,
		Limit:   &limit,
		Offset:  &offset,
		Context: context.Background(),
	}
	listReq.SetTimeout(30 * time.Second)
	listResp, err := edge.Service.ListServices(listReq, nil)
	if err != nil {
		return "", err
	}
	if len(listResp.Payload.Data) == 1 {
		return *(listResp.Payload.Data[0].ID), nil
	}
	return "", errors.Errorf("agent remote service '%v' not found", enrollmentToken)
}
