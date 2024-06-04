package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type BandwidthLimitJournalEntry struct {
	Model
	AccountId    int
	LimitClassId *int
	Action       LimitAction
	RxBytes      int64
	TxBytes      int64
}

func (str *Store) CreateBandwidthLimitJournalEntry(j *BandwidthLimitJournalEntry, trx *sqlx.Tx) (int, error) {
	stmt, err := trx.Prepare("insert into bandwidth_limit_journal (account_id, limit_class_id, action, rx_bytes, tx_bytes) values ($1, $2, $3, $4, $5) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing bandwidth_limit_journal insert statement")
	}
	var id int
	if err := stmt.QueryRow(j.AccountId, j.LimitClassId, j.Action, j.RxBytes, j.TxBytes).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing bandwidth_limit_journal insert statement")
	}
	return id, nil
}

func (str *Store) IsBandwidthLimitJournalEmpty(acctId int, trx *sqlx.Tx) (bool, error) {
	count := 0
	if err := trx.QueryRowx("select count(0) from bandwidth_limit_journal where account_id = $1", acctId).Scan(&count); err != nil {
		return false, err
	}
	return count == 0, nil
}

func (str *Store) FindLatestBandwidthLimitJournal(acctId int, trx *sqlx.Tx) (*BandwidthLimitJournalEntry, error) {
	j := &BandwidthLimitJournalEntry{}
	if err := trx.QueryRowx("select * from bandwidth_limit_journal where account_id = $1 order by id desc limit 1", acctId).StructScan(j); err != nil {
		return nil, errors.Wrap(err, "error finding bandwidth_limit_journal by account_id")
	}
	return j, nil
}

func (str *Store) IsBandwidthLimitJournalEmptyForGlobal(acctId int, trx *sqlx.Tx) (bool, error) {
	count := 0
	if err := trx.QueryRowx("select count(0) from bandwidth_limit_journal where account_id = $1 and limit_class_id is null", acctId).Scan(&count); err != nil {
		return false, err
	}
	return count == 0, nil
}

func (str *Store) FindLatestBandwidthLimitJournalForGlobal(acctId int, trx *sqlx.Tx) (*BandwidthLimitJournalEntry, error) {
	j := &BandwidthLimitJournalEntry{}
	if err := trx.QueryRowx("select * from bandwidth_limit_journal where account_id = $1 and limit_class_id is null order by id desc limit 1", acctId).Scan(&j); err != nil {
		return nil, errors.Wrap(err, "error finding bandwidth_limit_journal by account_id for global")
	}
	return j, nil
}

func (str *Store) IsBandwidthLimitJournalEmptyForLimitClass(acctId, lcId int, trx *sqlx.Tx) (bool, error) {
	count := 0
	if err := trx.QueryRowx("select count(0) from bandwidth_limit_journal where account_id = $1 and limit_class_id = $2", acctId, lcId).Scan(&count); err != nil {
		return false, err
	}
	return count == 0, nil
}

func (str *Store) FindLatestBandwidthLimitJournalForLimitClass(acctId, lcId int, trx *sqlx.Tx) (*BandwidthLimitJournalEntry, error) {
	j := &BandwidthLimitJournalEntry{}
	if err := trx.QueryRowx("select * from bandwidth_limit_journal where account_id = $1 and limit_class_id = $2 order by id desc limit 1", acctId, lcId).StructScan(j); err != nil {
		return nil, errors.Wrap(err, "error finding bandwidth_limit_journal by account_id and limit_class_id")
	}
	return j, nil
}

func (str *Store) FindAllBandwidthLimitJournal(trx *sqlx.Tx) ([]*BandwidthLimitJournalEntry, error) {
	rows, err := trx.Queryx("select * from bandwidth_limit_journal")
	if err != nil {
		return nil, errors.Wrap(err, "error finding all from bandwidth_limit_journal")
	}
	var jes []*BandwidthLimitJournalEntry
	for rows.Next() {
		je := &BandwidthLimitJournalEntry{}
		if err := rows.StructScan(je); err != nil {
			return nil, errors.Wrap(err, "error scanning bandwidth_limit_journal")
		}
		jes = append(jes, je)
	}
	return jes, nil
}

func (str *Store) FindAllLatestBandwidthLimitJournal(trx *sqlx.Tx) ([]*BandwidthLimitJournalEntry, error) {
	rows, err := trx.Queryx("select id, account_id, limit_class_id, action, rx_bytes, tx_bytes, created_at, updated_at from bandwidth_limit_journal where id in (select max(id) as id from bandwidth_limit_journal group by account_id)")
	if err != nil {
		return nil, errors.Wrap(err, "error finding all latest bandwidth_limit_journal")
	}
	var jes []*BandwidthLimitJournalEntry
	for rows.Next() {
		je := &BandwidthLimitJournalEntry{}
		if err := rows.StructScan(je); err != nil {
			return nil, errors.Wrap(err, "error scanning bandwidth_limit_journal")
		}
		jes = append(jes, je)
	}
	return jes, nil
}

func (str *Store) DeleteBandwidthLimitJournal(acctId int, trx *sqlx.Tx) error {
	if _, err := trx.Exec("delete from bandwidth_limit_journal where account_id = $1", acctId); err != nil {
		return errors.Wrapf(err, "error deleting from bandwidth_limit_journal for account_id = %d", acctId)
	}
	return nil
}

func (str *Store) DeleteBandwidthLimitJournalEntryForLimitClass(acctId int, lcId *int, trx *sqlx.Tx) error {
	if _, err := trx.Exec("delete from bandwidth_limit_journal where account_id = $1 and limit_class_id = $2", acctId, lcId); err != nil {
		return errors.Wrapf(err, "error deleting from bandwidth_limit_journal for account_id = %d and limit_class_id = %d", acctId, lcId)
	}
	return nil
}
