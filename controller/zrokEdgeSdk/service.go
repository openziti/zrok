package zrokEdgeSdk

import (
	"context"
	"github.com/openziti/edge/rest_management_api_client"
	edge_service "github.com/openziti/edge/rest_management_api_client/service"
	"github.com/openziti/edge/rest_model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

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
