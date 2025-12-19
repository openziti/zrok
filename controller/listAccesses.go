package controller

import (
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/metadata"
)

type listAccessesHandler struct{}

func newListAccessesHandler() *listAccessesHandler {
	return &listAccessesHandler{}
}

func (h *listAccessesHandler) Handle(params metadata.ListAccessesParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction for user '%v': %v", principal.Email, err)
		return metadata.NewListAccessesInternalServerError().WithPayload("error starting transaction")
	}
	defer func() { _ = trx.Rollback() }()

	// build filter from query parameters
	filter := &store.FrontendFilter{}

	if params.EnvZID != nil {
		filter.EnvZId = params.EnvZID
	}

	if params.ShareToken != nil {
		filter.ShareToken = params.ShareToken
	}

	if params.BindAddress != nil {
		filter.BindAddress = params.BindAddress
	}

	if params.Description != nil {
		filter.Description = params.Description
	}

	// parse date filters
	if params.CreatedAfter != nil {
		t, err := time.Parse(time.RFC3339, *params.CreatedAfter)
		if err != nil {
			dl.Errorf("invalid createdAfter format for user '%v': %v", principal.Email, err)
			return metadata.NewListAccessesBadRequest().WithPayload("invalid createdAfter date format, expected RFC3339")
		}
		filter.CreatedAfter = &t
	}

	if params.CreatedBefore != nil {
		t, err := time.Parse(time.RFC3339, *params.CreatedBefore)
		if err != nil {
			dl.Errorf("invalid createdBefore format for user '%v': %v", principal.Email, err)
			return metadata.NewListAccessesBadRequest().WithPayload("invalid createdBefore date format, expected RFC3339")
		}
		filter.CreatedBefore = &t
	}

	if params.UpdatedAfter != nil {
		t, err := time.Parse(time.RFC3339, *params.UpdatedAfter)
		if err != nil {
			dl.Errorf("invalid updatedAfter format for user '%v': %v", principal.Email, err)
			return metadata.NewListAccessesBadRequest().WithPayload("invalid updatedAfter date format, expected RFC3339")
		}
		filter.UpdatedAfter = &t
	}

	if params.UpdatedBefore != nil {
		t, err := time.Parse(time.RFC3339, *params.UpdatedBefore)
		if err != nil {
			dl.Errorf("invalid updatedBefore format for user '%v': %v", principal.Email, err)
			return metadata.NewListAccessesBadRequest().WithPayload("invalid updatedBefore date format, expected RFC3339")
		}
		filter.UpdatedBefore = &t
	}

	// query frontends with filter
	frontends, err := str.FindFrontendsForAccountWithFilter(int(principal.ID), filter, trx)
	if err != nil {
		dl.Errorf("error finding frontends for user '%v': %v", principal.Email, err)
		return metadata.NewListAccessesInternalServerError()
	}

	// check account limits
	isLimited := false
	if empty, err := str.IsBandwidthLimitJournalEmpty(int(principal.ID), trx); !empty && err == nil {
		alj, err := str.FindLatestBandwidthLimitJournal(int(principal.ID), trx)
		if err != nil {
			dl.Errorf("error finding account limit journal for '%v': %v", principal.Email, err)
		}
		isLimited = alj != nil && alj.Action == store.LimitLimitAction
	} else if err != nil {
		dl.Errorf("error finding limit journal for '%v': %v", principal.Email, err)
	}

	// build response
	response := &rest_model_zrok.AccessesList{}
	for _, fe := range frontends {
		summary := &rest_model_zrok.AccessSummary{
			ID:            int64(fe.Frontend.Id),
			FrontendToken: fe.Frontend.Token,
			Limited:       isLimited,
			CreatedAt:     fe.Frontend.CreatedAt.UnixMilli(),
			UpdatedAt:     fe.Frontend.UpdatedAt.UnixMilli(),
		}

		// add envZId if available
		if fe.EnvZId != nil {
			summary.EnvZID = *fe.EnvZId
		}

		// add shareToken if available
		if fe.ShareToken != nil {
			summary.ShareToken = *fe.ShareToken
		}

		// add backendMode if available
		if fe.BackendMode != nil {
			summary.BackendMode = *fe.BackendMode
		}

		// add bindAddress if available
		if fe.Frontend.BindAddress != nil {
			summary.BindAddress = *fe.Frontend.BindAddress
		}

		// add description if available
		if fe.Frontend.Description != nil {
			summary.Description = *fe.Frontend.Description
		}

		response.Accesses = append(response.Accesses, summary)
	}

	return metadata.NewListAccessesOK().WithPayload(response)
}
