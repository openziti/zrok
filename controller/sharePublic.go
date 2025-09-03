package controller

import (
	"github.com/openziti/edge-api/rest_management_api_client"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/rest_server_zrok/operations/share"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/openziti/zrok/util"
)

type publicResourceAllocator struct{}

func newPublicResourceAllocator() *publicResourceAllocator {
	return &publicResourceAllocator{}
}

func (a *publicResourceAllocator) allocate(envZId, shrToken string, frontendZIds, frontendTemplates []string, params share.ShareParams, interstitial bool, edge *rest_management_api_client.ZitiEdgeManagement) (shrZId string, frontendEndpoints []string, err error) {
	var authUsers []*sdk.AuthUserConfig
	for _, authUser := range params.Body.AuthUsers {
		authUsers = append(authUsers, &sdk.AuthUserConfig{Username: authUser.Username, Password: authUser.Password})
	}
	authScheme, err := sdk.ParseAuthScheme(params.Body.AuthScheme)
	if err != nil {
		return "", nil, err
	}
	options := &zrokEdgeSdk.FrontendOptions{
		Interstitial:   interstitial,
		AuthScheme:     authScheme,
		BasicAuthUsers: authUsers,
		Oauth: &sdk.OauthConfig{
			Provider:                   params.Body.OauthProvider,
			EmailDomains:               params.Body.OauthEmailDomains,
			AuthorizationCheckInterval: params.Body.OauthAuthorizationCheckInterval,
		},
	}
	cfgId, err := zrokEdgeSdk.CreateConfig(zrokProxyConfigId, envZId, shrToken, options, edge)
	if err != nil {
		return "", nil, err
	}

	shrZId, err = zrokEdgeSdk.CreateShareService(envZId, shrToken, cfgId, edge)
	if err != nil {
		return "", nil, err
	}

	if err := zrokEdgeSdk.CreateServicePolicyBind(envZId+"-"+shrZId+"-bind", shrZId, envZId, zrokEdgeSdk.ZrokShareTags(shrToken).SubTags, edge); err != nil {
		return "", nil, err
	}

	if err := zrokEdgeSdk.CreateServicePolicyDial(envZId+"-"+shrZId+"-dial", shrZId, frontendZIds, zrokEdgeSdk.ZrokShareTags(shrToken).SubTags, edge); err != nil {
		return "", nil, err
	}

	if err := zrokEdgeSdk.CreateShareServiceEdgeRouterPolicy(envZId, shrToken, shrZId, edge); err != nil {
		return "", nil, err
	}

	for _, frontendTemplate := range frontendTemplates {
		frontendEndpoints = append(frontendEndpoints, util.ExpandUrlTemplate(shrToken, frontendTemplate))
	}

	return shrZId, frontendEndpoints, nil
}
