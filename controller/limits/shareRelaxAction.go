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

func (a *shareRelaxAction) HandleShare(s *store.Share, rxBytes, txBytes int64, limit *BandwidthPerPeriod, trx *sqlx.Tx) error {
	logrus.Infof("relaxing '%v'", s.Token)

	if s.ShareMode == "public" {
		env, err := a.str.GetEnvironment(s.EnvironmentId, trx)
		if err != nil {
			return errors.Wrap(err, "error finding environment")
		}

		fe, err := a.str.FindFrontendPubliclyNamed(*s.FrontendSelection, trx)
		if err != nil {
			return errors.Wrapf(err, "error finding frontend name '%v' for '%v'", *s.FrontendSelection, s.Token)
		}

		if err := zrokEdgeSdk.CreateServicePolicyDial(env.ZId+"-"+s.ZId+"-dial", s.ZId, []string{fe.ZId}, zrokEdgeSdk.ZrokShareTags(s.Token).SubTags, a.edge); err != nil {
			return errors.Wrapf(err, "error creating dial service policy for '%v'", s.Token)
		}
		logrus.Infof("added dial service policy for '%v'", s.Token)

	} else if s.ShareMode == "private" {
		return errors.New("share relax for private share not implemented")
	}

	return nil
}
