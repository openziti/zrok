package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/openziti/edge-api/rest_management_api_client"
	restMgmtEdgeConfig "github.com/openziti/edge-api/rest_management_api_client/config"
	"github.com/openziti/edge-api/rest_management_api_client/edge_router_policy"
	"github.com/openziti/edge-api/rest_management_api_client/identity"
	"github.com/openziti/edge-api/rest_management_api_client/service"
	"github.com/openziti/edge-api/rest_management_api_client/service_policy"
	"github.com/openziti/edge-api/rest_model"
	restModelEdge "github.com/openziti/edge-api/rest_model"
	"github.com/openziti/edge-api/rest_util"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/zrok/controller/config"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type BootstrapConfig struct {
	SkipFrontend        bool
	SkipSecretsListener bool
}

func Bootstrap(bootCfg *BootstrapConfig, ctrlCfg *config.Config) error {
	if v, err := store.Open(ctrlCfg.Store); err == nil {
		str = v
	} else {
		return errors.Wrap(err, "error opening store")
	}

	logrus.Info("connecting to the ziti edge management api")
	edge, err := zrokEdgeSdk.Client(ctrlCfg.Ziti)
	if err != nil {
		return errors.Wrap(err, "error connecting to the ziti edge management api")
	}

	env, err := environment.LoadRoot()
	if err != nil {
		return err
	}

	if err := assertFrontendIdentity(bootCfg, env, edge); err != nil {
		return err
	}

	if err := assertSecretsListener(bootCfg, ctrlCfg, env, edge); err != nil {
		return err
	}

	if err := assertZrokProxyConfigType(edge); err != nil {
		return err
	}

	return nil
}

func assertFrontendIdentity(cfg *BootstrapConfig, env env_core.Root, edge *rest_management_api_client.ZitiEdgeManagement) error {
	var frontendZId string
	var err error
	if !cfg.SkipFrontend {
		logrus.Info("bootstrapping identity for public frontend access")

		if frontendZId, err = getIdentityId(env.PublicIdentityName()); err == nil {
			logrus.Infof("existing frontend identity: %v", frontendZId)
		} else {
			frontendZId, err = bootstrapIdentity(env.PublicIdentityName(), edge)
			if err != nil {
				panic(err)
			}
			logrus.Infof("created frontend identity (%v) '%v'", env.PublicIdentityName(), frontendZId)
		}
		if err := assertIdentity(frontendZId, edge); err != nil {
			panic(err)
		}
		if err := assertErpForIdentity(env.PublicIdentityName(), frontendZId, edge); err != nil {
			panic(err)
		}

		trx, err := str.Begin()
		if err != nil {
			panic(err)
		}
		defer trx.Rollback()
		publicFe, err := str.FindFrontendWithZId(frontendZId, trx)
		if err != nil {
			logrus.Error(err)
			logrus.Warnf("missing public frontend for ziti id '%v'; please use 'zrok admin create frontend %v public https://{token}.your.dns.name' to create a frontend instance", frontendZId, frontendZId)
		} else {
			if publicFe.PublicName != nil && publicFe.UrlTemplate != nil {
				logrus.Infof("found public frontend entry '%v' (%v) for ziti identity '%v'", *publicFe.PublicName, publicFe.Token, frontendZId)
			} else {
				logrus.Warnf("found frontend entry for ziti identity '%v'; missing either public name or url template", frontendZId)
			}
		}
	} else {
		logrus.Warnf("skipping frontend identity bootstrap")
	}
	return nil
}

