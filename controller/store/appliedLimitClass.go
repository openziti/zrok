package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type AppliedLimitClass struct {
	Model
	AccountId    int
	LimitClassId int
}

func (str *Store) ApplyLimitClass(lc *AppliedLimitClass, trx *sqlx.Tx) (int, error) {
	stmt, err := trx.Prepare("insert into applied_limit_classes (account_id, limit_class_id) values ($1, $2) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing applied_limit_classes insert statement")
	}
	var id int
	if err := stmt.QueryRow(lc.AccountId, lc.LimitClassId).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing applied_limit_classes insert statement")
	}
	return id, nil
}

func (str *Store) FindLimitClassesForAccount(acctId int, trx *sqlx.Tx) ([]*LimitClass, error) {
	rows, err := trx.Queryx("select limit_classes.* from applied_limit_classes, limit_classes where applied_limit_classes.account_id = $1 and applied_limit_classes.limit_class_id = limit_classes.id", acctId)
	if err != nil {
		return nil, errors.Wrap(err, "error finding limit classes for account")
	}
	var lcs []*LimitClass
	for rows.Next() {
		lc := &LimitClass{}
		if err := rows.StructScan(&lc); err != nil {
			return nil, errors.Wrap(err, "error scanning limit_classes")
		}
		lcs = append(lcs, lc)
	}
	return lcs, nil
}
