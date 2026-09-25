package zitifake

import (
	"context"
	"errors"
	"testing"

	"github.com/openziti/edge-api/rest_management_api_client/service"
	"github.com/openziti/edge-api/rest_management_api_client/service_policy"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/stretchr/testify/require"
)

func TestFilterAndUniqueNames(t *testing.T) {
	fake := New()
	defer fake.Close()
	edge := fake.Edge()
	require.NoError(t, zrokEdgeSdk.CreateServicePolicyDial("one", "service", []string{"identity"}, zrokEdgeSdk.ZrokShareTags("share-one").SubTags, edge))
	policy, err := zrokEdgeSdk.FindServicePolicyByName("one", edge)
	require.NoError(t, err)
	require.NotNil(t, policy)
	id := *policy.ID
	for _, filter := range []string{`name="one"`, `id="` + id + `"`, `tags.zrokShareToken="share-one" and type=1`, `tags.zrok != null`} {
		resp, err := edge.ServicePolicy.ListServicePolicies(&service_policy.ListServicePoliciesParams{Filter: &filter, Context: context.Background()}, nil)
		require.NoError(t, err, filter)
		require.Len(t, resp.Payload.Data, 1, filter)
	}
	filter := `tags.zrokShareToken="share-two" and type=2`
	resp, err := edge.ServicePolicy.ListServicePolicies(&service_policy.ListServicePoliciesParams{Filter: &filter, Context: context.Background()}, nil)
	require.NoError(t, err)
	require.Empty(t, resp.Payload.Data)
	require.NoError(t, zrokEdgeSdk.CreateServicePolicyBind("bind-one", "service", "identity", nil, edge))
	bind := rest_model.DialBindBind
	require.True(t, match(`type=2`, "id", "bind", nil, &bind))
	require.False(t, match(`type=1`, "id", "bind", nil, &bind))
	err = zrokEdgeSdk.CreateServicePolicyDial("one", "service", []string{"identity"}, nil, edge)
	var conflict *service_policy.CreateServicePolicyBadRequest
	require.True(t, errors.As(err, &conflict), "%v", err)
	require.Contains(t, conflict.Payload.Error.Message, "conflict")
	created, _, _, _ := fake.Counts()
	require.Equal(t, 2, created)
	detail, err := edge.ServicePolicy.DetailServicePolicy(&service_policy.DetailServicePolicyParams{ID: id}, nil)
	require.NoError(t, err)
	require.Equal(t, "one", *detail.Payload.Data.Name)
	_, err = edge.ServicePolicy.DeleteServicePolicy(&service_policy.DeleteServicePolicyParams{ID: id}, nil)
	require.NoError(t, err)
	_, err = edge.ServicePolicy.DetailServicePolicy(&service_policy.DetailServicePolicyParams{ID: id}, nil)
	var notFound *service_policy.DetailServicePolicyNotFound
	require.True(t, errors.As(err, &notFound), "%v", err)
}

func TestServiceRoutes(t *testing.T) {
	fake := New()
	defer fake.Close()
	edge := fake.Edge()
	id, err := zrokEdgeSdk.CreateService("service-one", nil, map[string]interface{}{"zrokShareToken": "share-one"}, edge)
	require.NoError(t, err)
	filter := `tags.zrokShareToken="share-one"`
	items, err := edge.Service.ListServices(&service.ListServicesParams{Filter: &filter, Context: context.Background()}, nil)
	require.NoError(t, err)
	require.Len(t, items.Payload.Data, 1)
	require.Equal(t, id, *items.Payload.Data[0].ID)
	_, err = zrokEdgeSdk.CreateService("service-one", nil, nil, edge)
	require.Error(t, err)
	detail, err := edge.Service.DetailService(&service.DetailServiceParams{ID: id}, nil)
	require.NoError(t, err)
	require.Equal(t, "service-one", *detail.Payload.Data.Name)
	require.NoError(t, zrokEdgeSdk.DeleteService("env", id, edge))
	_, _, creates, deletes := fake.Counts()
	require.Equal(t, 1, creates)
	require.Equal(t, 1, deletes)
}
