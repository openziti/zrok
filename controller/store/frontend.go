package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Frontend struct {
	Model
	EnvironmentId int
	Name          string
	ZId           string
	PublicName    *string
}

func (str *Store) CreateFrontend(envId int, f *Frontend, tx *sqlx.Tx) (int, error) {
	stmt, err := tx.Prepare("insert into frontends (environment_id, name, z_id, public_name) values ($1, $2, $3, $4) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing frontends insert statement")
	}
	var id int
	if err := stmt.QueryRow(envId, f.Name, f.ZId, f.PublicName).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing frontends insert statement")
	}
	return id, nil
}

func (str *Store) GetFrontend(id int, tx *sqlx.Tx) (*Frontend, error) {
	i := &Frontend{}
	if err := tx.QueryRowx("select * from frontends where id = $1", id).StructScan(i); err != nil {
		return nil, errors.Wrap(err, "error selecting frontend by id")
	}
	return i, nil
}

func (str *Store) FindFrontendNamed(name string, tx *sqlx.Tx) (*Frontend, error) {
	i := &Frontend{}
	if err := tx.QueryRowx("select frontends.* from frontends where name = $1", name).StructScan(i); err != nil {
		return nil, errors.Wrap(err, "error selecting frontend by name")
	}
	return i, nil
}

func (str *Store) FindFrontendsForEnvironment(envId int, tx *sqlx.Tx) ([]*Frontend, error) {
	rows, err := tx.Queryx("select frontends.* from frontends where environment_id = $1", envId)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting frontends by environment_id")
	}
	var is []*Frontend
	for rows.Next() {
		i := &Frontend{}
		if err := rows.StructScan(i); err != nil {
			return nil, errors.Wrap(err, "error scanning frontend")
		}
		is = append(is, i)
	}
	return is, nil
}

func (str *Store) DeleteFrontend(id int, tx *sqlx.Tx) error {
	stmt, err := tx.Prepare("delete from frontends where id = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing frontends delete statement")
	}
	_, err = stmt.Exec(id)
	if err != nil {
		return errors.Wrap(err, "error executing frontends delete statement")
	}
	return nil
}
