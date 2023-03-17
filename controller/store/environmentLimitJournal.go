package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type EnvironmentLimitJournal struct {
	Model
	EnvironmentId int
	Action        string
}

func (self *Store) CreateEnvironmentLimitJournal(j *EnvironmentLimitJournal, tx *sqlx.Tx) (int, error) {
	stmt, err := tx.Prepare("insert into environment_limit_journal (environment_id, action) values ($1, $2) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing environment_limit_journal insert statement")
	}
	var id int
	if err := stmt.QueryRow(j.EnvironmentId, j.Action).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing environment_limit_journal insert statement")
	}
	return id, nil
}

func (self *Store) FindLatestEnvironmentLimitJournal(envId int, tx *sqlx.Tx) (*EnvironmentLimitJournal, error) {
	j := &EnvironmentLimitJournal{}
	if err := tx.QueryRowx("select * from environment_limit_journal where environment_id = $1", envId).StructScan(j); err != nil {
		return nil, errors.Wrap(err, "error finding environment_limit_journal by environment_id")
	}
	return j, nil
}
