package limits

import (
	"github.com/jmoiron/sqlx"
	"github.com/openziti/edge-api/rest_management_api_client"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/sdk/golang/sdk"
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
			case string(sdk.PublicShareMode):
				if err := relaxPublicShare(a.str, edge, shr, trx); err != nil {
					return errors.Wrap(err, "error relaxing public share")
				}
			case string(sdk.PrivateShareMode):
				if err := relaxPrivateShare(a.str, edge, shr, trx); err != nil {
					return errors.Wrap(err, "error relaxing private share")
				}
			}
		}
	}

	return nil
}

func relaxPublicShare(str *store.Store, edge *rest_management_api_client.ZitiEdgeManagement, shr *store.Share, trx *sqlx.Tx) error {
	env, err := str.GetEnvironment(shr.EnvironmentId, trx)
	if err != nil {
		return errors.Wrap(err, "error finding environment")
	}

	fe, err := str.FindFrontendPubliclyNamed(*shr.FrontendSelection, trx)
	if err != nil {
		return errors.Wrapf(err, "error finding frontend name '%v' for '%v'", *shr.FrontendSelection, shr.Token)
	}

	if err := zrokEdgeSdk.CreateServicePolicyDial(env.ZId+"-"+shr.ZId+"-dial", shr.ZId, []string{fe.ZId}, zrokEdgeSdk.ZrokShareTags(shr.Token).SubTags, edge); err != nil {
		return errors.Wrapf(err, "error creating dial service policy for '%v'", shr.Token)
	}
	logrus.Infof("added dial service policy for '%v'", shr.Token)
	return nil
}

func relaxPrivateShare(str *store.Store, edge *rest_management_api_client.ZitiEdgeManagement, shr *store.Share, trx *sqlx.Tx) error {
	fes, err := str.FindFrontendsForPrivateShare(shr.Id, trx)
	if err != nil {
		return errors.Wrapf(err, "error finding frontends for share '%v'", shr.Token)
	}
	for _, fe := range fes {
		if fe.EnvironmentId != nil {
			env, err := str.GetEnvironment(*fe.EnvironmentId, trx)
			if err != nil {
				return errors.Wrapf(err, "error getting environment for frontend '%v'", fe.Token)
			}

			addlTags := map[string]interface{}{
				"zrokEnvironmentZId": env.ZId,
				"zrokFrontendToken":  fe.Token,
				"zrokShareToken":     shr.Token,
			}
			if err := zrokEdgeSdk.CreateServicePolicyDial(fe.Token+"-"+env.ZId+"-"+shr.ZId+"-dial", shr.ZId, []string{env.ZId}, addlTags, edge); err != nil {
				return errors.Wrapf(err, "unable to create dial policy for frontend '%v'", fe.Token)
			}

			logrus.Infof("added dial service policy for share '%v' to private frontend '%v'", shr.Token, fe.Token)
		}
	}
	return nil
}
