package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/share"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type shareHandler struct{}

func newShareHandler() *shareHandler {
	return &shareHandler{}
}

func (h *shareHandler) Handle(params share.ShareParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return share.NewShareInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	envZId := params.Body.EnvZID
	envId := 0
	envs, err := str.FindEnvironmentsForAccount(int(principal.ID), trx)
	if err == nil {
		found := false
		for _, env := range envs {
			if env.ZId == envZId {
				logrus.Debugf("found identity '%v' for user '%v'", envZId, principal.Email)
				envId = env.Id
				found = true
				break
			}
		}
		if !found {
			logrus.Errorf("environment '%v' not found for user '%v'", envZId, principal.Email)
			return share.NewShareUnauthorized()
		}
	} else {
		logrus.Errorf("error finding environments for account '%v'", principal.Email)
		return share.NewShareInternalServerError()
	}

	if err := h.checkLimits(envId, principal, trx); err != nil {
		logrus.Errorf("limits error: %v", err)
		return share.NewShareUnauthorized()
	}

	edge, err := zrokEdgeSdk.Client(cfg.Ziti)
	if err != nil {
		logrus.Error(err)
		return share.NewShareInternalServerError()
	}
	shrToken, err := createShareToken()
	if err != nil {
		logrus.Error(err)
		return share.NewShareInternalServerError()
	}

	var shrZId string
	var frontendEndpoints []string
	switch params.Body.ShareMode {
	case "public":
		if len(params.Body.FrontendSelection) < 1 {
			logrus.Info("no frontend selection provided")
			return share.NewShareNotFound()
		}

		var frontendZIds []string
		var frontendTemplates []string
		for _, frontendSelection := range params.Body.FrontendSelection {
			sfe, err := str.FindFrontendPubliclyNamed(frontendSelection, trx)
			if err != nil {
				logrus.Error(err)
				return share.NewShareNotFound()
			}
			if sfe != nil && sfe.UrlTemplate != nil {
				frontendZIds = append(frontendZIds, sfe.ZId)
				frontendTemplates = append(frontendTemplates, *sfe.UrlTemplate)
				logrus.Infof("added frontend selection '%v' with ziti identity '%v' for share '%v'", frontendSelection, sfe.ZId, shrToken)
			}
		}
		shrZId, frontendEndpoints, err = newPublicResourceAllocator().allocate(envZId, shrToken, frontendZIds, frontendTemplates, params, edge)
		if err != nil {
			logrus.Error(err)
			return share.NewShareInternalServerError()
		}

	case "private":
		shrZId, frontendEndpoints, err = newPrivateResourceAllocator().allocate(envZId, shrToken, params, edge)
		if err != nil {
			logrus.Error(err)
			return share.NewShareInternalServerError()
		}

	default:
		logrus.Errorf("unknown share mode '%v", params.Body.ShareMode)
		return share.NewShareInternalServerError()
	}

	logrus.Debugf("allocated share '%v'", shrToken)

	reserved := params.Body.Reserved
	sshr := &store.Share{
		ZId:                  shrZId,
		Token:                shrToken,
		ShareMode:            params.Body.ShareMode,
		BackendMode:          params.Body.BackendMode,
		BackendProxyEndpoint: &params.Body.BackendProxyEndpoint,
		Reserved:             reserved,
	}
	if len(params.Body.FrontendSelection) > 0 {
		sshr.FrontendSelection = &params.Body.FrontendSelection[0]
	}
	if len(frontendEndpoints) > 0 {
		sshr.FrontendEndpoint = &frontendEndpoints[0]
	} else if sshr.ShareMode == "private" {
		sshr.FrontendEndpoint = &sshr.ShareMode
	}

	sid, err := str.CreateShare(envId, sshr, trx)
	if err != nil {
		logrus.Errorf("error creating share record: %v", err)
		return share.NewShareInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		logrus.Errorf("error committing share record: %v", err)
		return share.NewShareInternalServerError()
	}
	logrus.Infof("recorded share '%v' with id '%v' for '%v'", shrToken, sid, principal.Email)

	return share.NewShareCreated().WithPayload(&rest_model_zrok.ShareResponse{
		FrontendProxyEndpoints: frontendEndpoints,
		ShrToken:               shrToken,
	})
}

func (h *shareHandler) checkLimits(envId int, principal *rest_model_zrok.Principal, trx *sqlx.Tx) error {
	if !principal.Limitless {
		if limitsAgent != nil {
			ok, err := limitsAgent.CanCreateShare(int(principal.ID), envId, trx)
			if err != nil {
				return errors.Wrapf(err, "error checking share limits for '%v'", principal.Email)
			}
			if !ok {
				return errors.Errorf("share limit check failed for '%v'", principal.Email)
			}
		}
	}
	return nil
}
