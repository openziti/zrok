package controller

import (
	"database/sql"

	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/account"
)

func mfaSetupHandler(params account.MfaSetupParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if cfg.Mfa == nil || cfg.Mfa.SecretKey == "" {
		dl.Error("MFA is not configured (missing secret key)")
		return account.NewMfaSetupInternalServerError()
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return account.NewMfaSetupInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	acct, err := str.GetAccount(int(principal.ID), trx)
	if err != nil {
		dl.Errorf("error finding account: %v", err)
		return account.NewMfaSetupUnauthorized()
	}

	// Check if MFA is already enabled
	existingMfa, err := str.FindAccountMfaByAccountId(acct.Id, trx)
	if err != nil && err != sql.ErrNoRows {
		dl.Errorf("error checking existing MFA for account '%v': %v", acct.Email, err)
		return account.NewMfaSetupInternalServerError()
	}
	if existingMfa != nil && existingMfa.Enabled {
		dl.Warnf("MFA already enabled for account '%v'", acct.Email)
		return account.NewMfaSetupBadRequest()
	}

	// Generate new TOTP secret
	totpSetup, err := GenerateTotpSecret(acct.Email)
	if err != nil {
		dl.Errorf("error generating TOTP secret for account '%v': %v", acct.Email, err)
		return account.NewMfaSetupInternalServerError()
	}

	// Encrypt the secret for storage
	encryptedSecret, err := EncryptTotpSecret(totpSetup.Secret, cfg.Mfa.SecretKey)
	if err != nil {
		dl.Errorf("error encrypting TOTP secret for account '%v': %v", acct.Email, err)
		return account.NewMfaSetupInternalServerError()
	}

	// Create or update the MFA record (not enabled yet - will be enabled after verification)
	if existingMfa != nil {
		existingMfa.TotpSecret = encryptedSecret
		existingMfa.Enabled = false
		if err := str.UpdateAccountMfa(existingMfa, trx); err != nil {
			dl.Errorf("error updating MFA record for account '%v': %v", acct.Email, err)
			return account.NewMfaSetupInternalServerError()
		}
	} else {
		_, err = str.CreateAccountMfa(&store.AccountMfa{
			AccountId:  acct.Id,
			TotpSecret: encryptedSecret,
			Enabled:    false,
		}, trx)
		if err != nil {
			dl.Errorf("error creating MFA record for account '%v': %v", acct.Email, err)
			return account.NewMfaSetupInternalServerError()
		}
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing transaction: %v", err)
		return account.NewMfaSetupInternalServerError()
	}

	dl.Infof("MFA setup initiated for account '%v'", acct.Email)

	return account.NewMfaSetupOK().WithPayload(&account.MfaSetupOKBody{
		Secret:          totpSetup.Secret,
		QrCode:          totpSetup.QrCode,
		ProvisioningURI: totpSetup.ProvisioningUri,
	})
}
