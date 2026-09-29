package controller

import (
	"github.com/jmoiron/sqlx"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/automation"
	"github.com/openziti/zrok/v2/controller/dynamicProxyController"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/pkg/errors"
)

// teardownShare is the only place a share row is deleted. it runs inside the caller's transaction and
// returns the unbind updates for the frontend mappings it deleted, which the caller publishes after it
// commits. any failure returns at once, so the caller's rollback leaves every row for a retry. ziti goes
// last, so a ziti failure leaves the store untouched.
func teardownShare(shr *store.Share, trx *sqlx.Tx, ziti *automation.ZitiAutomation) ([]pendingMappingUpdate, error) {
	// release the share's names: auto-allocated names are deleted with their mappings, reserved names
	// survive
	details, err := str.FindShareNameCleanupDetailsByShareId(shr.Id, trx)
	if err != nil {
		return nil, errors.Wrapf(err, "error finding share name mappings for share '%v'", shr.Token)
	}
	for _, detail := range details {
		if !detail.Reserved && !detail.NameDeleted {
			if err := str.DeleteName(detail.NameId, trx); err != nil {
				return nil, errors.Wrapf(err, "error deleting allocated name '%v' for share '%v'", detail.Name, shr.Token)
			}
			dl.Debugf("deleted allocated name '%v' for share '%v'", detail.Name, shr.Token)
		}
		if err := str.DeleteShareNameMapping(detail.MappingId, trx); err != nil {
			return nil, errors.Wrapf(err, "error deleting share name mapping '%v' for share '%v'", detail.MappingId, shr.Token)
		}
	}

	// frontend mappings are deleted whether or not a dynamic proxy controller is configured
	fms, err := str.FindFrontendMappingsByShareToken(shr.Token, trx)
	if err != nil {
		return nil, errors.Wrapf(err, "error finding frontend mappings for share '%v'", shr.Token)
	}
	if err := str.DeleteFrontendMappingsByShareToken(shr.Token, trx); err != nil {
		return nil, errors.Wrapf(err, "error deleting frontend mappings for share '%v'", shr.Token)
	}
	var updates []pendingMappingUpdate
	for _, fm := range fms {
		updates = append(updates, pendingMappingUpdate{
			frontendToken: fm.FrontendToken,
			mapping:       dynamicProxyController.Mapping{Operation: dynamicProxyController.OperationUnbind, Name: fm.Name},
		})
	}

	if err := str.DeleteAccessGrantsForShare(shr.Id, trx); err != nil {
		return nil, errors.Wrapf(err, "error deleting access grants for share '%v'", shr.Token)
	}

	if err := str.DeleteShare(shr.Id, trx); err != nil {
		return nil, errors.Wrapf(err, "error deleting share '%v'", shr.Token)
	}

	// a not-found inside is tolerated by the filter deletes; anything else fails the teardown
	if err := ziti.CleanupByTag("zrokShareToken", shr.Token); err != nil {
		return nil, errors.Wrapf(err, "error cleaning up ziti objects for share '%v'", shr.Token)
	}

	dl.Infof("tore down share '%v'", shr.Token)
	return updates, nil
}
