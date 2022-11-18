package controller

import (
	"context"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/controller/store"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/service"
	"github.com/openziti/edge/rest_management_api_client"
	edge_service "github.com/openziti/edge/rest_management_api_client/service"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

type unshareHandler struct {
}

func newUnshareHandler() *unshareHandler {
	return &unshareHandler{}
}

func (h *unshareHandler) Handle(params service.UnshareParams, principal *rest_model_zrok.Principal) middleware.Responder {
	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return service.NewUnshareInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()

	edge, err := edgeClient()
	if err != nil {
		logrus.Error(err)
		return service.NewUnshareInternalServerError()
	}
	svcName := params.Body.SvcName
	svcZId, err := h.findServiceZId(svcName, edge)
	if err != nil {
		logrus.Error(err)
		return service.NewUnshareInternalServerError()
	}
	var senv *store.Environment
	if envs, err := str.FindEnvironmentsForAccount(int(principal.ID), tx); err == nil {
		for _, env := range envs {
			if env.ZId == params.Body.ZID {
				senv = env
				break
			}
		}
		if senv == nil {
			err := errors.Errorf("environment with id '%v' not found for '%v", params.Body.ZID, principal.Email)
			logrus.Error(err)
			return service.NewUnshareNotFound()
		}
	} else {
		logrus.Errorf("error finding environments for account '%v': %v", principal.Email, err)
		return service.NewUnshareInternalServerError()
	}

	var ssvc *store.Service
	if svcs, err := str.FindServicesForEnvironment(senv.Id, tx); err == nil {
		for _, svc := range svcs {
			if svc.ZId == svcZId {
				ssvc = svc
				break
			}
		}
		if ssvc == nil {
			err := errors.Errorf("service with id '%v' not found for '%v'", svcZId, principal.Email)
			logrus.Error(err)
			return service.NewUnshareNotFound()
		}
	} else {
		logrus.Errorf("error finding services for account '%v': %v", principal.Email, err)
		return service.NewUnshareInternalServerError()
	}

	if err := deleteServiceEdgeRouterPolicy(senv.ZId, svcName, edge); err != nil {
		logrus.Error(err)
		return service.NewUnshareInternalServerError()
	}
	if err := deleteServicePolicyDial(senv.ZId, svcName, edge); err != nil {
		logrus.Error(err)
		return service.NewUnshareInternalServerError()
	}
	if err := deleteServicePolicyBind(senv.ZId, svcName, edge); err != nil {
		logrus.Error(err)
		return service.NewUnshareInternalServerError()
	}
	if err := deleteConfig(senv.ZId, svcName, edge); err != nil {
		logrus.Error(err)
		return service.NewUnshareInternalServerError()
	}
	if err := deleteService(senv.ZId, svcZId, edge); err != nil {
		logrus.Error(err)
		return service.NewUnshareInternalServerError()
	}

	logrus.Debugf("deallocated service '%v'", svcName)

	if err := str.DeleteService(ssvc.Id, tx); err != nil {
		logrus.Errorf("error deactivating service '%v': %v", svcZId, err)
		return service.NewUnshareInternalServerError()
	}
	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing: %v", err)
		return service.NewUnshareInternalServerError()
	}

	return service.NewUnshareOK()
}

func (h *unshareHandler) findServiceZId(svcName string, edge *rest_management_api_client.ZitiEdgeManagement) (string, error) {
	filter := fmt.Sprintf("name=\"%v\"", svcName)
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
	return "", errors.Errorf("service '%v' not found", svcName)
}
