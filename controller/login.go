package controller

import (
	"database/sql"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/account"
)

func loginHandler(params account.LoginParams) middleware.Responder {
	if params.Body.Email == "" || params.Body.Password == "" {
		dl.Errorf("missing email or password")
		return account.NewLoginUnauthorized()
	}

	dl.Infof("received login request for email '%v'", params.Body.Email)

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return account.NewLoginUnauthorized()
	}
	defer func() { _ = trx.Rollback() }()
	a, err := str.FindAccountWithEmail(params.Body.Email, trx)
	if err != nil {
		dl.Errorf("error finding account '%v': %v", params.Body.Email, err)
		return account.NewLoginUnauthorized()
	}
	hpwd, err := rehashPassword(params.Body.Password, a.Salt)
	if err != nil {
		dl.Errorf("error hashing password for '%v': %v", params.Body.Email, err)
		return account.NewLoginUnauthorized()
	}
	if a.Password != hpwd.Password {
		dl.Errorf("password mismatch for account '%v'", params.Body.Email)
		return account.NewLoginUnauthorized()
	}

	// Check if MFA is enabled for this account
	mfa, err := str.FindAccountMfaByAccountId(a.Id, trx)
	if err != nil && err != sql.ErrNoRows {
		dl.Errorf("error checking MFA for account '%v': %v", params.Body.Email, err)
		return account.NewLoginUnauthorized()
	}

	// If MFA is required globally but user hasn't set it up, return 403
	if cfg.Mfa != nil && cfg.Mfa.Required {
		if mfa == nil || !mfa.Enabled {
			dl.Warnf("MFA required but not enabled for account '%v'", params.Body.Email)
			return account.NewLoginForbidden()
		}
	}

	// If MFA is enabled for this account, create a pending auth and return 202
	if mfa != nil && mfa.Enabled {
		pendingToken, err := GenerateSecureToken(32)
		if err != nil {
			dl.Errorf("error generating pending token for '%v': %v", params.Body.Email, err)
			return account.NewLoginUnauthorized()
		}

		expiration := time.Minute * 5
		if cfg.Mfa != nil && cfg.Mfa.PendingAuthExpiration > 0 {
			expiration = cfg.Mfa.PendingAuthExpiration
		}

		_, err = str.CreateMfaPendingAuth(&store.MfaPendingAuth{
			AccountId:    a.Id,
			PendingToken: pendingToken,
			ExpiresAt:    time.Now().Add(expiration),
		}, trx)
		if err != nil {
			dl.Errorf("error creating pending auth for '%v': %v", params.Body.Email, err)
			return account.NewLoginUnauthorized()
		}

		if err := trx.Commit(); err != nil {
			dl.Errorf("error committing transaction: %v", err)
			return account.NewLoginUnauthorized()
		}

		dl.Infof("MFA required for login '%v'", params.Body.Email)
		return account.NewLoginAccepted().WithPayload(&account.LoginAcceptedBody{
			PendingToken: pendingToken,
		})
	}

	return account.NewLoginOK().WithPayload(a.Token)
}
