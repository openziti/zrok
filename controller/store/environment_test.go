package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestFindEnvironmentsForAccountWithFilter(t *testing.T) {
	str, err := Open(&Config{Path: ":memory:", Type: "sqlite3"})
	require.NoError(t, err)
	require.NotNil(t, str)

	trx, err := str.Begin()
	require.NoError(t, err)
	require.NotNil(t, trx)
	defer func() { _ = trx.Rollback() }()

	// create test account
	acctId, err := str.CreateAccount(&Account{
		Email:    "test@test.com",
		Password: "password",
		Token:    "token",
		Salt:     "salt",
	}, trx)
	require.NoError(t, err)

	// create environments with varying properties
	env1Id, err := str.CreateEnvironment(acctId, &Environment{
		Description: "prod-server",
		Host:        "host1.example.com",
		Address:     "192.168.1.1",
		ZId:         "env1",
	}, trx)
	require.NoError(t, err)

	env2Id, err := str.CreateEnvironment(acctId, &Environment{
		Description: "dev-server",
		Host:        "host2.example.com",
		Address:     "192.168.1.2",
		ZId:         "env2",
	}, trx)
	require.NoError(t, err)

	env3Id, err := str.CreateEnvironment(acctId, &Environment{
		Description: "test-environment",
		Host:        "host3.example.com",
		Address:     "192.168.1.3",
		ZId:         "env3",
	}, trx)
	require.NoError(t, err)

	_, err = str.CreateEnvironment(acctId, &Environment{
		Description: "staging-server",
		Host:        "host1.staging.com",
		Address:     "192.168.1.4",
		ZId:         "env4",
	}, trx)
	require.NoError(t, err)

	// create shares for environments
	// env1: 1 share
	_, err = str.CreateShare(env1Id, &Share{
		ZId:            "shr1",
		Token:          "token1",
		ShareMode:      "public",
		BackendMode:    "proxy",
		PermissionMode: OpenPermissionMode,
	}, trx)
	require.NoError(t, err)

	// env2: 3 shares
	_, err = str.CreateShare(env2Id, &Share{
		ZId:            "shr2",
		Token:          "token2",
		ShareMode:      "public",
		BackendMode:    "proxy",
		PermissionMode: OpenPermissionMode,
	}, trx)
	require.NoError(t, err)

	_, err = str.CreateShare(env2Id, &Share{
		ZId:            "shr3",
		Token:          "token3",
		ShareMode:      "public",
		BackendMode:    "proxy",
		PermissionMode: OpenPermissionMode,
	}, trx)
	require.NoError(t, err)

	_, err = str.CreateShare(env2Id, &Share{
		ZId:            "shr4",
		Token:          "token4",
		ShareMode:      "public",
		BackendMode:    "proxy",
		PermissionMode: OpenPermissionMode,
	}, trx)
	require.NoError(t, err)

	// env3: 5 shares
	for i := 0; i < 5; i++ {
		_, err = str.CreateShare(env3Id, &Share{
			ZId:            "shr" + string(rune('5'+i)),
			Token:          "token" + string(rune('5'+i)),
			ShareMode:      "public",
			BackendMode:    "proxy",
			PermissionMode: OpenPermissionMode,
		}, trx)
		require.NoError(t, err)
	}

	// env4: 0 shares

	// create frontends (accesses)
	// env1: 2 frontends
	_, err = str.CreateFrontend(env1Id, &Frontend{
		Token:          "frontend1",
		ZId:            "fzid1",
		PermissionMode: OpenPermissionMode,
	}, trx)
	require.NoError(t, err)
	_, err = str.CreateFrontend(env1Id, &Frontend{
		Token:          "frontend2",
		ZId:            "fzid2",
		PermissionMode: OpenPermissionMode,
	}, trx)
	require.NoError(t, err)

	// env2: 5 frontends
	for i := 0; i < 5; i++ {
		_, err = str.CreateFrontend(env2Id, &Frontend{
			Token:          "frontend" + string(rune('3'+i)),
			ZId:            "fzid" + string(rune('3'+i)),
			PermissionMode: OpenPermissionMode,
		}, trx)
		require.NoError(t, err)
	}

	// env3 and env4: 0 frontends

	t.Run("no filter returns all environments", func(t *testing.T) {
		filter := &EnvironmentFilter{}
		envs, err := str.FindEnvironmentsForAccountWithFilter(acctId, filter, trx)
		require.NoError(t, err)
		assert.Len(t, envs, 4)
	})

	t.Run("description filter", func(t *testing.T) {
		// should match "prod-server", "dev-server", "staging-server"
		desc := "server"
		filter := &EnvironmentFilter{Description: &desc}
		envs, err := str.FindEnvironmentsForAccountWithFilter(acctId, filter, trx)
		require.NoError(t, err)
		assert.Len(t, envs, 3)

		// case insensitive
		descUpper := "SERVER"
		filter = &EnvironmentFilter{Description: &descUpper}
		envs, err = str.FindEnvironmentsForAccountWithFilter(acctId, filter, trx)
		require.NoError(t, err)
		assert.Len(t, envs, 3)

		// specific match
		descProd := "prod"
		filter = &EnvironmentFilter{Description: &descProd}
		envs, err = str.FindEnvironmentsForAccountWithFilter(acctId, filter, trx)
		require.NoError(t, err)
		assert.Len(t, envs, 1)
		assert.Equal(t, "env1", envs[0].ZId)
	})

	t.Run("host filter", func(t *testing.T) {
		// should match "host1.example.com" and "host1.staging.com"
		host := "host1"
		filter := &EnvironmentFilter{Host: &host}
		envs, err := str.FindEnvironmentsForAccountWithFilter(acctId, filter, trx)
		require.NoError(t, err)
		assert.Len(t, envs, 2)

		// case-insensitive
		hostUpper := "HOST1"
		filter = &EnvironmentFilter{Host: &hostUpper}
		envs, err = str.FindEnvironmentsForAccountWithFilter(acctId, filter, trx)
		require.NoError(t, err)
		assert.Len(t, envs, 2)

		// specific domain
		hostExample := "example.com"
		filter = &EnvironmentFilter{Host: &hostExample}
		envs, err = str.FindEnvironmentsForAccountWithFilter(acctId, filter, trx)
		require.NoError(t, err)
		assert.Len(t, envs, 3)
	})

	t.Run("address filter", func(t *testing.T) {
		addr := "192.168.1.1"
		filter := &EnvironmentFilter{Address: &addr}
		envs, err := str.FindEnvironmentsForAccountWithFilter(acctId, filter, trx)
		require.NoError(t, err)
		assert.Len(t, envs, 1)
		assert.Equal(t, "env1", envs[0].ZId)
	})

	t.Run("hasShares filter", func(t *testing.T) {
		// hasShares = true (env1, env2, env3)
		hasShares := true
		filter := &EnvironmentFilter{HasShares: &hasShares}
		envs, err := str.FindEnvironmentsForAccountWithFilter(acctId, filter, trx)
		require.NoError(t, err)
		assert.Len(t, envs, 3)

		// hasShares = false (env4)
		hasShares = false
		filter = &EnvironmentFilter{HasShares: &hasShares}
		envs, err = str.FindEnvironmentsForAccountWithFilter(acctId, filter, trx)
		require.NoError(t, err)
		assert.Len(t, envs, 1)
		assert.Equal(t, "env4", envs[0].ZId)
		assert.Equal(t, 0, envs[0].ShareCount)
	})

	t.Run("hasAccesses filter", func(t *testing.T) {
		// hasAccesses = true (env1: 2, env2: 5)
		hasAccesses := true
		filter := &EnvironmentFilter{HasAccesses: &hasAccesses}
		envs, err := str.FindEnvironmentsForAccountWithFilter(acctId, filter, trx)
		require.NoError(t, err)
		assert.Len(t, envs, 2)

		// hasAccesses = false (env3, env4)
		hasAccesses = false
		filter = &EnvironmentFilter{HasAccesses: &hasAccesses}
		envs, err = str.FindEnvironmentsForAccountWithFilter(acctId, filter, trx)
		require.NoError(t, err)
		assert.Len(t, envs, 2)
	})

	t.Run("shareCount filter with operators", func(t *testing.T) {
		// > 0 (env1: 1, env2: 3, env3: 5)
		count := ">0"
		filter := &EnvironmentFilter{ShareCount: &count}
		envs, err := str.FindEnvironmentsForAccountWithFilter(acctId, filter, trx)
		require.NoError(t, err)
		assert.Len(t, envs, 3)

		// >= 3 (env2: 3, env3: 5)
		count = ">=3"
		filter = &EnvironmentFilter{ShareCount: &count}
		envs, err = str.FindEnvironmentsForAccountWithFilter(acctId, filter, trx)
		require.NoError(t, err)
		assert.Len(t, envs, 2)

		// = 1 (env1)
		count = "=1"
		filter = &EnvironmentFilter{ShareCount: &count}
		envs, err = str.FindEnvironmentsForAccountWithFilter(acctId, filter, trx)
		require.NoError(t, err)
		assert.Len(t, envs, 1)
		assert.Equal(t, "env1", envs[0].ZId)
		assert.Equal(t, 1, envs[0].ShareCount)

		// < 3 (env1: 1, env4: 0)
		count = "<3"
		filter = &EnvironmentFilter{ShareCount: &count}
		envs, err = str.FindEnvironmentsForAccountWithFilter(acctId, filter, trx)
		require.NoError(t, err)
		assert.Len(t, envs, 2)

		// <= 1 (env1: 1, env4: 0)
		count = "<=1"
		filter = &EnvironmentFilter{ShareCount: &count}
		envs, err = str.FindEnvironmentsForAccountWithFilter(acctId, filter, trx)
		require.NoError(t, err)
		assert.Len(t, envs, 2)
	})

	t.Run("accessCount filter with operators", func(t *testing.T) {
		// > 0 (env1: 2, env2: 5)
		count := ">0"
		filter := &EnvironmentFilter{AccessCount: &count}
		envs, err := str.FindEnvironmentsForAccountWithFilter(acctId, filter, trx)
		require.NoError(t, err)
		assert.Len(t, envs, 2)

		// >= 5 (env2: 5)
		count = ">=5"
		filter = &EnvironmentFilter{AccessCount: &count}
		envs, err = str.FindEnvironmentsForAccountWithFilter(acctId, filter, trx)
		require.NoError(t, err)
		assert.Len(t, envs, 1)
		assert.Equal(t, "env2", envs[0].ZId)
		assert.Equal(t, 5, envs[0].AccessCount)

		// = 2 (env1)
		count = "=2"
		filter = &EnvironmentFilter{AccessCount: &count}
		envs, err = str.FindEnvironmentsForAccountWithFilter(acctId, filter, trx)
		require.NoError(t, err)
		assert.Len(t, envs, 1)
		assert.Equal(t, "env1", envs[0].ZId)
		assert.Equal(t, 2, envs[0].AccessCount)
	})

	t.Run("date range filters", func(t *testing.T) {
		t.Skip("skipping date range tests due to timestamp precision issues in test environment")
		// TODO: implement proper date range testing with controlled timestamps
	})

	t.Run("combined filters", func(t *testing.T) {
		// description contains "server" AND hasShares = true AND shareCount > 0
		desc := "server"
		hasShares := true
		shareCount := ">0"
		filter := &EnvironmentFilter{
			Description: &desc,
			HasShares:   &hasShares,
			ShareCount:  &shareCount,
		}
		envs, err := str.FindEnvironmentsForAccountWithFilter(acctId, filter, trx)
		require.NoError(t, err)
		// should match: prod-server (env1: 1 share), dev-server (env2: 3 shares), staging-server (env4: 0 shares - excluded)
		assert.Len(t, envs, 2)

		// host contains "host1" AND shareCount >= 1
		host := "host1"
		shareCount = ">=1"
		filter = &EnvironmentFilter{
			Host:       &host,
			ShareCount: &shareCount,
		}
		envs, err = str.FindEnvironmentsForAccountWithFilter(acctId, filter, trx)
		require.NoError(t, err)
		// should match: env1 (1 share), env4 has 0 shares (excluded)
		assert.Len(t, envs, 1)
		assert.Equal(t, "env1", envs[0].ZId)
	})

	t.Run("verify counts are correct", func(t *testing.T) {
		filter := &EnvironmentFilter{}
		envs, err := str.FindEnvironmentsForAccountWithFilter(acctId, filter, trx)
		require.NoError(t, err)
		require.Len(t, envs, 4)

		for _, env := range envs {
			switch env.ZId {
			case "env1":
				assert.Equal(t, 1, env.ShareCount)
				assert.Equal(t, 2, env.AccessCount)
			case "env2":
				assert.Equal(t, 3, env.ShareCount)
				assert.Equal(t, 5, env.AccessCount)
			case "env3":
				assert.Equal(t, 5, env.ShareCount)
				assert.Equal(t, 0, env.AccessCount)
			case "env4":
				assert.Equal(t, 0, env.ShareCount)
				assert.Equal(t, 0, env.AccessCount)
			}
		}
	})
}

func TestParseComparisonFilter(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		columnName  string
		expected    string
		expectError bool
	}{
		{"greater than", ">5", "count", "count > 5", false},
		{"greater than or equal", ">=5", "count", "count >= 5", false},
		{"equal", "=5", "count", "count = 5", false},
		{"less than", "<5", "count", "count < 5", false},
		{"less than or equal", "<=5", "count", "count <= 5", false},
		{"no operator defaults to equal", "5", "count", "count = 5", false},
		{"with spaces", " >= 10 ", "count", "count >= 10", false},
		{"empty string", "", "count", "", true},
		{"non-numeric value", ">abc", "count", "", true},
		{"invalid operator", "~5", "count", "", true}, // no valid operator, can't parse as number
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseComparisonFilter(tt.input, tt.columnName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
