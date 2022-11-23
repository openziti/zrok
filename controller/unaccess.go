package controller

import (
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/controller/store"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/service"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type unaccessHandler struct{}

func newUnaccessHandler() *unaccessHandler {
	return &unaccessHandler{}
}

func (h *unaccessHandler) Handle(params service.UnaccessParams, principal *rest_model_zrok.Principal) middleware.Responder {
	svcName := params.Body.SvcName
	envZId := params.Body.ZID

	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return service.NewUnaccessInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()

	edge, err := edgeClient()
	if err != nil {
		logrus.Error(err)
		return service.NewUnaccessInternalServerError()
	}

	var senv *store.Environment
	if envs, err := str.FindEnvironmentsForAccount(int(principal.ID), tx); err == nil {
		for _, env := range envs {
			if env.ZId == params.Body.ZID {
				senv = env
				break
			}
		}
		if senv == nil {
			err := errors.Errorf("environment with id '%v' not found for '%v", params.Body.ZID, principal.Email)
			logrus.Error(err)
			return service.NewUnaccessUnauthorized()
		}
	} else {
		logrus.Errorf("error finding environments for account '%v': %v", principal.Email, err)
		return service.NewUnaccessUnauthorized()
	}

	if err := deleteServicePolicy(envZId, fmt.Sprintf("tags.zrokServiceName=\"%v\" and tags.zrokEnvironmentZId=\"%v\" and type=1", svcName, envZId), edge); err != nil {
		logrus.Errorf("error removing access to '%v' for '%v': %v", svcName, envZId, err)
		return service.NewUnaccessInternalServerError()
	}

	return service.NewUnaccessOK()
}
