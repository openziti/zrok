package store

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Invite struct {
	Model
	Token  string
	Status string `db:"token_status"`
}

const (
	INVITE_STATUS_UNUSED = "UNUSED"
	INVITE_STATUS_TAKEN  = "TAKEN"
)

func (str *Store) CreateInvites(invites []*Invite, tx *sqlx.Tx) error {
	sql := "insert into invites (token, token_status) values %s"
	invs := make([]any, len(invites)*2)
	queries := make([]string, len(invites))
	ct := 1
	for i, inv := range invites {
		invs[i] = inv.Token
		invs[i+1] = inv.Status
		queries[i] = fmt.Sprintf("($%d, $%d)", ct, ct+1)
		ct = ct + 2
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

func (str *Store) GetInvite(tx *sqlx.Tx) (*Invite, error) {
	invite := &Invite{}
	if err := tx.QueryRowx("select * from invites where token_status = $1 limit 1", INVITE_STATUS_UNUSED).StructScan(invite); err != nil {
		return nil, errors.Wrap(err, "error getting unused invite")
	}
	return invite, nil
}

func (str *Store) UpdateInvite(invite *Invite, tx *sqlx.Tx) error {
	stmt, err := tx.Prepare("update invites set token = $1, token_status = $2")
	if err != nil {
		return errors.Wrap(err, "error perparing invites update statement")
	}
	_, err = stmt.Exec(invite.Token, invite.Status)
	if err != nil {
		return errors.Wrap(err, "error executing invites update statement")
	}
	return nil
}
