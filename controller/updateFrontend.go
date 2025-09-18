package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
	"github.com/sirupsen/logrus"
)

type updateFrontendHandler struct{}

func newUpdateFrontendHandler() *updateFrontendHandler {
	return &updateFrontendHandler{}
}

func (h *updateFrontendHandler) Handle(params admin.UpdateFrontendParams, principal *rest_model_zrok.Principal) middleware.Responder {
	feToken := params.Body.FrontendToken
	publicName := params.Body.PublicName
	urlTemplate := params.Body.URLTemplate

	if !principal.Admin {
		logrus.Errorf("invalid admin principal")
		return admin.NewUpdateFrontendUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return admin.NewUpdateFrontendInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	fe, err := str.FindFrontendWithToken(feToken, trx)
	if err != nil {
		logrus.Errorf("error finding frontend with token '%v': %v", feToken, err)
		return admin.NewUpdateFrontendNotFound()
	}

	doUpdate := false
	if publicName != "" {
		if fe.PublicName == nil || (fe.PublicName != nil && *fe.PublicName != publicName) {
			fe.PublicName = &publicName
			doUpdate = true
		}
	}
	if urlTemplate != "" {
		if fe.UrlTemplate == nil || (fe.UrlTemplate != nil && *fe.UrlTemplate != urlTemplate) {
			fe.UrlTemplate = &urlTemplate
			doUpdate = true
		}
	}

	if doUpdate {
		if err := str.UpdateFrontend(fe, trx); err != nil {
			logrus.Errorf("error updating frontend: %v", err)
			return admin.NewUpdateFrontendInternalServerError()
		}

		if err := trx.Commit(); err != nil {
			logrus.Errorf("error committing frontend update: %v", err)
			return admin.NewUpdateFrontendInternalServerError()
		}
	}

	return admin.NewUpdateFrontendOK()
}
