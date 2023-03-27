package limits

import (
	"github.com/jmoiron/sqlx"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type accountLimitAction struct {
	str  *store.Store
	edge *rest_management_api_client.ZitiEdgeManagement
}

func newAccountLimitAction(str *store.Store, edge *rest_management_api_client.ZitiEdgeManagement) *accountLimitAction {
	return &accountLimitAction{str, edge}
}

func (a *accountLimitAction) HandleAccount(acct *store.Account, rxBytes, txBytes int64, limit *BandwidthPerPeriod, trx *sqlx.Tx) error {
	logrus.Infof("limiting '%v'", acct.Email)

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
			if err := zrokEdgeSdk.DeleteServicePolicyDial(env.ZId, shr.Token, a.edge); err != nil {
				return errors.Wrapf(err, "error deleting dial service policy for '%v'", shr.Token)
			}
			logrus.Infof("removed dial service policy for share '%v' of environment '%v'", shr.Token, env.ZId)
		}
	}

	return nil
}
