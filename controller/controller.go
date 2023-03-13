package controller

import (
	"context"
	"github.com/openziti/zrok/controller/config"
	"github.com/sirupsen/logrus"

	"github.com/go-openapi/loads"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/rest_server_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations"
	"github.com/openziti/zrok/rest_server_zrok/operations/account"
	"github.com/openziti/zrok/rest_server_zrok/operations/metadata"
	"github.com/pkg/errors"
)

var cfg *config.Config
var str *store.Store
var idb influxdb2.Client

func Run(inCfg *config.Config) error {
	cfg = inCfg

	swaggerSpec, err := loads.Embedded(rest_server_zrok.SwaggerJSON, rest_server_zrok.FlatSwaggerJSON)
	if err != nil {
		return errors.Wrap(err, "error loading embedded swagger spec")
	}

	api := operations.NewZrokAPI(swaggerSpec)
	api.KeyAuth = newZrokAuthenticator(cfg).authenticate
	api.AccountInviteHandler = newInviteHandler(cfg)
	api.AccountLoginHandler = account.LoginHandlerFunc(loginHandler)
	api.AccountRegisterHandler = newRegisterHandler()
	api.AccountResetPasswordHandler = newResetPasswordHandler()
	api.AccountResetPasswordRequestHandler = newResetPasswordRequestHandler()
	api.AccountVerifyHandler = newVerifyHandler()
	api.AdminCreateFrontendHandler = newCreateFrontendHandler()
	api.AdminCreateIdentityHandler = newCreateIdentityHandler()
	api.AdminDeleteFrontendHandler = newDeleteFrontendHandler()
	api.AdminInviteTokenGenerateHandler = newInviteTokenGenerateHandler()
	api.AdminListFrontendsHandler = newListFrontendsHandler()
	api.AdminUpdateFrontendHandler = newUpdateFrontendHandler()
	api.EnvironmentEnableHandler = newEnableHandler(cfg.Limits)
	api.EnvironmentDisableHandler = newDisableHandler()
	api.MetadataConfigurationHandler = newConfigurationHandler(cfg)
	api.MetadataGetEnvironmentDetailHandler = newEnvironmentDetailHandler()
	api.MetadataGetShareDetailHandler = newShareDetailHandler()
	api.MetadataOverviewHandler = metadata.OverviewHandlerFunc(overviewHandler)
	api.MetadataVersionHandler = metadata.VersionHandlerFunc(versionHandler)
	api.ShareAccessHandler = newAccessHandler()
	api.ShareShareHandler = newShareHandler(cfg.Limits)
	api.ShareUnaccessHandler = newUnaccessHandler()
	api.ShareUnshareHandler = newUnshareHandler()
	api.ShareUpdateShareHandler = newUpdateShareHandler()

	if err := controllerStartup(); err != nil {
		return err
	}

	if v, err := store.Open(cfg.Store); err == nil {
		str = v
	} else {
		return errors.Wrap(err, "error opening store")
	}

	if cfg.Metrics != nil && cfg.Metrics.Influx != nil {
		idb = influxdb2.NewClient(cfg.Metrics.Influx.Url, cfg.Metrics.Influx.Token)
	} else {
		logrus.Warn("skipping influx client; no configuration")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()

	if cfg.Maintenance != nil {
		if cfg.Maintenance.Registration != nil {
			go newRegistrationMaintenanceAgent(ctx, cfg.Maintenance.Registration).run()
		}
		if cfg.Maintenance.ResetPassword != nil {
			go newMaintenanceResetPasswordAgent(ctx, cfg.Maintenance.ResetPassword).run()
		}
	}

	server := rest_server_zrok.NewServer(api)
	defer func() { _ = server.Shutdown() }()
	server.Host = cfg.Endpoint.Host
	server.Port = cfg.Endpoint.Port
	server.ConfigureAPI()
	if err := server.Serve(); err != nil {
		return errors.Wrap(err, "api server error")
	}

	return nil
}
