package store

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAccountLimitJournal(t *testing.T) {
	str, err := Open(&Config{Path: ":memory:", Type: "sqlite3"})
	assert.Nil(t, err)
	assert.NotNil(t, str)

	trx, err := str.Begin()
	assert.Nil(t, err)
	assert.NotNil(t, trx)

	aljEmpty, err := str.IsAccountLimitJournalEmpty(1, trx)
	assert.Nil(t, err)
	assert.True(t, aljEmpty)

	acctId, err := str.CreateAccount(&Account{Email: "nobody@nowehere.com", Salt: "salt", Password: "password", Token: "token", Limitless: false, Deleted: false}, trx)
	assert.Nil(t, err)

	_, err = str.CreateAccountLimitJournal(&AccountLimitJournal{AccountId: acctId, RxBytes: 1024, TxBytes: 2048, Action: WarningAction}, trx)
	assert.Nil(t, err)

	aljEmpty, err = str.IsAccountLimitJournalEmpty(acctId, trx)
	assert.Nil(t, err)
	assert.False(t, aljEmpty)

	latestAlj, err := str.FindLatestAccountLimitJournal(acctId, trx)
	assert.Nil(t, err)
	assert.NotNil(t, latestAlj)
	assert.Equal(t, int64(1024), latestAlj.RxBytes)
	assert.Equal(t, int64(2048), latestAlj.TxBytes)
	assert.Equal(t, WarningAction, latestAlj.Action)

	_, err = str.CreateAccountLimitJournal(&AccountLimitJournal{AccountId: acctId, RxBytes: 2048, TxBytes: 4096, Action: LimitAction}, trx)
	assert.Nil(t, err)

	latestAlj, err = str.FindLatestAccountLimitJournal(acctId, trx)
	assert.Nil(t, err)
	assert.NotNil(t, latestAlj)
	assert.Equal(t, int64(2048), latestAlj.RxBytes)
	assert.Equal(t, int64(4096), latestAlj.TxBytes)
	assert.Equal(t, LimitAction, latestAlj.Action)
}
