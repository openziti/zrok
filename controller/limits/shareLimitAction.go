package limits

import (
	"github.com/jmoiron/sqlx"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/sirupsen/logrus"
)

type shareLimitAction struct {
	str  *store.Store
	edge *rest_management_api_client.ZitiEdgeManagement
}

func newShareLimitAction(str *store.Store, edge *rest_management_api_client.ZitiEdgeManagement) *shareLimitAction {
	return &shareLimitAction{str, edge}
}

func (a *shareLimitAction) HandleShare(s *store.Share, _, _ int64, _ *BandwidthPerPeriod, trx *sqlx.Tx) error {
	logrus.Infof("limiting '%v'", s.Token)

	env, err := a.str.GetEnvironment(s.EnvironmentId, trx)
	if err != nil {
		return err
	}

	if err := zrokEdgeSdk.DeleteServicePolicyDial(env.ZId, s.Token, a.edge); err != nil {
		return err
	}
	logrus.Infof("removed service dial policy for '%v'", s.Token)

	return nil
}
