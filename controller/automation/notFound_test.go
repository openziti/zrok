package automation

import (
	stderrors "errors"
	"testing"

	"github.com/openziti/edge-api/rest_management_api_client/config"
	"github.com/openziti/edge-api/rest_management_api_client/edge_router_policy"
	"github.com/openziti/edge-api/rest_management_api_client/identity"
	"github.com/openziti/edge-api/rest_management_api_client/service"
	"github.com/openziti/edge-api/rest_management_api_client/service_edge_router_policy"
	"github.com/openziti/edge-api/rest_management_api_client/service_policy"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/zrok/v2/controller/automation/zitifake"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestIsNotFoundByType(t *testing.T) {
	notFound := map[string]error{
		"identity":                   &identity.DeleteIdentityNotFound{},
		"service":                    &service.DeleteServiceNotFound{},
		"config":                     &config.DeleteConfigNotFound{},
		"config type":                &config.DeleteConfigTypeNotFound{},
		"service policy":             &service_policy.DeleteServicePolicyNotFound{},
		"service edge router policy": &service_edge_router_policy.DeleteServiceEdgeRouterPolicyNotFound{},
		"edge router policy":         &edge_router_policy.DeleteEdgeRouterPolicyNotFound{},
		"automation":                 NewNotFoundError("service", "GetByName", stderrors.New("missing")),
	}
	for name, err := range notFound {
		require.True(t, IsNotFound(err), name)
		require.True(t, IsNotFound(errors.Wrap(err, "error deleting")), "wrapped %s", name)
	}
	present := map[string]error{
		"nil":                   nil,
		"unauthorized":          &identity.DeleteIdentityUnauthorized{},
		"plain":                 stderrors.New("not found"),
		"automation permission": &AutomationError{Type: ErrorTypePermission, Resource: "service", Operation: "Delete", Cause: stderrors.New("denied")},
	}
	for name, err := range present {
		require.False(t, IsNotFound(err), name)
		require.False(t, IsNotFound(errors.Wrap(err, "error deleting")), "wrapped %s", name)
	}
}

func deleteFilterFixture(t *testing.T) (*zitifake.Server, *ZitiAutomation, []string) {
	t.Helper()
	fake := zitifake.New()
	t.Cleanup(fake.Close)
	ziti := NewZitiAutomationWithEdge(fake.Edge())
	var ids []string
	for _, name := range []string{"first", "second"} {
		id, err := ziti.ServicePolicies.Create(&ServicePolicyOptions{
			BaseOptions:   BaseOptions{Name: name, Tags: ZrokShareTags("vanishing")},
			IdentityRoles: []string{"@identity"}, ServiceRoles: []string{"@service"},
			PolicyType: rest_model.DialBindBind, Semantic: rest_model.SemanticAllOf,
		})
		require.NoError(t, err)
		ids = append(ids, id)
	}
	return fake, ziti, ids
}

func TestDeleteWithFilterContinuesPastVanished(t *testing.T) {
	fake, ziti, ids := deleteFilterFixture(t)
	deleter := func(id string) error {
		if id == ids[0] {
			// the object disappears between the listing and its delete.
			fake.Remove(zitifake.ServicePolicies, id)
		}
		return ziti.ServicePolicies.Delete(id)
	}
	require.NoError(t, DeleteWithFilter(ziti.ServicePolicies.Find, deleter, BuildTagFilter("zrokShareToken", "vanishing"), "service policy"))
	_, deleted := fake.Log()
	require.Equal(t, []string{zitifake.ServicePolicies + "/" + ids[1]}, deleted)
	require.Empty(t, fake.Tagged("zrokShareToken", "vanishing"))
}

func TestDeleteWithFilterPropagatesOtherErrors(t *testing.T) {
	fake, ziti, ids := deleteFilterFixture(t)
	deleter := func(id string) error {
		fake.RejectOperations(id == ids[0])
		return ziti.ServicePolicies.Delete(id)
	}
	err := DeleteWithFilter(ziti.ServicePolicies.Find, deleter, BuildTagFilter("zrokShareToken", "vanishing"), "service policy")
	var unauthorized *service_policy.DeleteServicePolicyUnauthorized
	require.True(t, errors.As(err, &unauthorized), "%v", err)
	_, deleted := fake.Log()
	require.Empty(t, deleted)
	require.Len(t, fake.Tagged("zrokShareToken", "vanishing"), 2)
}

func TestWrapEdgeErrorCarriesZitiMessage(t *testing.T) {
	fake := zitifake.New()
	t.Cleanup(fake.Close)
	ziti := NewZitiAutomationWithEdge(fake.Edge())

	err := ziti.Services.Delete("missing")

	require.Error(t, err)
	require.Contains(t, err.Error(), "error deleting service 'missing': ziti Not Found: service not found: missing")
	require.Contains(t, err.Error(), "[DELETE /services/{id}][404] deleteServiceNotFound")
	require.NotContains(t, err.Error(), "0x")
	require.True(t, IsNotFound(err))
	var generated *service.DeleteServiceNotFound
	require.True(t, errors.As(err, &generated))
}

func TestWrapEdgeErrorWithoutEnvelope(t *testing.T) {
	cause := stderrors.New("connection refused")

	err := wrapEdgeError(cause, "error deleting service '%s'", "id")

	require.Equal(t, errors.Wrapf(cause, "error deleting service '%s'", "id").Error(), err.Error())
	require.Equal(t, cause, errors.Cause(err))
}
