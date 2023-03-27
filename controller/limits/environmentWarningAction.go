package limits

import (
	"github.com/jmoiron/sqlx"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/zrok/controller/emailUi"
	"github.com/openziti/zrok/controller/store"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type environmentWarningAction struct {
	str  *store.Store
	edge *rest_management_api_client.ZitiEdgeManagement
	cfg  *emailUi.Config
}

func newEnvironmentWarningAction(cfg *emailUi.Config, str *store.Store, edge *rest_management_api_client.ZitiEdgeManagement) *environmentWarningAction {
	return &environmentWarningAction{str, edge, cfg}
}

func (a *environmentWarningAction) HandleEnvironment(env *store.Environment, rxBytes, txBytes int64, limit *BandwidthPerPeriod, trx *sqlx.Tx) error {
	logrus.Infof("warning '%v'", env.ZId)

	if env.AccountId != nil {
		acct, err := a.str.GetAccount(*env.AccountId, trx)
		if err != nil {
			return err
		}

		if err := sendLimitWarningEmail(a.cfg, acct.Email, limit, rxBytes, txBytes); err != nil {
			return errors.Wrapf(err, "error sending limit warning email to '%v'", acct.Email)
		}
	}

	return nil
}
