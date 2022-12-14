package controller

import (
	"github.com/openziti-test-kitchen/zrok/controller/zrokEdgeSdk"
	"github.com/openziti-test-kitchen/zrok/model"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/service"
	"github.com/openziti/edge/rest_management_api_client"
)

type privateResourceAllocator struct{}

func newPrivateResourceAllocator() *privateResourceAllocator {
	return &privateResourceAllocator{}
}

func (a *privateResourceAllocator) allocate(envZId, svcToken string, params service.ShareParams, edge *rest_management_api_client.ZitiEdgeManagement) (svcZId string, frontendEndpoints []string, err error) {
	var authUsers []*model.AuthUser
	for _, authUser := range params.Body.AuthUsers {
		authUsers = append(authUsers, &model.AuthUser{authUser.Username, authUser.Password})
	}
	cfgZId, err := zrokEdgeSdk.CreateConfig(zrokProxyConfigId, envZId, svcToken, params.Body.AuthScheme, authUsers, edge)
	if err != nil {
		return "", nil, err
	}

	svcZId, err = zrokEdgeSdk.CreateShareService(envZId, svcToken, cfgZId, edge)
	if err != nil {
		return "", nil, err
	}

	if err := zrokEdgeSdk.CreateServicePolicyBind(envZId, svcToken, svcZId, edge); err != nil {
		return "", nil, err
	}

	if err := zrokEdgeSdk.CreateShareServiceEdgeRouterPolicy(envZId, svcToken, svcZId, edge); err != nil {
		return "", nil, err
	}

	return svcZId, nil, nil
}
