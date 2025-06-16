package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Secrets struct {
	Model
	ShareId int
	Secrets []Secret
}

type Secret struct {
	Key   string
	Value string
}

func (str *Store) CreateSecrets(secrets Secrets, trx *sqlx.Tx) error {
	for _, secret := range secrets.Secrets {
		stmt, err := trx.Prepare("insert into secrets (share_id, key, value) values ($1, $2, $3)")
		if err != nil {
			return errors.Wrap(err, "error preparing secrets insert statement")
		}
		_, err = stmt.Exec(secrets.ShareId, secret.Key, secret.Value)
		if err != nil {
			return errors.Wrap(err, "error executing secrets insert statement")
		}
	}
	return nil
}

func (str *Store) GetSecrets(shareId int, trx *sqlx.Tx) (Secrets, error) {
	secrets := Secrets{}
	rows, err := trx.Queryx("select * from secrets where share_id = $1 and not deleted", shareId)
	if err != nil {
		return Secrets{}, errors.Wrap(err, "error getting all from secrets")
	}
	for rows.Next() {
		secret := Secret{}
		if err := rows.StructScan(&secret); err != nil {
			return Secrets{}, errors.Wrap(err, "error scanning secrets")
		}
		secrets.Secrets = append(secrets.Secrets, secret)
	}
	return secrets, nil
}
