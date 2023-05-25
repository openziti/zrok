package zrokEdgeSdk

import (
	"context"
	"fmt"
	"github.com/openziti/edge-api/rest_management_api_client"
	"github.com/openziti/edge-api/rest_management_api_client/edge_router_policy"
	rest_model_edge "github.com/openziti/edge-api/rest_model"
	"github.com/sirupsen/logrus"
	"time"
)

func CreateEdgeRouterPolicy(name, zId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	edgeRouterRoles := []string{"#all"}
	identityRoles := []string{fmt.Sprintf("@%v", zId)}
	semantic := rest_model_edge.SemanticAllOf
	erp := &rest_model_edge.EdgeRouterPolicyCreate{
		EdgeRouterRoles: edgeRouterRoles,
		IdentityRoles:   identityRoles,
		Name:            &name,
		Semantic:        &semantic,
		Tags:            ZrokTags(),
	}
	req := &edge_router_policy.CreateEdgeRouterPolicyParams{
		Policy:  erp,
		Context: context.Background(),
	}
	req.SetTimeout(30 * time.Second)
	resp, err := edge.EdgeRouterPolicy.CreateEdgeRouterPolicy(req, nil)
	if err != nil {
		return err
	}
	logrus.Infof("created edge router policy '%v' for ziti identity '%v'", resp.Payload.Data.ID, zId)
	return nil
}

func DeleteEdgeRouterPolicy(envZId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	filter := fmt.Sprintf("name=\"%v\"", envZId)
	limit := int64(0)
	offset := int64(0)
	listReq := &edge_router_policy.ListEdgeRouterPoliciesParams{
		Filter:  &filter,
		Limit:   &limit,
		Offset:  &offset,
		Context: context.Background(),
	}
	listReq.SetTimeout(30 * time.Second)
	listResp, err := edge.EdgeRouterPolicy.ListEdgeRouterPolicies(listReq, nil)
	if err != nil {
		return err
	}
	if len(listResp.Payload.Data) == 1 {
		erpId := *(listResp.Payload.Data[0].ID)
		req := &edge_router_policy.DeleteEdgeRouterPolicyParams{
			ID:      erpId,
			Context: context.Background(),
		}
		_, err := edge.EdgeRouterPolicy.DeleteEdgeRouterPolicy(req, nil)
		if err != nil {
			return err
		}
		logrus.Infof("deleted edge router policy '%v' for environment '%v'", erpId, envZId)
	} else {
		logrus.Infof("found '%d' edge router policies, expected 1", len(listResp.Payload.Data))
	}
	return nil
}
