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

func (str *Store) GrantSkipInterstitial(acctId int, trx *sqlx.Tx) error {
	granted, err := str.IsAccountGrantedSkipInterstitial(acctId, trx)
	if err != nil {
		return err
	}
	if granted {
		return nil
	}

	stmt, err := trx.Prepare("insert into skip_interstitial_grants (account_id) values ($1)")
	if err != nil {
		return errors.Wrap(err, "error preparing skip_interstitial_grants insert statement")
	}
	_, err = stmt.Exec(acctId)
	if err != nil {
		return errors.Wrap(err, "error executing skip_interstitial_grants insert statement")
	}
	return nil
}

func (str *Store) RevokeSkipInterstitial(acctId int, trx *sqlx.Tx) error {
	stmt, err := trx.Prepare("update skip_interstitial_grants set deleted = true, updated_at = current_timestamp where account_id = $1 and not deleted")
	if err != nil {
		return errors.Wrap(err, "error preparing skip_interstitial_grants update statement")
	}
	_, err = stmt.Exec(acctId)
	if err != nil {
		return errors.Wrap(err, "error executing skip_interstitial_grants update statement")
	}
	return nil
}
