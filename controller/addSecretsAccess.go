package controller

import (
	"fmt"

	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
	"github.com/sirupsen/logrus"
)

type addSecretsAccessHandler struct{}

func newAddSecretsAccessHandler() *addSecretsAccessHandler {
	return &addSecretsAccessHandler{}
}

func (h *addSecretsAccessHandler) Handle(params admin.AddSecretsAccessParams, principal *rest_model_zrok.Principal) middleware.Responder {
	secretsAccessIdentityZId := params.Body.SecretsAccessIdentityZID

	if !principal.Admin {
		logrus.Errorf("invalid admin principal")
		return admin.NewAddSecretsAccessUnauthorized()
	}

	edge, err := zrokEdgeSdk.Client(cfg.Ziti)
	if err != nil {
		logrus.Errorf("error getting edge client: %v", err)
		return admin.NewAddSecretsAccessInternalServerError()
	}

	serviceZId, err := getZIdForService(cfg.Secrets.ServiceName, edge)
	if err != nil {
		logrus.Errorf("error getting service ziti id for '%v': %v", cfg.Secrets.ServiceName, err)
		return admin.NewAddSecretsAccessInternalServerError()
	}

	spZId, err := getZIdForServicePolicy(serviceZId, secretsAccessIdentityZId, rest_model.DialBindDial, edge)
	if err != nil {
		logrus.Infof("could not assert service policy; creating")

		if err := zrokEdgeSdk.CreateServicePolicyDial(fmt.Sprintf("service-listener-dial-%v", secretsAccessIdentityZId), serviceZId, []string{secretsAccessIdentityZId}, nil, edge); err != nil {
			logrus.Errorf("error creating dial service policy for '@%v' -> '@%v': %v", secretsAccessIdentityZId, serviceZId, err)
			return admin.NewAddSecretsAccessInternalServerError()
		}
		logrus.Infof("created dial service policy for '@%v' -> '@%v'", secretsAccessIdentityZId, serviceZId)

	} else {
		logrus.Errorf("asserted existing service policy with ziti id '%v'", spZId)
		return admin.NewAddSecretsAccessBadRequest()
	}

	return admin.NewAddSecretsAccessOK()
}
