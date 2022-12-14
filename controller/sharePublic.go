package controller

import (
	"github.com/openziti-test-kitchen/zrok/controller/zrokEdgeSdk"
	"github.com/openziti-test-kitchen/zrok/model"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/service"
	"github.com/openziti/edge/rest_management_api_client"
)

type publicResourceAllocator struct{}

func newPublicResourceAllocator() *publicResourceAllocator {
	return &publicResourceAllocator{}
}

func (a *publicResourceAllocator) allocate(envZId, svcToken string, frontendZIds, frontendTemplates []string, params service.ShareParams, edge *rest_management_api_client.ZitiEdgeManagement) (svcZId string, frontendEndpoints []string, err error) {
	var authUsers []*model.AuthUser
	for _, authUser := range params.Body.AuthUsers {
		authUsers = append(authUsers, &model.AuthUser{authUser.Username, authUser.Password})
	}
	cfgId, err := zrokEdgeSdk.CreateConfig(zrokProxyConfigId, envZId, svcToken, params.Body.AuthScheme, authUsers, edge)
	if err != nil {
		return "", nil, err
	}

	svcZId, err = zrokEdgeSdk.CreateShareService(envZId, svcToken, cfgId, edge)
	if err != nil {
		return "", nil, err
	}

	if err := zrokEdgeSdk.CreateServicePolicyBind(envZId+"-bind", svcZId, envZId, zrokEdgeSdk.ZrokServiceTags(svcToken).SubTags, edge); err != nil {
		return "", nil, err
	}

	if err := zrokEdgeSdk.CreateServicePolicyDial(envZId+"-dial", svcZId, frontendZIds, zrokEdgeSdk.ZrokServiceTags(svcToken).SubTags, edge); err != nil {
		return "", nil, err
	}

	if err := zrokEdgeSdk.CreateServiceEdgeRouterPolicyForShare(envZId, svcToken, svcZId, edge); err != nil {
		return "", nil, err
	}

	for _, frontendTemplate := range frontendTemplates {
		frontendEndpoints = append(frontendEndpoints, proxyUrl(svcToken, frontendTemplate))
	}

	return svcZId, frontendEndpoints, nil
}
