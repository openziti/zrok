package controller

import (
	"database/sql"

	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/account"
)

func mfaRecoveryCodesHandler(params account.MfaRecoveryCodesParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if params.Body.Code == "" {
		dl.Error("missing code in MFA recovery codes request")
		return account.NewMfaRecoveryCodesBadRequest()
	}

	if cfg.Mfa == nil || cfg.Mfa.SecretKey == "" {
		dl.Error("MFA is not configured (missing secret key)")
		return account.NewMfaRecoveryCodesInternalServerError()
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return account.NewMfaRecoveryCodesInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	acct, err := str.GetAccount(int(principal.ID), trx)
	if err != nil {
		dl.Errorf("error finding account: %v", err)
		return account.NewMfaRecoveryCodesUnauthorized()
	}

	// Get MFA config
	mfa, err := str.FindAccountMfaByAccountId(acct.Id, trx)
	if err != nil {
		if err == sql.ErrNoRows {
			dl.Warnf("MFA not configured for account '%v'", acct.Email)
			return account.NewMfaRecoveryCodesBadRequest()
		}
		dl.Errorf("error finding MFA for account '%v': %v", acct.Email, err)
		return account.NewMfaRecoveryCodesInternalServerError()
	}

	if !mfa.Enabled {
		dl.Warnf("MFA not enabled for account '%v'", acct.Email)
		return account.NewMfaRecoveryCodesBadRequest()
	}

	// Validate the TOTP code
	secret, err := DecryptTotpSecret(mfa.TotpSecret, cfg.Mfa.SecretKey)
	if err != nil {
		dl.Errorf("error decrypting TOTP secret for account '%v': %v", acct.Email, err)
		return account.NewMfaRecoveryCodesInternalServerError()
	}

	if !ValidateTotpCode(secret, params.Body.Code) {
		dl.Warnf("invalid TOTP code for account '%v' during recovery code regeneration", acct.Email)
		return account.NewMfaRecoveryCodesBadRequest()
	}

	// Generate new recovery codes
	codes, hashes, err := GenerateRecoveryCodes()
	if err != nil {
		dl.Errorf("error generating recovery codes for account '%v': %v", acct.Email, err)
		return account.NewMfaRecoveryCodesInternalServerError()
	}

	// Delete old recovery codes and store new ones
	if err := str.DeleteMfaRecoveryCodesByAccountMfaId(mfa.Id, trx); err != nil {
		dl.Errorf("error deleting old recovery codes for account '%v': %v", acct.Email, err)
		return account.NewMfaRecoveryCodesInternalServerError()
	}
	if err := str.CreateMfaRecoveryCodes(mfa.Id, hashes, trx); err != nil {
		dl.Errorf("error storing recovery codes for account '%v': %v", acct.Email, err)
		return account.NewMfaRecoveryCodesInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing transaction: %v", err)
		return account.NewMfaRecoveryCodesInternalServerError()
	}

	dl.Infof("regenerated recovery codes for account '%v'", acct.Email)

	return account.NewMfaRecoveryCodesOK().WithPayload(&account.MfaRecoveryCodesOKBody{
		RecoveryCodes: codes,
	})
}
