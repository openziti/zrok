package zrokEdgeSdk

import (
	"context"
	"fmt"
	"github.com/openziti/edge-api/rest_management_api_client"
	"github.com/openziti/edge-api/rest_management_api_client/service_edge_router_policy"
	"github.com/openziti/edge-api/rest_model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

func CreateShareServiceEdgeRouterPolicy(envZId, shrToken, shrZId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	serpZId, err := CreateServiceEdgeRouterPolicy(shrToken, shrZId, ZrokShareTags(shrToken).SubTags, edge)
	if err != nil {
		return err
	}
	logrus.Infof("created service edge router policy '%v' for service '%v' for environment '%v'", serpZId, shrZId, envZId)
	return nil
}

func CreateServiceEdgeRouterPolicy(name, shrZId string, moreTags map[string]interface{}, edge *rest_management_api_client.ZitiEdgeManagement) (string, error) {
	edgeRouterRoles := []string{"#all"}
	semantic := rest_model.SemanticAllOf
	serviceRoles := []string{fmt.Sprintf("@%v", shrZId)}
	tags := ZrokTags()
	for k, v := range moreTags {
		tags.SubTags[k] = v
	}
	serp := &rest_model.ServiceEdgeRouterPolicyCreate{
		EdgeRouterRoles: edgeRouterRoles,
		Name:            &name,
		Semantic:        &semantic,
		ServiceRoles:    serviceRoles,
		Tags:            tags,
	}
	serpParams := &service_edge_router_policy.CreateServiceEdgeRouterPolicyParams{
		Policy:  serp,
		Context: context.Background(),
	}
	serpParams.SetTimeout(30 * time.Second)
	resp, err := edge.ServiceEdgeRouterPolicy.CreateServiceEdgeRouterPolicy(serpParams, nil)
	if err != nil {
		return "", errors.Wrapf(err, "error creating serp '%v' for service '%v'", name, shrZId)
	}
	return resp.Payload.Data.ID, nil
}

func DeleteServiceEdgeRouterPolicy(envZId, shrToken string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	filter := fmt.Sprintf("tags.zrokShareToken=\"%v\"", shrToken)
	limit := int64(1)
	offset := int64(0)
	listReq := &service_edge_router_policy.ListServiceEdgeRouterPoliciesParams{
		Filter:  &filter,
		Limit:   &limit,
		Offset:  &offset,
		Context: context.Background(),
	}
	listReq.SetTimeout(30 * time.Second)
	listResp, err := edge.ServiceEdgeRouterPolicy.ListServiceEdgeRouterPolicies(listReq, nil)
	if err != nil {
		return err
	}
	if len(listResp.Payload.Data) == 1 {
		serpId := *(listResp.Payload.Data[0].ID)
		req := &service_edge_router_policy.DeleteServiceEdgeRouterPolicyParams{
			ID:      serpId,
			Context: context.Background(),
		}
		req.SetTimeout(30 * time.Second)
		_, err := edge.ServiceEdgeRouterPolicy.DeleteServiceEdgeRouterPolicy(req, nil)
		if err != nil {
			return err
		}
		logrus.Infof("deleted service edge router policy '%v' for environment '%v'", serpId, envZId)
	} else {
		logrus.Infof("did not find a service edge router policy")
	}
	return nil
}
