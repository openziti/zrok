package automation

import (
	"fmt"
	"time"

	"github.com/openziti/edge-api/rest_management_api_client"
	"github.com/openziti/edge-api/rest_management_api_client/config"
	"github.com/openziti/edge-api/rest_management_api_client/edge_router_policy"
	"github.com/openziti/edge-api/rest_management_api_client/identity"
	"github.com/openziti/edge-api/rest_management_api_client/service"
	"github.com/openziti/edge-api/rest_management_api_client/service_edge_router_policy"
	"github.com/openziti/edge-api/rest_management_api_client/service_policy"
	"github.com/pkg/errors"
)

type Config struct {
	ApiEndpoint string
	Username    string
	Password    string `dd:"+secret"`
}

type ZitiAutomation struct {
	edge                      *rest_management_api_client.ZitiEdgeManagement
	Identities                *IdentityManager
	Services                  *ServiceManager
	Configs                   *ConfigManager
	ConfigTypes               *ConfigTypeManager
	EdgeRouterPolicies        *EdgeRouterPolicyManager
	ServiceEdgeRouterPolicies *ServiceEdgeRouterPolicyManager
	ServicePolicies           *ServicePolicyManager
}

func NewZitiAutomation(cfg *Config) (*ZitiAutomation, error) {
	return sharedSessions.get(cfg)
}

func (za *ZitiAutomation) Edge() *rest_management_api_client.ZitiEdgeManagement {
	return za.edge
}

// error helper methods to simplify error handling

func (za *ZitiAutomation) IsNotFound(err error) bool {
	return IsNotFound(err)
}

// IsNotFound reports whether err says the object is absent: a not-found from the package's read
// helpers, or the ziti api's not-found answer to one of the deletes this package issues. the pinned
// edge-api (v0.26.48) generates no Code() accessor on its response types, so this list of generated
// Delete*NotFound types is the contract; a delete added to the package adds its type here.
// unauthorized, forbidden, connectivity and validation errors are not absence and are not matched.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	var automationErr *AutomationError
	if errors.As(err, &automationErr) {
		return automationErr.IsNotFound()
	}
	return isError[*identity.DeleteIdentityNotFound](err) ||
		isError[*service.DeleteServiceNotFound](err) ||
		isError[*config.DeleteConfigNotFound](err) ||
		isError[*config.DeleteConfigTypeNotFound](err) ||
		isError[*service_policy.DeleteServicePolicyNotFound](err) ||
		isError[*service_edge_router_policy.DeleteServiceEdgeRouterPolicyNotFound](err) ||
		isError[*edge_router_policy.DeleteEdgeRouterPolicyNotFound](err)
}

// IsRateLimited reports whether err is a ziti rate limiter's refusal, from either the command rate
// limiter (seen by the session transport) or the authentication limiter.
func IsRateLimited(err error) bool {
	return isError[*RateLimitedError](err)
}

func isError[T error](err error) bool {
	var target T
	return errors.As(err, &target)
}

func (za *ZitiAutomation) ShouldRetry(err error) bool {
	var automationErr *AutomationError
	if errors.As(err, &automationErr) {
		return automationErr.IsRetryable()
	}
	return false
}

func (za *ZitiAutomation) CleanupByTag(tag, value string) error {
	var filter string
	if value == "*" {
		// cleanup all resources with the tag (any value)
		filter = fmt.Sprintf("tags.%s != null", tag)
	} else {
		// cleanup resources with specific tag value
		filter = BuildTagFilter(tag, value)
	}

	// delete service edge router policies
	if err := za.ServiceEdgeRouterPolicies.DeleteWithFilter(filter); err != nil {
		return errors.Wrap(err, "failed to delete service edge router policies")
	}

	// delete service policies
	if err := za.ServicePolicies.DeleteWithFilter(filter); err != nil {
		return errors.Wrap(err, "failed to delete service policies")
	}

	// delete configs
	if err := za.Configs.DeleteWithFilter(filter); err != nil {
		return errors.Wrap(err, "failed to delete configs")
	}

	// delete services
	if err := za.Services.DeleteWithFilter(filter); err != nil {
		return errors.Wrap(err, "failed to delete services")
	}

	// delete edge router policies
	if err := za.EdgeRouterPolicies.DeleteWithFilter(filter); err != nil {
		return errors.Wrap(err, "failed to delete edge router policies")
	}

	// find and delete identities
	if err := za.Identities.DeleteWithFilter(filter); err != nil {
		return errors.Wrap(err, "failed to delete identities")
	}

	return nil
}

const (
	// DefaultRequestTimeout is the default timeout for API requests
	DefaultRequestTimeout = 30 * time.Second

	// DefaultOperationTimeout is the default timeout for CRUD operations
	DefaultOperationTimeout = 30 * time.Second
)
