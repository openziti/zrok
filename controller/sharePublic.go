package controller

import (
	"github.com/openziti/edge-api/rest_management_api_client"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/rest_server_zrok/operations/share"
	"github.com/openziti/zrok/sdk"
)

type publicResourceAllocator struct{}

func newPublicResourceAllocator() *publicResourceAllocator {
	return &publicResourceAllocator{}
}

func (a *publicResourceAllocator) allocate(envZId, shrToken string, frontendZIds, frontendTemplates []string, params share.ShareParams, edge *rest_management_api_client.ZitiEdgeManagement) (shrZId string, frontendEndpoints []string, err error) {
	var authUsers []*sdk.AuthUser
	for _, authUser := range params.Body.AuthUsers {
		authUsers = append(authUsers, &sdk.AuthUser{authUser.Username, authUser.Password})
	}
	cfgId, err := zrokEdgeSdk.CreateConfig(zrokProxyConfigId, envZId, shrToken, params.Body.AuthScheme, authUsers, &zrokEdgeSdk.OauthOptions{
		Provider:                   params.Body.OauthProvider,
		EmailDomains:               params.Body.OauthEmailDomains,
		AuthorizationCheckInterval: params.Body.OauthAuthorizationCheckInterval,
	}, edge)
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
		frontendEndpoints = append(frontendEndpoints, proxyUrl(shrToken, frontendTemplate))
	}

	return shrZId, frontendEndpoints, nil
}
