package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/controller/store"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/service"
	"github.com/sirupsen/logrus"
)

type shareHandler struct{}

func newShareHandler() *shareHandler {
	return &shareHandler{}
}

func (h *shareHandler) Handle(params service.ShareParams, principal *rest_model_zrok.Principal) middleware.Responder {
	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return service.NewShareInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()

	envZId := params.Body.EnvZID
	envId := 0
	if envs, err := str.FindEnvironmentsForAccount(int(principal.ID), tx); err == nil {
		found := false
		for _, env := range envs {
			if env.ZId == envZId {
				logrus.Debugf("found identity '%v' for user '%v'", envZId, principal.Email)
				envId = env.Id
				found = true
				break
			}
		}
		if !found {
			logrus.Errorf("environment '%v' not found for user '%v'", envZId, principal.Email)
			return service.NewShareUnauthorized()
		}
	} else {
		logrus.Errorf("error finding environments for account '%v'", principal.Email)
		return service.NewShareInternalServerError()
	}

	edge, err := edgeClient()
	if err != nil {
		logrus.Error(err)
		return service.NewShareInternalServerError()
	}
	svcToken, err := createServiceToken()
	if err != nil {
		logrus.Error(err)
		return service.NewShareInternalServerError()
	}

	var svcZId string
	var frontendEndpoints []string
	switch params.Body.ShareMode {
	case "public":
		svcZId, frontendEndpoints, err = newPublicResourceAllocator().allocate(envZId, svcToken, params, edge)
		if err != nil {
			logrus.Error(err)
			return service.NewShareInternalServerError()
		}

	case "private":
		svcZId, frontendEndpoints, err = newPrivateResourceAllocator().allocate(envZId, svcToken, params, edge)
		if err != nil {
			logrus.Error(err)
			return service.NewShareInternalServerError()
		}

	default:
		logrus.Errorf("unknown share mode '%v", params.Body.ShareMode)
		return service.NewShareInternalServerError()
	}

	logrus.Debugf("allocated service '%v'", svcToken)

	reserved := params.Body.Reserved
	sid, err := str.CreateService(envId, &store.Service{
		ZId:                  svcZId,
		Token:                svcToken,
		ShareMode:            params.Body.ShareMode,
		BackendMode:          params.Body.BackendMode,
		FrontendEndpoint:     &frontendEndpoints[0],
		BackendProxyEndpoint: &params.Body.BackendProxyEndpoint,
		Reserved:             reserved,
	}, tx)
	if err != nil {
		logrus.Errorf("error creating service record: %v", err)
		return service.NewShareInternalServerError()
	}

	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing service record: %v", err)
		return service.NewShareInternalServerError()
	}
	logrus.Infof("recorded service '%v' with id '%v' for '%v'", svcToken, sid, principal.Email)

	return service.NewShareCreated().WithPayload(&rest_model_zrok.ShareResponse{
		FrontendProxyEndpoint: frontendEndpoints[0],
		SvcToken:              svcToken,
	})
}
