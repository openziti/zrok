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

func TestFindAllLatestAccountLimitJournal(t *testing.T) {
	str, err := Open(&Config{Path: ":memory:", Type: "sqlite3"})
	assert.Nil(t, err)
	assert.NotNil(t, str)

	trx, err := str.Begin()
	assert.Nil(t, err)
	assert.NotNil(t, trx)

	acctId1, err := str.CreateAccount(&Account{Email: "nobody@nowehere.com", Salt: "salt1", Password: "password1", Token: "token1", Limitless: false, Deleted: false}, trx)
	assert.Nil(t, err)

	_, err = str.CreateAccountLimitJournal(&AccountLimitJournal{AccountId: acctId1, RxBytes: 2048, TxBytes: 4096, Action: WarningAction}, trx)
	assert.Nil(t, err)
	_, err = str.CreateAccountLimitJournal(&AccountLimitJournal{AccountId: acctId1, RxBytes: 2048, TxBytes: 4096, Action: ClearAction}, trx)
	assert.Nil(t, err)
	aljId13, err := str.CreateAccountLimitJournal(&AccountLimitJournal{AccountId: acctId1, RxBytes: 2048, TxBytes: 4096, Action: LimitAction}, trx)
	assert.Nil(t, err)

	acctId2, err := str.CreateAccount(&Account{Email: "someone@somewhere.com", Salt: "salt2", Password: "password2", Token: "token2", Limitless: false, Deleted: false}, trx)
	assert.Nil(t, err)

	aljId21, err := str.CreateAccountLimitJournal(&AccountLimitJournal{AccountId: acctId2, RxBytes: 2048, TxBytes: 4096, Action: WarningAction}, trx)
	assert.Nil(t, err)

	aljs, err := str.FindAllLatestAccountLimitJournal(trx)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(aljs))
	assert.Equal(t, aljId13, aljs[0].Id)
	assert.Equal(t, aljId21, aljs[1].Id)
}
