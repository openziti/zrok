package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func (str *Store) IsAccountGrantedSkipInterstitial(acctId int, trx *sqlx.Tx) (bool, error) {
	stmt, err := trx.Prepare("select count(0) from skip_interstitial_grants where account_id = $1 and not deleted")
	if err != nil {
		return false, errors.Wrap(err, "error preparing skip_interstitial_grants select statement")
	}
	var count int
	if err := stmt.QueryRow(acctId).Scan(&count); err != nil {
		return false, errors.Wrap(err, "error querying skip_interstitial_grants count")
	}
	return count > 0, nil
}
