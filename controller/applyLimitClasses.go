package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/admin"
)

type applyLimitClassesHandler struct{}

func newApplyLimitClassesHandler() *applyLimitClassesHandler {
	return &applyLimitClassesHandler{}
}

func (h *applyLimitClassesHandler) Handle(params admin.ApplyLimitClassesParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		dl.Error("invalid admin principal")
		return admin.NewApplyLimitClassesUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return admin.NewApplyLimitClassesInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	acct, err := str.FindAccountWithEmail(params.Body.Email, trx)
	if err != nil {
		dl.Errorf("error finding account with email '%v': %v", params.Body.Email, err)
		return admin.NewApplyLimitClassesNotFound()
	}

	for _, lcId := range params.Body.LimitClassIds {
		if _, err := str.GetLimitClass(int(lcId), trx); err != nil {
			dl.Errorf("error finding limit class '%v': %v", lcId, err)
			return admin.NewApplyLimitClassesNotFound()
		}
		if _, err := str.ApplyLimitClass(&store.AppliedLimitClass{AccountId: acct.Id, LimitClassId: int(lcId)}, trx); err != nil {
			dl.Errorf("error applying limit class '%v' to '%v': %v", lcId, params.Body.Email, err)
			return admin.NewApplyLimitClassesInternalServerError()
		}
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing transaction: %v", err)
		return admin.NewApplyLimitClassesInternalServerError()
	}

	return admin.NewApplyLimitClassesOK()
}
