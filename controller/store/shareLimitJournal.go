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

func (str *Store) CreateShareLimitJournal(j *ShareLimitJournal, trx *sqlx.Tx) (int, error) {
	stmt, err := trx.Prepare("insert into share_limit_journal (share_id, rx_bytes, tx_bytes, action) values ($1, $2, $3, $4) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing share_limit_journal insert statement")
	}
	var id int
	if err := stmt.QueryRow(j.ShareId, j.RxBytes, j.TxBytes, j.Action).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing share_limit_journal insert statement")
	}
	return id, nil
}

func (str *Store) IsShareLimitJournalEmpty(shrId int, trx *sqlx.Tx) (bool, error) {
	count := 0
	if err := trx.QueryRowx("select count(0) from share_limit_journal where share_id = $1", shrId).Scan(&count); err != nil {
		return false, err
	}
	return count == 0, nil
}

func (str *Store) FindLatestShareLimitJournal(shrId int, trx *sqlx.Tx) (*ShareLimitJournal, error) {
	j := &ShareLimitJournal{}
	if err := trx.QueryRowx("select * from share_limit_journal where share_id = $1 order by created_at desc limit 1", shrId).StructScan(j); err != nil {
		return nil, errors.Wrap(err, "error finding share_limit_journal by share_id")
	}
	return j, nil
}

func (str *Store) FindAllLatestShareLimitJournal(trx *sqlx.Tx) ([]*ShareLimitJournal, error) {
	rows, err := trx.Queryx("select id, share_id, rx_bytes, tx_bytes, action, created_at, updated_at from share_limit_journal where id in (select max(id) as id from share_limit_journal group by share_id)")
	if err != nil {
		return nil, errors.Wrap(err, "error selecting all latest share_limit_journal")
	}
	var sljs []*ShareLimitJournal
	for rows.Next() {
		slj := &ShareLimitJournal{}
		if err := rows.StructScan(slj); err != nil {
			return nil, errors.Wrap(err, "error scanning share_limit_journal")
		}
		sljs = append(sljs, slj)
	}
	return sljs, nil
}

func (str *Store) DeleteShareLimitJournalForShare(shrId int, trx *sqlx.Tx) error {
	if _, err := trx.Exec("delete from share_limit_journal where share_id = $1", shrId); err != nil {
		return errors.Wrapf(err, "error deleting share_limit_journal for '#%d'", shrId)
	}
	return nil
}
