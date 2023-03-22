package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type ShareLimitJournal struct {
	Model
	ShareId int
	RxBytes int64
	TxBytes int64
	Action  LimitJournalAction
}

func (self *Store) CreateShareLimitJournal(j *ShareLimitJournal, tx *sqlx.Tx) (int, error) {
	stmt, err := tx.Prepare("insert into share_limit_journal (share_id, rx_bytes, tx_bytes, action) values ($1, $2, $3, $4) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing share_limit_journal insert statement")
	}
	var id int
	if err := stmt.QueryRow(j.ShareId, j.RxBytes, j.TxBytes, j.Action).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing share_limit_journal insert statement")
	}
	return id, nil
}

func (self *Store) IsShareLimitJournalEmpty(shrId int, trx *sqlx.Tx) (bool, error) {
	count := 0
	if err := trx.QueryRowx("select count(0) from share_limit_journal where share_id = $1", shrId).Scan(&count); err != nil {
		return false, err
	}
	return count == 0, nil
}

func (self *Store) FindLatestShareLimitJournal(shrId int, tx *sqlx.Tx) (*ShareLimitJournal, error) {
	j := &ShareLimitJournal{}
	if err := tx.QueryRowx("select * from share_limit_journal where share_id = $1 order by created_at desc limit 1", shrId).StructScan(j); err != nil {
		return nil, errors.Wrap(err, "error finding share_limit_journal by share_id")
	}
	return j, nil
}
