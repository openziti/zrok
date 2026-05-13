package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/admin"
)

type removeAppliedLimitClassesHandler struct{}

func newRemoveAppliedLimitClassesHandler() *removeAppliedLimitClassesHandler {
	return &removeAppliedLimitClassesHandler{}
}

func (h *removeAppliedLimitClassesHandler) Handle(params admin.RemoveAppliedLimitClassesParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		dl.Error("invalid admin principal")
		return admin.NewRemoveAppliedLimitClassesUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return admin.NewRemoveAppliedLimitClassesInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	acct, err := str.FindAccountWithEmail(params.Body.Email, trx)
	if err != nil {
		dl.Errorf("error finding account with email '%v': %v", params.Body.Email, err)
		return admin.NewRemoveAppliedLimitClassesNotFound()
	}

	for _, lcId := range params.Body.LimitClassIds {
		if err := str.RemoveAppliedLimitClass(acct.Id, int(lcId), trx); err != nil {
			dl.Errorf("error removing applied limit class '%v' from '%v': %v", lcId, params.Body.Email, err)
			return admin.NewRemoveAppliedLimitClassesInternalServerError()
		}
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing transaction: %v", err)
		return admin.NewRemoveAppliedLimitClassesInternalServerError()
	}

	return admin.NewRemoveAppliedLimitClassesOK()
}
