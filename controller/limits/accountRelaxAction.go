package limits

import (
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type accountRelaxAction struct {
	str  *store.Store
	zCfg *zrokEdgeSdk.Config
}

func newAccountRelaxAction(str *store.Store, zCfg *zrokEdgeSdk.Config) *accountRelaxAction {
	return &accountRelaxAction{str, zCfg}
}

func (a *accountRelaxAction) HandleAccount(acct *store.Account, _, _ int64, _ *BandwidthPerPeriod, trx *sqlx.Tx) error {
	logrus.Infof("relaxing '%v'", acct.Email)

	envs, err := a.str.FindEnvironmentsForAccount(acct.Id, trx)
	if err != nil {
		return errors.Wrapf(err, "error finding environments for account '%v'", acct.Email)
	}

	edge, err := zrokEdgeSdk.Client(a.zCfg)
	if err != nil {
		return err
	}

	for _, env := range envs {
		shrs, err := a.str.FindSharesForEnvironment(env.Id, trx)
		if err != nil {
			return errors.Wrapf(err, "error finding shares for environment '%v'", env.ZId)
		}

		for _, shr := range shrs {
			switch shr.ShareMode {
			case "public":
				if err := relaxPublicShare(a.str, edge, shr, trx); err != nil {
					return errors.Wrap(err, "error relaxing public share")
				}
			case "private":
				if err := relaxPrivateShare(a.str, edge, shr, trx); err != nil {
					return errors.Wrap(err, "error relaxing private share")
				}
			}
		}
	}

	return nil
}