func assertSecretsListener(bCfg *BootstrapConfig, ctrlCfg *config.Config, env env_core.Root, edge *rest_management_api_client.ZitiEdgeManagement) error {
	if !bCfg.SkipSecretsListener || ctrlCfg == nil || ctrlCfg.Secrets == nil {
		logrus.Info("bootstrapping secrets listener")

		if ctrlCfg.Secrets.ServiceName == "" {
			return errors.New("no secrets service name provided")
		}

		var secretsZId string
		var err error
		if ctrlCfg.Secrets.IdentityPath == "" || ctrlCfg.Secrets.ZId == "" {
			logrus.Warnf("no secrets identity path or ziti id provided; allocating a new identity")

			secretsZId, err = bootstrapIdentity("secretsListener", edge)
			if err != nil {
				return errors.Wrap(err, "error bootstrapping secrets identity")
			}
			logrus.Infof("created secrets identity '%v' (configure this into the 'secrets > z_id' field in the controller config)", secretsZId)

		} else {
			logrus.Infof("asserting existing secrets identity '%v'", ctrlCfg.Secrets.ZId)

			if err := assertIdentity(ctrlCfg.Secrets.ZId, edge); err != nil {
				return errors.Wrapf(err, "error asserting existing secrets identity '%v'", ctrlCfg.Secrets.ZId)
			}
			secretsZId = ctrlCfg.Secrets.ZId
			logrus.Infof("asserted secrets identity '%v'", ctrlCfg.Secrets.ZId)
		}

		if err := assertErpForIdentity("secretsListener", secretsZId, edge); err != nil {
			return errors.Wrapf(err, "error asserting erp for secrets identity (secrets) '%v'", secretsZId)
		}

		if err := assertSecretsService(ctrlCfg, edge); err != nil {
			return errors.Wrapf(err, "error asserting secrets service '%v'", ctrlCfg.Secrets.ServiceName)
		}

		if err := assertBindPolicyForIdentityAndService(ctrlCfg.Secrets.ServiceName, secretsZId, edge); err != nil {
			return errors.Wrapf(err, "error asserting bind policy for '@%v' -> '@%v'", ctrlCfg.Secrets.ServiceName, secretsZId)
		}

	} else {
		logrus.Warnf("skipping secrets listener bootstrap")
	}
	return nil
}

func assertZrokProxyConfigType(edge *rest_management_api_client.ZitiEdgeManagement) error {
	filter := fmt.Sprintf("name=\"%v\"", sdk.ZrokProxyConfig)
	limit := int64(100)
	offset := int64(0)
	listReq := &restMgmtEdgeConfig.ListConfigTypesParams{
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
		name := sdk.ZrokProxyConfig
		ct := &restModelEdge.ConfigTypeCreate{Name: &name}
		createReq := &restMgmtEdgeConfig.CreateConfigTypeParams{ConfigType: ct}
		createReq.SetTimeout(30 * time.Second)
		createResp, err := edge.Config.CreateConfigType(createReq, nil)
		if err != nil {
			return err
		}
		logrus.Infof("created '%v' config type with id '%v'", sdk.ZrokProxyConfig, createResp.Payload.Data.ID)
	} else if len(listResp.Payload.Data) > 1 {
		return errors.Errorf("found %d '%v' config types; expected 0 or 1", len(listResp.Payload.Data), sdk.ZrokProxyConfig)
	} else {
		logrus.Infof("found '%v' config type with id '%v'", sdk.ZrokProxyConfig, *(listResp.Payload.Data[0].ID))
	}
	return nil
}

