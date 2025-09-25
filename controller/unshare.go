package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/controller/automation"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/share"
	"github.com/openziti/zrok/util"
	"github.com/pkg/errors"
)

type unshareHandler struct{}

func newUnshareHandler() *unshareHandler {
	return &unshareHandler{}
}

func (h *unshareHandler) Handle(params share.UnshareParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction for '%v': %v", principal.Email, err)
		return share.NewUnshareInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	shrToken := params.Body.ShareToken
	envZId := params.Body.EnvZID

	// validate environment
	env, err := h.validateEnvironment(envZId, principal, trx)
	if err != nil {
		dl.Errorf("environment validation failed for '%v': %v", principal.Email, err)
		return share.NewUnshareNotFound()
	}

	// find and validate share
	shr, err := h.findAndValidateShare(shrToken, env, trx)
	if err != nil {
		dl.Errorf("share validation failed for '%v': %v", principal.Email, err)
		return share.NewUnshareNotFound()
	}

	// deallocate ziti resources using automation framework
	if err := h.deallocateResources(shrToken); err != nil {
		dl.Warnf("error deallocating ziti resources for share '%v': %v", shrToken, err)
	}

	// send unbind mapping updates before cleaning up share name mappings
	if err := h.processDynamicMappings(shr.Id, trx); err != nil {
		dl.Errorf("error sending unbind mapping updates for '%v': %v", shrToken, err)
	}

	// clean up share name mappings
	if err := h.cleanupShareNameMappings(shr.Id, trx); err != nil {
		dl.Errorf("error cleaning up share name mappings for '%v': %v", shrToken, err)
		return share.NewUnshareInternalServerError()
	}

	// clean up access grants
	if err := str.DeleteAccessGrantsForShare(shr.Id, trx); err != nil {
		dl.Errorf("error deleting access grants for share '%v': %v", shrToken, err)
		return share.NewUnshareInternalServerError()
	}

	// delete the share record
	if err := str.DeleteShare(shr.Id, trx); err != nil {
		dl.Errorf("error deleting share '%v': %v", shrToken, err)
		return share.NewUnshareInternalServerError()
	}

	// commit transaction
	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing transaction for '%v': %v", shrToken, err)
		return share.NewUnshareInternalServerError()
	}

	dl.Infof("successfully unshared '%v' for '%v'", shrToken, principal.Email)
	return share.NewUnshareOK()
}

func (h *unshareHandler) validateEnvironment(envZId string, principal *rest_model_zrok.Principal, trx *sqlx.Tx) (*store.Environment, error) {
	env, err := str.FindEnvironmentForAccount(envZId, int(principal.ID), trx)
	if err != nil {
		return nil, errors.Wrapf(err, "error finding environment '%v' for account '%v'", envZId, principal.Email)
	}
	return env, nil
}

func (h *unshareHandler) findAndValidateShare(shrToken string, env *store.Environment, trx *sqlx.Tx) (*store.Share, error) {
	shares, err := str.FindSharesForEnvironment(env.Id, trx)
	if err != nil {
		return nil, errors.Wrapf(err, "error finding shares for environment '%v'", env.ZId)
	}

	for _, share := range shares {
		if share.Token == shrToken {
			return share, nil
		}
	}

	return nil, errors.Errorf("share '%v' not found in environment '%v'", shrToken, env.ZId)
}

func (h *unshareHandler) deallocateResources(shrToken string) error {
	// get shared automation client
	za, err := automation.NewZitiAutomation(cfg.Ziti)
	if err != nil {
		return errors.Wrap(err, "error getting ziti automation client")
	}

	// use fluent workflow API for tag-based cleanup
	err = za.CleanupByTag("zrokShareToken", shrToken)
	if err != nil {
		return errors.Wrapf(err, "error cleaning up ziti resources for share '%v'", shrToken)
	}

	dl.Infof("deallocated ziti resources for share '%v'", shrToken)
	return nil
}

func (h *unshareHandler) cleanupShareNameMappings(shareId int, trx *sqlx.Tx) error {
	// find all share name mappings for this share
	mappings, err := str.FindShareNameMappingsByShareId(shareId, trx)
	if err != nil {
		return errors.Wrapf(err, "error finding share name mappings for share '%v'", shareId)
	}

	// delete each name that was dynamically allocated (not reserved)
	for _, mapping := range mappings {
		name, err := str.GetName(mapping.NameId, trx)
		if err != nil {
			dl.Warnf("error getting name '%v' for cleanup: %v", mapping.NameId, err)
			continue
		}

		// only delete names that are not reserved (dynamically allocated by share12)
		if !name.Reserved {
			if err := str.DeleteName(name.Id, trx); err != nil {
				dl.Warnf("error deleting name '%v': %v", name.Name, err)
			} else {
				dl.Debugf("deleted dynamically allocated name '%v'", name.Name)
			}
		}

		// delete the share name mapping
		if err := str.DeleteShareNameMapping(mapping.Id, trx); err != nil {
			dl.Warnf("error deleting share name mapping '%v': %v", mapping.Id, err)
		}
	}

	return nil
}

func (h *unshareHandler) processDynamicMappings(shareId int, trx *sqlx.Tx) error {
	// only send updates if dynamic proxy controller is enabled
	if dPCtrl == nil {
		return nil
	}

	// find all share name mappings for this share
	mappings, err := str.FindShareNameMappingsByShareId(shareId, trx)
	if err != nil {
		return errors.Wrapf(err, "error finding share name mappings for share '%v'", shareId)
	}

	for _, mapping := range mappings {
		// find name record to get the name and namespace
		name, err := str.GetName(mapping.NameId, trx)
		if err != nil {
			dl.Warnf("error finding name with id '%v' for unbind update: %v", mapping.NameId, err)
			continue
		}

		// find namespace
		ns, err := str.GetNamespace(name.NamespaceId, trx)
		if err != nil {
			dl.Warnf("error finding namespace with id '%v' for unbind update: %v", name.NamespaceId, err)
			continue
		}

		// find dynamic frontends for this namespace
		frontends, err := str.FindDynamicFrontendsForNamespace(ns.Id, trx)
		if err != nil {
			dl.Warnf("error finding dynamic frontends for namespace '%v': %v", ns.Token, err)
			continue
		}

		// send unbind mapping updates to each dynamic frontend
		for _, frontend := range frontends {
			frontendName := util.ExpandUrlTemplate(name.Name, ns.Name)

			if err := dPCtrl.UnbindFrontendMapping(frontend.Token, frontendName, trx); err != nil {
				dl.Errorf("error unbinding frontend mapping from frontend '%v': %v", frontend.Token, err)
				// continue with other frontends rather than failing completely
			} else {
				dl.Debugf("unbound frontend mapping '%v' from dynamic frontend '%v'", frontendName, frontend.Token)
			}
		}
	}

	return nil
}
