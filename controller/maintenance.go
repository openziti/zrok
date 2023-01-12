package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type maintenanceAgent struct {
	*MaintenanceConfig
	ctx context.Context
}

func newMaintenanceAgent(ctx context.Context, cfg *MaintenanceConfig) *maintenanceAgent {
	return &maintenanceAgent{
		MaintenanceConfig: cfg,
		ctx:               ctx,
	}
}

func (ma *maintenanceAgent) run() {
	logrus.Info("starting")
	defer logrus.Info("stopping")

	ticker := time.NewTicker(ma.Registration.CheckFrequency)
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

func (ma *maintenanceAgent) deleteExpiredAccountRequests() error {
	tx, err := str.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	timeout := time.Now().UTC().Add(-ma.Registration.ExpirationTimeout)
	accountRequests, err := str.FindExpiredAccountRequests(timeout, ma.Registration.BatchLimit, tx)
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
