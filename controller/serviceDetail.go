package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/controller/store"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/metadata"
	"github.com/sirupsen/logrus"
)

type serviceDetailHandler struct{}

func newServiceDetailHandler() *serviceDetailHandler {
	return &serviceDetailHandler{}
}

func (h *serviceDetailHandler) Handle(params metadata.GetServiceDetailParams, principal *rest_model_zrok.Principal) middleware.Responder {
	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return metadata.NewGetServiceDetailInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()
	shr, err := str.FindShareWithToken(params.SvcToken, tx)
	if err != nil {
		logrus.Errorf("error finding service '%v': %v", params.SvcToken, err)
		return metadata.NewGetServiceDetailNotFound()
	}
	envs, err := str.FindEnvironmentsForAccount(int(principal.ID), tx)
	if err != nil {
		logrus.Errorf("error finding environments for account '%v': %v", principal.Email, err)
		return metadata.NewGetServiceDetailInternalServerError()
	}
	found := false
	for _, env := range envs {
		if shr.EnvironmentId == env.Id {
			found = true
			break
		}
	}
	if !found {
		logrus.Errorf("environment not matched for service '%v' for account '%v'", params.SvcToken, principal.Email)
		return metadata.NewGetServiceDetailNotFound()
	}
	var sparkData map[string][]int64
	if cfg.Influx != nil {
		sparkData, err = sparkDataForServices([]*store.Share{shr})
		if err != nil {
			logrus.Errorf("error querying spark data for services: %v", err)
			return metadata.NewGetEnvironmentDetailInternalServerError()
		}
	}
	feEndpoint := ""
	if shr.FrontendEndpoint != nil {
		feEndpoint = *shr.FrontendEndpoint
	}
	feSelection := ""
	if shr.FrontendSelection != nil {
		feSelection = *shr.FrontendSelection
	}
	beProxyEndpoint := ""
	if shr.BackendProxyEndpoint != nil {
		beProxyEndpoint = *shr.BackendProxyEndpoint
	}
	return metadata.NewGetServiceDetailOK().WithPayload(&rest_model_zrok.Service{
		Token:                shr.Token,
		ZID:                  shr.ZId,
		ShareMode:            shr.ShareMode,
		BackendMode:          shr.BackendMode,
		FrontendSelection:    feSelection,
		FrontendEndpoint:     feEndpoint,
		BackendProxyEndpoint: beProxyEndpoint,
		Reserved:             shr.Reserved,
		Metrics:              sparkData[shr.Token],
		CreatedAt:            shr.CreatedAt.UnixMilli(),
		UpdatedAt:            shr.UpdatedAt.UnixMilli(),
	})
}
