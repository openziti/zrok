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
	elm, err := newEnvironmentsLimitedMap(envs, trx)
	if err != nil {
		logrus.Errorf("error finding limited environments for '%v': %v", principal.Email, err)
		return metadata.NewOverviewInternalServerError()
	}
	accountLimited, err := h.isAccountLimited(principal, trx)
	if err != nil {
		logrus.Errorf("error checking account limited for '%v': %v", principal.Email, err)
	}
	ovr := &rest_model_zrok.Overview{AccountLimited: accountLimited}
	for _, env := range envs {
		shrs, err := str.FindSharesForEnvironment(env.Id, trx)
		if err != nil {
			logrus.Errorf("error finding shares for environment '%v': %v", env.ZId, err)
			return metadata.NewOverviewInternalServerError()
		}
		slm, err := newSharesLimitedMap(shrs, trx)
		if err != nil {
			logrus.Errorf("error finding limited shares for environment '%v': %v", env.ZId, err)
			return metadata.NewOverviewInternalServerError()
		}
		envRes := &rest_model_zrok.EnvironmentAndResources{
			Environment: &rest_model_zrok.Environment{
				Address:     env.Address,
				Description: env.Description,
				Host:        env.Host,
				ZID:         env.ZId,
				Limited:     elm.isLimited(env),
				CreatedAt:   env.CreatedAt.UnixMilli(),
				UpdatedAt:   env.UpdatedAt.UnixMilli(),
			},
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
				Token:                shr.Token,
				ZID:                  shr.ZId,
				ShareMode:            shr.ShareMode,
				BackendMode:          shr.BackendMode,
				FrontendSelection:    feSelection,
				FrontendEndpoint:     feEndpoint,
				BackendProxyEndpoint: beProxyEndpoint,
				Reserved:             shr.Reserved,
				Limited:              slm.isLimited(shr),
				CreatedAt:            shr.CreatedAt.UnixMilli(),
				UpdatedAt:            shr.UpdatedAt.UnixMilli(),
			}
			envRes.Shares = append(envRes.Shares, envShr)
		}
		ovr.Environments = append(ovr.Environments, envRes)
	}
	return metadata.NewOverviewOK().WithPayload(ovr)
}

func (h *overviewHandler) isAccountLimited(principal *rest_model_zrok.Principal, trx *sqlx.Tx) (bool, error) {
	var alj *store.AccountLimitJournal
	aljEmpty, err := str.IsAccountLimitJournalEmpty(int(principal.ID), trx)
	if err != nil {
		return false, err
	}
	if !aljEmpty {
		alj, err = str.FindLatestAccountLimitJournal(int(principal.ID), trx)
		if err != nil {
			return false, err
		}
	}
	return alj != nil && alj.Action == store.LimitAction, nil
}

type sharesLimitedMap struct {
	v map[int]struct{}
}

func newSharesLimitedMap(shrs []*store.Share, trx *sqlx.Tx) (*sharesLimitedMap, error) {
	var shrIds []int
	for i := range shrs {
		shrIds = append(shrIds, shrs[i].Id)
	}
	shrsLimited, err := str.FindSelectedLatestShareLimitjournal(shrIds, trx)
	if err != nil {
		return nil, err
	}
	slm := &sharesLimitedMap{v: make(map[int]struct{})}
	for i := range shrsLimited {
		if shrsLimited[i].Action == store.LimitAction {
			slm.v[shrsLimited[i].ShareId] = struct{}{}
		}
	}
	return slm, nil
}

func (m *sharesLimitedMap) isLimited(shr *store.Share) bool {
	_, limited := m.v[shr.Id]
	return limited
}

type environmentsLimitedMap struct {
	v map[int]struct{}
}

func newEnvironmentsLimitedMap(envs []*store.Environment, trx *sqlx.Tx) (*environmentsLimitedMap, error) {
	var envIds []int
	for i := range envs {
		envIds = append(envIds, envs[i].Id)
	}
	envsLimited, err := str.FindSelectedLatestEnvironmentLimitJournal(envIds, trx)
	if err != nil {
		return nil, err
	}
	elm := &environmentsLimitedMap{v: make(map[int]struct{})}
	for i := range envsLimited {
		if envsLimited[i].Action == store.LimitAction {
			elm.v[envsLimited[i].EnvironmentId] = struct{}{}
		}
	}
	return elm, nil
}

func (m *environmentsLimitedMap) isLimited(env *store.Environment) bool {
	_, limited := m.v[env.Id]
	return limited
}
