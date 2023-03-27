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

func (a *shareLimitAction) HandleShare(shr *store.Share, _, _ int64, _ *BandwidthPerPeriod, trx *sqlx.Tx) error {
	logrus.Infof("limiting '%v'", shr.Token)

	env, err := a.str.GetEnvironment(shr.EnvironmentId, trx)
	if err != nil {
		return err
	}

	if err := zrokEdgeSdk.DeleteServicePolicyDial(env.ZId, shr.Token, a.edge); err != nil {
		return err
	}
	logrus.Infof("removed dial service policy for '%v'", shr.Token)

	return nil
}
