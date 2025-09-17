package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/controller/automation"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/sirupsen/logrus"
)

type grantsHandler struct{}

func newGrantsHandler() *grantsHandler {
	return &grantsHandler{}
}

func (h *grantsHandler) Handle(params admin.GrantsParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		logrus.Errorf("invalid admin principal")
		return admin.NewGrantsUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return admin.NewGrantsInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	acct, err := str.FindAccountWithEmail(params.Body.Email, trx)
	if err != nil {
		logrus.Errorf("error finding account with email '%v': %v", params.Body.Email, err)
		return admin.NewGrantsNotFound()
	}

	acctSkipInterstitial, err := str.IsAccountGrantedSkipInterstitial(acct.Id, trx)
	if err != nil {
		logrus.Errorf("error checking account '%v' granted skip interstitial: %v", acct.Email, err)
	}

	envs, err := str.FindEnvironmentsForAccount(acct.Id, trx)
	if err != nil {
		logrus.Errorf("error finding environments for '%v': %v", acct.Email, err)
		return admin.NewGrantsInternalServerError()
	}

	ziti, err := automation.NewZitiAutomation(cfg.Ziti)
	if err != nil {
		logrus.Errorf("error connecting to ziti: %v", err)
		return admin.NewGrantsInternalServerError()
	}

	for _, env := range envs {
		shrs, err := str.FindSharesForEnvironment(env.Id, trx)
		if err != nil {
			logrus.Errorf("error finding shares for '%v': %v", acct.Email, err)
			return admin.NewGrantsInternalServerError()
		}

		for _, shr := range shrs {
			if shr.ShareMode == string(sdk.PublicShareMode) && shr.BackendMode != string(sdk.DriveBackendMode) {
				// find config by zrokShareToken tag
				filterOpts := &automation.FilterOptions{
					Filter: "tags.zrokShareToken=\"" + shr.Token + "\"",
					Limit:  0,
					Offset: 0,
				}
				configs, err := ziti.Configs.Find(filterOpts)
				if err != nil {
					logrus.Errorf("error finding config for share '%v': %v", shr.Token, err)
					return admin.NewGrantsInternalServerError()
				}
				if len(configs) != 1 {
					logrus.Errorf("expected 1 configuration for share '%v', found %v", shr.Token, len(configs))
					return admin.NewGrantsInternalServerError()
				}
				config := configs[0]
				if config.ConfigType.Name != sdk.ZrokProxyConfig {
					logrus.Errorf("expected '%v' config type for share '%v', found '%v'", sdk.ZrokProxyConfig, shr.Token, config.ConfigType.Name)
					return admin.NewGrantsInternalServerError()
				}

				// parse the config data
				var shrCfg *sdk.FrontendConfig
				if v, ok := config.Data.(map[string]interface{}); ok {
					shrCfg, err = sdk.FrontendConfigFromMap(v)
					if err != nil {
						logrus.Errorf("error parsing config data for share '%v': %v", shr.Token, err)
						return admin.NewGrantsInternalServerError()
					}
				} else {
					logrus.Errorf("unexpected config data type for share '%v'", shr.Token)
					return admin.NewGrantsInternalServerError()
				}

				if shrCfg.Interstitial != !acctSkipInterstitial {
					shrCfg.Interstitial = !acctSkipInterstitial

					// update config using automation
					configOpts := &automation.ConfigOptions{
						BaseOptions: automation.BaseOptions{
							Name: shr.Token,
							Tags: automation.ZrokShareTags(shr.Token),
						},
						ConfigTypeID: config.ConfigType.ID,
						Data:         shrCfg,
					}
					err := ziti.Configs.Update(*config.ID, configOpts)
					if err != nil {
						logrus.Errorf("error updating config for '%v': %v", shr.Token, err)
						return admin.NewGrantsInternalServerError()
					}
				} else {
					logrus.Infof("skipping config update for '%v'", shr.Token)
				}
			} else {
				logrus.Debugf("skipping share mode %v, backend mode %v", shr.ShareMode, shr.BackendMode)
			}
		}
	}

	return admin.NewGrantsOK()
}
