package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type ShareLimitJournal struct {
	Model
	ShareId int
	Action  string
}

func (self *Store) CreateShareLimitJournal(j *ShareLimitJournal, tx *sqlx.Tx) (int, error) {
	stmt, err := tx.Prepare("insert into share_limit_journal (share_id, action) values ($1, $2) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing share_limit_journal insert statement")
	}
	var id int
	if err := stmt.QueryRow(j.ShareId, j.Action).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing share_limit_journal insert statement")
	}
	return id, nil
}

func (self *Store) FindLatestShareLimitJournal(shrId int, tx *sqlx.Tx) (*ShareLimitJournal, error) {
	j := &ShareLimitJournal{}
	if err := tx.QueryRowx("select * from share_limit_journal where share_id = $1", shrId).StructScan(j); err != nil {
		return nil, errors.Wrap(err, "error finding share_limit_journal by share_id")
	}
	return j, nil
}
