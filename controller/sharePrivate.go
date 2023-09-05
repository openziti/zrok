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
	var authUsers []*sdk.AuthUser
	for _, authUser := range params.Body.AuthUsers {
		authUsers = append(authUsers, &sdk.AuthUser{authUser.Username, authUser.Password})
	}
	cfgZId, err := zrokEdgeSdk.CreateConfig(zrokProxyConfigId, envZId, shrToken, params.Body.AuthScheme, authUsers, params.Body.OauthProvider, params.Body.OauthEmailDomains, edge)
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
