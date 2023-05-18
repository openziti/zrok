package zrokEdgeSdk

import (
	"context"
	"fmt"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/edge/rest_management_api_client/service_policy"
	"github.com/openziti/edge/rest_model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	servicePolicyDial = 1
	servicePolicyBind = 2
)

func CreateServicePolicyBind(name, shrZId, bindZId string, addlTags map[string]interface{}, edge *rest_management_api_client.ZitiEdgeManagement) error {
	semantic := rest_model.SemanticAllOf
	identityRoles := []string{"@" + bindZId}
	serviceRoles := []string{"@" + shrZId}
	spZId, err := createServicePolicy(name, semantic, identityRoles, serviceRoles, addlTags, servicePolicyBind, edge)
	if err != nil {
		return errors.Wrapf(err, "error creating bind service policy for service '%v' for identity '%v'", shrZId, bindZId)
	}
	logrus.Infof("created bind service policy '%v' for service '%v' for identity '%v'", spZId, shrZId, bindZId)
	return nil
}

func CreateServicePolicyDial(name, shrZId string, dialZIds []string, addlTags map[string]interface{}, edge *rest_management_api_client.ZitiEdgeManagement) error {
	semantic := rest_model.SemanticAllOf
	var identityRoles []string
	for _, zId := range dialZIds {
		identityRoles = append(identityRoles, "@"+zId)
	}
	serviceRoles := []string{"@" + shrZId}
	spZId, err := createServicePolicy(name, semantic, identityRoles, serviceRoles, addlTags, servicePolicyDial, edge)
	if err != nil {
		return errors.Wrapf(err, "error creating dial service policy for service '%v' for identities '%v'", shrZId, dialZIds)
	}
	logrus.Infof("created dial service policy '%v' for service '%v' for identities '%v'", spZId, shrZId, dialZIds)
	return nil
}

func createServicePolicy(name string, semantic rest_model.Semantic, identityRoles, serviceRoles []string, addlTags map[string]interface{}, dialBind int, edge *rest_management_api_client.ZitiEdgeManagement) (spZId string, err error) {
	var dialBindType rest_model.DialBind
	switch dialBind {
	case servicePolicyBind:
		dialBindType = rest_model.DialBindBind
	case servicePolicyDial:
		dialBindType = rest_model.DialBindDial
	default:
		return "", errors.Errorf("invalid dial bind type")
	}

	spc := &rest_model.ServicePolicyCreate{
		IdentityRoles:     identityRoles,
		Name:              &name,
		PostureCheckRoles: make([]string, 0),
		Semantic:          &semantic,
		ServiceRoles:      serviceRoles,
		Tags:              MergeTags(ZrokTags(), addlTags),
		Type:              &dialBindType,
	}

	req := &service_policy.CreateServicePolicyParams{
		Policy:  spc,
		Context: context.Background(),
	}
	req.SetTimeout(30 * time.Second)

	resp, err := edge.ServicePolicy.CreateServicePolicy(req, nil)
	if err != nil {
		return "", errors.Wrap(err, "error creating service policy")
	}

	return resp.Payload.Data.ID, nil
}

func DeleteServicePoliciesBind(envZId, shrToken string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	return DeleteServicePolicies(envZId, fmt.Sprintf("tags.zrokShareToken=\"%v\" and type=%d", shrToken, servicePolicyBind), edge)
}

func DeleteServicePoliciesDial(envZId, shrToken string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	return DeleteServicePolicies(envZId, fmt.Sprintf("tags.zrokShareToken=\"%v\" and type=%d", shrToken, servicePolicyDial), edge)
}

func DeleteServicePolicies(envZId, filter string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	limit := int64(0)
	offset := int64(0)
	listReq := &service_policy.ListServicePoliciesParams{
		Filter:  &filter,
		Limit:   &limit,
		Offset:  &offset,
		Context: context.Background(),
	}
	listReq.SetTimeout(30 * time.Second)
	listResp, err := edge.ServicePolicy.ListServicePolicies(listReq, nil)
	if err != nil {
		return err
	}
	logrus.Infof("found %d service policies to delete for '%v'", len(listResp.Payload.Data), filter)
	for i := range listResp.Payload.Data {
		spId := *(listResp.Payload.Data[i].ID)
		req := &service_policy.DeleteServicePolicyParams{
			ID:      spId,
			Context: context.Background(),
		}
		req.SetTimeout(30 * time.Second)
		_, err := edge.ServicePolicy.DeleteServicePolicy(req, nil)
		if err != nil {
			return err
		}
		logrus.Infof("deleted service policy '%v' for environment '%v'", spId, envZId)
	}
	if len(listResp.Payload.Data) < 1 {
		logrus.Warnf("did not find any service policies to delete for '%v'", filter)
	}
	return nil
}