func getIdentityId(identityName string) (string, error) {
	env, err := environment.LoadRoot()
	if err != nil {
		return "", errors.Wrap(err, "error opening environment root")
	}
	zif, err := env.ZitiIdentityNamed(identityName)
	if err != nil {
		return "", errors.Wrapf(err, "error opening identity '%v' from environment", identityName)
	}
	zcfg, err := ziti.NewConfigFromFile(zif)
	if err != nil {
		return "", errors.Wrapf(err, "error loading ziti config from file '%v'", zif)
	}
	zctx, err := ziti.NewContext(zcfg)
	if err != nil {
		return "", errors.Wrap(err, "error loading ziti context")
	}
	id, err := zctx.GetCurrentIdentity()
	if err != nil {
		return "", errors.Wrapf(err, "error getting current identity from '%v'", zif)
	}
	if id.ID != nil {
		return *id.ID, nil
	}
	return "", nil
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
	env, err := environment.LoadRoot()
	if err != nil {
		return "", errors.Wrap(err, "error loading environment root")
	}

	idc, err := zrokEdgeSdk.CreateIdentity(name, restModelEdge.IdentityTypeDevice, nil, edge)
	if err != nil {
		return "", errors.Wrapf(rest_util.WrapErr(err), "error creating '%v' identity", name)
	}

	zId := idc.Payload.Data.ID
	cfg, err := zrokEdgeSdk.EnrollIdentity(zId, edge)
	if err != nil {
		return "", errors.Wrapf(rest_util.WrapErr(err), "error enrolling '%v' identity", name)
	}

	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	err = enc.Encode(&cfg)
	if err != nil {
		return "", errors.Wrapf(err, "error encoding identity config '%v'", name)
	}
	if err := env.SaveZitiIdentityNamed(name, out.String()); err != nil {
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

func assertSecretsService(ctrlCfg *config.Config, edge *rest_management_api_client.ZitiEdgeManagement) error {
	filter := fmt.Sprintf("name=\"%v\"", ctrlCfg.Secrets.ServiceName)
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
		return errors.Wrapf(err, "error listing services for '%v'", ctrlCfg.Secrets.ServiceName)
	}
	if len(listResp.Payload.Data) == 1 {
		logrus.Infof("found service '%v' with zId '%v'", ctrlCfg.Secrets.ServiceName, listResp.Payload.Data[0].ID)

	} else if len(listResp.Payload.Data) < 1 {
		serviceZId, err := zrokEdgeSdk.CreateService(ctrlCfg.Secrets.ServiceName, nil, nil, edge)
		if err != nil {
			return errors.Wrapf(err, "error creating service '%v'", ctrlCfg.Secrets.ServiceName)
		}
		logrus.Infof("created service '%v' with zId '%v'", ctrlCfg.Secrets.ServiceName, serviceZId)

	} else if len(listResp.Payload.Data) > 1 {
		return errors.Errorf("found %d services for '%v'", len(listResp.Payload.Data), ctrlCfg.Secrets.ServiceName)
	}
	return nil
}

func assertBindPolicyForIdentityAndService(serviceName, zId string, edge *rest_management_api_client.ZitiEdgeManagement) error {
	serviceZId, err := getZIdForService(serviceName, edge)
	if err != nil {
		return errors.Wrapf(err, "error getting ziti id for service '%v'", serviceName)
	}
	spZId, err := getZIdForServicePolicy(serviceZId, zId, rest_model.DialBindBind, edge)
	if err != nil {
		logrus.Warnf("no existing bind policy for '@%v' -> '@%v'; creating one", zId, serviceName)
		if err := zrokEdgeSdk.CreateServicePolicyBind(fmt.Sprintf("service-listener-bind-%v", zId), serviceZId, zId, nil, edge); err != nil {
			return errors.Wrapf(err, "error creating bind policy for '%v' -> '%v'", zId, serviceName)
		}
		logrus.Infof("created bind policy for '@%v' -> '@%v' with zId '%v'", zId, serviceName, spZId)
	} else {
		logrus.Infof("found existing bind policy for '@%v' -> '@%v' with zId '%v'", zId, serviceName, spZId)
	}
	return nil
}

func getZIdForServicePolicy(serviceZId, zId string, dialBind rest_model.DialBind, edge *rest_management_api_client.ZitiEdgeManagement) (string, error) {
	dialBindValue := 1
	if dialBind == "Bind" {
		dialBindValue = 2
	}
	filter := fmt.Sprintf("allOf(serviceRoles)=\"@%v\" and allOf(identityRoles)=\"@%v\" and type=%d", serviceZId, zId, dialBindValue)
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
		return "", errors.Wrapf(err, "error listing service policies for '%v' (%v) and '%v'", serviceZId, zId, dialBind)
	}
	if len(listResp.Payload.Data) != 1 {
		return "", errors.Errorf("found %d service policies for '%v' (%v) and '%v'", len(listResp.Payload.Data), serviceZId, zId, dialBind)
	}
	return *listResp.Payload.Data[0].ID, nil
}

func getZIdForService(serviceName string, edge *rest_management_api_client.ZitiEdgeManagement) (string, error) {
	filter := fmt.Sprintf("name=\"%v\"", serviceName)
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
		return "", errors.Wrapf(err, "error listing services for '%v'", serviceName)
	}
	if len(listResp.Payload.Data) != 1 {
		return "", errors.Errorf("found %d services for '%v'", len(listResp.Payload.Data), serviceName)
	}
	return *listResp.Payload.Data[0].ID, nil
}
