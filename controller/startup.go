package controller

import (
	"context"
	"fmt"
	"github.com/openziti-test-kitchen/zrok/model"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/edge/rest_management_api_client/config"
	"github.com/openziti/edge/rest_model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

var zrokProxyConfigId string

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
	if err := ensureZrokProxyConfigType(edge); err != nil {
		return err
	}

	return nil
}

func ensureZrokProxyConfigType(edge *rest_management_api_client.ZitiEdgeManagement) error {
	filter := fmt.Sprintf("name=\"%v\"", model.ZrokProxyConfig)
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
		name := model.ZrokProxyConfig
		ct := &rest_model.ConfigTypeCreate{Name: &name}
		createReq := &config.CreateConfigTypeParams{ConfigType: ct}
		createReq.SetTimeout(30 * time.Second)
		createResp, err := edge.Config.CreateConfigType(createReq, nil)
		if err != nil {
			return err
		}
		logrus.Infof("created '%v' config type with id '%v'", model.ZrokProxyConfig, createResp.Payload.Data.ID)
		zrokProxyConfigId = createResp.Payload.Data.ID
	} else if len(listResp.Payload.Data) > 1 {
		return errors.Errorf("found %d '%v' config types; expected 0 or 1", len(listResp.Payload.Data), model.ZrokProxyConfig)
	} else {
		logrus.Infof("found '%v' config type with id '%v'", model.ZrokProxyConfig, *(listResp.Payload.Data[0].ID))
		zrokProxyConfigId = *(listResp.Payload.Data[0].ID)
	}
	return nil
}
