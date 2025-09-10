package controller

import (
	"bytes"
	"encoding/json"

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
	automationCfg := &automation.Config{
		ApiEndpoint: cfg.Ziti.ApiEndpoint,
		Username:    cfg.Ziti.Username,
		Password:    cfg.Ziti.Password,
	}

	za, err := automation.NewZitiAutomation(automationCfg)
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
			// still need to ensure edge router policy for existing identity
			if err := za.EnsureEdgeRouterPolicyForIdentity(env.PublicIdentityName(), frontendZId); err != nil {
				panic(err)
			}
		} else {
			frontendZId, err = createBootstrapIdentity(env.PublicIdentityName(), za)
			if err != nil {
				panic(err)
			}
		}
		if err := assertIdentity(frontendZId, za); err != nil {
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

	if err := za.EnsureConfigType(sdk.ZrokProxyConfig); err != nil {
		return err
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

func assertIdentity(zId string, za *automation.ZitiAutomation) error {
	identity, err := za.FindIdentityByID(zId)
	if err != nil {
		return errors.Wrapf(err, "error finding identity '%v'", zId)
	}
	if identity == nil {
		return errors.Errorf("identity '%v' not found", zId)
	}
	logrus.Infof("asserted identity '%v'", zId)
	return nil
}

func createBootstrapIdentity(name string, za *automation.ZitiAutomation) (string, error) {
	env, err := environment.LoadRoot()
	if err != nil {
		return "", errors.Wrap(err, "error loading environment root")
	}

	// create identity with complete setup
	zId, cfg, err := za.CreateBootstrapIdentity(name)
	if err != nil {
		return "", errors.Wrapf(err, "error creating bootstrap identity '%v'", name)
	}

	// save the config to the environment for bootstrap identities
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
