package controller

import (
	"database/sql"
	"strings"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/account"
)

func mfaAuthenticateHandler(params account.MfaAuthenticateParams) middleware.Responder {
	if params.Body.PendingToken == "" || params.Body.Code == "" {
		dl.Error("missing pending token or code")
		return account.NewMfaAuthenticateUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return account.NewMfaAuthenticateInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	// Find the pending auth
	pendingAuth, err := str.FindMfaPendingAuthByToken(params.Body.PendingToken, trx)
	if err != nil {
		if err == sql.ErrNoRows {
			dl.Warn("pending auth token not found")
		} else {
			dl.Errorf("error finding pending auth: %v", err)
		}
		return account.NewMfaAuthenticateUnauthorized()
	}

	// Check if expired
	if time.Now().After(pendingAuth.ExpiresAt) {
		dl.Warn("pending auth token expired")
		// Clean up expired token
		_ = str.DeleteMfaPendingAuth(pendingAuth.Id, trx)
		_ = trx.Commit()
		return account.NewMfaAuthenticateUnauthorized()
	}

	// Get the account
	acct, err := str.GetAccount(pendingAuth.AccountId, trx)
	if err != nil {
		dl.Errorf("error finding account for pending auth: %v", err)
		return account.NewMfaAuthenticateUnauthorized()
	}

	// Get MFA config
	mfa, err := str.FindAccountMfaByAccountId(acct.Id, trx)
	if err != nil {
		dl.Errorf("error finding MFA for account '%v': %v", acct.Email, err)
		return account.NewMfaAuthenticateUnauthorized()
	}

	if !mfa.Enabled {
		dl.Warnf("MFA not enabled for account '%v'", acct.Email)
		return account.NewMfaAuthenticateUnauthorized()
	}

	// Try to validate as TOTP code first
	var validCode bool
	if cfg.Mfa != nil && cfg.Mfa.SecretKey != "" {
		secret, err := DecryptTotpSecret(mfa.TotpSecret, cfg.Mfa.SecretKey)
		if err != nil {
			dl.Errorf("error decrypting TOTP secret for account '%v': %v", acct.Email, err)
			return account.NewMfaAuthenticateInternalServerError()
		}
		validCode = ValidateTotpCode(secret, params.Body.Code)
	}

	// If not valid as TOTP, try as recovery code
	if !validCode {
		// Normalize the code and hash it
		normalizedCode := strings.ToUpper(strings.ReplaceAll(params.Body.Code, "-", ""))
		codeHash := HashRecoveryCode(normalizedCode)

		recoveryCode, err := str.FindMfaRecoveryCodeByHash(mfa.Id, codeHash, trx)
		if err == nil && recoveryCode != nil {
			// Mark the recovery code as used
			if err := str.MarkMfaRecoveryCodeUsed(recoveryCode.Id, trx); err != nil {
				dl.Errorf("error marking recovery code as used: %v", err)
				return account.NewMfaAuthenticateInternalServerError()
			}
			validCode = true
			dl.Infof("recovery code used for account '%v'", acct.Email)
		}
	}

	if !validCode {
		dl.Warnf("invalid MFA code for account '%v'", acct.Email)
		return account.NewMfaAuthenticateUnauthorized()
	}

	// Delete the pending auth token (it's been used)
	if err := str.DeleteMfaPendingAuth(pendingAuth.Id, trx); err != nil {
		dl.Errorf("error deleting pending auth: %v", err)
		return account.NewMfaAuthenticateInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing transaction: %v", err)
		return account.NewMfaAuthenticateInternalServerError()
	}

	dl.Infof("MFA authentication successful for '%v'", acct.Email)
	return account.NewMfaAuthenticateOK().WithPayload(acct.Token)
}
