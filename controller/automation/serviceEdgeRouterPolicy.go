package automation

import (
	"github.com/openziti/edge-api/rest_management_api_client/service_edge_router_policy"
	"github.com/openziti/edge-api/rest_model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ServiceEdgeRouterPolicyManager struct {
	*BaseResourceManager[rest_model.ServiceEdgeRouterPolicyDetail]
}

func NewServiceEdgeRouterPolicyManager(client *Client) *ServiceEdgeRouterPolicyManager {
	return &ServiceEdgeRouterPolicyManager{
		BaseResourceManager: NewBaseResourceManager[rest_model.ServiceEdgeRouterPolicyDetail](client),
	}
}

type ServiceEdgeRouterPolicyOptions struct {
	BaseOptions
	ServiceRoles    []string
	EdgeRouterRoles []string
	Semantic        rest_model.Semantic
}

func (serpm *ServiceEdgeRouterPolicyManager) Create(opts *ServiceEdgeRouterPolicyOptions) (string, error) {
	serp := &rest_model.ServiceEdgeRouterPolicyCreate{
		EdgeRouterRoles: opts.EdgeRouterRoles,
		Name:            &opts.Name,
		Semantic:        &opts.Semantic,
		ServiceRoles:    opts.ServiceRoles,
		Tags:            opts.GetTags(),
	}

	req := &service_edge_router_policy.CreateServiceEdgeRouterPolicyParams{
		Policy:  serp,
		Context: serpm.Context(),
	}
	req.SetTimeout(opts.GetTimeout())

	resp, err := serpm.Edge().ServiceEdgeRouterPolicy.CreateServiceEdgeRouterPolicy(req, nil)
	if err != nil {
		return "", errors.Wrapf(err, "error creating service edge router policy '%s'", opts.Name)
	}

	logrus.Infof("created service edge router policy '%s' with id '%s'", opts.Name, resp.Payload.Data.ID)
	return resp.Payload.Data.ID, nil
}

func (serpm *ServiceEdgeRouterPolicyManager) Delete(id string) error {
	req := &service_edge_router_policy.DeleteServiceEdgeRouterPolicyParams{
		ID:      id,
		Context: serpm.Context(),
	}
	req.SetTimeout(DefaultOperationTimeout)

	_, err := serpm.Edge().ServiceEdgeRouterPolicy.DeleteServiceEdgeRouterPolicy(req, nil)
	if err != nil {
		return errors.Wrapf(err, "error deleting service edge router policy '%s'", id)
	}

	logrus.Infof("deleted service edge router policy '%s'", id)
	return nil
}

func (serpm *ServiceEdgeRouterPolicyManager) Find(opts *FilterOptions) ([]*rest_model.ServiceEdgeRouterPolicyDetail, error) {
	req := &service_edge_router_policy.ListServiceEdgeRouterPoliciesParams{
		Filter:  &opts.Filter,
		Limit:   &opts.Limit,
		Offset:  &opts.Offset,
		Context: serpm.Context(),
	}
	req.SetTimeout(opts.GetTimeout())

	resp, err := serpm.Edge().ServiceEdgeRouterPolicy.ListServiceEdgeRouterPolicies(req, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error listing service edge router policies")
	}

	return resp.Payload.Data, nil
}

func (serpm *ServiceEdgeRouterPolicyManager) GetByID(id string) (*rest_model.ServiceEdgeRouterPolicyDetail, error) {
	return GetByID(serpm.Find, id, "service edge router policy")
}

func (serpm *ServiceEdgeRouterPolicyManager) GetByName(name string) (*rest_model.ServiceEdgeRouterPolicyDetail, error) {
	return GetByName(serpm.Find, name, "service edge router policy")
}

func (serpm *ServiceEdgeRouterPolicyManager) DeleteWithFilter(filter string) error {
	return DeleteWithFilter(serpm.Find, serpm.Delete, filter, "service edge router policy")
}

// ensure ServiceEdgeRouterPolicyManager implements the interface
var _ IResourceManager[rest_model.ServiceEdgeRouterPolicyDetail, *ServiceEdgeRouterPolicyOptions] = (*ServiceEdgeRouterPolicyManager)(nil)