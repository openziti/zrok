package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/controller/config"

	"github.com/pkg/errors"
)

type maintenanceRegistrationAgent struct {
	cfg *config.RegistrationMaintenanceConfig
	ctx context.Context
}

func newRegistrationMaintenanceAgent(ctx context.Context, cfg *config.RegistrationMaintenanceConfig) *maintenanceRegistrationAgent {
	return &maintenanceRegistrationAgent{
		cfg: cfg,
		ctx: ctx,
	}
}

func (ma *maintenanceRegistrationAgent) run() {
	dl.Infof("started")
	defer dl.Info("exited")

	ticker := time.NewTicker(ma.cfg.CheckFrequency)
	for {
		select {
		case <-ma.ctx.Done():
			{
				ticker.Stop()
				return
			}
		case <-ticker.C:
			{
				if err := ma.deleteExpiredAccountRequests(); err != nil {
					dl.Error(err)
				}
			}
		}
	}
}

func (ma *maintenanceRegistrationAgent) deleteExpiredAccountRequests() error {
	trx, err := str.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = trx.Rollback() }()

	timeout := time.Now().UTC().Add(-ma.cfg.ExpirationTimeout)
	accountRequests, err := str.FindExpiredAccountRequests(timeout, ma.cfg.BatchLimit, trx)
	if err != nil {
		return errors.Wrapf(err, "error finding expire account requests before %v", timeout)
	}
	if len(accountRequests) > 0 {
		dl.Infof("found %d expired account requests to remove", len(accountRequests))
		acctStrings := make([]string, len(accountRequests))
		ids := make([]int, len(accountRequests))
		for i, acct := range accountRequests {
			ids[i] = acct.Id
			acctStrings[i] = fmt.Sprintf("{%d:%s}", acct.Id, acct.Email)
		}

		dl.Infof("deleting expired account requests: %v", strings.Join(acctStrings, ","))
		if err := str.DeleteMultipleAccountRequests(ids, trx); err != nil {
			return errors.Wrapf(err, "error deleting expired account requests before %v", timeout)
		}
		if err := trx.Commit(); err != nil {
			return errors.Wrapf(err, "error committing expired acount requests deletion")
		}
	}

	return nil
}

type maintenanceResetPasswordAgent struct {
	cfg *config.ResetPasswordMaintenanceConfig
	ctx context.Context
}

func newMaintenanceResetPasswordAgent(ctx context.Context, cfg *config.ResetPasswordMaintenanceConfig) *maintenanceResetPasswordAgent {
	return &maintenanceResetPasswordAgent{
		cfg: cfg,
		ctx: ctx,
	}
}

func (ma *maintenanceResetPasswordAgent) run() {
	dl.Infof("started")
	defer dl.Info("exited")

	ticker := time.NewTicker(ma.cfg.CheckFrequency)
	for {
		select {
		case <-ma.ctx.Done():
			{
				ticker.Stop()
				return
			}
		case <-ticker.C:
			{
				if err := ma.deleteExpiredForgetPasswordRequests(); err != nil {
					dl.Error(err)
				}
			}
		}
	}
}
func (ma *maintenanceResetPasswordAgent) deleteExpiredForgetPasswordRequests() error {
	trx, err := str.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = trx.Rollback() }()

	timeout := time.Now().UTC().Add(-ma.cfg.ExpirationTimeout)
	passwordResetRequests, err := str.FindExpiredPasswordResetRequests(timeout, ma.cfg.BatchLimit, trx)
	if err != nil {
		return errors.Wrapf(err, "error finding expired password reset requests before %v", timeout)
	}
	if len(passwordResetRequests) > 0 {
		dl.Infof("found %d expired password reset requests to remove", len(passwordResetRequests))
		acctStrings := make([]string, len(passwordResetRequests))
		ids := make([]int, len(passwordResetRequests))
		for i, acct := range passwordResetRequests {
			ids[i] = acct.Id
			acctStrings[i] = fmt.Sprintf("{id:%d}", acct.Id)
		}

		dl.Infof("deleting expired password reset requests: %v", strings.Join(acctStrings, ","))
		if err := str.DeleteMultiplePasswordResetRequests(ids, trx); err != nil {
			return errors.Wrapf(err, "error deleting expired password reset requests before %v", timeout)
		}
		if err := trx.Commit(); err != nil {
			return errors.Wrapf(err, "error committing expired password reset requests deletion")
		}
	}
	return nil
}
