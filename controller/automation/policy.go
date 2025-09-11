package automation

import (
	"context"
	"time"

	"github.com/openziti/edge-api/rest_management_api_client"
	"github.com/openziti/edge-api/rest_management_api_client/edge_router_policy"
	"github.com/openziti/edge-api/rest_management_api_client/service_edge_router_policy"
	"github.com/openziti/edge-api/rest_management_api_client/service_policy"
	"github.com/openziti/edge-api/rest_model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type PolicyManager struct {
	client *Client
}

func NewPolicyManager(client *Client) *PolicyManager {
	return &PolicyManager{
		client: client,
	}
}

func (pm *PolicyManager) Edge() *rest_management_api_client.ZitiEdgeManagement {
	return pm.client.Edge()
}

func (pm *PolicyManager) Context() context.Context {
	return context.Background()
}

type PolicyBuilder struct {
	name            string
	semantic        rest_model.Semantic
	identityRoles   []string
	serviceRoles    []string
	edgeRouterRoles []string
	tags            TagStrategy
	tagContext      map[string]interface{}
}

func NewPolicyBuilder(name string) *PolicyBuilder {
	return &PolicyBuilder{
		name:     name,
		semantic: rest_model.SemanticAllOf,
	}
}

func (pb *PolicyBuilder) WithSemantic(semantic rest_model.Semantic) *PolicyBuilder {
	pb.semantic = semantic
	return pb
}

func (pb *PolicyBuilder) WithIdentityRoles(roles ...string) *PolicyBuilder {
	pb.identityRoles = append(pb.identityRoles, roles...)
	return pb
}

func (pb *PolicyBuilder) WithIdentityIDs(ids ...string) *PolicyBuilder {
	for _, id := range ids {
		pb.identityRoles = append(pb.identityRoles, "@"+id)
	}
	return pb
}

func (pb *PolicyBuilder) WithServiceRoles(roles ...string) *PolicyBuilder {
	pb.serviceRoles = append(pb.serviceRoles, roles...)
	return pb
}

func (pb *PolicyBuilder) WithServiceIDs(ids ...string) *PolicyBuilder {
	for _, id := range ids {
		pb.serviceRoles = append(pb.serviceRoles, "@"+id)
	}
	return pb
}

func (pb *PolicyBuilder) WithEdgeRouterRoles(roles ...string) *PolicyBuilder {
	pb.edgeRouterRoles = append(pb.edgeRouterRoles, roles...)
	return pb
}

func (pb *PolicyBuilder) WithAllEdgeRouters() *PolicyBuilder {
	pb.edgeRouterRoles = append(pb.edgeRouterRoles, "#all")
	return pb
}

func (pb *PolicyBuilder) WithTags(strategy TagStrategy, context map[string]interface{}) *PolicyBuilder {
	pb.tags = strategy
	pb.tagContext = context
	return pb
}

func (pb *PolicyBuilder) getTags() *rest_model.Tags {
	if pb.tags != nil {
		return pb.tags.GenerateTags(pb.tagContext)
	}
	return &rest_model.Tags{SubTags: make(map[string]interface{})}
}

// edge router policies
func (pm *PolicyManager) CreateEdgeRouterPolicy(builder *PolicyBuilder) (string, error) {
	erp := &rest_model.EdgeRouterPolicyCreate{
		EdgeRouterRoles: builder.edgeRouterRoles,
		IdentityRoles:   builder.identityRoles,
		Name:            &builder.name,
		Semantic:        &builder.semantic,
		Tags:            builder.getTags(),
	}

	req := &edge_router_policy.CreateEdgeRouterPolicyParams{
		Policy:  erp,
		Context: pm.Context(),
	}
	req.SetTimeout(30 * time.Second)

	resp, err := pm.Edge().EdgeRouterPolicy.CreateEdgeRouterPolicy(req, nil)
	if err != nil {
		return "", errors.Wrapf(err, "error creating edge router policy '%s'", builder.name)
	}

	logrus.Infof("created edge router policy '%s' with id '%s'", builder.name, resp.Payload.Data.ID)
	return resp.Payload.Data.ID, nil
}

func (pm *PolicyManager) DeleteEdgeRouterPolicy(id string) error {
	req := &edge_router_policy.DeleteEdgeRouterPolicyParams{
		ID:      id,
		Context: pm.Context(),
	}
	req.SetTimeout(30 * time.Second)

	_, err := pm.Edge().EdgeRouterPolicy.DeleteEdgeRouterPolicy(req, nil)
	if err != nil {
		return errors.Wrapf(err, "error deleting edge router policy '%s'", id)
	}

	logrus.Infof("deleted edge router policy '%s'", id)
	return nil
}

// service edge router policies
func (pm *PolicyManager) CreateServiceEdgeRouterPolicy(builder *PolicyBuilder) (string, error) {
	serp := &rest_model.ServiceEdgeRouterPolicyCreate{
		EdgeRouterRoles: builder.edgeRouterRoles,
		Name:            &builder.name,
		Semantic:        &builder.semantic,
		ServiceRoles:    builder.serviceRoles,
		Tags:            builder.getTags(),
	}

	req := &service_edge_router_policy.CreateServiceEdgeRouterPolicyParams{
		Policy:  serp,
		Context: pm.Context(),
	}
	req.SetTimeout(30 * time.Second)

	resp, err := pm.Edge().ServiceEdgeRouterPolicy.CreateServiceEdgeRouterPolicy(req, nil)
	if err != nil {
		return "", errors.Wrapf(err, "error creating service edge router policy '%s'", builder.name)
	}

	logrus.Infof("created service edge router policy '%s' with id '%s'", builder.name, resp.Payload.Data.ID)
	return resp.Payload.Data.ID, nil
}

