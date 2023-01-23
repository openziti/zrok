package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Account struct {
	Model
	Email     string
	Salt      string
	Password  string
	Token     string
	Limitless bool
}

func (self *Store) CreateAccount(a *Account, tx *sqlx.Tx) (int, error) {
	stmt, err := tx.Prepare("insert into accounts (email, salt, password, token, limitless) values ($1, $2, $3, $4, $5) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing accounts insert statement")
	}
	var id int
	if err := stmt.QueryRow(a.Email, a.Salt, a.Password, a.Token, a.Limitless).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing accounts insert statement")
	}
	return id, nil
}

func (self *Store) GetAccount(id int, tx *sqlx.Tx) (*Account, error) {
	a := &Account{}
	if err := tx.QueryRowx("select * from accounts where id = $1", id).StructScan(a); err != nil {
		return nil, errors.Wrap(err, "error selecting account by id")
	}
	return a, nil
}

func (self *Store) FindAccountWithEmail(email string, tx *sqlx.Tx) (*Account, error) {
	a := &Account{}
	if err := tx.QueryRowx("select * from accounts where email = $1", email).StructScan(a); err != nil {
		return nil, errors.Wrap(err, "error selecting account by email")
	}
	return a, nil
}

func (self *Store) FindAccountWithToken(token string, tx *sqlx.Tx) (*Account, error) {
	a := &Account{}
	if err := tx.QueryRowx("select * from accounts where token = $1", token).StructScan(a); err != nil {
		return nil, errors.Wrap(err, "error selecting account by token")
	}
	return a, nil
}

func (self *Store) UpdateAccount(a *Account, tx *sqlx.Tx) (int, error) {
	stmt, err := tx.Prepare("update accounts set email=$1, salt=$2, password=$3, token=$4, limitless=$5 where id = $6")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing accounts update statement")
	}
	var id int
	if _, err := stmt.Exec(a.Email, a.Salt, a.Password, a.Token, a.Limitless, a.Id); err != nil {
		return 0, errors.Wrap(err, "error executing accounts update statement")
	}
	return id, nil
}
