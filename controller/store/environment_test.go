package store

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEphemeralEnvironment(t *testing.T) {
	str, err := Open(&Config{Path: ":memory:", Type: "sqlite3"})
	assert.Nil(t, err)
	assert.NotNil(t, str)

	tx, err := str.Begin()
	assert.Nil(t, err)
	assert.NotNil(t, tx)

	envId, err := str.CreateEphemeralEnvironment(&Environment{
		Description: "description",
		Host:        "host",
		Address:     "address",
		ZId:         "zId0",
	}, tx)
	assert.Nil(t, err)

	env, err := str.GetEnvironment(envId, tx)
	assert.Nil(t, err)
	assert.NotNil(t, env)
	assert.Nil(t, env.AccountId)
}
