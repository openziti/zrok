package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/admin"
)

type listLimitClassesHandler struct{}

func newListLimitClassesHandler() *listLimitClassesHandler {
	return &listLimitClassesHandler{}
}

func (h *listLimitClassesHandler) Handle(params admin.ListLimitClassesParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		dl.Error("invalid admin principal")
		return admin.NewListLimitClassesUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return admin.NewListLimitClassesInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	lcs, err := str.FindLimitClassesByLabel(params.Body.Label, trx)
	if err != nil {
		dl.Errorf("error finding limit classes by label '%v': %v", params.Body.Label, err)
		return admin.NewListLimitClassesInternalServerError()
	}

	return admin.NewListLimitClassesOK().WithPayload(limitClassesToApi(lcs))
}

func limitClassToApi(lc *store.LimitClass) *rest_model_zrok.LimitClass {
	out := &rest_model_zrok.LimitClass{
		ID:             int64(lc.Id),
		Environments:   int64(lc.Environments),
		Shares:         int64(lc.Shares),
		ReservedShares: int64(lc.ReservedShares),
		UniqueNames:    int64(lc.UniqueNames),
		ShareFrontends: int64(lc.ShareFrontends),
		PeriodMinutes:  int64(lc.PeriodMinutes),
		RxBytes:        lc.RxBytes,
		TxBytes:        lc.TxBytes,
		TotalBytes:     lc.TotalBytes,
		LimitAction:    string(lc.LimitAction),
		CreatedAt:      lc.CreatedAt.UnixMilli(),
		UpdatedAt:      lc.UpdatedAt.UnixMilli(),
	}
	if lc.Label != nil {
		out.Label = *lc.Label
	}
	if lc.BackendMode != nil {
		out.BackendMode = string(*lc.BackendMode)
	}
	return out
}

func limitClassesToApi(lcs []*store.LimitClass) []*rest_model_zrok.LimitClass {
	var out []*rest_model_zrok.LimitClass
	for _, lc := range lcs {
		out = append(out, limitClassToApi(lc))
	}
	return out
}
