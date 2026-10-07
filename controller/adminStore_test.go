package controller

import (
	"path/filepath"
	"testing"

	controllerConfig "github.com/openziti/zrok/v2/controller/config"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/stretchr/testify/require"
)

// behindSchemaStore writes a sqlite store migrated fully and then rolled back past the v2 name tables, and
// returns its config (auto-migration left enabled, as a controller config would have it) and the number
// of migrations it has applied.
func behindSchemaStore(t *testing.T) (*store.Config, int) {
	t.Helper()
	cfg := &store.Config{Path: filepath.Join(t.TempDir(), "behind.db"), Type: "sqlite3"}
	v, err := store.Open(cfg)
	require.NoError(t, err)
	require.NoError(t, v.MigrateDown(cfg, 6))
	require.NoError(t, v.Close())
	return cfg, appliedMigrations(t, cfg)
}

func appliedMigrations(t *testing.T, cfg *store.Config) int {
	t.Helper()
	v, err := store.Open(&store.Config{Path: cfg.Path, Type: cfg.Type, DisableAutoMigration: true})
	require.NoError(t, err)
	defer func() { require.NoError(t, v.Close()) }()
	trx, err := v.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()
	var count int
	require.NoError(t, trx.QueryRow("select count(*) from migrations").Scan(&count))
	return count
}

func TestAdminStoreNeverMigrates(t *testing.T) {
	storeCfg, applied := behindSchemaStore(t)

	v, err := openAdminStore(storeCfg)
	require.NoError(t, err)
	require.NoError(t, v.Close())

	require.Equal(t, applied, appliedMigrations(t, storeCfg))
	require.False(t, storeCfg.DisableAutoMigration, "the caller's config is not modified")
}

func TestRepairStoreFailsOnBehindSchemaStore(t *testing.T) {
	prevStore, prevCfg := str, cfg
	t.Cleanup(func() { str, cfg = prevStore, prevCfg })
	storeCfg, applied := behindSchemaStore(t)
	c := controllerConfig.DefaultConfig()
	c.Store = storeCfg

	err := RepairStore(c, RepairStoreOptions{Batch: DefaultRepairStoreBatch})

	require.Error(t, err)
	require.Contains(t, err.Error(), "no such table")
	require.Equal(t, applied, appliedMigrations(t, storeCfg))
}
