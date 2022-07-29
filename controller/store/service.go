package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Service struct {
	Model
	AccountId int
	ZitiId    string
	Endpoint  string
	Active    bool
}

func (self *Store) CreateService(accountId int, svc *Service, tx *sqlx.Tx) (int, error) {
	stmt, err := tx.Prepare("insert into services (account_id, ziti_id, endpoint, active) values (?, ?, ?, true)")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing services insert statement")
	}
	res, err := stmt.Exec(accountId, svc.ZitiId, svc.Endpoint)
	if err != nil {
		return 0, errors.Wrap(err, "error executing services insert statement")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(err, "error retrieving last services insert id")
	}
	return int(id), nil
}

func (self *Store) GetService(id int, tx *sqlx.Tx) (*Service, error) {
	svc := &Service{}
	if err := tx.QueryRowx("select * from services where id = ?", id).StructScan(svc); err != nil {
		return nil, errors.Wrap(err, "error selecting service by id")
	}
	return svc, nil
}

func (self *Store) FindServicesForAccount(accountId int, tx *sqlx.Tx) ([]*Service, error) {
	rows, err := tx.Queryx("select services.* from services where account_id = ?", accountId)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting services by account id")
	}
	var svcs []*Service
	for rows.Next() {
		svc := &Service{}
		if err := rows.StructScan(svc); err != nil {
			return nil, errors.Wrap(err, "error scanning service")
		}
		svcs = append(svcs, svc)
	}
	return svcs, nil
}

func (self *Store) DeactivateService(id int, tx *sqlx.Tx) error {
	stmt, err := tx.Prepare("update services set active=false where id = ?")
	if err != nil {
		return errors.Wrap(err, "error preparing services deactivate statement")
	}
	_, err = stmt.Exec(id)
	if err != nil {
		return errors.Wrap(err, "error executing services deactivate statement")
	}
	return nil
}

func (self *Store) DeleteService(id int, tx *sqlx.Tx) error {
	stmt, err := tx.Prepare("delete from services where id = ?")
	if err != nil {
		return errors.Wrap(err, "error preparing services delete statement")
	}
	_, err = stmt.Exec(id)
	if err != nil {
		return errors.Wrap(err, "error executing services delete statement")
	}
	return nil
}
