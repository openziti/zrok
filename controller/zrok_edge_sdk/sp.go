package zrok_edge_sdk

import (
	"context"
	"fmt"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/edge/rest_management_api_client/service_policy"
	"github.com/openziti/edge/rest_model"
	"github.com/sirupsen/logrus"
	"time"
)

func CreateServicePolicyBind(envZId, svcToken, svcZId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	semantic := rest_model.SemanticAllOf
	identityRoles := []string{fmt.Sprintf("@%v", envZId)}
	name := fmt.Sprintf("%v-backend", svcToken)
	var postureCheckRoles []string
	serviceRoles := []string{fmt.Sprintf("@%v", svcZId)}
	dialBind := rest_model.DialBindBind
	svcp := &rest_model.ServicePolicyCreate{
		IdentityRoles:     identityRoles,
		Name:              &name,
		PostureCheckRoles: postureCheckRoles,
		Semantic:          &semantic,
		ServiceRoles:      serviceRoles,
		Type:              &dialBind,
		Tags:              ZrokServiceTags(svcToken),
	}
	req := &service_policy.CreateServicePolicyParams{
		Policy:  svcp,
		Context: context.Background(),
	}
	req.SetTimeout(30 * time.Second)
	resp, err := edge.ServicePolicy.CreateServicePolicy(req, nil)
	if err != nil {
		return err
	}
	logrus.Infof("created bind service policy '%v' for service '%v' for environment '%v'", resp.Payload.Data.ID, svcZId, envZId)
	return nil
}

func CreateNamedBindServicePolicy(name, svcZId, idZId string, edge *rest_management_api_client.ZitiEdgeManagement, tags ...*rest_model.Tags) error {
	allTags := &rest_model.Tags{SubTags: make(rest_model.SubTags)}
	for _, t := range tags {
		for k, v := range t.SubTags {
			allTags.SubTags[k] = v
		}
	}
	identityRoles := []string{"@" + idZId}
	var postureCheckRoles []string
	semantic := rest_model.SemanticAllOf
	serviceRoles := []string{"@" + svcZId}
	dialBind := rest_model.DialBindBind
	sp := &rest_model.ServicePolicyCreate{
		IdentityRoles:     identityRoles,
		Name:              &name,
		PostureCheckRoles: postureCheckRoles,
		Semantic:          &semantic,
		ServiceRoles:      serviceRoles,
		Type:              &dialBind,
		Tags:              allTags,
	}
	req := &service_policy.CreateServicePolicyParams{
		Policy:  sp,
		Context: context.Background(),
	}
	req.SetTimeout(30 * time.Second)
	_, err := edge.ServicePolicy.CreateServicePolicy(req, nil)
	if err != nil {
		return err
	}
	return nil
}

func DeleteServicePolicyBind(envZId, svcToken string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	// type=2 == "Bind"
	return DeleteServicePolicy(envZId, fmt.Sprintf("tags.zrokServiceToken=\"%v\" and type=2", svcToken), edge)
}

func CreateServicePolicyDial(envZId, svcToken, svcZId string, dialZIds []string, edge *rest_management_api_client.ZitiEdgeManagement, tags ...*rest_model.Tags) error {
	allTags := ZrokServiceTags(svcToken)
	for _, t := range tags {
		for k, v := range t.SubTags {
			allTags.SubTags[k] = v
		}
	}

	var identityRoles []string
	for _, proxyIdentity := range dialZIds {
		identityRoles = append(identityRoles, "@"+proxyIdentity)
		logrus.Infof("added proxy identity role '%v'", proxyIdentity)
	}
	name := fmt.Sprintf("%v-dial", svcToken)
	var postureCheckRoles []string
	semantic := rest_model.SemanticAllOf
	serviceRoles := []string{fmt.Sprintf("@%v", svcZId)}
	dialBind := rest_model.DialBindDial
	svcp := &rest_model.ServicePolicyCreate{
		IdentityRoles:     identityRoles,
		Name:              &name,
		PostureCheckRoles: postureCheckRoles,
		Semantic:          &semantic,
		ServiceRoles:      serviceRoles,
		Type:              &dialBind,
		Tags:              allTags,
	}
	req := &service_policy.CreateServicePolicyParams{
		Policy:  svcp,
		Context: context.Background(),
	}
	req.SetTimeout(30 * time.Second)
	resp, err := edge.ServicePolicy.CreateServicePolicy(req, nil)
	if err != nil {
		return err
	}
	logrus.Infof("created dial service policy '%v' for service '%v' for environment '%v'", resp.Payload.Data.ID, svcZId, envZId)
	return nil
}

func CreateNamedDialServicePolicy(name, svcZId, idZId string, edge *rest_management_api_client.ZitiEdgeManagement, tags ...*rest_model.Tags) error {
	allTags := &rest_model.Tags{SubTags: make(rest_model.SubTags)}
	for _, t := range tags {
		for k, v := range t.SubTags {
			allTags.SubTags[k] = v
		}
	}
	identityRoles := []string{"@" + idZId}
	var postureCheckRoles []string
	semantic := rest_model.SemanticAllOf
	serviceRoles := []string{"@" + svcZId}
	dialBind := rest_model.DialBindDial
	sp := &rest_model.ServicePolicyCreate{
		IdentityRoles:     identityRoles,
		Name:              &name,
		PostureCheckRoles: postureCheckRoles,
		Semantic:          &semantic,
		ServiceRoles:      serviceRoles,
		Type:              &dialBind,
		Tags:              allTags,
	}
	req := &service_policy.CreateServicePolicyParams{
		Policy:  sp,
		Context: context.Background(),
	}
	req.SetTimeout(30 * time.Second)
	_, err := edge.ServicePolicy.CreateServicePolicy(req, nil)
	if err != nil {
		return err
	}
	return nil
}

func DeleteServicePolicyDial(envZId, svcToken string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	// type=1 == "Dial"
	return DeleteServicePolicy(envZId, fmt.Sprintf("tags.zrokServiceToken=\"%v\" and type=1", svcToken), edge)
}

func DeleteServicePolicy(envZId, filter string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	limit := int64(1)
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
	if len(listResp.Payload.Data) == 1 {
		spId := *(listResp.Payload.Data[0].ID)
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
	} else {
		logrus.Infof("did not find a service policy")
	}
	return nil
}
