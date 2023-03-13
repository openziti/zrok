package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/edge/rest_management_api_client/config"
	"github.com/openziti/edge/rest_management_api_client/edge_router_policy"
	"github.com/openziti/edge/rest_management_api_client/identity"
	"github.com/openziti/edge/rest_model"
	rest_model_edge "github.com/openziti/edge/rest_model"
	"github.com/openziti/sdk-golang/ziti"
	ziti_config "github.com/openziti/sdk-golang/ziti/config"
	zrok_config "github.com/openziti/zrok/controller/config"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/model"
	"github.com/openziti/zrok/zrokdir"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

func Bootstrap(skipCtrl, skipFrontend bool, inCfg *zrok_config.Config) error {
	cfg = inCfg

	if v, err := store.Open(cfg.Store); err == nil {
		str = v
	} else {
		return errors.Wrap(err, "error opening store")
	}

	logrus.Info("connecting to the ziti edge management api")
	edge, err := zrokEdgeSdk.Client(cfg.Ziti)
	if err != nil {
		return errors.Wrap(err, "error connecting to the ziti edge management api")
	}

	var ctrlZId string
	if !skipCtrl {
		logrus.Info("creating identity for controller ziti access")

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
		logrus.Info("creating identity for frontend ziti access")

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

		tx, err := str.Begin()
		if err != nil {
			panic(err)
		}
		defer func() { _ = tx.Rollback() }()
		publicFe, err := str.FindFrontendWithZId(frontendZId, tx)
		if err != nil {
			logrus.Warnf("missing public frontend for ziti id '%v'; please use 'zrok admin create frontend %v public https://{token}.your.dns.name' to create a frontend instance", frontendZId, frontendZId)
		} else {
			if publicFe.PublicName != nil && publicFe.UrlTemplate != nil {
				logrus.Infof("found public frontend entry '%v' (%v) for ziti identity '%v'", *publicFe.PublicName, publicFe.Token, frontendZId)
			} else {
				logrus.Warnf("found frontend entry for ziti identity '%v'; missing either public name or url template", frontendZId)
			}
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
	zcfg, err := ziti_config.NewFromFile(zif)
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
	idc, err := zrokEdgeSdk.CreateIdentity(name, rest_model_edge.IdentityTypeDevice, nil, edge)
	if err != nil {
		return "", errors.Wrapf(err, "error creating '%v' identity", name)
	}

	zId := idc.Payload.Data.ID
	cfg, err := zrokEdgeSdk.EnrollIdentity(zId, edge)
	if err != nil {
		return "", errors.Wrapf(err, "error enrolling '%v' identity", name)
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
		if err := zrokEdgeSdk.CreateEdgeRouterPolicy(name, zId, edge); err != nil {
			return errors.Wrapf(err, "error creating erp for '%v' (%v)", name, zId)
		}
	}
	logrus.Infof("asserted erps for '%v' (%v)", name, zId)
	return nil
}
