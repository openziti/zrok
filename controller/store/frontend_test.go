package store

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPublicFrontend(t *testing.T) {
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

	feName := "public"
	feId, err := str.CreateFrontend(envId, &Frontend{
		Token:      "token",
		ZId:        "zId0",
		PublicName: &feName,
	}, trx)
	assert.Nil(t, err)

	fe, err := str.GetFrontend(feId, trx)
	assert.Nil(t, err)
	assert.NotNil(t, fe)
	assert.Equal(t, envId, *fe.EnvironmentId)
	assert.Equal(t, feName, *fe.PublicName)
	assert.False(t, fe.Deleted)

	fe0, err := str.FindFrontendPubliclyNamed(feName, trx)
	assert.Nil(t, err)
	assert.NotNil(t, fe0)
	assert.EqualValues(t, fe, fe0)

	err = str.DeleteFrontend(fe.Id, trx)
	assert.Nil(t, err)

	fe0, err = str.FindFrontendWithToken(feName, trx)
	assert.NotNil(t, err)
	assert.Nil(t, fe0)

	fe0, err = str.GetFrontend(fe.Id, trx)
	assert.Nil(t, err)
	assert.NotNil(t, fe0)
	assert.True(t, fe0.Deleted)
}
