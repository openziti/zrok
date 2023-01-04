package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/service"
	"github.com/sirupsen/logrus"
)

func getServiceHandler(params service.GetServiceParams, principal *rest_model_zrok.Principal) middleware.Responder {
	envZId := params.Body.EnvZID
	svcToken := params.Body.SvcToken

	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return service.NewGetServiceInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()

	sshr, err := str.FindShareWithToken(svcToken, tx)
	if err != nil {
		logrus.Errorf("error finding service with token '%v': %v", svcToken, err)
		return service.NewGetServiceNotFound()
	}
	senvs, err := str.FindEnvironmentsForAccount(int(principal.ID), tx)
	if err != nil {
		logrus.Errorf("error listing environments for account '%v': %v", principal.Email, err)
		return service.NewGetServiceInternalServerError()
	}
	envFound := false
	for _, senv := range senvs {
		if senv.Id == sshr.EnvironmentId && senv.ZId == envZId {
			envFound = true
			break
		}
	}
	if !envFound {
		logrus.Errorf("service '%v' not in environment '%v'", svcToken, envZId)
		return service.NewGetServiceNotFound()
	}

	shr := &rest_model_zrok.Service{
		Token:       sshr.Token,
		ZID:         sshr.ZId,
		ShareMode:   sshr.ShareMode,
		BackendMode: sshr.BackendMode,
		Reserved:    sshr.Reserved,
		CreatedAt:   sshr.CreatedAt.UnixMilli(),
		UpdatedAt:   sshr.UpdatedAt.UnixMilli(),
	}
	if sshr.FrontendSelection != nil {
		shr.FrontendSelection = *sshr.FrontendSelection
	}
	if sshr.FrontendEndpoint != nil {
		shr.FrontendEndpoint = *sshr.FrontendEndpoint
	}
	if sshr.BackendProxyEndpoint != nil {
		shr.BackendProxyEndpoint = *sshr.BackendProxyEndpoint
	}

	return service.NewGetServiceOK().WithPayload(shr)
}
