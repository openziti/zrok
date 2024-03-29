package store

import (
	"fmt"
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

func (str *Store) CreateEnvironmentLimitJournal(j *EnvironmentLimitJournal, trx *sqlx.Tx) (int, error) {
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

func (str *Store) IsEnvironmentLimitJournalEmpty(envId int, trx *sqlx.Tx) (bool, error) {
	count := 0
	if err := trx.QueryRowx("select count(0) from environment_limit_journal where environment_id = $1", envId).Scan(&count); err != nil {
		return false, err
	}
	return count == 0, nil
}

func (str *Store) FindLatestEnvironmentLimitJournal(envId int, trx *sqlx.Tx) (*EnvironmentLimitJournal, error) {
	j := &EnvironmentLimitJournal{}
	if err := trx.QueryRowx("select * from environment_limit_journal where environment_id = $1 order by created_at desc limit 1", envId).StructScan(j); err != nil {
		return nil, errors.Wrap(err, "error finding environment_limit_journal by environment_id")
	}
	return j, nil
}

func (str *Store) FindSelectedLatestEnvironmentLimitJournal(envIds []int, trx *sqlx.Tx) ([]*EnvironmentLimitJournal, error) {
	if len(envIds) < 1 {
		return nil, nil
	}
	in := "("
	for i := range envIds {
		if i > 0 {
			in += ", "
		}
		in += fmt.Sprintf("%d", envIds[i])
	}
	in += ")"
	rows, err := trx.Queryx("select id, environment_id, rx_bytes, tx_bytes, action, created_at, updated_at from environment_limit_journal where id in (select max(id) as id from environment_limit_journal group by environment_id) and environment_id in " + in)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting all latest environment_limit_journal")
	}
	var eljs []*EnvironmentLimitJournal
	for rows.Next() {
		elj := &EnvironmentLimitJournal{}
		if err := rows.StructScan(elj); err != nil {
			return nil, errors.Wrap(err, "error scanning environment_limit_journal")
		}
		eljs = append(eljs, elj)
	}
	return eljs, nil
}

func (str *Store) FindAllLatestEnvironmentLimitJournal(trx *sqlx.Tx) ([]*EnvironmentLimitJournal, error) {
	rows, err := trx.Queryx("select id, environment_id, rx_bytes, tx_bytes, action, created_at, updated_at from environment_limit_journal where id in (select max(id) as id from environment_limit_journal group by environment_id)")
	if err != nil {
		return nil, errors.Wrap(err, "error selecting all latest environment_limit_journal")
	}
	var eljs []*EnvironmentLimitJournal
	for rows.Next() {
		elj := &EnvironmentLimitJournal{}
		if err := rows.StructScan(elj); err != nil {
			return nil, errors.Wrap(err, "error scanning environment_limit_journal")
		}
		eljs = append(eljs, elj)
	}
	return eljs, nil
}

func (str *Store) DeleteEnvironmentLimitJournalForEnvironment(envId int, trx *sqlx.Tx) error {
	if _, err := trx.Exec("delete from environment_limit_journal where environment_id = $1", envId); err != nil {
		return errors.Wrapf(err, "error deleteing environment_limit_journal for '#%d'", envId)
	}
	return nil
}
