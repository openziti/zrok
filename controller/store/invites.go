package store

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Invite struct {
	Model
	Token string
}

func (str *Store) CreateInvites(invites []*Invite, tx *sqlx.Tx) error {
	sql := "insert into invites (token) values %s"
	invs := make([]any, len(invites))
	queries := make([]string, len(invites))
	for i, inv := range invites {
		invs[i] = inv.Token
		queries[i] = fmt.Sprintf("($%d)", i+1)
	}
	stmt, err := tx.Prepare(fmt.Sprintf(sql, strings.Join(queries, ",")))
	if err != nil {
		return errors.Wrap(err, "error preparing invites insert statement")
	}
	if _, err := stmt.Exec(invs...); err != nil {
		return errors.Wrap(err, "error executing invites insert statement")
	}
	return nil
}

func (str *Store) GetInviteByToken(token string, tx *sqlx.Tx) (*Invite, error) {
	invite := &Invite{}
	if err := tx.QueryRowx("select * from invites where token = $1", token).StructScan(invite); err != nil {
		return nil, errors.Wrap(err, "error getting unused invite")
	}
	return invite, nil
}

func (str *Store) UpdateInvite(invite *Invite, tx *sqlx.Tx) error {
	stmt, err := tx.Prepare("update invites set token = $1")
	if err != nil {
		return errors.Wrap(err, "error perparing invites update statement")
	}
	_, err = stmt.Exec(invite.Token)
	if err != nil {
		return errors.Wrap(err, "error executing invites update statement")
	}
	return nil
}

func (str *Store) DeleteInvite(id int, tx *sqlx.Tx) error {
	stmt, err := tx.Prepare("delete from invites where id = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing invites delete statement")
	}
	_, err = stmt.Exec(id)
	if err != nil {
		return errors.Wrap(err, "error executing invites delete statement")
	}
	return nil
}
