package controller

import (
	"database/sql"

	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/account"
)

func mfaDisableHandler(params account.MfaDisableParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if params.Body.Password == "" || params.Body.Code == "" {
		dl.Error("missing password or code in MFA disable request")
		return account.NewMfaDisableBadRequest()
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return account.NewMfaDisableInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	acct, err := str.GetAccount(int(principal.ID), trx)
	if err != nil {
		dl.Errorf("error finding account: %v", err)
		return account.NewMfaDisableUnauthorized()
	}

	// Verify password
	hpwd, err := rehashPassword(params.Body.Password, acct.Salt)
	if err != nil {
		dl.Errorf("error hashing password for '%v': %v", acct.Email, err)
		return account.NewMfaDisableBadRequest()
	}
	if acct.Password != hpwd.Password {
		dl.Warnf("password mismatch for account '%v' during MFA disable", acct.Email)
		return account.NewMfaDisableBadRequest()
	}

	// Get MFA config
	mfa, err := str.FindAccountMfaByAccountId(acct.Id, trx)
	if err != nil {
		if err == sql.ErrNoRows {
			dl.Warnf("MFA not configured for account '%v'", acct.Email)
			return account.NewMfaDisableBadRequest()
		}
		dl.Errorf("error finding MFA for account '%v': %v", acct.Email, err)
		return account.NewMfaDisableInternalServerError()
	}

	if !mfa.Enabled {
		dl.Warnf("MFA not enabled for account '%v'", acct.Email)
		return account.NewMfaDisableBadRequest()
	}

	// Validate the TOTP code
	if cfg.Mfa == nil || cfg.Mfa.SecretKey == "" {
		dl.Error("MFA is not configured (missing secret key)")
		return account.NewMfaDisableInternalServerError()
	}

	secret, err := DecryptTotpSecret(mfa.TotpSecret, cfg.Mfa.SecretKey)
	if err != nil {
		dl.Errorf("error decrypting TOTP secret for account '%v': %v", acct.Email, err)
		return account.NewMfaDisableInternalServerError()
	}

	if !ValidateTotpCode(secret, params.Body.Code) {
		dl.Warnf("invalid TOTP code for account '%v' during MFA disable", acct.Email)
		return account.NewMfaDisableBadRequest()
	}

	// Delete the MFA record (this will cascade delete recovery codes)
	if err := str.DeleteAccountMfa(mfa.Id, trx); err != nil {
		dl.Errorf("error deleting MFA for account '%v': %v", acct.Email, err)
		return account.NewMfaDisableInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing transaction: %v", err)
		return account.NewMfaDisableInternalServerError()
	}

	dl.Infof("MFA disabled for account '%v'", acct.Email)
	return account.NewMfaDisableOK()
}
