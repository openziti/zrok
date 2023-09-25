package zrokEdgeSdk

import (
	"context"
	"fmt"
	"github.com/openziti/edge-api/rest_management_api_client"
	"github.com/openziti/edge-api/rest_management_api_client/config"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/zrok/sdk"
	"github.com/sirupsen/logrus"
	"time"
)

type FrontendOptions struct {
	AuthScheme string
	AuthUsers  []*sdk.AuthUserConfig
	OAuth      *sdk.OAuthConfig
}

func CreateConfig(cfgTypeZId, envZId, shrToken string, options *FrontendOptions, edge *rest_management_api_client.ZitiEdgeManagement) (cfgZId string, err error) {
	authScheme, err := sdk.ParseAuthScheme(options.AuthScheme)
	if err != nil {
		return "", err
	}
	cfg := &sdk.FrontendConfig{
		AuthScheme: authScheme,
	}
	if cfg.AuthScheme == sdk.Basic {
		cfg.BasicAuth = &sdk.BasicAuthConfig{}
		for _, authUser := range options.AuthUsers {
			cfg.BasicAuth.Users = append(cfg.BasicAuth.Users, &sdk.AuthUserConfig{Username: authUser.Username, Password: authUser.Password})
		}
	}
	if cfg.AuthScheme == sdk.Oauth && options.OAuth != nil {
		cfg.OAuthAuth = &sdk.OAuthConfig{
			Provider:                   options.OAuth.Provider,
			EmailDomains:               options.OAuth.EmailDomains,
			AuthorizationCheckInterval: options.OAuth.AuthorizationCheckInterval,
		}
	}
	cfgCrt := &rest_model.ConfigCreate{
		ConfigTypeID: &cfgTypeZId,
		Data:         cfg,
		Name:         &shrToken,
		Tags:         ZrokShareTags(shrToken),
	}
	cfgReq := &config.CreateConfigParams{
		Config:  cfgCrt,
		Context: context.Background(),
	}
	cfgReq.SetTimeout(30 * time.Second)
	cfgResp, err := edge.Config.CreateConfig(cfgReq, nil)
	if err != nil {
		return "", err
	}
	logrus.Infof("created config '%v' for environment '%v'", cfgResp.Payload.Data.ID, envZId)
	return cfgResp.Payload.Data.ID, nil
}

func DeleteConfig(envZId, shrToken string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	filter := fmt.Sprintf("tags.zrokShareToken=\"%v\"", shrToken)
	limit := int64(0)
	offset := int64(0)
	listReq := &config.ListConfigsParams{
		Filter:  &filter,
		Limit:   &limit,
		Offset:  &offset,
		Context: context.Background(),
	}
	listReq.SetTimeout(30 * time.Second)
	listResp, err := edge.Config.ListConfigs(listReq, nil)
	if err != nil {
		return err
	}
	for _, cfg := range listResp.Payload.Data {
		deleteReq := &config.DeleteConfigParams{
			ID:      *cfg.ID,
			Context: context.Background(),
		}
		deleteReq.SetTimeout(30 * time.Second)
		_, err := edge.Config.DeleteConfig(deleteReq, nil)
		if err != nil {
			return err
		}
		logrus.Infof("deleted config '%v' for '%v'", *cfg.ID, envZId)
	}
	return nil
}
