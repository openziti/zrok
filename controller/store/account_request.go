package store

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type AccountRequest struct {
	Model
	Token         string
	Email         string
	SourceAddress string
}

func (self *Store) CreateAccountRequest(ar *AccountRequest, tx *sqlx.Tx) (int, error) {
	stmt, err := tx.Prepare("insert into account_requests (token, email, source_address) values ($1, $2, $3) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing account_requests insert statement")
	}
	var id int
	if err := stmt.QueryRow(ar.Token, ar.Email, ar.SourceAddress).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing account_requests insert statement")
	}
	return id, nil
}

func (self *Store) GetAccountRequest(id int, tx *sqlx.Tx) (*AccountRequest, error) {
	ar := &AccountRequest{}
	if err := tx.QueryRowx("select * from account_requests where id = $1", id).StructScan(ar); err != nil {
		return nil, errors.Wrap(err, "error selecting account_request by id")
	}
	return ar, nil
}

func (self *Store) FindAccountRequestWithToken(token string, tx *sqlx.Tx) (*AccountRequest, error) {
	ar := &AccountRequest{}
	if err := tx.QueryRowx("select * from account_requests where token = $1", token).StructScan(ar); err != nil {
		return nil, errors.Wrap(err, "error selecting account_request by token")
	}
	return ar, nil
}

func (self *Store) FindExpiredAccountRequests(before time.Time, limit int, tx *sqlx.Tx) ([]*AccountRequest, error) {
	rows, err := tx.Queryx(fmt.Sprintf("select * from account_requests where created_at < $1 limit %d for update", limit), before)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting expired account_requests")
	}
	var ars []*AccountRequest
	for rows.Next() {
		ar := &AccountRequest{}
		if err := rows.StructScan(ar); err != nil {
			return nil, errors.Wrap(err, "error scanning account_request")
		}
		ars = append(ars, ar)
	}
	return ars, nil
}

func (self *Store) FindAccountRequestWithEmail(email string, tx *sqlx.Tx) (*AccountRequest, error) {
	ar := &AccountRequest{}
	if err := tx.QueryRowx("select * from account_requests where email = $1", email).StructScan(ar); err != nil {
		return nil, errors.Wrap(err, "error selecting account_request by email")
	}
	return ar, nil
}

func (self *Store) DeleteAccountRequest(id int, tx *sqlx.Tx) error {
	stmt, err := tx.Prepare("delete from account_requests where id = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing account_requests delete statement")
	}
	_, err = stmt.Exec(id)
	if err != nil {
		return errors.Wrap(err, "error executing account_requests delete statement")
	}
	return nil
}

func (self *Store) DeleteMultipleAccountRequests(ids []int, tx *sqlx.Tx) error {
	if len(ids) == 0 {
		return nil
	}

	anyIds := make([]any, len(ids))
	indexes := make([]string, len(ids))

	for i, id := range ids {
		anyIds[i] = id
		indexes[i] = fmt.Sprintf("$%d", i+1)
	}

	stmt, err := tx.Prepare(fmt.Sprintf("delete from account_requests where id in (%s)", strings.Join(indexes, ",")))
	if err != nil {
		return errors.Wrap(err, "error preparing account_requests delete multiple statement")
	}
	_, err = stmt.Exec(anyIds...)
	if err != nil {
		return errors.Wrap(err, "error executing account_requests delete multiple statement")
	}
	return nil
}
