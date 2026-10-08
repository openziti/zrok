package controller

import (
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/pkg/errors"
)

// openAdminStore opens the store for an admin repair command (gc, repair-dial-policies, repair-store)
// with auto-migration forced off, whatever the configuration says. these commands default to a dry run,
// and a dry run must never change the store, its schema included; a store behind this binary's schema
// makes the command fail instead of being migrated by it. migrations belong to the controller and to
// 'zrok2 admin migrate'.
func openAdminStore(cfg *store.Config) (*store.Store, error) {
	if cfg == nil {
		return nil, errors.New("no store configured")
	}
	noMigrate := *cfg
	noMigrate.DisableAutoMigration = true
	v, err := store.Open(&noMigrate)
	if err != nil {
		return nil, errors.Wrap(err, "error opening store")
	}
	return v, nil
}
