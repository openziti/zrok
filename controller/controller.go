package controller

import (
	"context"

	"github.com/go-openapi/loads"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/openziti-test-kitchen/zrok/controller/store"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/account"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/metadata"
	"github.com/pkg/errors"
)

var cfg *Config
var str *store.Store
var mtr *metricsAgent
var idb influxdb2.Client

func Run(inCfg *Config) error {
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
	api.AccountVerifyHandler = newVerifyHandler()
	api.AdminCreateFrontendHandler = newCreateFrontendHandler()
	api.AdminCreateIdentityHandler = newCreateIdentityHandler()
	api.AdminDeleteFrontendHandler = newDeleteFrontendHandler()
	api.AdminInviteTokenGenerateHandler = newInviteTokenGenerateHandler()
	api.AdminListFrontendsHandler = newListFrontendsHandler()
	api.AdminUpdateFrontendHandler = newUpdateFrontendHandler()
	api.EnvironmentEnableHandler = newEnableHandler()
	api.EnvironmentDisableHandler = newDisableHandler()
	api.MetadataGetEnvironmentDetailHandler = newEnvironmentDetailHandler()
	api.MetadataGetShareDetailHandler = newShareDetailHandler()
	api.MetadataOverviewHandler = metadata.OverviewHandlerFunc(overviewHandler)
	api.MetadataVersionHandler = metadata.VersionHandlerFunc(versionHandler)
	api.ShareAccessHandler = newAccessHandler()
	api.ShareShareHandler = newShareHandler()
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

	if cfg.Influx != nil {
		idb = influxdb2.NewClient(cfg.Influx.Url, cfg.Influx.Token)
	}

	if cfg.Metrics != nil {
		mtr = newMetricsAgent()
		go mtr.run()
		defer func() {
			mtr.stop()
			mtr.join()
		}()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()

	if cfg.Maintenance != nil && cfg.Maintenance.Registration != nil {
		go newMaintenanceAgent(ctx, cfg.Maintenance).run()
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
