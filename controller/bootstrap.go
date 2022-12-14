package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/openziti-test-kitchen/zrok/controller/edge_ctrl"
	"github.com/openziti-test-kitchen/zrok/controller/store"
	"github.com/openziti-test-kitchen/zrok/model"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/edge/rest_management_api_client/config"
	"github.com/openziti/edge/rest_management_api_client/edge_router_policy"
	"github.com/openziti/edge/rest_management_api_client/identity"
	"github.com/openziti/edge/rest_management_api_client/service"
	"github.com/openziti/edge/rest_management_api_client/service_edge_router_policy"
	"github.com/openziti/edge/rest_management_api_client/service_policy"
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

	if v, err := store.Open(cfg.Store); err == nil {
		str = v
	} else {
		return errors.Wrap(err, "error opening store")
	}

	logrus.Info("connecting to the ziti edge management api")
	edge, err := edgeClient()
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
			logrus.Warnf("missing public frontend for ziti id '%v'; please use 'zrok admin create frontend %v public https://{svcToken}.your.dns.name' to create a frontend instance", frontendZId, frontendZId)
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

	var metricsSvcZId string
	if metricsSvcZId, err = assertMetricsService(cfg, edge); err != nil {
		return err
	}

	if err := assertMetricsSerp(metricsSvcZId, cfg, edge); err != nil {
		return err
	}

	if !skipCtrl {
		if err := assertCtrlMetricsBind(ctrlZId, metricsSvcZId, edge); err != nil {
			return err
		}
	}

	if !skipFrontend {
		if err := assertFrontendMetricsDial(frontendZId, metricsSvcZId, edge); err != nil {
			return err
		}
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
		return "", errors.Wrapf(err, "error creating '%v' identity", name)
	}

	zId := idc.Payload.Data.ID
	cfg, err := enrollIdentity(zId, edge)
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
		if err := createEdgeRouterPolicy(name, zId, edge); err != nil {
			return errors.Wrapf(err, "error creating erp for '%v' (%v)", name, zId)
		}
	}
	logrus.Infof("asserted erps for '%v' (%v)", name, zId)
	return nil
}

func assertMetricsService(cfg *Config, edge *rest_management_api_client.ZitiEdgeManagement) (string, error) {
	filter := fmt.Sprintf("name=\"%v\" and tags.zrok != null", cfg.Metrics.ServiceName)
	limit := int64(0)
	offset := int64(0)
	listReq := &service.ListServicesParams{
		Filter: &filter,
		Limit:  &limit,
		Offset: &offset,
	}
	listReq.SetTimeout(30 * time.Second)
	listResp, err := edge.Service.ListServices(listReq, nil)
	if err != nil {
		return "", errors.Wrapf(err, "error listing '%v' service", cfg.Metrics.ServiceName)
	}
	var svcZId string
	if len(listResp.Payload.Data) != 1 {
		logrus.Infof("creating '%v' service", cfg.Metrics.ServiceName)
		svcZId, err = edge_ctrl.CreateService("metrics", nil, nil, edge)
		if err != nil {
			return "", errors.Wrapf(err, "error creating '%v' service", cfg.Metrics.ServiceName)
		}
	} else {
		svcZId = *listResp.Payload.Data[0].ID
	}

	logrus.Infof("asserted '%v' service (%v)", cfg.Metrics.ServiceName, svcZId)
	return svcZId, nil
}

func assertMetricsSerp(metricsSvcZId string, cfg *Config, edge *rest_management_api_client.ZitiEdgeManagement) error {
	filter := fmt.Sprintf("allOf(serviceRoles) = \"@%v\" and allOf(edgeRouterRoles) = \"#all\" and tags.zrok != null", metricsSvcZId)
	limit := int64(0)
	offset := int64(0)
	listReq := &service_edge_router_policy.ListServiceEdgeRouterPoliciesParams{
		Filter: &filter,
		Limit:  &limit,
		Offset: &offset,
	}
	listReq.SetTimeout(30 * time.Second)
	listResp, err := edge.ServiceEdgeRouterPolicy.ListServiceEdgeRouterPolicies(listReq, nil)
	if err != nil {
		return errors.Wrapf(err, "error listing '%v' serps", cfg.Metrics.ServiceName)
	}
	if len(listResp.Payload.Data) != 1 {
		logrus.Infof("creating '%v' serp", cfg.Metrics.ServiceName)
		_, err := createServiceEdgeRouterPolicy(cfg.Metrics.ServiceName, metricsSvcZId, nil, edge)
		if err != nil {
			return errors.Wrapf(err, "error creating '%v' serp", cfg.Metrics.ServiceName)
		}
	}
	logrus.Infof("asserted '%v' serp", cfg.Metrics.ServiceName)
	return nil
}

func assertCtrlMetricsBind(ctrlZId, metricsSvcZId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	filter := fmt.Sprintf("allOf(serviceRoles) = \"@%v\" and allOf(identityRoles) = \"@%v\" and type = 2 and tags.zrok != null", metricsSvcZId, ctrlZId)
	limit := int64(0)
	offset := int64(0)
	listReq := &service_policy.ListServicePoliciesParams{
		Filter: &filter,
		Limit:  &limit,
		Offset: &offset,
	}
	listReq.SetTimeout(30 * time.Second)
	listResp, err := edge.ServicePolicy.ListServicePolicies(listReq, nil)
	if err != nil {
		return errors.Wrapf(err, "error listing 'ctrl-metrics-bind' service policy")
	}
	if len(listResp.Payload.Data) != 1 {
		logrus.Info("creating 'ctrl-metrics-bind' service policy")
		if err := createNamedBindServicePolicy("ctrl-metrics-bind", metricsSvcZId, ctrlZId, edge, edge_ctrl.ZrokTags()); err != nil {
			return errors.Wrap(err, "error creating 'ctrl-metrics-bind' service policy")
		}
	}
	logrus.Infof("asserted 'ctrl-metrics-bind' service policy")
	return nil
}

func assertFrontendMetricsDial(frontendZId, metricsSvcZId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	filter := fmt.Sprintf("allOf(serviceRoles) = \"@%v\" and allOf(identityRoles) = \"@%v\" and type = 1 and tags.zrok != null", metricsSvcZId, frontendZId)
	limit := int64(0)
	offset := int64(0)
	listReq := &service_policy.ListServicePoliciesParams{
		Filter: &filter,
		Limit:  &limit,
		Offset: &offset,
	}
	listReq.SetTimeout(30 * time.Second)
	listResp, err := edge.ServicePolicy.ListServicePolicies(listReq, nil)
	if err != nil {
		return errors.Wrapf(err, "error listing 'frontend-metrics-dial' service policy")
	}
	if len(listResp.Payload.Data) != 1 {
		logrus.Info("creating 'frontend-metrics-dial' service policy")
		if err := createNamedDialServicePolicy("frontend-metrics-dial", metricsSvcZId, frontendZId, edge, edge_ctrl.ZrokTags()); err != nil {
			return errors.Wrap(err, "error creating 'frontend-metrics-dial' service policy")
		}
	}
	logrus.Infof("asserted 'frontend-metrics-dial' service policy")
	return nil
}
