package controller

import (
	"github.com/openziti/edge-api/rest_management_api_client"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/rest_server_zrok/operations/share"
	"github.com/openziti/zrok/sdk"
)

type privateResourceAllocator struct{}

func newPrivateResourceAllocator() *privateResourceAllocator {
	return &privateResourceAllocator{}
}

func (a *privateResourceAllocator) allocate(envZId, shrToken string, params share.ShareParams, edge *rest_management_api_client.ZitiEdgeManagement) (shrZId string, frontendEndpoints []string, err error) {
	var authUsers []*sdk.AuthUserConfig
	for _, authUser := range params.Body.AuthUsers {
		authUsers = append(authUsers, &sdk.AuthUserConfig{Username: authUser.Username, Password: authUser.Password})
	}
	authScheme, err := sdk.ParseAuthScheme(params.Body.AuthScheme)
	if err != nil {
		return "", nil, err
	}
	options := &zrokEdgeSdk.FrontendOptions{
		AuthScheme:     authScheme,
		BasicAuthUsers: authUsers,
		Oauth: &sdk.OauthConfig{
			Provider:                   params.Body.OauthProvider,
			EmailDomains:               params.Body.OauthEmailDomains,
			AuthorizationCheckInterval: params.Body.OauthAuthorizationCheckInterval,
		},
	}
	cfgZId, err := zrokEdgeSdk.CreateConfig(zrokProxyConfigId, envZId, shrToken, options, edge)
	if err != nil {
		return "", nil, err
	}

	shrZId, err = zrokEdgeSdk.CreateShareService(envZId, shrToken, cfgZId, edge)
	if err != nil {
		return "", nil, err
	}

	if err := zrokEdgeSdk.CreateServicePolicyBind(envZId+"-"+shrZId+"-bind", shrZId, envZId, zrokEdgeSdk.ZrokShareTags(shrToken).SubTags, edge); err != nil {
		return "", nil, err
	}

	if err := zrokEdgeSdk.CreateShareServiceEdgeRouterPolicy(envZId, shrToken, shrZId, edge); err != nil {
		return "", nil, err
	}

	return shrZId, nil, nil
}
