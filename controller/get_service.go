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

	ssvc, err := str.FindServiceWithToken(svcToken, tx)
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
		if senv.Id == ssvc.EnvironmentId && senv.ZId == envZId {
			envFound = true
			break
		}
	}
	if !envFound {
		logrus.Errorf("service '%v' not in environment '%v'", svcToken, envZId)
		return service.NewGetServiceNotFound()
	}

	svc := &rest_model_zrok.Service{
		Token:       ssvc.Token,
		ZID:         ssvc.ZId,
		ShareMode:   ssvc.ShareMode,
		BackendMode: ssvc.BackendMode,
		Reserved:    ssvc.Reserved,
		CreatedAt:   ssvc.CreatedAt.UnixMilli(),
		UpdatedAt:   ssvc.UpdatedAt.UnixMilli(),
	}
	if ssvc.FrontendSelection != nil {
		svc.FrontendSelection = *ssvc.FrontendSelection
	}
	if ssvc.FrontendEndpoint != nil {
		svc.FrontendEndpoint = *ssvc.FrontendEndpoint
	}
	if ssvc.BackendProxyEndpoint != nil {
		svc.BackendProxyEndpoint = *ssvc.BackendProxyEndpoint
	}

	return service.NewGetServiceOK().WithPayload(svc)
}
