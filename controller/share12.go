package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/share"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type share12Handler struct{}

func newShare12Handler() *share12Handler {
	return &share12Handler{}
}

func (h *share12Handler) Handle(params share.Share12Params, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return share.NewShare12InternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	// validate environment
	envZId := params.Body.EnvZID
	envId, err := h.validateEnvironment(envZId, principal, trx)
	if err != nil {
		logrus.Errorf("environment validation failed: %v", err)
		return share.NewShare12Unauthorized()
	}

	// check limits
	if err := h.checkLimits(envId, principal, params, trx); err != nil {
		logrus.Errorf("limits error: %v", err)
		return share.NewShare12Unauthorized()
	}

	// process namespace selections
	frontendEndpoints, err := h.processNamespaceSelections(params.Body.NamespaceSelections, principal, trx)
	if err != nil {
		logrus.Errorf("namespace selection processing failed: %v", err)
		return share.NewShare12NotFound()
	}

	// create share token
	shrToken, err := createShareToken()
	if err != nil {
		logrus.Error(err)
		return share.NewShare12InternalServerError()
	}

	// allocate resources based on share mode
	var shrZId string
	switch params.Body.ShareMode {
	case "public":
		shrZId, frontendEndpoints, err = h.allocatePublicResources(envZId, shrToken, params, trx)
	case "private":
		shrZId, frontendEndpoints, err = h.allocatePrivateResources(envZId, shrToken, params, trx)
	default:
		logrus.Errorf("unknown share mode '%v'", params.Body.ShareMode)
		return share.NewShare12InternalServerError()
	}

	if err != nil {
		logrus.Error(err)
		return share.NewShare12InternalServerError()
	}

	// create share record
	shareId, err := h.createShareRecord(envId, shrZId, shrToken, params, frontendEndpoints, trx)
	if err != nil {
		logrus.Errorf("error creating share record: %v", err)
		return share.NewShare12InternalServerError()
	}

	// handle access grants if closed permission mode
	if err := h.processAccessGrants(shareId, params.Body.AccessGrants, params.Body.PermissionMode, principal, trx); err != nil {
		logrus.Errorf("error processing access grants: %v", err)
		return share.NewShare12InternalServerError()
	}

	if err := trx.Commit(); err != nil {
		logrus.Errorf("error committing share record: %v", err)
		return share.NewShare12InternalServerError()
	}

	logrus.Infof("recorded share '%v' with id '%v' for '%v'", shrToken, shareId, principal.Email)

	return share.NewShare12Created().WithPayload(&rest_model_zrok.ShareResponse{
		FrontendProxyEndpoints: frontendEndpoints,
		ShareToken:             shrToken,
	})
}

func (h *share12Handler) validateEnvironment(envZId string, principal *rest_model_zrok.Principal, trx *sqlx.Tx) (int, error) {
	env, err := str.FindEnvironmentForAccount(envZId, int(principal.ID), trx)
	if err != nil {
		return 0, errors.Wrapf(err, "error finding environment '%v' for account '%v'", envZId, principal.Email)
	}
	return env.Id, nil
}

func (h *share12Handler) checkLimits(envId int, principal *rest_model_zrok.Principal, params share.Share12Params, trx *sqlx.Tx) error {
	if !principal.Limitless {
		if limitsAgent != nil {
			shareMode := sdk.ShareMode(params.Body.ShareMode)
			backendMode := sdk.BackendMode(params.Body.BackendMode)

			// we're going to skip reservation checking because we're moving name creation outside the scope of share
			// creation. the limits check for name creation will happen in the `/share/name` endpoint instead.
			ok, err := limitsAgent.CanCreateShare(int(principal.ID), envId, false, false, shareMode, backendMode, trx)
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

func (h *share12Handler) processNamespaceSelections(selections []*rest_model_zrok.NamespaceSelection, principal *rest_model_zrok.Principal, trx interface{}) ([]string, error) {
	// TODO: implement namespace selection processing
	// this is the key difference from the original share endpoint
	return nil, nil
}

func (h *share12Handler) allocatePublicResources(envZId, shrToken string, params share.Share12Params, trx interface{}) (string, []string, error) {
	// TODO: implement public resource allocation for share12
	return "", nil, nil
}

func (h *share12Handler) allocatePrivateResources(envZId, shrToken string, params share.Share12Params, trx interface{}) (string, []string, error) {
	// TODO: implement private resource allocation for share12
	return "", nil, nil
}

func (h *share12Handler) createShareRecord(envId int, shrZId, shrToken string, params share.Share12Params, frontendEndpoints []string, trx interface{}) (int, error) {
	// TODO: implement share record creation with new share12 fields
	// note: target instead of backendProxyEndpoint, basicAuthUsers instead of authUsers, etc.
	return 0, nil
}

func (h *share12Handler) processAccessGrants(shareId int, accessGrants []string, permissionMode string, principal *rest_model_zrok.Principal, trx interface{}) error {
	// TODO: implement access grants processing similar to existing share handler
	return nil
}
