package controller

import (
	"github.com/openziti-test-kitchen/zrok/controller/store"
	"github.com/openziti/edge/rest_management_api_client"
)

type unsharePublicHandler struct{}

func newUnsharePublicHandler() *unsharePublicHandler {
	return &unsharePublicHandler{}
}

func (h *unsharePublicHandler) Handle(senv *store.Environment, ssvc *store.Service, svcName, svcZId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	if err := deleteServiceEdgeRouterPolicy(senv.ZId, svcName, edge); err != nil {
		return err
	}
	if err := deleteServicePolicyDial(senv.ZId, svcName, edge); err != nil {
		return err
	}
	if err := deleteServicePolicyBind(senv.ZId, svcName, edge); err != nil {
		return err
	}
	if err := deleteConfig(senv.ZId, svcName, edge); err != nil {
		return err
	}
	if err := deleteService(senv.ZId, svcZId, edge); err != nil {
		return err
	}
	return nil
}
