package limits

import (
	"github.com/jmoiron/sqlx"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
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
			if shr.ShareMode == "public" {
				env, err := a.str.GetEnvironment(shr.EnvironmentId, trx)
				if err != nil {
					return errors.Wrap(err, "error finding environment")
				}

				fe, err := a.str.FindFrontendPubliclyNamed(*shr.FrontendSelection, trx)
				if err != nil {
					return errors.Wrapf(err, "error finding frontend name '%v' for '%v'", *shr.FrontendSelection, shr.Token)
				}

				if err := zrokEdgeSdk.CreateServicePolicyDial(env.ZId+"-"+shr.ZId+"-dial", shr.ZId, []string{fe.ZId}, zrokEdgeSdk.ZrokShareTags(shr.Token).SubTags, a.edge); err != nil {
					return errors.Wrapf(err, "error creating dial service policy for '%v'", shr.Token)
				}
				logrus.Infof("added dial service policy for '%v'", shr.Token)

			} else if shr.ShareMode == "private" {
				return errors.New("share relax for private share not implemented")
			}
		}
	}

	return nil
}
