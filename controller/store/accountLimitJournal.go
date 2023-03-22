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
	Action    LimitJournalAction
}

func (str *Store) CreateAccountLimitJournal(j *AccountLimitJournal, trx *sqlx.Tx) (int, error) {
	stmt, err := trx.Prepare("insert into account_limit_journal (account_id, rx_bytes, tx_bytes, action) values ($1, $2, $3, $4) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing account_limit_journal insert statement")
	}
	var id int
	if err := stmt.QueryRow(j.AccountId, j.RxBytes, j.TxBytes, j.Action).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing account_limit_journal insert statement")
	}
	return id, nil
}

func (str *Store) IsAccountLimitJournalEmpty(acctId int, trx *sqlx.Tx) (bool, error) {
	count := 0
	if err := trx.QueryRowx("select count(0) from account_limit_journal where account_id = $1", acctId).Scan(&count); err != nil {
		return false, err
	}
	return count == 0, nil
}

func (str *Store) FindLatestAccountLimitJournal(acctId int, trx *sqlx.Tx) (*AccountLimitJournal, error) {
	j := &AccountLimitJournal{}
	if err := trx.QueryRowx("select * from account_limit_journal where account_id = $1 order by id desc limit 1", acctId).StructScan(j); err != nil {
		return nil, errors.Wrap(err, "error finding account_limit_journal by account_id")
	}
	return j, nil
}

func (str *Store) FindAllLatestAccountLimitJournal(trx *sqlx.Tx) ([]*AccountLimitJournal, error) {
	rows, err := trx.Queryx("select id, account_id, rx_bytes, tx_bytes, action, created_at, updated_at from account_limit_journal where id in (select max(id) as id from account_limit_journal group by account_id)")
	if err != nil {
		return nil, errors.Wrap(err, "error selecting distinct account_limit_jounal")
	}
	var is []*AccountLimitJournal
	for rows.Next() {
		i := &AccountLimitJournal{}
		if err := rows.StructScan(i); err != nil {
			return nil, errors.Wrap(err, "error scanning account_limit_journal")
		}
		is = append(is, i)
	}
	return is, nil
}
