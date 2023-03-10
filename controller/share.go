package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/controller/limits"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/share"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type shareHandler struct {
	cfg *limits.Config
}

func newShareHandler(cfg *limits.Config) *shareHandler {
	return &shareHandler{cfg: cfg}
}

func (h *shareHandler) Handle(params share.ShareParams, principal *rest_model_zrok.Principal) middleware.Responder {
	logrus.Infof("handling")

	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return share.NewShareInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()

	envZId := params.Body.EnvZID
	envId := 0
	envs, err := str.FindEnvironmentsForAccount(int(principal.ID), tx)
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

	if err := h.checkLimits(principal, envs, tx); err != nil {
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
			sfe, err := str.FindFrontendPubliclyNamed(frontendSelection, tx)
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
	if len(frontendEndpoints) > 0 {
		sshr.FrontendEndpoint = &frontendEndpoints[0]
	} else if sshr.ShareMode == "private" {
		sshr.FrontendEndpoint = &sshr.ShareMode
	}

	sid, err := str.CreateShare(envId, sshr, tx)
	if err != nil {
		logrus.Errorf("error creating share record: %v", err)
		return share.NewShareInternalServerError()
	}

	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing share record: %v", err)
		return share.NewShareInternalServerError()
	}
	logrus.Infof("recorded share '%v' with id '%v' for '%v'", shrToken, sid, principal.Email)

	return share.NewShareCreated().WithPayload(&rest_model_zrok.ShareResponse{
		FrontendProxyEndpoints: frontendEndpoints,
		ShrToken:               shrToken,
	})
}

func (h *shareHandler) checkLimits(principal *rest_model_zrok.Principal, envs []*store.Environment, tx *sqlx.Tx) error {
	if !principal.Limitless && h.cfg.Shares > limits.Unlimited {
		total := 0
		for i := range envs {
			shrs, err := str.FindSharesForEnvironment(envs[i].Id, tx)
			if err != nil {
				return errors.Errorf("unable to find shares for environment '%v': %v", envs[i].ZId, err)
			}
			total += len(shrs)
			if total+1 > h.cfg.Shares {
				return errors.Errorf("would exceed shares limit of %d for '%v'", h.cfg.Shares, principal.Email)
			}
		}
	}
	return nil
}
