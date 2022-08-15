package controller

import (
	"context"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/edge/rest_management_api_client/config"
	"github.com/openziti/edge/rest_model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

var zrokAuthV1Id string

func controllerStartup(cfg *Config) error {
	if err := inspectZiti(cfg); err != nil {
		return err
	}
	return nil
}

func inspectZiti(cfg *Config) error {
	logrus.Infof("inspecting ziti controller configuration")

	edge, err := edgeClient(cfg.Ziti)
	if err != nil {
		return errors.Wrap(err, "error getting ziti edge client")
	}
	if err := ensureZrokAuthConfigType(edge); err != nil {
		return err
	}

	return nil
}

func ensureZrokAuthConfigType(edge *rest_management_api_client.ZitiEdgeManagement) error {
	filter := "name=\"zrok.auth.v1\""
	limit := int64(100)
	offset := int64(0)
	listReq := &config.ListConfigTypesParams{
		Filter:  &filter,
		Limit:   &limit,
		Offset:  &offset,
		Context: context.Background(),
	}
	listReq.SetTimeout(30 * time.Second)
	listResp, err := edge.Config.ListConfigTypes(listReq, nil)
	if err != nil {
		return err
	}
	if len(listResp.Payload.Data) < 1 {
		name := "zrok.auth.v1"
		ct := &rest_model.ConfigTypeCreate{Name: &name}
		createReq := &config.CreateConfigTypeParams{ConfigType: ct}
		createReq.SetTimeout(30 * time.Second)
		createResp, err := edge.Config.CreateConfigType(createReq, nil)
		if err != nil {
			return err
		}
		logrus.Infof("created 'zrok.auth.v1' config type with id '%v'", createResp.Payload.Data.ID)
		zrokAuthV1Id = createResp.Payload.Data.ID
	} else if len(listResp.Payload.Data) > 1 {
		return errors.Errorf("found %d 'zrok.auth.v1' config types; expected 0 or 1", len(listResp.Payload.Data))
	} else {
		logrus.Infof("found 'zrok.auth.v1' config type with id '%v'", *(listResp.Payload.Data[0].ID))
		zrokAuthV1Id = *(listResp.Payload.Data[0].ID)
	}
	return nil
}
