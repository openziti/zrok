package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/openziti-test-kitchen/zrok/model"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/edge/rest_management_api_client/config"
	"github.com/openziti/edge/rest_management_api_client/edge_router_policy"
	"github.com/openziti/edge/rest_management_api_client/identity"
	"github.com/openziti/edge/rest_model"
	rest_model_edge "github.com/openziti/edge/rest_model"
	"github.com/openziti/sdk-golang/ziti"
	config2 "github.com/openziti/sdk-golang/ziti/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

func Bootstrap(skipCtrl, skipFrontend bool, inCfg *Config) error {
	cfg = inCfg

	edge, err := edgeClient()
	if err != nil {
		return err
	}

	var ctrlZId string
	if !skipCtrl {
		if ctrlZId, err = getIdentityId("ctrl"); err == nil {
			logrus.Infof("controller identity: %v", ctrlZId)
		} else {
			ctrlZId, err = bootstrapIdentity("ctrl", edge)
			if err != nil {
				panic(err)
			}
		}
		if err := assertIdentity(ctrlZId, edge); err != nil {
			panic(err)
		}
		if err := assertErpForIdentity("ctrl", ctrlZId, edge); err != nil {
			panic(err)
		}
	}

	var frontendZId string
	if !skipFrontend {
		if frontendZId, err = getIdentityId("frontend"); err == nil {
			logrus.Infof("frontend identity: %v", frontendZId)
		} else {
			frontendZId, err = bootstrapIdentity("frontend", edge)
			if err != nil {
				panic(err)
			}
		}
		if err := assertIdentity(frontendZId, edge); err != nil {
			panic(err)
		}
		if err := assertErpForIdentity("frontend", frontendZId, edge); err != nil {
			panic(err)
		}
	}

	if err := assertZrokProxyConfigType(edge); err != nil {
		return err
	}

	return nil
}

func assertZrokProxyConfigType(edge *rest_management_api_client.ZitiEdgeManagement) error {
	filter := fmt.Sprintf("name=\"%v\"", model.ZrokProxyConfig)
	limit := int64(100)
	offset := int64(0)
	listReq := &config.ListConfigTypesParams{
		Filter:  &filter,
		Limit:   &limit,
		Offset:  &offset,
		Context: context.Background(),
	}
	listReq.SetTimeout(30 * time.Second)
	listResp, err := edge.Config.ListConfigTypes(listReq, nil)
	if err != nil {
		return err
	}
	if len(listResp.Payload.Data) < 1 {
		name := model.ZrokProxyConfig
		ct := &rest_model.ConfigTypeCreate{Name: &name}
		createReq := &config.CreateConfigTypeParams{ConfigType: ct}
		createReq.SetTimeout(30 * time.Second)
		createResp, err := edge.Config.CreateConfigType(createReq, nil)
		if err != nil {
			return err
		}
		logrus.Infof("created '%v' config type with id '%v'", model.ZrokProxyConfig, createResp.Payload.Data.ID)
	} else if len(listResp.Payload.Data) > 1 {
		return errors.Errorf("found %d '%v' config types; expected 0 or 1", len(listResp.Payload.Data), model.ZrokProxyConfig)
	} else {
		logrus.Infof("found '%v' config type with id '%v'", model.ZrokProxyConfig, *(listResp.Payload.Data[0].ID))
	}
	return nil
}

func getIdentityId(identityName string) (string, error) {
	zif, err := zrokdir.ZitiIdentityFile(identityName)
	if err != nil {
		return "", errors.Wrapf(err, "error opening identity '%v' from zrokdir", identityName)
	}
	zcfg, err := config2.NewFromFile(zif)
	if err != nil {
		return "", errors.Wrapf(err, "error loading ziti config from file '%v'", zif)
	}
	zctx := ziti.NewContextWithConfig(zcfg)
	id, err := zctx.GetCurrentIdentity()
	if err != nil {
		return "", errors.Wrapf(err, "error getting current identity from '%v'", zif)
	}
	return id.Id, nil
}

func assertIdentity(zId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	filter := fmt.Sprintf("id=\"%v\"", zId)
	limit := int64(0)
	offset := int64(0)
	listReq := &identity.ListIdentitiesParams{
		Filter: &filter,
		Limit:  &limit,
		Offset: &offset,
	}
	listReq.SetTimeout(30 * time.Second)
	listResp, err := edge.Identity.ListIdentities(listReq, nil)
	if err != nil {
		return errors.Wrapf(err, "error listing identities for '%v'", zId)
	}
	if len(listResp.Payload.Data) != 1 {
		return errors.Wrapf(err, "found %d identities for '%v'", len(listResp.Payload.Data), zId)
	}
	logrus.Infof("asserted identity '%v'", zId)
	return nil
}

func bootstrapIdentity(name string, edge *rest_management_api_client.ZitiEdgeManagement) (string, error) {
	idc, err := createIdentity(name, rest_model_edge.IdentityTypeDevice, nil, edge)
	if err != nil {
		return "", errors.Wrap(err, "error creating 'ctrl' identity")
	}

	zId := idc.Payload.Data.ID
	cfg, err := enrollIdentity(zId, edge)
	if err != nil {
		return "", errors.Wrap(err, "error enrolling 'ctrl' identity")
	}

	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	err = enc.Encode(&cfg)
	if err != nil {
		return "", errors.Wrapf(err, "error encoding identity config '%v'", name)
	}
	if err := zrokdir.SaveZitiIdentity(name, out.String()); err != nil {
		return "", errors.Wrapf(err, "error saving identity config '%v'", name)
	}
	return zId, nil
}

func assertErpForIdentity(name, zId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	logrus.Infof("asserting erps for '%v'", name)
	filter := fmt.Sprintf("name=\"%v\" and tags.zrok != null", name)
	limit := int64(0)
	offset := int64(0)
	listReq := &edge_router_policy.ListEdgeRouterPoliciesParams{
		Filter: &filter,
		Limit:  &limit,
		Offset: &offset,
	}
	listReq.SetTimeout(30 * time.Second)
	listResp, err := edge.EdgeRouterPolicy.ListEdgeRouterPolicies(listReq, nil)
	if err != nil {
		return errors.Wrapf(err, "error listing edge router policies for '%v' (%v)", name, zId)
	}
	if len(listResp.Payload.Data) != 1 {
		logrus.Infof("creating erp for '%v' (%v)", name, zId)
		if err := createEdgeRouterPolicy(name, zId, edge); err != nil {
			return errors.Wrapf(err, "error creating erp for '%v' (%v)", name, zId)
		}
	}
	logrus.Infof("asserted erps for '%v' (%v)", name, zId)
	return nil
}
