package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/metadata"
	"github.com/sirupsen/logrus"
)

type overviewHandler struct{}

func newOverviewHandler() *overviewHandler {
	return &overviewHandler{}
}

func (h *overviewHandler) Handle(_ metadata.OverviewParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return metadata.NewOverviewInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	envs, err := str.FindEnvironmentsForAccount(int(principal.ID), trx)
	if err != nil {
		logrus.Errorf("error finding environments for '%v': %v", principal.Email, err)
		return metadata.NewOverviewInternalServerError()
	}

	accountLimited, err := h.isAccountLimited(principal, trx)
	if err != nil {
		logrus.Errorf("error checking account limited for '%v': %v", principal.Email, err)
	}

	ovr := &rest_model_zrok.Overview{AccountLimited: accountLimited}
	for _, env := range envs {
		ear := &rest_model_zrok.EnvironmentAndResources{
			Environment: &rest_model_zrok.Environment{
				Address:     env.Address,
				Description: env.Description,
				Host:        env.Host,
				ZID:         env.ZId,
				Limited:     accountLimited,
				CreatedAt:   env.CreatedAt.UnixMilli(),
				UpdatedAt:   env.UpdatedAt.UnixMilli(),
			},
		}

		shrs, err := str.FindSharesForEnvironment(env.Id, trx)
		if err != nil {
			logrus.Errorf("error finding shares for environment '%v': %v", env.ZId, err)
			return metadata.NewOverviewInternalServerError()
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
			envShr := &rest_model_zrok.Share{
				ShareToken:           shr.Token,
				ZID:                  shr.ZId,
				ShareMode:            shr.ShareMode,
				BackendMode:          shr.BackendMode,
				FrontendSelection:    feSelection,
				FrontendEndpoint:     feEndpoint,
				BackendProxyEndpoint: beProxyEndpoint,
				Reserved:             shr.Reserved,
				Limited:              accountLimited,
				CreatedAt:            shr.CreatedAt.UnixMilli(),
				UpdatedAt:            shr.UpdatedAt.UnixMilli(),
			}
			ear.Shares = append(ear.Shares, envShr)
		}
		fes, err := str.FindFrontendsForEnvironment(env.Id, trx)
		if err != nil {
			logrus.Errorf("error finding frontends for environment '%v': %v", env.ZId, err)
			return metadata.NewOverviewInternalServerError()
		}
		for _, fe := range fes {
			envFe := &rest_model_zrok.Frontend{
				ID:            int64(fe.Id),
				FrontendToken: fe.Token,
				ZID:           fe.ZId,
				CreatedAt:     fe.CreatedAt.UnixMilli(),
				UpdatedAt:     fe.UpdatedAt.UnixMilli(),
			}
			if fe.BindAddress != nil {
				envFe.BindAddress = *fe.BindAddress
			}
			if fe.Description != nil {
				envFe.Description = *fe.Description
			}
			if fe.PrivateShareId != nil {
				feShr, err := str.GetShare(*fe.PrivateShareId, trx)
				if err != nil {
					logrus.Errorf("error getting share for frontend '%v': %v", fe.ZId, err)
					return metadata.NewOverviewInternalServerError()
				}
				envFe.ShareToken = feShr.Token
				envFe.BackendMode = feShr.BackendMode
			}
			ear.Frontends = append(ear.Frontends, envFe)
		}

		ovr.Environments = append(ovr.Environments, ear)
	}

	return metadata.NewOverviewOK().WithPayload(ovr)
}

func (h *overviewHandler) isAccountLimited(principal *rest_model_zrok.Principal, trx *sqlx.Tx) (bool, error) {
	var je *store.BandwidthLimitJournalEntry
	jEmpty, err := str.IsBandwidthLimitJournalEmpty(int(principal.ID), trx)
	if err != nil {
		return false, err
	}
	if !jEmpty {
		je, err = str.FindLatestBandwidthLimitJournal(int(principal.ID), trx)
		if err != nil {
			return false, err
		}
	}
	return je != nil && je.Action == store.LimitLimitAction, nil
}
