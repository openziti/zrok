package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type AgentEnrollment struct {
	Model
	EnvironmentId int
	Token         string
}

func (str *Store) CreateAgentEnrollment(envId int, token string, trx *sqlx.Tx) (int, error) {
	stmt, err := trx.Prepare("insert into agent_enrollments (environment_id, token) values ($1, $2) returing id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing agent enrollments insert statement")
	}
	var id int
	if err := stmt.QueryRow(envId, token).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing agent enrollments insert statement")
	}
	return id, nil
}

func (str *Store) FindAgentEnrollmentForEnvironment(envId int, trx *sqlx.Tx) (*AgentEnrollment, error) {
	ae := &AgentEnrollment{}
	if err := trx.QueryRowx("select * from agent_enrollments where environment_id = $1", envId).StructScan(ae); err != nil {
		return nil, errors.Wrap(err, "error finding agent enrollment")
	}
	return ae, nil
}
