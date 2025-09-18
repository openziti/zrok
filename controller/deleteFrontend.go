package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
	"github.com/sirupsen/logrus"
)

type deleteFrontendHandler struct{}

func newDeleteFrontendHandler() *deleteFrontendHandler {
	return &deleteFrontendHandler{}
}

func (h *deleteFrontendHandler) Handle(params admin.DeleteFrontendParams, principal *rest_model_zrok.Principal) middleware.Responder {
	feToken := params.Body.FrontendToken

	if !principal.Admin {
		logrus.Errorf("invalid admin principal")
		return admin.NewDeleteFrontendUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return admin.NewDeleteFrontendInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	fe, err := str.FindFrontendWithToken(feToken, trx)
	if err != nil {
		logrus.Errorf("error finding frontend with token '%v': %v", feToken, err)
		return admin.NewDeleteFrontendNotFound()
	}

	if err := str.DeleteFrontend(fe.Id, trx); err != nil {
		logrus.Errorf("error deleting frontend '%v': %v", feToken, err)
		return admin.NewDeleteFrontendInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		logrus.Errorf("error committing frontend '%v' deletion: %v", feToken, err)
		return admin.NewDeleteFrontendInternalServerError()
	}

	return admin.NewDeleteFrontendOK()

}
