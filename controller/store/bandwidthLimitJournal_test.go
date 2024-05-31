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
		LimitScope:    AccountLimitScope,
		LimitAction:   LimitLimitAction,
		ShareMode:     sdk.PrivateShareMode,
		BackendMode:   sdk.VpnBackendMode,
		PeriodMinutes: 60,
		RxBytes:       4096,
		TxBytes:       8192,
		TotalBytes:    10240,
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
