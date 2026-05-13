package store

import (
	stderrors "errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindAccountWithEmailReturnsNotFoundSentinel(t *testing.T) {
	str, err := Open(&Config{Path: ":memory:", Type: "sqlite3"})
	require.NoError(t, err)

	trx, err := str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()

	account, err := str.FindAccountWithEmail("missing@test.com", trx)
	require.Nil(t, account)
	require.ErrorIs(t, err, ErrAccountNotFound)
	require.True(t, stderrors.Is(err, ErrAccountNotFound))
}

func TestFindAccountWithEmailFindsExistingAccount(t *testing.T) {
	str, err := Open(&Config{Path: ":memory:", Type: "sqlite3"})
	require.NoError(t, err)

	trx, err := str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()

	_, err = str.CreateAccount(&Account{
		Email:    "test@test.com",
		Password: "password",
		Token:    "token",
		Salt:     "salt",
	}, trx)
	require.NoError(t, err)

	account, err := str.FindAccountWithEmail("TEST@test.com", trx)
	require.NoError(t, err)
	require.NotNil(t, account)
	require.Equal(t, "test@test.com", account.Email)
}
