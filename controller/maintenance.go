package controller

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type maintenanceAgent struct {
	ctx        context.Context
	frequency  time.Duration
	expiration time.Duration
}

func newMaintenanceAgent(ctx context.Context, frequency, expiration time.Duration) *maintenanceAgent {
	return &maintenanceAgent{
		ctx:        ctx,
		frequency:  frequency,
		expiration: expiration,
	}
}

func (ma *maintenanceAgent) run() {
	ticker := time.NewTicker(ma.frequency)
	for {
		select {
		case <-ma.ctx.Done():
			{
				ticker.Stop()
				return
			}
		case <-ticker.C:
			{
				logrus.Info("TICK")
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

	if err := str.DeleteExpiredAccountRequests(time.Now().UTC().Add(-ma.expiration), tx); err != nil {
		return errors.Wrapf(err, "error deleting expired account requests")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrapf(err, "error committing expired acount requests deletion")
	}

	return nil
}
