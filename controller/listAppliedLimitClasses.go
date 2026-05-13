package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/admin"
)

type listAppliedLimitClassesHandler struct{}

func newListAppliedLimitClassesHandler() *listAppliedLimitClassesHandler {
	return &listAppliedLimitClassesHandler{}
}

func (h *listAppliedLimitClassesHandler) Handle(params admin.ListAppliedLimitClassesParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		dl.Error("invalid admin principal")
		return admin.NewListAppliedLimitClassesUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return admin.NewListAppliedLimitClassesInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	acct, err := str.FindAccountWithEmail(params.Body.Email, trx)
	if err != nil {
		dl.Errorf("error finding account with email '%v': %v", params.Body.Email, err)
		return admin.NewListAppliedLimitClassesNotFound()
	}

	lcs, err := str.FindAppliedLimitClassesForAccount(acct.Id, trx)
	if err != nil {
		dl.Errorf("error finding applied limit classes for '%v': %v", params.Body.Email, err)
		return admin.NewListAppliedLimitClassesInternalServerError()
	}

	return admin.NewListAppliedLimitClassesOK().WithPayload(limitClassesToApi(lcs))
}
