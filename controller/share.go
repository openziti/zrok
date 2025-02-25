package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/share"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/openziti/zrok/util"
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
				logrus.Debugf("found identity '%v' for account '%v'", envZId, principal.Email)
				envId = env.Id
				found = true
				break
			}
		}
		if !found {
			logrus.Errorf("environment '%v' not found for account '%v'", envZId, principal.Email)
			return share.NewShareUnauthorized()
		}
	} else {
		logrus.Errorf("error finding environments for account '%v'", principal.Email)
		return share.NewShareInternalServerError()
	}

	shareMode := sdk.ShareMode(params.Body.ShareMode)
	backendMode := sdk.BackendMode(params.Body.BackendMode)
	if err := h.checkLimits(envId, principal, params.Body.Reserved, params.Body.UniqueName != "", shareMode, backendMode, trx); err != nil {
		logrus.Errorf("limits error: %v", err)
		return share.NewShareUnauthorized()
	}

	var accessGrantAcctIds []int
	if store.PermissionMode(params.Body.PermissionMode) == store.ClosedPermissionMode {
		for _, email := range params.Body.AccessGrants {
			acct, err := str.FindAccountWithEmail(email, trx)
			if err != nil {
				logrus.Errorf("unable to find account '%v' for share request from '%v'", email, principal.Email)
				return share.NewShareNotFound()
			}
			logrus.Debugf("found id '%d' for '%v'", acct.Id, acct.Email)
			accessGrantAcctIds = append(accessGrantAcctIds, acct.Id)
		}
	}

	edge, err := zrokEdgeSdk.Client(cfg.Ziti)
	if err != nil {
		logrus.Error(err)
		return share.NewShareInternalServerError()
	}

	reserved := params.Body.Reserved
	uniqueName := params.Body.UniqueName
	shrToken, err := createShareToken()
	if err != nil {
		logrus.Error(err)
		return share.NewShareInternalServerError()
	}
	if reserved && uniqueName != "" {
		if !util.IsValidUniqueName(uniqueName) {
			logrus.Errorf("invalid unique name '%v' for account '%v'", uniqueName, principal.Email)
			return share.NewShareUnprocessableEntity()
		}
		shareExists, err := str.ShareWithTokenExists(uniqueName, trx)
		if err != nil {
			logrus.Errorf("error checking share for token collision: %v", err)
			return share.NewUpdateShareInternalServerError()
		}
		if shareExists {
			logrus.Errorf("token '%v' already exists; cannot create share", uniqueName)
			return share.NewShareConflict()
		}
		shrToken = uniqueName
	}

	var shrZId string
	var frontendEndpoints []string
	switch params.Body.ShareMode {
	case string(sdk.PublicShareMode):
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
			if sfe.PermissionMode == store.ClosedPermissionMode {
				granted, err := str.IsFrontendGrantedToAccount(int(principal.ID), sfe.Id, trx)
				if err != nil {
					logrus.Error(err)
					return share.NewShareInternalServerError()
				}
				if !granted {
					logrus.Errorf("'%v' is not granted access to frontend '%v'", principal.Email, frontendSelection)
					return share.NewShareNotFound()
				}
			}
			if sfe != nil && sfe.UrlTemplate != nil {
				frontendZIds = append(frontendZIds, sfe.ZId)
				frontendTemplates = append(frontendTemplates, *sfe.UrlTemplate)
				logrus.Infof("added frontend selection '%v' with ziti identity '%v' for share '%v'", frontendSelection, sfe.ZId, shrToken)
			}
		}
		var skipInterstitial bool
		if backendMode != sdk.DriveBackendMode {
			skipInterstitial, err = str.IsAccountGrantedSkipInterstitial(int(principal.ID), trx)
			if err != nil {
				logrus.Errorf("error checking skip interstitial for account '%v': %v", principal.Email, err)
				return share.NewShareInternalServerError()
			}
		} else {
			skipInterstitial = true
		}
		shrZId, frontendEndpoints, err = newPublicResourceAllocator().allocate(envZId, shrToken, frontendZIds, frontendTemplates, params, !skipInterstitial, edge)
		if err != nil {
			logrus.Error(err)
			return share.NewShareInternalServerError()
		}

	case string(sdk.PrivateShareMode):
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

	sshr := &store.Share{
		ZId:                  shrZId,
		Token:                shrToken,
		ShareMode:            params.Body.ShareMode,
		BackendMode:          params.Body.BackendMode,
		BackendProxyEndpoint: &params.Body.BackendProxyEndpoint,
		Reserved:             reserved,
		UniqueName:           reserved && uniqueName != "",
		PermissionMode:       store.OpenPermissionMode,
	}
	if params.Body.PermissionMode != "" {
		sshr.PermissionMode = store.PermissionMode(params.Body.PermissionMode)
	}
	if len(params.Body.FrontendSelection) > 0 {
		sshr.FrontendSelection = &params.Body.FrontendSelection[0]
	}
	if len(frontendEndpoints) > 0 {
		sshr.FrontendEndpoint = &frontendEndpoints[0]
	} else if sshr.ShareMode == string(sdk.PrivateShareMode) {
		sshr.FrontendEndpoint = &sshr.ShareMode
	}

	sid, err := str.CreateShare(envId, sshr, trx)
	if err != nil {
		logrus.Errorf("error creating share record: %v", err)
		return share.NewShareInternalServerError()
	}

	if sshr.PermissionMode == store.ClosedPermissionMode {
		for _, acctId := range accessGrantAcctIds {
			_, err := str.CreateAccessGrant(sid, acctId, trx)
			if err != nil {
				logrus.Errorf("error creating access grant for '%v': %v", principal.Email, err)
				return share.NewShareInternalServerError()
			}
		}
	}

	if err := trx.Commit(); err != nil {
		logrus.Errorf("error committing share record: %v", err)
		return share.NewShareInternalServerError()
	}
	logrus.Infof("recorded share '%v' with id '%v' for '%v'", shrToken, sid, principal.Email)

	return share.NewShareCreated().WithPayload(&rest_model_zrok.ShareResponse{
		FrontendProxyEndpoints: frontendEndpoints,
		ShareToken:             shrToken,
	})
}

func (h *shareHandler) checkLimits(envId int, principal *rest_model_zrok.Principal, reserved, uniqueName bool, shareMode sdk.ShareMode, backendMode sdk.BackendMode, trx *sqlx.Tx) error {
	if !principal.Limitless {
		if limitsAgent != nil {
			ok, err := limitsAgent.CanCreateShare(int(principal.ID), envId, reserved, uniqueName, shareMode, backendMode, trx)
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
