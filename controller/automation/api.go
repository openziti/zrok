package automation

import (
	"github.com/openziti/edge-api/rest_model"
	"github.com/pkg/errors"
)

type ZitiAutomation struct {
	client     *Client
	Identities *IdentityManager
	Services   *ServiceManager
	Policies   *PolicyManager
	Configs    *ConfigManager
}

func NewZitiAutomation(cfg *Config) (*ZitiAutomation, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client")
	}

	return &ZitiAutomation{
		client:     client,
		Identities: NewIdentityManager(client),
		Services:   NewServiceManager(client),
		Policies:   NewPolicyManager(client),
		Configs:    NewConfigManager(client),
	}, nil
}

func (za *ZitiAutomation) Client() *Client {
	return za.client
}

// convenience methods for common patterns

func (za *ZitiAutomation) CreateUserIdentityWithTag(name, tag, value string) (string, error) {
	identityType := rest_model.IdentityTypeUser
	tags := NewSimpleTagStrategy(map[string]interface{}{tag: value})

	opts := &IdentityOptions{
		ResourceOptions: &ResourceOptions{
			Name: name,
			Tags: tags,
		},
		Type:    identityType,
		IsAdmin: false,
	}

	return za.Identities.Create(opts)
}

func (za *ZitiAutomation) CreateServiceWithConfig(name, configID string, tags TagStrategy) (string, error) {
	opts := &ServiceOptions{
		ResourceOptions: &ResourceOptions{
			Name: name,
			Tags: tags,
		},
		Configs:            []string{configID},
		EncryptionRequired: true,
	}

	return za.Services.Create(opts)
}

func (za *ZitiAutomation) CreateBindPolicy(name, serviceID, identityID string, tags TagStrategy) (string, error) {
	builder := NewPolicyBuilder(name).
		WithServiceIDs(serviceID).
		WithIdentityIDs(identityID).
		WithTags(tags, nil)

	return za.Policies.CreateServicePolicyBind(builder)
}

func (za *ZitiAutomation) CreateDialPolicy(name, serviceID string, identityIDs []string, tags TagStrategy) (string, error) {
	builder := NewPolicyBuilder(name).
		WithServiceIDs(serviceID).
		WithIdentityIDs(identityIDs...).
		WithTags(tags, nil)

	return za.Policies.CreateServicePolicyDial(builder)
}

func (za *ZitiAutomation) CreateEdgeRouterPolicy(name, identityID string, tags TagStrategy) (string, error) {
	builder := NewPolicyBuilder(name).
		WithIdentityIDs(identityID).
		WithAllEdgeRouters().
		WithTags(tags, nil)

	return za.Policies.CreateEdgeRouterPolicy(builder)
}

func (za *ZitiAutomation) CreateServiceEdgeRouterPolicy(name, serviceID string, tags TagStrategy) (string, error) {
	builder := NewPolicyBuilder(name).
		WithServiceIDs(serviceID).
		WithAllEdgeRouters().
		WithTags(tags, nil)

	return za.Policies.CreateServiceEdgeRouterPolicy(builder)
}

// cleanup methods

func (za *ZitiAutomation) CleanupByTagFilter(tag, value string) error {
	filter := BuildTagFilter(tag, value)

	// delete service policies first
	if err := za.Policies.DeleteServicePoliciesWithFilter(filter); err != nil {
		return errors.Wrap(err, "failed to delete service policies")
	}

	// then delete configs
	if err := za.Configs.DeleteWithFilter(filter); err != nil {
		return errors.Wrap(err, "failed to delete configs")
	}

	// find and delete services
	services, err := za.Services.Find(&FilterOptions{Filter: filter})
	if err != nil {
		return errors.Wrap(err, "failed to find services")
	}

	for _, service := range services {
		if err := za.Services.Delete(*service.ID); err != nil {
			return errors.Wrapf(err, "failed to delete service %s", *service.ID)
		}
	}

	// find and delete identities
	identities, err := za.Identities.Find(&FilterOptions{Filter: filter})
	if err != nil {
		return errors.Wrap(err, "failed to find identities")
	}

	for _, identity := range identities {
		if err := za.Identities.Delete(*identity.ID); err != nil {
			return errors.Wrapf(err, "failed to delete identity %s", *identity.ID)
		}
	}

	return nil
}
