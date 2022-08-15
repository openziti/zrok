package controller

import (
	"context"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/edge/rest_management_api_client/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

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
	logrus.Infof("found %d zrok.auth.v1 config types", len(listResp.Payload.Data))
	return nil
}
