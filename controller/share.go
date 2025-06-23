package controller

import (
	"encoding/json"

	"github.com/go-openapi/runtime/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/openziti/edge-api/rest_management_api_client"
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
	logrus.Info("handling")

	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return share.NewShareInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	envId, err := h.validateEnvironment(params.Body.EnvZID, principal, trx)
	if err != nil {
		return h.handleEnvironmentError(err)
	}

	if err := h.checkLimits(envId, principal, params.Body.Reserved, params.Body.UniqueName != "", sdk.ShareMode(params.Body.ShareMode), sdk.BackendMode(params.Body.BackendMode), trx); err != nil {
		logrus.Errorf("limits error: %v", err)
		return share.NewShareUnauthorized()
	}

	accessGrantAcctIds, responder := h.processAccessGrants(params, principal, trx)
	if responder != nil {
		return responder
	}

	edge, err := zrokEdgeSdk.Client(cfg.Ziti)
	if err != nil {
		logrus.Error(err)
		return share.NewShareInternalServerError()
	}

	shrToken, responder := h.createShareToken(params.Body.Reserved, params.Body.UniqueName, trx)
	if responder != nil {
		return responder
	}

	shrZId, frontendEndpoints, responder := h.allocateResources(params, principal, edge, shrToken, trx)
	if responder != nil {
		return responder
	}

	sshr := h.createShareRecord(shrZId, shrToken, params, frontendEndpoints)

	sid, responder := h.saveShareAndGrants(sshr, envId, accessGrantAcctIds, trx)
	if responder != nil {
		return responder
	}

	if responder := h.handleAuthSecrets(params, sid, sshr, trx); responder != nil {
		return responder
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

func (h *shareHandler) validateEnvironment(envZId string, principal *rest_model_zrok.Principal, trx *sqlx.Tx) (int, error) {
	envs, err := str.FindEnvironmentsForAccount(int(principal.ID), trx)
	if err != nil {
		logrus.Errorf("error finding environments for account '%v'", principal.Email)
		return 0, err
	}

	for _, env := range envs {
		if env.ZId == envZId {
			logrus.Debugf("found identity '%v' for account '%v'", envZId, principal.Email)
			return env.Id, nil
		}
	}

	logrus.Errorf("environment '%v' not found for account '%v'", envZId, principal.Email)
	return 0, errors.New("environment not found")
}

func (h *shareHandler) handleEnvironmentError(err error) middleware.Responder {
	if err.Error() == "environment not found" {
		return share.NewShareUnauthorized()
	}
	return share.NewShareInternalServerError()
}

func (h *shareHandler) processAccessGrants(params share.ShareParams, principal *rest_model_zrok.Principal, trx *sqlx.Tx) ([]int, middleware.Responder) {
	var accessGrantAcctIds []int
	if store.PermissionMode(params.Body.PermissionMode) == store.ClosedPermissionMode {
		for _, email := range params.Body.AccessGrants {
			acct, err := str.FindAccountWithEmail(email, trx)
			if err != nil {
				logrus.Errorf("unable to find account '%v' for share request from '%v'", email, principal.Email)
				return nil, share.NewShareNotFound()
			}
			logrus.Debugf("found id '%d' for '%v'", acct.Id, acct.Email)
			accessGrantAcctIds = append(accessGrantAcctIds, acct.Id)
		}
	}
	return accessGrantAcctIds, nil
}

func (h *shareHandler) createShareToken(reserved bool, uniqueName string, trx *sqlx.Tx) (string, middleware.Responder) {
	if !reserved || uniqueName == "" {
		token, err := createShareToken()
		if err != nil {
			logrus.Error(err)
			return "", share.NewShareInternalServerError()
		}
		return token, nil
	}

	if !util.IsValidUniqueName(uniqueName) {
		logrus.Errorf("invalid unique name '%v'", uniqueName)
		return "", share.NewShareUnprocessableEntity()
	}

	shareExists, err := str.ShareWithTokenExists(uniqueName, trx)
	if err != nil {
		logrus.Errorf("error checking share for token collision: %v", err)
		return "", share.NewUpdateShareInternalServerError()
	}
	if shareExists {
		logrus.Errorf("token '%v' already exists; cannot create share", uniqueName)
		return "", share.NewShareConflict()
	}

	return uniqueName, nil
}

func (h *shareHandler) allocateResources(params share.ShareParams, principal *rest_model_zrok.Principal, edge *rest_management_api_client.ZitiEdgeManagement, shrToken string, trx *sqlx.Tx) (string, []string, middleware.Responder) {
	var shrZId string
	var frontendEndpoints []string
	var err error

	switch params.Body.ShareMode {
	case string(sdk.PublicShareMode):
		shrZId, frontendEndpoints, err = h.allocatePublicResources(params, principal, edge, shrToken, trx)
	case string(sdk.PrivateShareMode):
		shrZId, frontendEndpoints, err = h.allocatePrivateResources(params, edge, shrToken)
	default:
		logrus.Errorf("unknown share mode '%v'", params.Body.ShareMode)
		return "", nil, share.NewShareInternalServerError()
	}

	if err != nil {
		logrus.Error(err)
		return "", nil, share.NewShareInternalServerError()
	}

	return shrZId, frontendEndpoints, nil
}

func (h *shareHandler) allocatePublicResources(params share.ShareParams, principal *rest_model_zrok.Principal, edge *rest_management_api_client.ZitiEdgeManagement, shrToken string, trx *sqlx.Tx) (string, []string, error) {
	if len(params.Body.FrontendSelection) < 1 {
		logrus.Info("no frontend selection provided")
		return "", nil, errors.New("no frontend selection")
	}

	var frontendZIds []string
	var frontendTemplates []string
	for _, frontendSelection := range params.Body.FrontendSelection {
		sfe, err := str.FindFrontendPubliclyNamed(frontendSelection, trx)
		if err != nil {
			return "", nil, err
		}
		if sfe.PermissionMode == store.ClosedPermissionMode {
			granted, err := str.IsFrontendGrantedToAccount(int(principal.ID), sfe.Id, trx)
			if err != nil {
				return "", nil, err
			}
			if !granted {
				return "", nil, errors.Errorf("'%v' is not granted access to frontend '%v'", principal.Email, frontendSelection)
			}
		}
		if sfe.UrlTemplate != nil {
			frontendZIds = append(frontendZIds, sfe.ZId)
			frontendTemplates = append(frontendTemplates, *sfe.UrlTemplate)
			logrus.Infof("added frontend selection '%v' with ziti identity '%v' for share '%v'", frontendSelection, sfe.ZId, shrToken)
		}
	}

	skipInterstitial, err := h.determineSkipInterstitial(params, principal, trx)
	if err != nil {
		return "", nil, err
	}

	logrus.Infof("allocating public resources for '%v'", shrToken)
	return newPublicResourceAllocator().allocate(params.Body.EnvZID, shrToken, frontendZIds, frontendTemplates, params, !skipInterstitial, edge)
}

func (h *shareHandler) determineSkipInterstitial(params share.ShareParams, principal *rest_model_zrok.Principal, trx *sqlx.Tx) (bool, error) {
	if sdk.BackendMode(params.Body.BackendMode) != sdk.DriveBackendMode {
		skipInterstitial, err := str.IsAccountGrantedSkipInterstitial(int(principal.ID), trx)
		if err != nil {
			logrus.Errorf("error checking skip interstitial for account '%v': %v", principal.Email, err)
			return false, err
		}
		return skipInterstitial, nil
	}
	return true, nil
}

func (h *shareHandler) allocatePrivateResources(params share.ShareParams, edge *rest_management_api_client.ZitiEdgeManagement, shrToken string) (string, []string, error) {
	return newPrivateResourceAllocator().allocate(params.Body.EnvZID, shrToken, params, edge)
}

func (h *shareHandler) createShareRecord(shrZId string, shrToken string, params share.ShareParams, frontendEndpoints []string) *store.Share {
	sshr := &store.Share{
		ZId:                  shrZId,
		Token:                shrToken,
		ShareMode:            params.Body.ShareMode,
		BackendMode:          params.Body.BackendMode,
		BackendProxyEndpoint: &params.Body.BackendProxyEndpoint,
		Reserved:             params.Body.Reserved,
		UniqueName:           params.Body.Reserved && params.Body.UniqueName != "",
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
	return sshr
}

func (h *shareHandler) saveShareAndGrants(sshr *store.Share, envId int, accessGrantAcctIds []int, trx *sqlx.Tx) (int, middleware.Responder) {
	sid, err := str.CreateShare(envId, sshr, trx)
	if err != nil {
		logrus.Errorf("error creating share record: %v", err)
		return 0, share.NewShareInternalServerError()
	}

	if sshr.PermissionMode == store.ClosedPermissionMode {
		for _, acctId := range accessGrantAcctIds {
			_, err := str.CreateAccessGrant(sid, acctId, trx)
			if err != nil {
				logrus.Errorf("error creating access grant: %v", err)
				return 0, share.NewShareInternalServerError()
			}
		}
	}

	return sid, nil
}

func (h *shareHandler) handleAuthSecrets(params share.ShareParams, sid int, sshr *store.Share, trx *sqlx.Tx) middleware.Responder {
	if sshr.ShareMode == string(sdk.PublicShareMode) && params.Body.AuthScheme == string(sdk.Basic) {
		logrus.Infof("writing basic auth secrets for '%v'", sshr.Token)
		authUsersMap := make(map[string]string)
		for _, authUser := range params.Body.AuthUsers {
			authUsersMap[authUser.Username] = authUser.Password
		}
		authUsersMapJson, err := json.Marshal(authUsersMap)
		if err != nil {
			logrus.Errorf("error marshalling auth secrets for '%v': %v", sshr.Token, err)
			return share.NewShareInternalServerError()
		}
		secrets := store.Secrets{
			ShareId: sid,
			Secrets: []store.Secret{
				{Key: "auth_scheme", Value: string(sdk.Basic)},
				{Key: "auth_users", Value: string(authUsersMapJson)},
			},
		}
		if err := str.CreateSecrets(secrets, trx); err != nil {
			logrus.Errorf("error creating secrets: %v", err)
			return share.NewShareInternalServerError()
		}
		logrus.Infof("wrote auth secrets for '%v'", sshr.Token)
	}
	return nil
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
