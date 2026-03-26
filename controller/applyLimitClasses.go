package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/admin"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
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

	existingLcs, err := str.FindAppliedLimitClassesForAccount(acct.Id, trx)
	if err != nil {
		dl.Errorf("error finding applied limit classes for '%v': %v", params.Body.Email, err)
		return admin.NewApplyLimitClassesInternalServerError()
	}

	seenIds := make(map[int]bool)
	for _, lc := range existingLcs {
		seenIds[lc.Id] = true
	}

	occupiedSlots := newLimitClassSlots()
	for _, lc := range existingLcs {
		if conflict := occupiedSlots.add(lc); conflict {
			dl.Errorf("existing applied limit classes for '%v' are already in a conflicting state", params.Body.Email)
			return admin.NewApplyLimitClassesInternalServerError()
		}
	}

	var toApply []*store.AppliedLimitClass
	for _, lcId := range params.Body.LimitClassIds {
		if seenIds[int(lcId)] {
			continue
		}

		lc, err := str.GetLimitClass(int(lcId), trx)
		if err != nil {
			dl.Errorf("error finding limit class '%v': %v", lcId, err)
			return admin.NewApplyLimitClassesNotFound()
		}

		if conflict := occupiedSlots.add(lc); conflict {
			dl.Errorf("applying limit class '%v' to '%v' would create conflicting effective limits", lcId, params.Body.Email)
			return admin.NewApplyLimitClassesInternalServerError()
		}

		seenIds[lc.Id] = true
		toApply = append(toApply, &store.AppliedLimitClass{AccountId: acct.Id, LimitClassId: lc.Id})
	}

	for _, applied := range toApply {
		if _, err := str.ApplyLimitClass(applied, trx); err != nil {
			dl.Errorf("error applying limit class '%v' to '%v': %v", applied.LimitClassId, params.Body.Email, err)
			return admin.NewApplyLimitClassesInternalServerError()
		}
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing transaction: %v", err)
		return admin.NewApplyLimitClassesInternalServerError()
	}

	return admin.NewApplyLimitClassesOK()
}

type limitClassSlots struct {
	resource  *store.LimitClass
	bwWarning *store.LimitClass
	bwLimit   *store.LimitClass
	scopes    map[sdk.BackendMode]*store.LimitClass
}

func newLimitClassSlots() *limitClassSlots {
	return &limitClassSlots{
		scopes: make(map[sdk.BackendMode]*store.LimitClass),
	}
}

func (slots *limitClassSlots) add(lc *store.LimitClass) bool {
	if lc == nil {
		return false
	}

	if lc.IsResourceCountClass() {
		if slots.resource != nil && slots.resource.Id != lc.Id {
			return true
		}
		slots.resource = lc
		return false
	}

	if lc.IsUnscopedBandwidthClass() {
		if lc.LimitAction == store.WarningLimitAction {
			if slots.bwWarning != nil && slots.bwWarning.Id != lc.Id {
				return true
			}
			slots.bwWarning = lc
			return false
		}
		if slots.bwLimit != nil && slots.bwLimit.Id != lc.Id {
			return true
		}
		slots.bwLimit = lc
		return false
	}

	if lc.IsScopedBandwidthClass() {
		backendMode := *lc.BackendMode
		if existing, found := slots.scopes[backendMode]; found && existing.Id != lc.Id {
			return true
		}
		slots.scopes[backendMode] = lc
	}

	return false
}
