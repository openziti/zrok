package automation

import (
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/edge-api/rest_management_api_client/edge_router_policy"
	"github.com/openziti/edge-api/rest_model"
	"github.com/pkg/errors"
)

type EdgeRouterPolicyManager struct {
	*BaseResourceManager[rest_model.EdgeRouterPolicyDetail]
}

func NewEdgeRouterPolicyManager(ziti *ZitiAutomation) *EdgeRouterPolicyManager {
	return &EdgeRouterPolicyManager{
		BaseResourceManager: NewBaseResourceManager[rest_model.EdgeRouterPolicyDetail](ziti),
	}
}

type EdgeRouterPolicyOptions struct {
	BaseOptions
	IdentityRoles   []string
	EdgeRouterRoles []string
	Semantic        rest_model.Semantic
}

func (erpm *EdgeRouterPolicyManager) Create(opts *EdgeRouterPolicyOptions) (string, error) {
	erp := &rest_model.EdgeRouterPolicyCreate{
		EdgeRouterRoles: opts.EdgeRouterRoles,
		IdentityRoles:   opts.IdentityRoles,
		Name:            &opts.Name,
		Semantic:        &opts.Semantic,
		Tags:            opts.GetTags(),
	}

	req := &edge_router_policy.CreateEdgeRouterPolicyParams{
		Policy:  erp,
		Context: erpm.Context(),
	}
	req.SetTimeout(opts.GetTimeout())

	resp, err := erpm.Edge().EdgeRouterPolicy.CreateEdgeRouterPolicy(req, nil)
	if err != nil {
		return "", errors.Wrapf(err, "error creating edge router policy '%s'", opts.Name)
	}

	dl.Infof("created edge router policy '%s' with id '%s'", opts.Name, resp.Payload.Data.ID)
	return resp.Payload.Data.ID, nil
}

func (erpm *EdgeRouterPolicyManager) Delete(id string) error {
	req := &edge_router_policy.DeleteEdgeRouterPolicyParams{
		ID:      id,
		Context: erpm.Context(),
	}
	req.SetTimeout(DefaultOperationTimeout)

	_, err := erpm.Edge().EdgeRouterPolicy.DeleteEdgeRouterPolicy(req, nil)
	if err != nil {
		return errors.Wrapf(err, "error deleting edge router policy '%s'", id)
	}

	dl.Infof("deleted edge router policy '%s'", id)
	return nil
}

func (erpm *EdgeRouterPolicyManager) Find(opts *FilterOptions) ([]*rest_model.EdgeRouterPolicyDetail, error) {
	req := &edge_router_policy.ListEdgeRouterPoliciesParams{
		Filter:  &opts.Filter,
		Limit:   &opts.Limit,
		Offset:  &opts.Offset,
		Context: erpm.Context(),
	}
	req.SetTimeout(opts.GetTimeout())

	resp, err := erpm.Edge().EdgeRouterPolicy.ListEdgeRouterPolicies(req, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error listing edge router policies")
	}

	return resp.Payload.Data, nil
}

func (erpm *EdgeRouterPolicyManager) GetByID(id string) (*rest_model.EdgeRouterPolicyDetail, error) {
	return GetByID(erpm.Find, id, "edge router policy")
}

func (erpm *EdgeRouterPolicyManager) GetByName(name string) (*rest_model.EdgeRouterPolicyDetail, error) {
	return GetByName(erpm.Find, name, "edge router policy")
}

func (erpm *EdgeRouterPolicyManager) DeleteWithFilter(filter string) error {
	return DeleteWithFilter(erpm.Find, erpm.Delete, filter, "edge router policy")
}

// ensure EdgeRouterPolicyManager implements the interface
var _ IResourceManager[rest_model.EdgeRouterPolicyDetail, *EdgeRouterPolicyOptions] = (*EdgeRouterPolicyManager)(nil)
