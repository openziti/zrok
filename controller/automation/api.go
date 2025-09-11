package automation

import (
	"fmt"
	"sync"
	"time"

	"github.com/openziti/zrok/controller/config"
	"github.com/pkg/errors"
)

type ZitiAutomation struct {
	client                    *Client
	Identities                *IdentityManager
	Services                  *ServiceManager
	Configs                   *ConfigManager
	ConfigTypes               *ConfigTypeManager
	EdgeRouterPolicies        *EdgeRouterPolicyManager
	ServiceEdgeRouterPolicies *ServiceEdgeRouterPolicyManager
	ServicePolicies           *ServicePolicyManager
}

func NewZitiAutomation(cfg *Config) (*ZitiAutomation, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client")
	}

	return &ZitiAutomation{
		client:                    client,
		Identities:                client.Identity,
		Services:                  client.Service,
		Configs:                   client.Config,
		ConfigTypes:               client.ConfigType,
		EdgeRouterPolicies:        client.EdgeRouterPolicy,
		ServiceEdgeRouterPolicies: client.ServiceEdgeRouterPolicy,
		ServicePolicies:           client.ServicePolicy,
	}, nil
}

func (za *ZitiAutomation) Client() *Client {
	return za.client
}

// GetZitiAutomation returns a shared automation client instance
func GetZitiAutomation(cfg *config.Config) (*ZitiAutomation, error) {
	automationClientOnce.Do(func() {
		if cfg == nil {
			automationClientErr = errors.New("controller config is nil")
			return
		}

		automationCfg := &Config{
			ApiEndpoint: cfg.Ziti.ApiEndpoint,
			Username:    cfg.Ziti.Username,
			Password:    cfg.Ziti.Password,
		}

		automationClient, automationClientErr = NewZitiAutomation(automationCfg)
	})

	return automationClient, automationClientErr
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

var (
	automationClientOnce sync.Once
	automationClient     *ZitiAutomation
	automationClientErr  error
)
