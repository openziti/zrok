package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/metadata"
	"github.com/sirupsen/logrus"
)

type environmentDetailHandler struct{}

func newEnvironmentDetailHandler() *environmentDetailHandler {
	return &environmentDetailHandler{}
}

func (h *environmentDetailHandler) Handle(params metadata.GetEnvironmentDetailParams, principal *rest_model_zrok.Principal) middleware.Responder {
	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return metadata.NewGetEnvironmentDetailInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()
	senv, err := str.FindEnvironmentForAccount(params.EnvZID, int(principal.ID), tx)
	if err != nil {
		logrus.Errorf("environment '%v' not found for account '%v': %v", params.EnvZID, principal.Email, err)
		return metadata.NewGetEnvironmentDetailNotFound()
	}
	es := &rest_model_zrok.EnvironmentServices{
		Environment: &rest_model_zrok.Environment{
			Address:     senv.Address,
			CreatedAt:   senv.CreatedAt.UnixMilli(),
			Description: senv.Description,
			Host:        senv.Host,
			UpdatedAt:   senv.UpdatedAt.UnixMilli(),
			ZID:         senv.ZId,
		},
	}
	svcs, err := str.FindServicesForEnvironment(senv.Id, tx)
	if err != nil {
		logrus.Errorf("error finding services for environment '%v': %v", senv.ZId, err)
		return metadata.NewGetEnvironmentDetailInternalServerError()
	}
	var sparkData map[string][]int64
	if cfg.Influx != nil {
		sparkData, err = sparkDataForServices(svcs)
		if err != nil {
			logrus.Errorf("error querying spark data for services: %v", err)
			return metadata.NewGetEnvironmentDetailInternalServerError()
		}
	}
	for _, svc := range svcs {
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
		es.Services = append(es.Services, &rest_model_zrok.Service{
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
	return metadata.NewGetEnvironmentDetailOK().WithPayload(es)
}
