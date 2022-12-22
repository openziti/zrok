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
	svc, err := str.FindServiceWithToken(params.SvcToken, tx)
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
		if svc.EnvironmentId == env.Id {
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
		sparkData, err = sparkDataForServices([]*store.Service{svc})
		if err != nil {
			logrus.Errorf("error querying spark data for services: %v", err)
			return metadata.NewGetEnvironmentDetailInternalServerError()
		}
	}
	feEndpoint := ""
	if svc.FrontendEndpoint != nil {
		feEndpoint = *svc.FrontendEndpoint
	}
	feSelection := ""
	if svc.FrontendSelection != nil {
		feSelection = *svc.FrontendSelection
	}
	beProxyEndpoint := ""
	if svc.BackendProxyEndpoint != nil {
		beProxyEndpoint = *svc.BackendProxyEndpoint
	}
	return metadata.NewGetServiceDetailOK().WithPayload(&rest_model_zrok.Service{
		Token:                svc.Token,
		ZID:                  svc.ZId,
		ShareMode:            svc.ShareMode,
		BackendMode:          svc.BackendMode,
		FrontendSelection:    feSelection,
		FrontendEndpoint:     feEndpoint,
		BackendProxyEndpoint: beProxyEndpoint,
		Reserved:             svc.Reserved,
		Metrics:              sparkData[svc.Token],
		CreatedAt:            svc.CreatedAt.UnixMilli(),
		UpdatedAt:            svc.UpdatedAt.UnixMilli(),
	})
}
