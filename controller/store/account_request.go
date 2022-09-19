package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type AccountRequest struct {
	Model
	Token         string
	Email         string
	Verified      bool
	SourceAddress string
}

func (self *Store) CreateAccountRequest(ar *AccountRequest, tx *sqlx.Tx) (int, error) {
	stmt, err := tx.Prepare("insert into account_requests (token, email, source_address) values (?, ?, ?)")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing account_requests insert statement")
	}
	res, err := stmt.Exec(ar.Token, ar.Email, ar.SourceAddress)
	if err != nil {
		return 0, errors.Wrap(err, "error executing account_requests insert statement")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(err, "error retrieving last account_requests insert id")
	}
	return int(id), nil
}

func (self *Store) GetAccountRequest(id int, tx *sqlx.Tx) (*AccountRequest, error) {
	ar := &AccountRequest{}
	if err := tx.QueryRowx("select * from account_requests where id = ?", id).StructScan(ar); err != nil {
		return nil, errors.Wrap(err, "error selecting account_request by id")
	}
	return ar, nil
}

func (self *Store) FindAccountRequestWithToken(token string, tx *sqlx.Tx) (*AccountRequest, error) {
	ar := &AccountRequest{}
	if err := tx.QueryRowx("select * from account_requests where token = ?", token).StructScan(ar); err != nil {
		return nil, errors.Wrap(err, "error selecting account_request by token")
	}
	return ar, nil
}
