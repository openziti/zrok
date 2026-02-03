package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/admin"
)

type deleteFrontendHandler struct{}

func newDeleteFrontendHandler() *deleteFrontendHandler {
	return &deleteFrontendHandler{}
}

func (h *deleteFrontendHandler) Handle(params admin.DeleteFrontendParams, principal *rest_model_zrok.Principal) middleware.Responder {
	feToken := params.Body.FrontendToken

	if !principal.Admin {
		dl.Errorf("invalid admin principal")
		return admin.NewDeleteFrontendUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return admin.NewDeleteFrontendInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	fe, err := str.FindFrontendWithToken(feToken, trx)
	if err != nil {
		dl.Errorf("error finding frontend with token '%v': %v", feToken, err)
		return admin.NewDeleteFrontendNotFound()
	}

	if err := str.DeleteFrontend(fe.Id, trx); err != nil {
		dl.Errorf("error deleting frontend '%v': %v", feToken, err)
		return admin.NewDeleteFrontendInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing frontend '%v' deletion: %v", feToken, err)
		return admin.NewDeleteFrontendInternalServerError()
	}

	return admin.NewDeleteFrontendOK()

}
