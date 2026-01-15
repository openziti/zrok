package controller

import (
	"database/sql"

	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/account"
)

func mfaStatusHandler(params account.MfaStatusParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return account.NewMfaStatusInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	acct, err := str.GetAccount(int(principal.ID), trx)
	if err != nil {
		dl.Errorf("error finding account: %v", err)
		return account.NewMfaStatusUnauthorized()
	}

	mfa, err := str.FindAccountMfaByAccountId(acct.Id, trx)
	if err != nil {
		if err == sql.ErrNoRows {
			// No MFA record means MFA is not enabled
			return account.NewMfaStatusOK().WithPayload(&account.MfaStatusOKBody{
				Enabled:                 false,
				RecoveryCodesRemaining: 0,
			})
		}
		dl.Errorf("error finding MFA for account '%v': %v", acct.Email, err)
		return account.NewMfaStatusInternalServerError()
	}

	if !mfa.Enabled {
		return account.NewMfaStatusOK().WithPayload(&account.MfaStatusOKBody{
			Enabled:                 false,
			RecoveryCodesRemaining: 0,
		})
	}

	// Count remaining recovery codes
	remaining, err := str.CountUnusedMfaRecoveryCodes(mfa.Id, trx)
	if err != nil {
		dl.Errorf("error counting recovery codes for account '%v': %v", acct.Email, err)
		return account.NewMfaStatusInternalServerError()
	}

	return account.NewMfaStatusOK().WithPayload(&account.MfaStatusOKBody{
		Enabled:                 true,
		RecoveryCodesRemaining: int64(remaining),
	})
}
