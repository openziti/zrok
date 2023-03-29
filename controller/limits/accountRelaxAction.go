package limits

import (
	"github.com/jmoiron/sqlx"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/zrok/controller/store"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type accountRelaxAction struct {
	str  *store.Store
	edge *rest_management_api_client.ZitiEdgeManagement
}

func newAccountRelaxAction(str *store.Store, edge *rest_management_api_client.ZitiEdgeManagement) *accountRelaxAction {
	return &accountRelaxAction{str, edge}
}

func (a *accountRelaxAction) HandleAccount(acct *store.Account, _, _ int64, _ *BandwidthPerPeriod, trx *sqlx.Tx) error {
	logrus.Infof("relaxing '%v'", acct.Email)

	envs, err := a.str.FindEnvironmentsForAccount(acct.Id, trx)
	if err != nil {
		return errors.Wrapf(err, "error finding environments for account '%v'", acct.Email)
	}

	for _, env := range envs {
		shrs, err := a.str.FindSharesForEnvironment(env.Id, trx)
		if err != nil {
			return errors.Wrapf(err, "error finding shares for environment '%v'", env.ZId)
		}

		for _, shr := range shrs {
			switch shr.ShareMode {
			case "public":
				if err := relaxPublicShare(a.str, a.edge, shr, trx); err != nil {
					return err
				}
			case "private":
				if err := relaxPrivateShare(a.str, a.edge, shr, trx); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
