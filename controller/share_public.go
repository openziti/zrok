package controller

import (
	"github.com/openziti-test-kitchen/zrok/controller/zrok_edge_sdk"
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
	cfgId, err := zrok_edge_sdk.CreateConfig(zrokProxyConfigId, envZId, svcToken, params.Body.AuthScheme, authUsers, edge)
	if err != nil {
		return "", nil, err
	}

	svcZId, err = zrok_edge_sdk.CreateShareService(envZId, svcToken, cfgId, edge)
	if err != nil {
		return "", nil, err
	}

	if err := zrok_edge_sdk.CreateServicePolicyBind(envZId, svcToken, svcZId, edge); err != nil {
		return "", nil, err
	}

	if err := zrok_edge_sdk.CreateServicePolicyDial(envZId, svcToken, svcZId, frontendZIds, edge); err != nil {
		return "", nil, err
	}

	if err := zrok_edge_sdk.CreateShareServiceEdgeRouterPolicy(envZId, svcToken, svcZId, edge); err != nil {
		return "", nil, err
	}

	for _, frontendTemplate := range frontendTemplates {
		frontendEndpoints = append(frontendEndpoints, proxyUrl(svcToken, frontendTemplate))
	}

	return svcZId, frontendEndpoints, nil
}