func (pm *PolicyManager) DeleteServiceEdgeRouterPolicy(id string) error {
	req := &service_edge_router_policy.DeleteServiceEdgeRouterPolicyParams{
		ID:      id,
		Context: pm.Context(),
	}
	req.SetTimeout(30 * time.Second)

	_, err := pm.Edge().ServiceEdgeRouterPolicy.DeleteServiceEdgeRouterPolicy(req, nil)
	if err != nil {
		return errors.Wrapf(err, "error deleting service edge router policy '%s'", id)
	}

	logrus.Infof("deleted service edge router policy '%s'", id)
	return nil
}

// service policies
func (pm *PolicyManager) CreateServicePolicy(builder *PolicyBuilder, policyType rest_model.DialBind) (string, error) {
	spc := &rest_model.ServicePolicyCreate{
		IdentityRoles:     builder.identityRoles,
		Name:              &builder.name,
		PostureCheckRoles: make([]string, 0),
		Semantic:          &builder.semantic,
		ServiceRoles:      builder.serviceRoles,
		Tags:              builder.getTags(),
		Type:              &policyType,
	}

	req := &service_policy.CreateServicePolicyParams{
		Policy:  spc,
		Context: pm.Context(),
	}
	req.SetTimeout(30 * time.Second)

	resp, err := pm.Edge().ServicePolicy.CreateServicePolicy(req, nil)
	if err != nil {
		return "", errors.Wrapf(err, "error creating service policy '%s'", builder.name)
	}

	logrus.Infof("created service policy '%s' with id '%s'", builder.name, resp.Payload.Data.ID)
	return resp.Payload.Data.ID, nil
}

func (pm *PolicyManager) CreateServicePolicyBind(builder *PolicyBuilder) (string, error) {
	return pm.CreateServicePolicy(builder, rest_model.DialBindBind)
}

func (pm *PolicyManager) CreateServicePolicyDial(builder *PolicyBuilder) (string, error) {
	return pm.CreateServicePolicy(builder, rest_model.DialBindDial)
}

func (pm *PolicyManager) DeleteServicePolicy(id string) error {
	req := &service_policy.DeleteServicePolicyParams{
		ID:      id,
		Context: pm.Context(),
	}
	req.SetTimeout(30 * time.Second)

	_, err := pm.Edge().ServicePolicy.DeleteServicePolicy(req, nil)
	if err != nil {
		return errors.Wrapf(err, "error deleting service policy '%s'", id)
	}

	logrus.Infof("deleted service policy '%s'", id)
	return nil
}

func (pm *PolicyManager) FindServicePolicies(opts *FilterOptions) ([]*rest_model.ServicePolicyDetail, error) {
	req := &service_policy.ListServicePoliciesParams{
		Filter:  &opts.Filter,
		Limit:   &opts.Limit,
		Offset:  &opts.Offset,
		Context: pm.Context(),
	}
	req.SetTimeout(opts.GetTimeout())

	resp, err := pm.Edge().ServicePolicy.ListServicePolicies(req, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error listing service policies")
	}

	return resp.Payload.Data, nil
}

func (pm *PolicyManager) DeleteServicePoliciesWithFilter(filter string) error {
	opts := &FilterOptions{Filter: filter}
	policies, err := pm.FindServicePolicies(opts)
	if err != nil {
		return err
	}

	logrus.Infof("found %d service policies to delete for filter '%s'", len(policies), filter)

	for _, policy := range policies {
		if err := pm.DeleteServicePolicy(*policy.ID); err != nil {
			return err
		}
	}

	if len(policies) == 0 {
		logrus.Warnf("no service policies found for filter '%s'", filter)
	}

	return nil
}

func (pm *PolicyManager) FindEdgeRouterPolicies(opts *FilterOptions) ([]*rest_model.EdgeRouterPolicyDetail, error) {
	req := &edge_router_policy.ListEdgeRouterPoliciesParams{
		Filter:  &opts.Filter,
		Limit:   &opts.Limit,
		Offset:  &opts.Offset,
		Context: pm.Context(),
	}
	req.SetTimeout(opts.GetTimeout())

	resp, err := pm.Edge().EdgeRouterPolicy.ListEdgeRouterPolicies(req, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error listing edge router policies")
	}

	return resp.Payload.Data, nil
}

func (pm *PolicyManager) FindServiceEdgeRouterPolicies(opts *FilterOptions) ([]*rest_model.ServiceEdgeRouterPolicyDetail, error) {
	req := &service_edge_router_policy.ListServiceEdgeRouterPoliciesParams{
		Filter:  &opts.Filter,
		Limit:   &opts.Limit,
		Offset:  &opts.Offset,
		Context: pm.Context(),
	}
	req.SetTimeout(opts.GetTimeout())

	resp, err := pm.Edge().ServiceEdgeRouterPolicy.ListServiceEdgeRouterPolicies(req, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error listing service edge router policies")
	}

	return resp.Payload.Data, nil
}
