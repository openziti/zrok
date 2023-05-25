package zrokEdgeSdk

import (
	"context"
	"fmt"
	"time"

	"github.com/openziti/edge-api/rest_management_api_client"
	edge_service "github.com/openziti/edge-api/rest_management_api_client/service"
	"github.com/openziti/edge-api/rest_model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func FindShareService(svcZId string, edge *rest_management_api_client.ZitiEdgeManagement) (string, error) {
	filter := fmt.Sprintf("id=\"%v\"", svcZId)
	limit := int64(0)
	offset := int64(0)
	listReq := &edge_service.ListServicesParams{
		Filter:  &filter,
		Limit:   &limit,
		Offset:  &offset,
		Context: context.Background(),
	}
	listReq.SetTimeout(30 * time.Second)
	listResp, err := edge.Service.ListServices(listReq, nil)
	if err != nil {
		return "", errors.Wrapf(err, "error listing service '%v'", svcZId)
	}
	if len(listResp.Payload.Data) == 1 {
		return *listResp.Payload.Data[0].Name, nil
	}
	return "", errors.Errorf("service with ziti id '%v' not found", svcZId)
}

func CreateShareService(envZId, shrToken, cfgZId string, edge *rest_management_api_client.ZitiEdgeManagement) (shrZId string, err error) {
	shrZId, err = CreateService(shrToken, []string{cfgZId}, map[string]interface{}{"zrokShareToken": shrToken}, edge)
	if err != nil {
		return "", errors.Wrapf(err, "error creating share '%v'", shrToken)
	}
	logrus.Infof("created share '%v' (with ziti id '%v') for environment '%v'", shrToken, shrZId, envZId)
	return shrZId, nil
}

func CreateService(name string, cfgZIds []string, addlTags map[string]interface{}, edge *rest_management_api_client.ZitiEdgeManagement) (shrZId string, err error) {
	encryptionRequired := true
	svc := &rest_model.ServiceCreate{
		EncryptionRequired: &encryptionRequired,
		Name:               &name,
		Tags:               MergeTags(ZrokTags(), addlTags),
	}
	if cfgZIds != nil {
		svc.Configs = cfgZIds
	}
	req := &edge_service.CreateServiceParams{
		Service: svc,
		Context: context.Background(),
	}
	req.SetTimeout(30 * time.Second)
	resp, err := edge.Service.CreateService(req, nil)
	if err != nil {
		return "", err
	}
	return resp.Payload.Data.ID, nil
}

func DeleteService(envZId, shrZId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	req := &edge_service.DeleteServiceParams{
		ID:      shrZId,
		Context: context.Background(),
	}
	req.SetTimeout(30 * time.Second)
	_, err := edge.Service.DeleteService(req, nil)
	if err != nil {
		return err
	}
	logrus.Infof("deleted service '%v' for environment '%v'", shrZId, envZId)
	return nil
}
