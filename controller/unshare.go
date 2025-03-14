package controller

import (
	"context"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/edge-api/rest_management_api_client"
	edge_service "github.com/openziti/edge-api/rest_management_api_client/service"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/share"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

type unshareHandler struct{}

func newUnshareHandler() *unshareHandler {
	return &unshareHandler{}
}

func (h *unshareHandler) Handle(params share.UnshareParams, principal *rest_model_zrok.Principal) middleware.Responder {
	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction for '%v': %v", principal.Email, err)
		return share.NewUnshareInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()

	edge, err := zrokEdgeSdk.Client(cfg.Ziti)
	if err != nil {
		logrus.Errorf("error getting edge client for '%v': %v", principal.Email, err)
		return share.NewUnshareInternalServerError()
	}
	shrToken := params.Body.ShareToken
	shrZId, err := h.findShareZId(shrToken, edge)
	if err != nil {
		logrus.Errorf("error finding share identity for '%v' (%v): %v", shrToken, principal.Email, err)
		return share.NewUnshareNotFound()
	}
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
			if shr.ZId == shrZId {
				sshr = shr
				break
			}
		}
		if sshr == nil {
			logrus.Errorf("share with id '%v' not found for '%v'", shrZId, principal.Email)
			return share.NewUnshareNotFound()
		}
	} else {
		logrus.Errorf("error finding shares for account '%v': %v", principal.Email, err)
		return share.NewUnshareInternalServerError()
	}

	if sshr.Reserved == params.Body.Reserved {
		// single tag-based share deallocator; should work regardless of sharing mode
		h.deallocateResources(senv, shrToken, shrZId, edge)
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
			logrus.Errorf("error committing transaction for '%v': %v", shrZId, err)
			return share.NewUnshareInternalServerError()
		}

	} else {
		logrus.Infof("share '%v' is reserved, skipping deallocation", shrToken)
	}

	return share.NewUnshareOK()
}

func (h *unshareHandler) findShareZId(shrToken string, edge *rest_management_api_client.ZitiEdgeManagement) (string, error) {
	filter := fmt.Sprintf("name=\"%v\"", shrToken)
	limit := int64(1)
	offset := int64(0)
	listReq := &edge_service.ListServicesParams{
		Filter:  &filter,
		Limit:   &limit,
		Offset:  &offset,
		Context: context.Background(),
	}
	listReq.SetTimeout(30 * time.Second)
	listResp, err := edge.Service.ListServices(listReq, nil)
	if err != nil {
		return "", err
	}
	if len(listResp.Payload.Data) == 1 {
		return *(listResp.Payload.Data[0].ID), nil
	}
	return "", errors.Errorf("share '%v' not found", shrToken)
}

func (h *unshareHandler) deallocateResources(senv *store.Environment, shrToken, shrZId string, edge *rest_management_api_client.ZitiEdgeManagement) {
	if err := zrokEdgeSdk.DeleteServiceEdgeRouterPolicy(senv.ZId, shrToken, edge); err != nil {
		logrus.Warnf("error deleting service edge router policies for share '%v' in environment '%v': %v", shrToken, senv.ZId, err)
	}
	if err := zrokEdgeSdk.DeleteServicePoliciesDial(senv.ZId, shrToken, edge); err != nil {
		logrus.Warnf("error deleting dial service policies for share '%v' in environment '%v': %v", shrToken, senv.ZId, err)
	}
	if err := zrokEdgeSdk.DeleteServicePoliciesBind(senv.ZId, shrToken, edge); err != nil {
		logrus.Warnf("error deleting bind service policies for share '%v' in environment '%v': %v", shrToken, senv.ZId, err)
	}
	if err := zrokEdgeSdk.DeleteConfig(senv.ZId, shrToken, edge); err != nil {
		logrus.Warnf("error deleting config for share '%v' in environment '%v': %v", shrToken, senv.ZId, err)
	}
	if err := zrokEdgeSdk.DeleteService(senv.ZId, shrZId, edge); err != nil {
		logrus.Warnf("error deleting service '%v' for share '%v' in environment '%v': %v", shrZId, shrToken, senv.ZId, err)
	}
}
