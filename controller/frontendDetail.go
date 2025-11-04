package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/metadata"
)

type getFrontendDetailHandler struct{}

func newGetFrontendDetailHandler() *getFrontendDetailHandler {
	return &getFrontendDetailHandler{}
}

func (h *getFrontendDetailHandler) Handle(params metadata.GetFrontendDetailParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return metadata.NewGetFrontendDetailInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()
	fe, err := str.GetFrontend(int(params.FrontendID), trx)
	if err != nil {
		dl.Errorf("error finding share '%d': %v", params.FrontendID, err)
		return metadata.NewGetFrontendDetailNotFound()
	}
	envs, err := str.FindEnvironmentsForAccount(int(principal.ID), trx)
	if err != nil {
		dl.Errorf("error finding environments for account '%v': %v", principal.Email, err)
		return metadata.NewGetFrontendDetailInternalServerError()
	}
	found := false
	if fe.EnvironmentId == nil {
		dl.Errorf("non owned environment '%d' for '%v'", fe.Id, principal.Email)
		return metadata.NewGetFrontendDetailNotFound()
	}
	for _, env := range envs {
		if *fe.EnvironmentId == env.Id {
			found = true
			break
		}
	}
	if !found {
		dl.Errorf("environment not matched for frontend '%d' for account '%v'", fe.Id, principal.Email)
		return metadata.NewGetFrontendDetailNotFound()
	}
	payload := &rest_model_zrok.Frontend{
		ID:            int64(fe.Id),
		FrontendToken: fe.Token,
		ZID:           fe.ZId,
		CreatedAt:     fe.CreatedAt.UnixMilli(),
		UpdatedAt:     fe.UpdatedAt.UnixMilli(),
	}
	if fe.BindAddress != nil {
		payload.BindAddress = *fe.BindAddress
	}
	if fe.Description != nil {
		payload.Description = *fe.Description
	}
	if fe.PrivateShareId != nil {
		shr, err := str.GetShare(*fe.PrivateShareId, trx)
		if err != nil {
			dl.Errorf("error getting share for frontend '%d': %v", fe.Id, err)
			return metadata.NewGetFrontendDetailInternalServerError()
		}
		payload.ShareToken = shr.Token
		payload.BackendMode = shr.BackendMode
	}
	return metadata.NewGetFrontendDetailOK().WithPayload(payload)
}
