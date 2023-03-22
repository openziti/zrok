package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type EnvironmentLimitJournal struct {
	Model
	EnvironmentId int
	RxBytes       int64
	TxBytes       int64
	Action        LimitJournalAction
}

func (self *Store) CreateEnvironmentLimitJournal(j *EnvironmentLimitJournal, trx *sqlx.Tx) (int, error) {
	stmt, err := trx.Prepare("insert into environment_limit_journal (environment_id, rx_bytes, tx_bytes, action) values ($1, $2, $3, $4) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing environment_limit_journal insert statement")
	}
	var id int
	if err := stmt.QueryRow(j.EnvironmentId, j.RxBytes, j.TxBytes, j.Action).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing environment_limit_journal insert statement")
	}
	return id, nil
}

func (self *Store) IsEnvironmentLimitJournalEmpty(envId int, trx *sqlx.Tx) (bool, error) {
	count := 0
	if err := trx.QueryRowx("select count(0) from environment_limit_journal where environment_id = $1", envId).Scan(&count); err != nil {
		return false, err
	}
	return count == 0, nil
}

func (self *Store) FindLatestEnvironmentLimitJournal(envId int, trx *sqlx.Tx) (*EnvironmentLimitJournal, error) {
	j := &EnvironmentLimitJournal{}
	if err := trx.QueryRowx("select * from environment_limit_journal where environment_id = $1 order by created_at desc limit 1", envId).StructScan(j); err != nil {
		return nil, errors.Wrap(err, "error finding environment_limit_journal by environment_id")
	}
	return j, nil
}
