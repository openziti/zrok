package controller

import (
	"context"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/controller/store"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/service"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/edge/rest_management_api_client/service_policy"
	"github.com/openziti/edge/rest_model"
	rest_model_edge "github.com/openziti/edge/rest_model"
	"github.com/sirupsen/logrus"
	"time"
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
		return service.NewAccessNotFound()
	}

	svcName := params.Body.SvcName
	ssvc, err := str.FindServiceWithToken(svcName, tx)
	if err != nil {
		logrus.Errorf("error finding service")
		return service.NewAccessNotFound()
	}
	if ssvc == nil {
		logrus.Errorf("unable to find service '%v' for user '%v'", params.Body.SvcName, principal.Email)
		return service.NewAccessNotFound()
	}

	feToken, err := createToken()
	if err != nil {
		logrus.Error(err)
		return service.NewAccessInternalServerError()
	}

	if _, err := str.CreateFrontend(envId, &store.Frontend{Token: feToken, ZId: envZId}, tx); err != nil {
		logrus.Errorf("error creating frontend record: %v", err)
		return service.NewAccessInternalServerError()
	}

	edge, err := edgeClient()
	if err != nil {
		logrus.Error(err)
		return service.NewAccessInternalServerError()
	}
	extraTags := &rest_model_edge.Tags{SubTags: map[string]interface{}{
		"zrokEnvironmentZId": envZId,
		"zrokFrontendToken":  feToken,
	}}
	if err := createServicePolicyDialForEnvironment(envZId, ssvc.Token, ssvc.ZId, edge, extraTags); err != nil {
		logrus.Errorf("unable to create dial policy: %v", err)
		return service.NewAccessInternalServerError()
	}

	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing frontend record: %v", err)
		return service.NewAccessInternalServerError()
	}

	return service.NewAccessCreated().WithPayload(&rest_model_zrok.AccessResponse{FrontendName: feToken})
}

func createServicePolicyDialForEnvironment(envZId, svcToken, svcZId string, edge *rest_management_api_client.ZitiEdgeManagement, tags ...*rest_model.Tags) error {
	allTags := zrokTags(svcToken)
	for _, t := range tags {
		for k, v := range t.SubTags {
			allTags.SubTags[k] = v
		}
	}

	identityRoles := []string{"@" + envZId}
	name := fmt.Sprintf("%v-%v-dial", envZId, svcToken)
	var postureCheckRoles []string
	semantic := rest_model.SemanticAllOf
	serviceRoles := []string{fmt.Sprintf("@%v", svcZId)}
	dialBind := rest_model.DialBindDial
	svcp := &rest_model.ServicePolicyCreate{
		IdentityRoles:     identityRoles,
		Name:              &name,
		PostureCheckRoles: postureCheckRoles,
		Semantic:          &semantic,
		ServiceRoles:      serviceRoles,
		Type:              &dialBind,
		Tags:              allTags,
	}
	req := &service_policy.CreateServicePolicyParams{
		Policy:  svcp,
		Context: context.Background(),
	}
	req.SetTimeout(30 * time.Second)
	resp, err := edge.ServicePolicy.CreateServicePolicy(req, nil)
	if err != nil {
		return err
	}
	logrus.Infof("created dial service policy '%v' for service '%v' for environment '%v'", resp.Payload.Data.ID, svcZId, envZId)
	return nil
}
