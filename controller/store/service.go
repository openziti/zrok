package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Service struct {
	Model
	EnvironmentId int
	ZitiServiceId string
	ZrokServiceId string
	Frontend      string
	Backend       string
	Active        bool
}

func (self *Store) CreateService(envId int, svc *Service, tx *sqlx.Tx) (int, error) {
	stmt, err := tx.Prepare("insert into services (environment_id, ziti_service_id, zrok_service_id, frontend, backend, active) values (?, ?, ?, ?, ?, true)")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing services insert statement")
	}
	res, err := stmt.Exec(envId, svc.ZitiServiceId, svc.ZrokServiceId, svc.Frontend, svc.Backend)
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

func (self *Store) GetAllServices(tx *sqlx.Tx) ([]*Service, error) {
	rows, err := tx.Queryx("select * from services order by id")
	if err != nil {
		return nil, errors.Wrap(err, "error selecting all services")
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

func (self *Store) FindServicesForEnvironment(envId int, tx *sqlx.Tx) ([]*Service, error) {
	rows, err := tx.Queryx("select services.* from services where environment_id = ?", envId)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting services by environment id")
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

func (self *Store) UpdateService(svc *Service, tx *sqlx.Tx) error {
	sql := "update services set ziti_service_id = ?, zrok_service_id = ?, frontend = ?, backend = ?, active = ?, updated_at = strftime('%Y-%m-%d %H:%M:%f', 'now') where id = ?"
	stmt, err := tx.Prepare(sql)
	if err != nil {
		return errors.Wrap(err, "error preparing services update statement")
	}
	_, err = stmt.Exec(svc.ZitiServiceId, svc.ZrokServiceId, svc.Frontend, svc.Backend, svc.Active, svc.Id)
	if err != nil {
		return errors.Wrap(err, "error executing services update statement")
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
