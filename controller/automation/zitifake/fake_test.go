package zitifake

import (
	"errors"
	"testing"

	"github.com/openziti/edge-api/rest_management_api_client/service_policy"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/zrok/v2/controller/automation"
	"github.com/stretchr/testify/require"
)

func TestFilterAndUniqueNames(t *testing.T) {
	fake := New()
	defer fake.Close()
	ziti := automation.NewZitiAutomationWithEdge(fake.Edge())
	opts := &automation.ServicePolicyOptions{
		BaseOptions:   automation.BaseOptions{Name: "one", Tags: automation.ZrokShareTags("share-one")},
		IdentityRoles: []string{"@identity"}, ServiceRoles: []string{"@service"}, Semantic: rest_model.SemanticAllOf,
	}
	id, err := ziti.ServicePolicies.CreateDial(opts)
	require.NoError(t, err)
	for _, filter := range []string{`name="one"`, `id="` + id + `"`, `tags.zrokShareToken="share-one" and type=1`, `tags.zrok != null`} {
		items, err := ziti.ServicePolicies.Find(&automation.FilterOptions{Filter: filter})
		require.NoError(t, err, filter)
		require.Len(t, items, 1, filter)
	}
	items, err := ziti.ServicePolicies.Find(&automation.FilterOptions{Filter: `tags.zrokShareToken="share-two" and type=2`})
	require.NoError(t, err)
	require.Empty(t, items)
	_, err = ziti.ServicePolicies.CreateDial(opts)
	var conflict *service_policy.CreateServicePolicyBadRequest
	require.True(t, errors.As(err, &conflict), "%v", err)
	require.Contains(t, conflict.Payload.Error.Message, "conflict")
	created, _, _, _ := fake.Counts()
	require.Equal(t, 1, created)
	detail, err := ziti.Edge().ServicePolicy.DetailServicePolicy(&service_policy.DetailServicePolicyParams{ID: id}, nil)
	require.NoError(t, err)
	require.Equal(t, "one", *detail.Payload.Data.Name)
	require.NoError(t, ziti.ServicePolicies.Delete(id))
	_, err = ziti.Edge().ServicePolicy.DetailServicePolicy(&service_policy.DetailServicePolicyParams{ID: id}, nil)
	var notFound *service_policy.DetailServicePolicyNotFound
	require.True(t, errors.As(err, &notFound), "%v", err)
}

func TestServiceRoutes(t *testing.T) {
	fake := New()
	defer fake.Close()
	ziti := automation.NewZitiAutomationWithEdge(fake.Edge())
	id, err := ziti.Services.Create(&automation.ServiceOptions{BaseOptions: automation.BaseOptions{Name: "service-one", Tags: automation.ZrokShareTags("share-one")}})
	require.NoError(t, err)
	items, err := ziti.Services.Find(&automation.FilterOptions{Filter: `tags.zrokShareToken="share-one"`})
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, id, *items[0].ID)
	_, err = ziti.Services.Create(&automation.ServiceOptions{BaseOptions: automation.BaseOptions{Name: "service-one"}})
	require.Error(t, err)
	require.NoError(t, ziti.Services.Delete(id))
	_, _, creates, deletes := fake.Counts()
	require.Equal(t, 1, creates)
	require.Equal(t, 1, deletes)
}
