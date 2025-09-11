package automation

import (
	"github.com/openziti/edge-api/rest_management_api_client/service_policy"
	"github.com/openziti/edge-api/rest_model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ServicePolicyManager struct {
	*BaseResourceManager[rest_model.ServicePolicyDetail]
}

func NewServicePolicyManager(client *Client) *ServicePolicyManager {
	return &ServicePolicyManager{
		BaseResourceManager: NewBaseResourceManager[rest_model.ServicePolicyDetail](client),
	}
}

type ServicePolicyOptions struct {
	BaseOptions
	IdentityRoles []string
	ServiceRoles  []string
	PolicyType    rest_model.DialBind
	Semantic      rest_model.Semantic
}

func (spm *ServicePolicyManager) Create(opts *ServicePolicyOptions) (string, error) {
	spc := &rest_model.ServicePolicyCreate{
		IdentityRoles:     opts.IdentityRoles,
		Name:              &opts.Name,
		PostureCheckRoles: make([]string, 0),
		Semantic:          &opts.Semantic,
		ServiceRoles:      opts.ServiceRoles,
		Tags:              opts.GetTags(),
		Type:              &opts.PolicyType,
	}

	req := &service_policy.CreateServicePolicyParams{
		Policy:  spc,
		Context: spm.Context(),
	}
	req.SetTimeout(opts.GetTimeout())

	resp, err := spm.Edge().ServicePolicy.CreateServicePolicy(req, nil)
	if err != nil {
		return "", errors.Wrapf(err, "error creating service policy '%s'", opts.Name)
	}

	logrus.Infof("created service policy '%s' with id '%s'", opts.Name, resp.Payload.Data.ID)
	return resp.Payload.Data.ID, nil
}

func (spm *ServicePolicyManager) Delete(id string) error {
	req := &service_policy.DeleteServicePolicyParams{
		ID:      id,
		Context: spm.Context(),
	}
	req.SetTimeout(DefaultOperationTimeout)

	_, err := spm.Edge().ServicePolicy.DeleteServicePolicy(req, nil)
	if err != nil {
		return errors.Wrapf(err, "error deleting service policy '%s'", id)
	}

	logrus.Infof("deleted service policy '%s'", id)
	return nil
}

func (spm *ServicePolicyManager) Find(opts *FilterOptions) ([]*rest_model.ServicePolicyDetail, error) {
	req := &service_policy.ListServicePoliciesParams{
		Filter:  &opts.Filter,
		Limit:   &opts.Limit,
		Offset:  &opts.Offset,
		Context: spm.Context(),
	}
	req.SetTimeout(opts.GetTimeout())

	resp, err := spm.Edge().ServicePolicy.ListServicePolicies(req, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error listing service policies")
	}

	return resp.Payload.Data, nil
}

func (spm *ServicePolicyManager) GetByID(id string) (*rest_model.ServicePolicyDetail, error) {
	return GetByID(spm.Find, id, "service policy")
}

func (spm *ServicePolicyManager) GetByName(name string) (*rest_model.ServicePolicyDetail, error) {
	return GetByName(spm.Find, name, "service policy")
}

func (spm *ServicePolicyManager) DeleteWithFilter(filter string) error {
	return DeleteWithFilter(spm.Find, spm.Delete, filter, "service policy")
}

// convenience methods for specific policy types
func (spm *ServicePolicyManager) CreateBind(opts *ServicePolicyOptions) (string, error) {
	opts.PolicyType = rest_model.DialBindBind
	return spm.Create(opts)
}

func (spm *ServicePolicyManager) CreateDial(opts *ServicePolicyOptions) (string, error) {
	opts.PolicyType = rest_model.DialBindDial
	return spm.Create(opts)
}

// ensure ServicePolicyManager implements the interface
var _ IResourceManager[rest_model.ServicePolicyDetail, *ServicePolicyOptions] = (*ServicePolicyManager)(nil)