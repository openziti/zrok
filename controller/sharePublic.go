package controller

import (
	"github.com/openziti-test-kitchen/zrok/controller/zrokEdgeSdk"
	"github.com/openziti-test-kitchen/zrok/model"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/share"
	"github.com/openziti/edge/rest_management_api_client"
)

type publicResourceAllocator struct{}

func newPublicResourceAllocator() *publicResourceAllocator {
	return &publicResourceAllocator{}
}

func (a *publicResourceAllocator) allocate(envZId, shrToken string, frontendZIds, frontendTemplates []string, params share.ShareParams, edge *rest_management_api_client.ZitiEdgeManagement) (shrZId string, frontendEndpoints []string, err error) {
	var authUsers []*model.AuthUser
	for _, authUser := range params.Body.AuthUsers {
		authUsers = append(authUsers, &model.AuthUser{authUser.Username, authUser.Password})
	}
	cfgId, err := zrokEdgeSdk.CreateConfig(zrokProxyConfigId, envZId, shrToken, params.Body.AuthScheme, authUsers, edge)
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
