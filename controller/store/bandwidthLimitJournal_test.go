package store

import (
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBandwidthLimitJournal(t *testing.T) {
	str, err := Open(&Config{Path: ":memory:", Type: "sqlite3"})
	assert.NoError(t, err)
	assert.NotNil(t, str)

	trx, err := str.Begin()
	assert.NoError(t, err)
	assert.NotNil(t, trx)

	jEmpty, err := str.IsBandwidthLimitJournalEmpty(1, trx)
	assert.NoError(t, err)
	assert.True(t, jEmpty)

	acctId, err := str.CreateAccount(&Account{Email: "nobody@nowhere.com", Salt: "salt", Password: "password", Token: "token"}, trx)
	assert.NoError(t, err)

	_, err = str.CreateBandwidthLimitJournalEntry(&BandwidthLimitJournalEntry{AccountId: acctId, Action: WarningLimitAction, RxBytes: 1024, TxBytes: 2048}, trx)
	assert.NoError(t, err)

	jEmpty, err = str.IsBandwidthLimitJournalEmpty(acctId, trx)
	assert.NoError(t, err)
	assert.False(t, jEmpty)

	latestJe, err := str.FindLatestBandwidthLimitJournal(acctId, trx)
	assert.NoError(t, err)
	assert.NotNil(t, latestJe)
	assert.Nil(t, latestJe.LimitClassId)
	assert.Equal(t, WarningLimitAction, latestJe.Action)
	assert.Equal(t, int64(1024), latestJe.RxBytes)
	assert.Equal(t, int64(2048), latestJe.TxBytes)

	lcId, err := str.CreateLimitClass(&LimitClass{
		ShareMode:     sdk.PrivateShareMode,
		BackendMode:   sdk.VpnBackendMode,
		PeriodMinutes: 60,
		RxBytes:       4096,
		TxBytes:       8192,
		TotalBytes:    10240,
		LimitAction:   LimitLimitAction,
	}, trx)
	assert.NoError(t, err)

	_, err = str.CreateBandwidthLimitJournalEntry(&BandwidthLimitJournalEntry{AccountId: acctId, LimitClassId: &lcId, Action: LimitLimitAction, RxBytes: 10240, TxBytes: 20480}, trx)
	assert.NoError(t, err)

	latestJe, err = str.FindLatestBandwidthLimitJournal(acctId, trx)
	assert.NoError(t, err)
	assert.NotNil(t, latestJe)
	assert.NotNil(t, latestJe.LimitClassId)
	assert.Equal(t, lcId, *latestJe.LimitClassId)
	assert.Equal(t, LimitLimitAction, latestJe.Action)
	assert.Equal(t, int64(10240), latestJe.RxBytes)
	assert.Equal(t, int64(20480), latestJe.TxBytes)
}

func TestFindAllBandwidthLimitJournal(t *testing.T) {
	str, err := Open(&Config{Path: ":memory:", Type: "sqlite3"})
	assert.Nil(t, err)
	assert.NotNil(t, str)

	trx, err := str.Begin()
	assert.Nil(t, err)
	assert.NotNil(t, trx)

	acctId1, err := str.CreateAccount(&Account{Email: "nobody@nowehere.com", Salt: "salt1", Password: "password1", Token: "token1", Limitless: false, Deleted: false}, trx)
	assert.Nil(t, err)

	_, err = str.CreateBandwidthLimitJournalEntry(&BandwidthLimitJournalEntry{AccountId: acctId1, Action: WarningLimitAction, RxBytes: 2048, TxBytes: 4096}, trx)
	assert.Nil(t, err)
	_, err = str.CreateBandwidthLimitJournalEntry(&BandwidthLimitJournalEntry{AccountId: acctId1, Action: LimitLimitAction, RxBytes: 2048, TxBytes: 4096}, trx)
	assert.Nil(t, err)
	aljId13, err := str.CreateBandwidthLimitJournalEntry(&BandwidthLimitJournalEntry{AccountId: acctId1, Action: LimitLimitAction, RxBytes: 8192, TxBytes: 10240}, trx)
	assert.Nil(t, err)

	acctId2, err := str.CreateAccount(&Account{Email: "someone@somewhere.com", Salt: "salt2", Password: "password2", Token: "token2", Limitless: false, Deleted: false}, trx)
	assert.Nil(t, err)

	aljId21, err := str.CreateBandwidthLimitJournalEntry(&BandwidthLimitJournalEntry{AccountId: acctId2, Action: WarningLimitAction, RxBytes: 2048, TxBytes: 4096}, trx)
	assert.Nil(t, err)

	aljs, err := str.FindAllLatestBandwidthLimitJournal(trx)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(aljs))
	assert.Equal(t, aljId13, aljs[0].Id)
	assert.Equal(t, aljId21, aljs[1].Id)
}
