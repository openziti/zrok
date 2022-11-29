package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Service struct {
	Model
	EnvironmentId        int
	ZId                  string
	Name                 string
	ShareMode            string
	BackendMode          string
	FrontendSelection    *string
	FrontendEndpoint     *string
	BackendProxyEndpoint *string
}

func (self *Store) CreateService(envId int, svc *Service, tx *sqlx.Tx) (int, error) {
	stmt, err := tx.Prepare("insert into services (environment_id, z_id, name, share_mode, backend_mode, frontend_selection, frontend_endpoint, backend_proxy_endpoint) values ($1, $2, $3, $4, $5, $6, $7, $8) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing services insert statement")
	}
	var id int
	if err := stmt.QueryRow(envId, svc.ZId, svc.Name, svc.ShareMode, svc.BackendMode, svc.FrontendSelection, svc.FrontendEndpoint, svc.BackendProxyEndpoint).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing services insert statement")
	}
	return id, nil
}

func (self *Store) GetService(id int, tx *sqlx.Tx) (*Service, error) {
	svc := &Service{}
	if err := tx.QueryRowx("select * from services where id = $1", id).StructScan(svc); err != nil {
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

func (self *Store) FindServiceWithName(svcName string, tx *sqlx.Tx) (*Service, error) {
	svc := &Service{}
	if err := tx.QueryRowx("select * from services where name = $1", svcName).StructScan(svc); err != nil {
		return nil, errors.Wrap(err, "error selecting service by name")
	}
	return svc, nil
}

func (self *Store) FindServicesForEnvironment(envId int, tx *sqlx.Tx) ([]*Service, error) {
	rows, err := tx.Queryx("select services.* from services where environment_id = $1", envId)
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
	sql := "update services set z_id = $1, name = $2, share_mode = $3, backend_mode = $4, frontend_selection = $5, frontend_endpoint = $6, backend_proxy_endpoint = $7, updated_at = strftime('%Y-%m-%d %H:%M:%f', 'now') where id = $8"
	stmt, err := tx.Prepare(sql)
	if err != nil {
		return errors.Wrap(err, "error preparing services update statement")
	}
	_, err = stmt.Exec(svc.ZId, svc.Name, svc.ShareMode, svc.BackendMode, svc.FrontendSelection, svc.FrontendEndpoint, svc.BackendProxyEndpoint, svc.Id)
	if err != nil {
		return errors.Wrap(err, "error executing services update statement")
	}
	return nil
}

func (self *Store) DeleteService(id int, tx *sqlx.Tx) error {
	stmt, err := tx.Prepare("delete from services where id = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing services delete statement")
	}
	_, err = stmt.Exec(id)
	if err != nil {
		return errors.Wrap(err, "error executing services delete statement")
	}
	return nil
}
