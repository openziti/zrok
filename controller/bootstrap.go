package controller

import (
	"bytes"
	"encoding/json"
	"fmt"

	restModelEdge "github.com/openziti/edge-api/rest_model"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/zrok/controller/automation"
	"github.com/openziti/zrok/controller/config"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func Bootstrap(skipFrontend bool, inCfg *config.Config) error {
	cfg = inCfg

	if v, err := store.Open(cfg.Store); err == nil {
		str = v
	} else {
		return errors.Wrap(err, "error opening store")
	}

	logrus.Info("connecting to the ziti edge management api")
	ziti, err := automation.GetZitiAutomation(cfg)
	if err != nil {
		return errors.Wrap(err, "error connecting to the ziti edge management api")
	}

	env, err := environment.LoadRoot()
	if err != nil {
		return err
	}

	var frontendZId string
	if !skipFrontend {
		logrus.Info("creating identity for public frontend access")

		if frontendZId, err = getIdentityId(env.PublicIdentityName()); err == nil {
			logrus.Infof("frontend identity: %v", frontendZId)
		} else {
			frontendZId, err = bootstrapIdentity(env.PublicIdentityName(), ziti)
			if err != nil {
				panic(err)
			}
		}
		if err := assertIdentity(frontendZId, ziti); err != nil {
			panic(err)
		}
		if err := assertErpForIdentity(env.PublicIdentityName(), frontendZId, ziti); err != nil {
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

	if err := assertZrokProxyConfigType(ziti); err != nil {
		return err
	}

	return nil
}

func assertZrokProxyConfigType(auto *automation.ZitiAutomation) error {
	_, err := auto.ConfigTypes.EnsureExists(sdk.ZrokProxyConfig)
	if err != nil {
		return errors.Wrapf(err, "error ensuring '%v' config type exists", sdk.ZrokProxyConfig)
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

func assertIdentity(zId string, auto *automation.ZitiAutomation) error {
	_, err := auto.Identities.GetByID(zId)
	if err != nil {
		return errors.Wrapf(err, "error asserting identity '%v'", zId)
	}
	logrus.Infof("asserted identity '%v'", zId)
	return nil
}

func bootstrapIdentity(name string, auto *automation.ZitiAutomation) (string, error) {
	env, err := environment.LoadRoot()
	if err != nil {
		return "", errors.Wrap(err, "error loading environment root")
	}

	opts := &automation.IdentityOptions{
		BaseOptions: automation.BaseOptions{
			Name: name,
		},
		Type:    restModelEdge.IdentityTypeDevice,
		IsAdmin: false,
	}

	zId, err := auto.Identities.Create(opts)
	if err != nil {
		return "", errors.Wrapf(err, "error creating '%v' identity", name)
	}

	cfg, err := auto.Identities.Enroll(zId)
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
	if err := env.SaveZitiIdentityNamed(name, out.String()); err != nil {
		return "", errors.Wrapf(err, "error saving identity config '%v'", name)
	}
	return zId, nil
}

func assertErpForIdentity(name, zId string, auto *automation.ZitiAutomation) error {
	filter := fmt.Sprintf("name=\"%v\" and tags.zrok != null", name)
	opts := &automation.FilterOptions{Filter: filter}

	erps, err := auto.EdgeRouterPolicies.Find(opts)
	if err != nil {
		return errors.Wrapf(err, "error listing edge router policies for '%v' (%v)", name, zId)
	}

	if len(erps) != 1 {
		logrus.Infof("creating erp for '%v' (%v)", name, zId)

		erpOpts := &automation.EdgeRouterPolicyOptions{
			BaseOptions: automation.BaseOptions{
				Name: name,
				Tags: automation.ZrokTags(),
			},
			EdgeRouterRoles: []string{"#all"},
			IdentityRoles:   []string{fmt.Sprintf("@%v", zId)},
			Semantic:        restModelEdge.SemanticAllOf,
		}

		_, err := auto.EdgeRouterPolicies.Create(erpOpts)
		if err != nil {
			return errors.Wrapf(err, "error creating erp for '%v' (%v)", name, zId)
		}
	}
	logrus.Infof("asserted erps for '%v' (%v)", name, zId)
	return nil
}
