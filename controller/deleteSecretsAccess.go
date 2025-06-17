package controller

import (
	"context"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/edge-api/rest_management_api_client/service_policy"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
	"github.com/sirupsen/logrus"
)

type deleteSecretsAccessHandler struct{}

func newDeleteSecretsAccessHandler() *deleteSecretsAccessHandler {
	return &deleteSecretsAccessHandler{}
}

func (h *deleteSecretsAccessHandler) Handle(params admin.DeleteSecretsAccessParams, principal *rest_model_zrok.Principal) middleware.Responder {
	secretsAccessIdentityZId := params.Body.SecretsIdentityZID

	if !principal.Admin {
		logrus.Errorf("invalid admin principal")
		return admin.NewDeleteSecretsAccessUnauthorized()
	}

	edge, err := zrokEdgeSdk.Client(cfg.Ziti)
	if err != nil {
		logrus.Errorf("error getting edge client: %v", err)
		return admin.NewDeleteSecretsAccessInternalServerError()
	}

	serviceZId, err := getZIdForService(cfg.Secrets.ServiceName, edge)
	if err != nil {
		logrus.Errorf("error getting service ziti id for '%v': %v", cfg.Secrets.ServiceName, err)
		return admin.NewDeleteSecretsAccessInternalServerError()
	}

	spZId, err := getZIdForServicePolicy(serviceZId, secretsAccessIdentityZId, rest_model.DialBindDial, edge)
	if err == nil {
		req := &service_policy.DeleteServicePolicyParams{
			ID:      spZId,
			Context: context.Background(),
		}
		req.SetTimeout(30 * time.Second)
		_, err := edge.ServicePolicy.DeleteServicePolicy(req, nil)
		if err != nil {
			logrus.Errorf("error deleting service policy '%v': %v", spZId, err)
			return admin.NewDeleteSecretsAccessInternalServerError()
		}
		logrus.Infof("removed dial service policy for '@%v' -> '@%v", secretsAccessIdentityZId, serviceZId)

	} else {
		logrus.Errorf("error getting dial service policy ziti id: %v", err)
		return admin.NewDeleteSecretsAccessBadRequest()
	}

	return admin.NewDeleteSecretsAccessOK()
}
