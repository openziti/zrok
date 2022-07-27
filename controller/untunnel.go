package controller

import (
	"context"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/tunnel"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/edge/rest_management_api_client/edge_router_policy"
	"github.com/sirupsen/logrus"
	"time"
)

func untunnelHandler(params tunnel.UntunnelParams) middleware.Responder {
	edge, err := edgeClient()
	if err != nil {
		logrus.Error(err)
		return tunnel.NewUntunnelInternalServerError()
	}

	svcId := params.Body.Service
	if err := deleteEdgeRouterPolicy(svcId, edge); err != nil {
		logrus.Error(err)
		return middleware.Error(500, err.Error())
	}

	return tunnel.NewUntunnelOK()
}

func deleteEdgeRouterPolicy(svcId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	filter := fmt.Sprintf("name=\"%v\"", svcId)
	limit := int64(1)
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
		logrus.Infof("found edge router policy '%v'", erpId)
		deleteReq := &edge_router_policy.DeleteEdgeRouterPolicyParams{
			ID:      erpId,
			Context: context.Background(),
		}
		deleteReq.SetTimeout(30 * time.Second)
		_, err := edge.EdgeRouterPolicy.DeleteEdgeRouterPolicy(deleteReq, nil)
		if err != nil {
			return err
		}
		logrus.Infof("deleted edge router policy '%v'", erpId)
	} else {
		logrus.Infof("did not find an edge router policy")
	}
	return nil
}
