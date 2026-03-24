package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/automation"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/admin"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
)

// getSkipInterstitialGrant

type getSkipInterstitialGrantHandler struct{}

func newGetSkipInterstitialGrantHandler() *getSkipInterstitialGrantHandler {
	return &getSkipInterstitialGrantHandler{}
}

func (h *getSkipInterstitialGrantHandler) Handle(params admin.GetSkipInterstitialGrantParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		dl.Error("invalid admin principal")
		return admin.NewGetSkipInterstitialGrantUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return admin.NewGetSkipInterstitialGrantInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	acct, err := str.FindAccountWithEmail(params.Email, trx)
	if err != nil {
		dl.Errorf("error finding account with email '%v': %v", params.Email, err)
		return admin.NewGetSkipInterstitialGrantNotFound()
	}

	granted, err := str.IsAccountGrantedSkipInterstitial(acct.Id, trx)
	if err != nil {
		dl.Errorf("error checking skip interstitial grant for '%v': %v", params.Email, err)
		return admin.NewGetSkipInterstitialGrantInternalServerError()
	}

	return admin.NewGetSkipInterstitialGrantOK().WithPayload(&admin.GetSkipInterstitialGrantOKBody{
		Email:   acct.Email,
		Granted: granted,
	})
}

// grantSkipInterstitial

type grantSkipInterstitialHandler struct{}

func newGrantSkipInterstitialHandler() *grantSkipInterstitialHandler {
	return &grantSkipInterstitialHandler{}
}

func (h *grantSkipInterstitialHandler) Handle(params admin.GrantSkipInterstitialParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		dl.Error("invalid admin principal")
		return admin.NewGrantSkipInterstitialUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return admin.NewGrantSkipInterstitialInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	acct, err := str.FindAccountWithEmail(params.Body.Email, trx)
	if err != nil {
		dl.Errorf("error finding account with email '%v': %v", params.Body.Email, err)
		return admin.NewGrantSkipInterstitialNotFound()
	}

	if err := str.GrantSkipInterstitial(acct.Id, trx); err != nil {
		dl.Errorf("error granting skip interstitial for '%v': %v", params.Body.Email, err)
		return admin.NewGrantSkipInterstitialInternalServerError()
	}

	if err := syncSkipInterstitialForAccount(acct, true); err != nil {
		dl.Errorf("error syncing skip interstitial for '%v': %v", params.Body.Email, err)
		return admin.NewGrantSkipInterstitialInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing transaction: %v", err)
		return admin.NewGrantSkipInterstitialInternalServerError()
	}

	return admin.NewGrantSkipInterstitialOK()
}

// revokeSkipInterstitial

type revokeSkipInterstitialHandler struct{}

func newRevokeSkipInterstitialHandler() *revokeSkipInterstitialHandler {
	return &revokeSkipInterstitialHandler{}
}

func (h *revokeSkipInterstitialHandler) Handle(params admin.RevokeSkipInterstitialParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		dl.Error("invalid admin principal")
		return admin.NewRevokeSkipInterstitialUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return admin.NewRevokeSkipInterstitialInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	acct, err := str.FindAccountWithEmail(params.Body.Email, trx)
	if err != nil {
		dl.Errorf("error finding account with email '%v': %v", params.Body.Email, err)
		return admin.NewRevokeSkipInterstitialNotFound()
	}

	if err := str.RevokeSkipInterstitial(acct.Id, trx); err != nil {
		dl.Errorf("error revoking skip interstitial for '%v': %v", params.Body.Email, err)
		return admin.NewRevokeSkipInterstitialInternalServerError()
	}

	if err := syncSkipInterstitialForAccount(acct, false); err != nil {
		dl.Errorf("error syncing skip interstitial for '%v': %v", params.Body.Email, err)
		return admin.NewRevokeSkipInterstitialInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing transaction: %v", err)
		return admin.NewRevokeSkipInterstitialInternalServerError()
	}

	return admin.NewRevokeSkipInterstitialOK()
}

// syncSkipInterstitialForAccount best-effort synchronizes the interstitial
// setting on existing public (non-drive) share Ziti configs for the given
// account. Failures while processing individual shares are logged and skipped,
// since affected shares can be recreated to correct transient issues.
func syncSkipInterstitialForAccount(acct *store.Account, skipInterstitial bool) error {
	trx, err := str.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = trx.Rollback() }()

	envs, err := str.FindEnvironmentsForAccount(acct.Id, trx)
	if err != nil {
		return err
	}

	ziti, err := automation.NewZitiAutomation(cfg.Ziti)
	if err != nil {
		return err
	}

	for _, env := range envs {
		shrs, err := str.FindSharesForEnvironment(env.Id, trx)
		if err != nil {
			return err
		}

		for _, shr := range shrs {
			if shr.ShareMode == string(sdk.PublicShareMode) && shr.BackendMode != string(sdk.DriveBackendMode) {
				filterOpts := &automation.FilterOptions{
					Filter: "tags.zrokShareToken=\"" + shr.Token + "\"",
					Limit:  0,
					Offset: 0,
				}
				configs, err := ziti.Configs.Find(filterOpts)
				if err != nil {
					dl.Errorf("error finding config for share '%v': %v", shr.Token, err)
					return err
				}
				if len(configs) != 1 {
					dl.Errorf("expected 1 configuration for share '%v', found %v", shr.Token, len(configs))
					continue
				}
				config := configs[0]
				if config.ConfigType.Name != sdk.ZrokProxyConfig {
					dl.Errorf("expected '%v' config type for share '%v', found '%v'", sdk.ZrokProxyConfig, shr.Token, config.ConfigType.Name)
					continue
				}

				v, ok := config.Data.(map[string]interface{})
				if !ok {
					dl.Errorf("unexpected config data type for share '%v'", shr.Token)
					continue
				}
				shrCfg, err := sdk.FrontendConfigFromMap(v)
				if err != nil {
					dl.Errorf("error parsing config data for share '%v': %v", shr.Token, err)
					continue
				}

				if shrCfg.Interstitial != !skipInterstitial {
					shrCfg.Interstitial = !skipInterstitial
					configOpts := &automation.ConfigOptions{
						BaseOptions: automation.BaseOptions{
							Name: shr.Token,
							Tags: automation.ZrokShareTags(shr.Token),
						},
						ConfigTypeID: config.ConfigType.ID,
						Data:         shrCfg,
					}
					if err := ziti.Configs.Update(*config.ID, configOpts); err != nil {
						dl.Errorf("error updating config for '%v': %v", shr.Token, err)
						return err
					}
				} else {
					dl.Infof("skipping config update for '%v'", shr.Token)
				}
			}
		}
	}

	return nil
}
