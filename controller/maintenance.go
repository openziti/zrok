package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type maintenanceRegistrationAgent struct {
	*RegistrationMaintenanceConfig
	ctx context.Context
}

func newRegistrationMaintenanceAgent(ctx context.Context, cfg *RegistrationMaintenanceConfig) *maintenanceRegistrationAgent {
	return &maintenanceRegistrationAgent{
		RegistrationMaintenanceConfig: cfg,
		ctx:                           ctx,
	}
}

func (ma *maintenanceRegistrationAgent) run() {
	logrus.Infof("starting maintenance registration agent")
	defer logrus.Info("stopping maintenance registration agent")

	ticker := time.NewTicker(ma.CheckFrequency)
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
					logrus.Error(err)
				}
			}
		}
	}
}

func (ma *maintenanceRegistrationAgent) deleteExpiredAccountRequests() error {
	tx, err := str.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	timeout := time.Now().UTC().Add(-ma.ExpirationTimeout)
	accountRequests, err := str.FindExpiredAccountRequests(timeout, ma.BatchLimit, tx)
	if err != nil {
		return errors.Wrapf(err, "error finding expire account requests before %v", timeout)
	}
	if len(accountRequests) > 0 {
		logrus.Infof("found %d expired account requests to remove", len(accountRequests))
		acctStrings := make([]string, len(accountRequests))
		ids := make([]int, len(accountRequests))
		for i, acct := range accountRequests {
			ids[i] = acct.Id
			acctStrings[i] = fmt.Sprintf("{%d:%s}", acct.Id, acct.Email)
		}

		logrus.Infof("deleting expired account requests: %v", strings.Join(acctStrings, ","))
		if err := str.DeleteMultipleAccountRequests(ids, tx); err != nil {
			return errors.Wrapf(err, "error deleting expired account requests before %v", timeout)
		}
		if err := tx.Commit(); err != nil {
			return errors.Wrapf(err, "error committing expired acount requests deletion")
		}
	}

	return nil
}

type maintenanceAccountAgent struct {
	*AccountMaintenanceConfig
	ctx context.Context
}

func newAccountMaintenanceAgent(ctx context.Context, cfg *AccountMaintenanceConfig) *maintenanceAccountAgent {
	return &maintenanceAccountAgent{
		AccountMaintenanceConfig: cfg,
		ctx:                      ctx,
	}
}

func (ma *maintenanceAccountAgent) run() {
	logrus.Infof("starting maintenance account agent")
	defer logrus.Info("stopping maintenance account agent")

	ticker := time.NewTicker(ma.CheckFrequency)
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
					logrus.Error(err)
				}
			}
		}
	}
}
func (ma *maintenanceAccountAgent) deleteExpiredForgetPasswordRequests() error {
	tx, err := str.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	timeout := time.Now().UTC().Add(-ma.ExpirationTimeout)
	passwordResetRequests, err := str.FindExpiredPasswordResetRequests(timeout, ma.BatchLimit, tx)
	if err != nil {
		return errors.Wrapf(err, "error finding expired password reset requests before %v", timeout)
	}
	if len(passwordResetRequests) > 0 {
		logrus.Infof("found %d expired password reset requests to remove", len(passwordResetRequests))
		acctStrings := make([]string, len(passwordResetRequests))
		ids := make([]int, len(passwordResetRequests))
		for i, acct := range passwordResetRequests {
			ids[i] = acct.Id
			acctStrings[i] = fmt.Sprintf("{id:%d}", acct.Id)
		}

		logrus.Infof("deleting expired password reset requests: %v", strings.Join(acctStrings, ","))
		if err := str.DeleteMultiplePasswordResetRequests(ids, tx); err != nil {
			return errors.Wrapf(err, "error deleting expired password reset requests before %v", timeout)
		}
		if err := tx.Commit(); err != nil {
			return errors.Wrapf(err, "error committing expired password reset requests deletion")
		}
	}
	return nil
}
