package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/controller/store"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/service"
	rest_model_edge "github.com/openziti/edge/rest_model"
	"github.com/sirupsen/logrus"
)

type accessHandler struct{}

func newAccessHandler() *accessHandler {
	return &accessHandler{}
}

func (h *accessHandler) Handle(params service.AccessParams, principal *rest_model_zrok.Principal) middleware.Responder {
	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return service.NewAccessInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()

	envZId := params.Body.ZID
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
			return service.NewAccessUnauthorized()
		}
	} else {
		logrus.Errorf("error finding environments for account '%v'", principal.Email)
		return service.NewAccessInternalServerError()
	}

	ssvcs, err := str.FindServicesForEnvironment(envId, tx)
	if err != nil {
		logrus.Errorf("error finding services for environment")
		return service.NewAccessInternalServerError()
	}
	var ssvc *store.Service
	for _, v := range ssvcs {
		if v.Name == params.Body.SvcName {
			ssvc = v
			break
		}
	}
	if ssvc == nil {
		logrus.Errorf("unable to find service '%v' for user '%v'", params.Body.SvcName, principal.Email)
		return service.NewAccessNotFound()
	}

	edge, err := edgeClient()
	if err != nil {
		logrus.Error(err)
		return service.NewAccessInternalServerError()
	}

	extraTags := &rest_model_edge.Tags{SubTags: map[string]interface{}{"zrokEnvironmentZId": envZId}}
	if err := createServicePolicyDial(envZId, ssvc.Name, ssvc.ZId, edge, extraTags); err != nil {
		logrus.Errorf("unable to create dial policy: %v", err)
		return service.NewAccessInternalServerError()
	}

	return service.NewAccessCreated()
}
