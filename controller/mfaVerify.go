package controller

import (
	"database/sql"

	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/account"
)

func mfaVerifyHandler(params account.MfaVerifyParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if cfg.Mfa == nil || cfg.Mfa.SecretKey == "" {
		dl.Error("MFA is not configured (missing secret key)")
		return account.NewMfaVerifyInternalServerError()
	}

	if params.Body.Code == "" {
		dl.Error("missing code in MFA verify request")
		return account.NewMfaVerifyBadRequest()
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return account.NewMfaVerifyInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	acct, err := str.GetAccount(int(principal.ID), trx)
	if err != nil {
		dl.Errorf("error finding account: %v", err)
		return account.NewMfaVerifyUnauthorized()
	}

	// Get the pending MFA setup
	mfa, err := str.FindAccountMfaByAccountId(acct.Id, trx)
	if err != nil {
		if err == sql.ErrNoRows {
			dl.Warnf("no pending MFA setup for account '%v'", acct.Email)
			return account.NewMfaVerifyBadRequest()
		}
		dl.Errorf("error finding MFA for account '%v': %v", acct.Email, err)
		return account.NewMfaVerifyInternalServerError()
	}

	if mfa.Enabled {
		dl.Warnf("MFA already enabled for account '%v'", acct.Email)
		return account.NewMfaVerifyBadRequest()
	}

	// Decrypt the secret and validate the code
	secret, err := DecryptTotpSecret(mfa.TotpSecret, cfg.Mfa.SecretKey)
	if err != nil {
		dl.Errorf("error decrypting TOTP secret for account '%v': %v", acct.Email, err)
		return account.NewMfaVerifyInternalServerError()
	}

	if !ValidateTotpCode(secret, params.Body.Code) {
		dl.Warnf("invalid TOTP code for account '%v'", acct.Email)
		return account.NewMfaVerifyBadRequest()
	}

	// Enable MFA
	mfa.Enabled = true
	if err := str.UpdateAccountMfa(mfa, trx); err != nil {
		dl.Errorf("error enabling MFA for account '%v': %v", acct.Email, err)
		return account.NewMfaVerifyInternalServerError()
	}

	// Generate recovery codes
	codes, hashes, err := GenerateRecoveryCodes()
	if err != nil {
		dl.Errorf("error generating recovery codes for account '%v': %v", acct.Email, err)
		return account.NewMfaVerifyInternalServerError()
	}

	// Delete any existing recovery codes and store new ones
	if err := str.DeleteMfaRecoveryCodesByAccountMfaId(mfa.Id, trx); err != nil {
		dl.Errorf("error deleting old recovery codes for account '%v': %v", acct.Email, err)
		return account.NewMfaVerifyInternalServerError()
	}
	if err := str.CreateMfaRecoveryCodes(mfa.Id, hashes, trx); err != nil {
		dl.Errorf("error storing recovery codes for account '%v': %v", acct.Email, err)
		return account.NewMfaVerifyInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing transaction: %v", err)
		return account.NewMfaVerifyInternalServerError()
	}

	dl.Infof("MFA enabled for account '%v'", acct.Email)

	return account.NewMfaVerifyOK().WithPayload(&account.MfaVerifyOKBody{
		RecoveryCodes: codes,
	})
}
