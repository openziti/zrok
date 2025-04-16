package controller

import (
	"context"
	"fmt"
	"github.com/openziti/edge-api/rest_management_api_client"
	apiConfig "github.com/openziti/edge-api/rest_management_api_client/config"
	"github.com/openziti/zrok/controller/config"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/sirupsen/logrus"
)

func Unbootstrap(cfg *config.Config) error {
	edge, err := zrokEdgeSdk.Client(cfg.Ziti)
	if err != nil {
		return err
	}
	if err := unbootstrapConfigs(edge); err != nil {
		logrus.Errorf("error unbootrapping configs: %v", err)
	}
	return nil
}

func unbootstrapConfigs(edge *rest_management_api_client.ZitiEdgeManagement) error {
	filter := fmt.Sprintf("tags.zrok != null")
	limit := int64(100)
	offset := int64(0)
	listReq := &apiConfig.ListConfigsParams{
		Filter:  &filter,
		Limit:   &limit,
		Offset:  &offset,
		Context: context.Background(),
	}
	listResp, err := edge.Config.ListConfigs(listReq, nil)
	if err != nil {
		return err
	}
	for _, listCfg := range listResp.Payload.Data {
		logrus.Infof("found config: %v", *listCfg.ID)
	}
	return nil
}
