package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type AccountLimitJournal struct {
	Model
	AccountId int
	RxBytes   int64
	TxBytes   int64
	Action    string
}

func (self *Store) CreateAccountLimitJournal(j *AccountLimitJournal, tx *sqlx.Tx) (int, error) {
	stmt, err := tx.Prepare("insert into account_limit_journal (account_id, rx_bytes, tx_bytes, action) values ($1, $2, $3, $4) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing account_limit_journal insert statement")
	}
	var id int
	if err := stmt.QueryRow(j.AccountId, j.RxBytes, j.TxBytes, j.Action).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing account_limit_journal insert statement")
	}
	return id, nil
}

func (self *Store) IsAccountLimitJournalEmpty(acctId int, tx *sqlx.Tx) (bool, error) {
	count := 0
	if err := tx.QueryRowx("select count(0) from account_limit_journal where account_id = $1", acctId).Scan(&count); err != nil {
		return false, err
	}
	return count == 0, nil
}

func (self *Store) FindLatestAccountLimitJournal(acctId int, tx *sqlx.Tx) (*AccountLimitJournal, error) {
	j := &AccountLimitJournal{}
	if err := tx.QueryRowx("select * from account_limit_journal where account_id = $1 order by created_at desc limit 1", acctId).StructScan(j); err != nil {
		return nil, errors.Wrap(err, "error finding account_limit_journal by account_id")
	}
	return j, nil
}
