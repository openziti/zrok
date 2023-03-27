package limits

import (
	"github.com/jmoiron/sqlx"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type shareRelaxAction struct {
	str  *store.Store
	edge *rest_management_api_client.ZitiEdgeManagement
}

func newShareRelaxAction(str *store.Store, edge *rest_management_api_client.ZitiEdgeManagement) *shareRelaxAction {
	return &shareRelaxAction{str, edge}
}

func (a *shareRelaxAction) HandleShare(shr *store.Share, rxBytes, txBytes int64, limit *BandwidthPerPeriod, trx *sqlx.Tx) error {
	logrus.Infof("relaxing '%v'", shr.Token)

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

	return nil
}
