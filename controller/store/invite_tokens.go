package store

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type InviteToken struct {
	Model
	Token   string
	Deleted bool
}

func (str *Store) CreateInviteTokens(inviteTokens []*InviteToken, tx *sqlx.Tx) error {
	sql := "insert into invite_tokens (token) values %s"
	invs := make([]any, len(inviteTokens))
	queries := make([]string, len(inviteTokens))
	for i, inv := range inviteTokens {
		invs[i] = inv.Token
		queries[i] = fmt.Sprintf("($%d)", i+1)
	}
	stmt, err := tx.Prepare(fmt.Sprintf(sql, strings.Join(queries, ",")))
	if err != nil {
		return errors.Wrap(err, "error preparing invite_tokens insert statement")
	}
	if _, err := stmt.Exec(invs...); err != nil {
		return errors.Wrap(err, "error executing invites_tokens insert statement")
	}
	return nil
}

func (str *Store) GetInviteTokenByToken(token string, tx *sqlx.Tx) (*InviteToken, error) {
	inviteToken := &InviteToken{}
	if err := tx.QueryRowx("select * from invite_tokens where token = $1", token).StructScan(inviteToken); err != nil {
		return nil, errors.Wrap(err, "error getting unused invite_token")
	}
	return inviteToken, nil
}

func (str *Store) DeleteInviteToken(id int, tx *sqlx.Tx) error {
	stmt, err := tx.Prepare("delete from invite_tokens where id = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing invite_tokens delete statement")
	}
	_, err = stmt.Exec(id)
	if err != nil {
		return errors.Wrap(err, "error executing invite_tokens delete statement")
	}
	return nil
}
