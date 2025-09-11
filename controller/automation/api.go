package automation

import (
	"fmt"

	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ZitiAutomation struct {
	client      *Client
	Identities  *IdentityManager
	Services    *ServiceManager
	Policies    *PolicyManager
	Configs     *ConfigManager
	ConfigTypes *ConfigTypeManager
}

func NewZitiAutomation(cfg *Config) (*ZitiAutomation, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client")
	}

	return &ZitiAutomation{
		client:      client,
		Identities:  NewIdentityManager(client),
		Services:    NewServiceManager(client),
		Policies:    NewPolicyManager(client),
		Configs:     NewConfigManager(client),
		ConfigTypes: NewConfigTypeManager(client),
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
		BaseOptions: BaseOptions{
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
		BaseOptions: BaseOptions{
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

func (za *ZitiAutomation) CleanupByTagFilter(tag, value string) error {
	var filter string
	if value == "*" {
		// cleanup all resources with the tag (any value)
		filter = fmt.Sprintf("tags.%s != null", tag)
	} else {
		// cleanup resources with specific tag value
		filter = BuildTagFilter(tag, value)
	}

	// delete service edge router policies first
	if err := za.deleteServiceEdgeRouterPoliciesWithFilter(filter); err != nil {
		return errors.Wrap(err, "failed to delete service edge router policies")
	}

	// delete service policies
	if err := za.Policies.DeleteServicePoliciesWithFilter(filter); err != nil {
		return errors.Wrap(err, "failed to delete service policies")
	}

	// delete configs
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

	// delete edge router policies
	if err := za.deleteEdgeRouterPoliciesWithFilter(filter); err != nil {
		return errors.Wrap(err, "failed to delete edge router policies")
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

func (za *ZitiAutomation) CreateIdentity(name string, identityType rest_model.IdentityType, tags TagStrategy) (string, *ziti.Config, error) {
	opts := &IdentityOptions{
		BaseOptions: BaseOptions{
			Name: name,
			Tags: tags,
		},
		Type:    identityType,
		IsAdmin: false,
	}

	// create identity
	zId, err := za.Identities.Create(opts)
	if err != nil {
		return "", nil, err
	}

	// enroll identity
	cfg, err := za.Identities.Enroll(zId)
	if err != nil {
		return zId, nil, err
	}

	// create edge router policy
	if err := za.EnsureEdgeRouterPolicyForIdentity(name, zId); err != nil {
		// identity was created but policy failed - log but don't fail completely
		logrus.Warnf("failed to create edge router policy for identity '%s' (%s): %v", name, zId, err)
	}

	return zId, cfg, nil
}

func (za *ZitiAutomation) CreateBootstrapIdentity(name string) (string, *ziti.Config, error) {
	tags := ZrokBaseTags()
	return za.CreateIdentity(name, rest_model.IdentityTypeDevice, tags)
}

func (za *ZitiAutomation) CreateEnvironmentIdentity(uniqueToken, accountEmail, envDescription string) (string, *ziti.Config, error) {
	name := accountEmail + "-" + uniqueToken + "-" + envDescription
	tags := NewZrokTagStrategy().WithTag("zrokEmail", accountEmail)
	return za.CreateIdentity(name, rest_model.IdentityTypeUser, tags)
}

func (za *ZitiAutomation) EnsureConfigType(name string) error {
	_, err := za.ConfigTypes.EnsureExists(name)
	return err
}

func (za *ZitiAutomation) FindIdentityByID(id string) (*rest_model.IdentityDetail, error) {
	return za.Identities.GetByID(id)
}

func (za *ZitiAutomation) EnsureEdgeRouterPolicyForIdentity(name, identityID string) error {
	// check if policy already exists
	filter := BuildFilter("name", name) + " and tags.zrok != null"
	opts := &FilterOptions{Filter: filter}

	policies, err := za.Policies.FindEdgeRouterPolicies(opts)
	if err != nil {
		return errors.Wrapf(err, "error listing edge router policies for '%s' (%s)", name, identityID)
	}

	if len(policies) == 1 {
		logrus.Infof("found existing edge router policy for '%s' (%s)", name, identityID)
		return nil
	}

	if len(policies) > 1 {
		return errors.Errorf("found %d edge router policies for '%s' (%s); expected 0 or 1", len(policies), name, identityID)
	}

	// create policy
	logrus.Infof("creating edge router policy for '%s' (%s)", name, identityID)
	zrokTags := ZrokBaseTags()
	_, err = za.CreateEdgeRouterPolicy(name, identityID, zrokTags)
	if err != nil {
		return errors.Wrapf(err, "error creating edge router policy for '%s' (%s)", name, identityID)
	}

	logrus.Infof("asserted edge router policy for '%s' (%s)", name, identityID)
	return nil
}

// helper methods for cleanup operations

func (za *ZitiAutomation) deleteServiceEdgeRouterPoliciesWithFilter(filter string) error {
	opts := &FilterOptions{Filter: filter}
	policies, err := za.findServiceEdgeRouterPolicies(opts)
	if err != nil {
		return err
	}

	logrus.Infof("found %d service edge router policies to delete for filter '%s'", len(policies), filter)

	for _, policy := range policies {
		if err := za.Policies.DeleteServiceEdgeRouterPolicy(*policy.ID); err != nil {
			return err
		}
	}

	return nil
}

func (za *ZitiAutomation) deleteEdgeRouterPoliciesWithFilter(filter string) error {
	opts := &FilterOptions{Filter: filter}
	policies, err := za.Policies.FindEdgeRouterPolicies(opts)
	if err != nil {
		return err
	}

	logrus.Infof("found %d edge router policies to delete for filter '%s'", len(policies), filter)

	for _, policy := range policies {
		if err := za.Policies.DeleteEdgeRouterPolicy(*policy.ID); err != nil {
			return err
		}
	}

	return nil
}

func (za *ZitiAutomation) findServiceEdgeRouterPolicies(opts *FilterOptions) ([]*rest_model.ServiceEdgeRouterPolicyDetail, error) {
	// this method needs to be implemented in the policy manager
	return za.Policies.FindServiceEdgeRouterPolicies(opts)
}
