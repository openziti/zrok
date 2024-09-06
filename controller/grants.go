package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
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

	edge, err := zrokEdgeSdk.Client(cfg.Ziti)
	if err != nil {
		logrus.Errorf("error connecting to ziti: %v", err)
		return admin.NewGrantsInternalServerError()
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

	for _, env := range envs {
		shrs, err := str.FindSharesForEnvironment(env.Id, trx)
		if err != nil {
			logrus.Errorf("error finding shares for '%v': %v", acct.Email, err)
			return admin.NewGrantsInternalServerError()
		}

		for _, shr := range shrs {
			if shr.ShareMode == string(sdk.PublicShareMode) && shr.BackendMode != string(sdk.DriveBackendMode) {
				cfgZId, shrCfg, err := zrokEdgeSdk.GetConfig(shr.Token, edge)
				if err != nil {
					logrus.Errorf("error getting config for share '%v': %v", shr.Token, err)
					return admin.NewGrantsInternalServerError()
				}

				if shrCfg.Interstitial != !acctSkipInterstitial {
					shrCfg.Interstitial = !acctSkipInterstitial
					err := zrokEdgeSdk.UpdateConfig(shr.Token, cfgZId, shrCfg, edge)
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
