package automation

import (
	"fmt"
	"time"

	"github.com/openziti/edge-api/rest_management_api_client"
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
	var automationErr *AutomationError
	if errors.As(err, &automationErr) {
		return automationErr.IsNotFound()
	}
	return false
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
