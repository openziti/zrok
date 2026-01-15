package controller

import (
	"database/sql"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/account"
)

func mfaChallengeHandler(params account.MfaChallengeParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if params.Body.Code == "" {
		dl.Error("missing code in MFA challenge request")
		return account.NewMfaChallengeBadRequest()
	}

	if cfg.Mfa == nil || cfg.Mfa.SecretKey == "" {
		dl.Error("MFA is not configured (missing secret key)")
		return account.NewMfaChallengeInternalServerError()
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return account.NewMfaChallengeInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	acct, err := str.GetAccount(int(principal.ID), trx)
	if err != nil {
		dl.Errorf("error finding account: %v", err)
		return account.NewMfaChallengeUnauthorized()
	}

	// Get MFA config
	mfa, err := str.FindAccountMfaByAccountId(acct.Id, trx)
	if err != nil {
		if err == sql.ErrNoRows {
			dl.Warnf("MFA not configured for account '%v'", acct.Email)
			return account.NewMfaChallengeBadRequest()
		}
		dl.Errorf("error finding MFA for account '%v': %v", acct.Email, err)
		return account.NewMfaChallengeInternalServerError()
	}

	if !mfa.Enabled {
		dl.Warnf("MFA not enabled for account '%v'", acct.Email)
		return account.NewMfaChallengeBadRequest()
	}

	// Validate the TOTP code
	secret, err := DecryptTotpSecret(mfa.TotpSecret, cfg.Mfa.SecretKey)
	if err != nil {
		dl.Errorf("error decrypting TOTP secret for account '%v': %v", acct.Email, err)
		return account.NewMfaChallengeInternalServerError()
	}

	if !ValidateTotpCode(secret, params.Body.Code) {
		dl.Warnf("invalid TOTP code for account '%v' during MFA challenge", acct.Email)
		return account.NewMfaChallengeBadRequest()
	}

	// Generate challenge token
	challengeToken, err := GenerateSecureToken(32)
	if err != nil {
		dl.Errorf("error generating challenge token for '%v': %v", acct.Email, err)
		return account.NewMfaChallengeInternalServerError()
	}

	expiration := time.Minute * 10
	if cfg.Mfa.ChallengeExpiration > 0 {
		expiration = cfg.Mfa.ChallengeExpiration
	}
	expiresAt := time.Now().Add(expiration)

	_, err = str.CreateMfaChallengeToken(&store.MfaChallengeToken{
		AccountId:      acct.Id,
		ChallengeToken: challengeToken,
		ExpiresAt:      expiresAt,
	}, trx)
	if err != nil {
		dl.Errorf("error creating challenge token for '%v': %v", acct.Email, err)
		return account.NewMfaChallengeInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing transaction: %v", err)
		return account.NewMfaChallengeInternalServerError()
	}

	dl.Infof("MFA challenge token issued for account '%v'", acct.Email)

	return account.NewMfaChallengeOK().WithPayload(&account.MfaChallengeOKBody{
		ChallengeToken: challengeToken,
		ExpiresAt:      strfmt.DateTime(expiresAt),
	})
}
