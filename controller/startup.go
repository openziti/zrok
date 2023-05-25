package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/openziti/edge-api/rest_management_api_client"
	"github.com/openziti/edge-api/rest_management_api_client/config"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var zrokProxyConfigId string

func controllerStartup() error {
	if err := inspectZiti(); err != nil {
		return err
	}
	return nil
}

func inspectZiti() error {
	logrus.Infof("inspecting ziti controller configuration")

	edge, err := zrokEdgeSdk.Client(cfg.Ziti)
	if err != nil {
		return errors.Wrap(err, "error getting ziti edge client")
	}
	if err := findZrokProxyConfigType(edge); err != nil {
		return err
	}

	return nil
}

func findZrokProxyConfigType(edge *rest_management_api_client.ZitiEdgeManagement) error {
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
	if len(listResp.Payload.Data) != 1 {
		return errors.Errorf("expected 1 zrok proxy config type, found %d", len(listResp.Payload.Data))
	}
	logrus.Infof("found '%v' config type with id '%v'", model.ZrokProxyConfig, *(listResp.Payload.Data[0].ID))
	zrokProxyConfigId = *(listResp.Payload.Data[0].ID)

	return nil
}
