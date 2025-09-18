package store

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEphemeralEnvironment(t *testing.T) {
	str, err := Open(&Config{Path: ":memory:", Type: "sqlite3"})
	assert.Nil(t, err)
	assert.NotNil(t, str)

	trx, err := str.Begin()
	assert.Nil(t, err)
	assert.NotNil(t, trx)

	envId, err := str.CreateEphemeralEnvironment(&Environment{
		Description: "description",
		Host:        "host",
		Address:     "address",
		ZId:         "zId0",
	}, trx)
	assert.Nil(t, err)

	env, err := str.GetEnvironment(envId, trx)
	assert.Nil(t, err)
	assert.NotNil(t, env)
	assert.Nil(t, env.AccountId)
	assert.False(t, env.Deleted)
}

func TestEnvironment(t *testing.T) {
	str, err := Open(&Config{Path: ":memory:", Type: "sqlite3"})
	assert.Nil(t, err)
	assert.NotNil(t, str)

	trx, err := str.Begin()
	assert.Nil(t, err)
	assert.NotNil(t, trx)

	acctId, err := str.CreateAccount(&Account{
		Email:    "test@test.com",
		Password: "password",
		Token:    "token",
	}, trx)
	assert.Nil(t, err)

	envId, err := str.CreateEnvironment(acctId, &Environment{
		Description: "description",
		Host:        "host",
		Address:     "address",
		ZId:         "zId0",
	}, trx)
	assert.Nil(t, err)

	env, err := str.GetEnvironment(envId, trx)
	assert.Nil(t, err)
	assert.NotNil(t, env)
	assert.NotNil(t, env.AccountId)
	assert.Equal(t, acctId, *env.AccountId)
	assert.False(t, env.Deleted)
}
