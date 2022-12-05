package controller

import (
	"context"
	"fmt"
	"github.com/openziti-test-kitchen/zrok/model"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/edge/rest_management_api_client/config"
	"github.com/openziti/edge/rest_model"
	"github.com/openziti/sdk-golang/ziti"
	config2 "github.com/openziti/sdk-golang/ziti/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

func Bootstrap(inCfg *Config) error {
	cfg = inCfg

	if ctrlZId, err := getIdentityId("ctrl"); err == nil {
		logrus.Infof("controller identity: %v", ctrlZId)
	} else {
		panic(err)
	}

	if frontendZId, err := getIdentityId("frontend"); err == nil {
		logrus.Infof("frontend identity: %v", frontendZId)
	} else {
		panic(err)
	}

	edge, err := edgeClient()
	if err != nil {
		return err
	}

	if err := assertZrokProxyConfigType(edge); err != nil {
		return err
	}

	return nil
}

func assertZrokProxyConfigType(edge *rest_management_api_client.ZitiEdgeManagement) error {
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
	} else if len(listResp.Payload.Data) > 1 {
		return errors.Errorf("found %d '%v' config types; expected 0 or 1", len(listResp.Payload.Data), model.ZrokProxyConfig)
	} else {
		logrus.Infof("found '%v' config type with id '%v'", model.ZrokProxyConfig, *(listResp.Payload.Data[0].ID))
	}
	return nil
}

func getIdentityId(identityName string) (string, error) {
	zif, err := zrokdir.ZitiIdentityFile(identityName)
	if err != nil {
		return "", errors.Wrapf(err, "error opening identity '%v' from zrokdir", identityName)
	}
	zcfg, err := config2.NewFromFile(zif)
	if err != nil {
		return "", errors.Wrapf(err, "error loading ziti config from file '%v'", zif)
	}
	zctx := ziti.NewContextWithConfig(zcfg)
	id, err := zctx.GetCurrentIdentity()
	if err != nil {
		return "", errors.Wrapf(err, "error getting current identity from '%v'", zif)
	}
	return id.Id, nil
}
