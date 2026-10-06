package controller

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/dynamicProxyController"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/util"
	"github.com/pkg/errors"
)

// severedNameConflict reports the live share holding a name that healSeveredName was asked to release.
type severedNameConflict struct {
	name       string
	namespace  string
	shareToken string
}

func (c *severedNameConflict) message() string {
	return fmt.Sprintf("name '%v' in namespace '%v' is in use by share '%v'; run 'zrok2 delete share %v' to release it", c.name, c.namespace, c.shareToken, c.shareToken)
}

// healSeveredName releases a name whose share no longer exists. share name mappings and dynamic frontend
// mappings for the name that point at a deleted share are removed; a mapping that points at a live share
// (or, for a frontend mapping, at no share row at all) is a conflict, returned without modifying anything.
// this is a repair of rows left behind before teardownShare released names, not a teardown path: it never
// touches ziti. it runs inside the caller's transaction, so a failure later in the request restores
// everything it removed. it returns one unbind per frontend mapping removed, for the caller to publish
// after its commit.
func healSeveredName(ns *store.Namespace, name *store.Name, trx *sqlx.Tx) (*severedNameConflict, []pendingMappingUpdate, error) {
	mappings, err := str.FindShareNameMappingsByNameIdWithShare(name.Id, trx)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "error finding share name mappings for name '%v' in namespace '%v'", name.Name, ns.Token)
	}
	for _, mapping := range mappings {
		if !mapping.ShareDeleted {
			return &severedNameConflict{name: name.Name, namespace: ns.Token, shareToken: mapping.ShareToken}, nil, nil
		}
	}

	frontends, err := str.FindDynamicFrontendsForNamespace(ns.Id, trx)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "error finding dynamic frontends for namespace '%v'", ns.Token)
	}
	frontendName := util.NameInNamespace(name.Name, ns.Name)
	var staleFrontends []*store.FrontendMappingWithShareState
	for _, frontend := range frontends {
		fm, err := str.FindFrontendMappingByFrontendTokenAndNameWithShareState(frontend.Token, frontendName, trx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return nil, nil, errors.Wrapf(err, "error finding frontend mapping for name '%v' on frontend '%v'", frontendName, frontend.Token)
		}
		if fm.ShareDeleted == nil || !*fm.ShareDeleted {
			return &severedNameConflict{name: name.Name, namespace: ns.Token, shareToken: fm.ShareToken}, nil, nil
		}
		staleFrontends = append(staleFrontends, fm)
	}

	for _, mapping := range mappings {
		if err := str.DeleteShareNameMapping(mapping.Id, trx); err != nil {
			return nil, nil, errors.Wrapf(err, "error deleting severed share name mapping '%v' for name '%v' in namespace '%v'", mapping.Id, name.Name, ns.Token)
		}
		dl.Infof("healed severed name '%v' in namespace '%v': released mapping '%v' to deleted share '%v'", name.Name, ns.Token, mapping.Id, mapping.ShareToken)
	}
	var unbinds []pendingMappingUpdate
	for _, fm := range staleFrontends {
		if err := str.DeleteFrontendMappingsByFrontendTokenAndName(fm.FrontendToken, fm.Name, trx); err != nil {
			return nil, nil, errors.Wrapf(err, "error deleting stale frontend mapping for name '%v' on frontend '%v'", fm.Name, fm.FrontendToken)
		}
		dl.Infof("healed severed name '%v' in namespace '%v': removed frontend mapping '%v' on frontend '%v' for deleted share '%v'", name.Name, ns.Token, fm.Name, fm.FrontendToken, fm.ShareToken)
		unbinds = append(unbinds, pendingMappingUpdate{
			frontendToken: fm.FrontendToken,
			mapping:       dynamicProxyController.Mapping{Operation: dynamicProxyController.OperationUnbind, Name: fm.Name},
		})
	}

	return nil, unbinds, nil
}
