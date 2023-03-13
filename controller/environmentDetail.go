package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/metadata"
	"github.com/sirupsen/logrus"
)

type environmentDetailHandler struct{}

func newEnvironmentDetailHandler() *environmentDetailHandler {
	return &environmentDetailHandler{}
}

func (h *environmentDetailHandler) Handle(params metadata.GetEnvironmentDetailParams, principal *rest_model_zrok.Principal) middleware.Responder {
	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction for user '%v': %v", principal.Email, err)
		return metadata.NewGetEnvironmentDetailInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()
	senv, err := str.FindEnvironmentForAccount(params.EnvZID, int(principal.ID), tx)
	if err != nil {
		logrus.Errorf("environment '%v' not found for account '%v': %v", params.EnvZID, principal.Email, err)
		return metadata.NewGetEnvironmentDetailNotFound()
	}
	es := &rest_model_zrok.EnvironmentShares{
		Environment: &rest_model_zrok.Environment{
			Address:     senv.Address,
			CreatedAt:   senv.CreatedAt.UnixMilli(),
			Description: senv.Description,
			Host:        senv.Host,
			UpdatedAt:   senv.UpdatedAt.UnixMilli(),
			ZID:         senv.ZId,
		},
	}
	shrs, err := str.FindSharesForEnvironment(senv.Id, tx)
	if err != nil {
		logrus.Errorf("error finding shares for environment '%v' for user '%v': %v", senv.ZId, principal.Email, err)
		return metadata.NewGetEnvironmentDetailInternalServerError()
	}
	var sparkData map[string][]int64
	if cfg.Metrics != nil && cfg.Metrics.Influx != nil {
		sparkData, err = sparkDataForShares(shrs)
		if err != nil {
			logrus.Errorf("error querying spark data for shares for user '%v': %v", principal.Email, err)
		}
	} else {
		logrus.Debug("skipping spark data for shares; no influx configuration")
	}
	for _, shr := range shrs {
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
		es.Shares = append(es.Shares, &rest_model_zrok.Share{
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
	return metadata.NewGetEnvironmentDetailOK().WithPayload(es)
}
