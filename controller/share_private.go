package controller

import (
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
	cfgId, err := createConfig(envZId, svcToken, params.Body.AuthScheme, authUsers, edge)
	if err != nil {
		return "", nil, err
	}

	svcZId, err = createShareService(envZId, svcToken, cfgId, edge)
	if err != nil {
		return "", nil, err
	}

	if err := createServicePolicyBind(envZId, svcToken, svcZId, edge); err != nil {
		return "", nil, err
	}

	if err := createShareServiceEdgeRouterPolicy(envZId, svcToken, svcZId, edge); err != nil {
		return "", nil, err
	}

	return svcZId, []string{proxyUrl(svcToken)}, nil
}
