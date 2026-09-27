package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/edge-api/rest_management_api_client"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/share"
	"github.com/sirupsen/logrus"
)

type unshareHandler struct {
	edge func() (*rest_management_api_client.ZitiEdgeManagement, error)
}

func newUnshareHandler() *unshareHandler {
	return &unshareHandler{edge: func() (*rest_management_api_client.ZitiEdgeManagement, error) {
		return zrokEdgeSdk.Client(cfg.Ziti)
	}}
}

func (h *unshareHandler) Handle(params share.UnshareParams, principal *rest_model_zrok.Principal) middleware.Responder {
	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction for '%v': %v", principal.Email, err)
		return share.NewUnshareInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()

	// the store is the authority for the share; a missing share never reaches ziti.
	shrToken := params.Body.ShareToken
	var senv *store.Environment
	if envs, err := str.FindEnvironmentsForAccount(int(principal.ID), tx); err == nil {
		for _, env := range envs {
			if env.ZId == params.Body.EnvZID {
				senv = env
				break
			}
		}
		if senv == nil {
			logrus.Errorf("environment with id '%v' not found for '%v", params.Body.EnvZID, principal.Email)
			return share.NewUnshareNotFound()
		}
	} else {
		logrus.Errorf("error finding environments for account '%v': %v", principal.Email, err)
		return share.NewUnshareNotFound()
	}

	var sshr *store.Share
	if shrs, err := str.FindSharesForEnvironment(senv.Id, tx); err == nil {
		for _, shr := range shrs {
			if shr.Token == shrToken {
				sshr = shr
				break
			}
		}
		if sshr == nil {
			logrus.Errorf("share '%v' not found for '%v'", shrToken, principal.Email)
			return share.NewUnshareNotFound()
		}
	} else {
		logrus.Errorf("error finding shares for account '%v': %v", principal.Email, err)
		return share.NewUnshareInternalServerError()
	}

	if sshr.Reserved == params.Body.Reserved {
		edge, err := h.edge()
		if err != nil {
			logrus.Errorf("error getting edge client for '%v': %v", principal.Email, err)
			return share.NewUnshareInternalServerError()
		}

		// single tag-based share deallocator; should work regardless of sharing mode
		h.deallocateResources(senv, shrToken, sshr.ZId, edge)
		logrus.Debugf("deallocated share '%v'", shrToken)

		if err := str.DeleteAccessGrantsForShare(sshr.Id, tx); err != nil {
			logrus.Errorf("error deleting access grants for share '%v': %v", shrToken, err)
			return share.NewUnshareInternalServerError()
		}
		if err := str.DeleteShare(sshr.Id, tx); err != nil {
			logrus.Errorf("error deleting share '%v': %v", shrToken, err)
			return share.NewUnshareInternalServerError()
		}
		if err := tx.Commit(); err != nil {
			logrus.Errorf("error committing transaction for '%v': %v", sshr.ZId, err)
			return share.NewUnshareInternalServerError()
		}

	} else {
		logrus.Infof("share '%v' is reserved, skipping deallocation", shrToken)
	}

	return share.NewUnshareOK()
}

func (h *unshareHandler) deallocateResources(senv *store.Environment, shrToken, shrZId string, edge *rest_management_api_client.ZitiEdgeManagement) {
	if err := zrokEdgeSdk.DeleteServiceEdgeRouterPolicyForShare(senv.ZId, shrToken, edge); err != nil {
		logrus.Warnf("error deleting service edge router policies for share '%v' in environment '%v': %v", shrToken, senv.ZId, err)
	}
	if err := zrokEdgeSdk.DeleteServicePoliciesDialForShare(senv.ZId, shrToken, edge); err != nil {
		logrus.Warnf("error deleting dial service policies for share '%v' in environment '%v': %v", shrToken, senv.ZId, err)
	}
	if err := zrokEdgeSdk.DeleteServicePoliciesBindForShare(senv.ZId, shrToken, edge); err != nil {
		logrus.Warnf("error deleting bind service policies for share '%v' in environment '%v': %v", shrToken, senv.ZId, err)
	}
	if err := zrokEdgeSdk.DeleteConfig(senv.ZId, shrToken, edge); err != nil {
		logrus.Warnf("error deleting config for share '%v' in environment '%v': %v", shrToken, senv.ZId, err)
	}
	if err := zrokEdgeSdk.DeleteService(senv.ZId, shrZId, edge); err != nil {
		logrus.Warnf("error deleting service '%v' for share '%v' in environment '%v': %v", shrZId, shrToken, senv.ZId, err)
	}
}
