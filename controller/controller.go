package controller

import (
	"context"
	"github.com/go-openapi/loads"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/jessevdk/go-flags"
	"github.com/openziti/zrok/controller/config"
	"github.com/openziti/zrok/controller/limits"
	"github.com/openziti/zrok/controller/metrics"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/rest_server_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations"
	"github.com/openziti/zrok/rest_server_zrok/operations/account"
	"github.com/openziti/zrok/rest_server_zrok/operations/metadata"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	_ "net/http/pprof"
)

var (
	cfg         *config.Config
	str         *store.Store
	idb         influxdb2.Client
	limitsAgent *limits.Agent
)

func Run(inCfg *config.Config) error {
	cfg = inCfg

	if cfg.Admin != nil && cfg.Admin.ProfileEndpoint != "" {
		go func() {
			log.Println(http.ListenAndServe(cfg.Admin.ProfileEndpoint, nil))
		}()
	}

	swaggerSpec, err := loads.Embedded(rest_server_zrok.SwaggerJSON, rest_server_zrok.FlatSwaggerJSON)
	if err != nil {
		return errors.Wrap(err, "error loading embedded swagger spec")
	}

	api := operations.NewZrokAPI(swaggerSpec)
	api.KeyAuth = newZrokAuthenticator(cfg).authenticate
	api.AccountChangePasswordHandler = newChangePasswordHandler(cfg)
	api.AccountInviteHandler = newInviteHandler(cfg)
	api.AccountLoginHandler = account.LoginHandlerFunc(loginHandler)
	api.AccountRegenerateAccountTokenHandler = newRegenerateAccountTokenHandler()
	api.AccountRegisterHandler = newRegisterHandler(cfg)
	api.AccountResetPasswordHandler = newResetPasswordHandler(cfg)
	api.AccountResetPasswordRequestHandler = newResetPasswordRequestHandler()
	api.AccountVerifyHandler = newVerifyHandler()
	api.AdminAddOrganizationMemberHandler = newAddOrganizationMemberHandler()
	api.AdminCreateAccountHandler = newCreateAccountHandler()
	api.AdminCreateFrontendHandler = newCreateFrontendHandler()
	api.AdminCreateIdentityHandler = newCreateIdentityHandler()
	api.AdminCreateOrganizationHandler = newCreateOrganizationHandler()
	api.AdminDeleteFrontendHandler = newDeleteFrontendHandler()
	api.AdminDeleteOrganizationHandler = newDeleteOrganizationHandler()
	api.AdminGrantsHandler = newGrantsHandler()
	api.AdminInviteTokenGenerateHandler = newInviteTokenGenerateHandler()
	api.AdminListFrontendsHandler = newListFrontendsHandler()
	api.AdminListOrganizationMembersHandler = newListOrganizationMembersHandler()
	api.AdminListOrganizationsHandler = newListOrganizationsHandler()
	api.AdminRemoveOrganizationMemberHandler = newRemoveOrganizationMemberHandler()
	api.AdminUpdateFrontendHandler = newUpdateFrontendHandler()
	api.EnvironmentEnableHandler = newEnableHandler()
	api.EnvironmentDisableHandler = newDisableHandler()
	api.MetadataConfigurationHandler = newConfigurationHandler(cfg)
	api.MetadataClientVersionCheckHandler = metadata.ClientVersionCheckHandlerFunc(clientVersionCheckHandler)
	api.MetadataGetAccountDetailHandler = newAccountDetailHandler()
	api.MetadataGetSparklinesHandler = newSparklinesHandler()
	if cfg.Metrics != nil && cfg.Metrics.Influx != nil {
		api.MetadataGetAccountMetricsHandler = newGetAccountMetricsHandler(cfg.Metrics.Influx)
		api.MetadataGetEnvironmentMetricsHandler = newGetEnvironmentMetricsHandler(cfg.Metrics.Influx)
		api.MetadataGetShareMetricsHandler = newGetShareMetricsHandler(cfg.Metrics.Influx)
	}
	api.MetadataGetEnvironmentDetailHandler = newEnvironmentDetailHandler()
	api.MetadataGetFrontendDetailHandler = newGetFrontendDetailHandler()
	api.MetadataGetShareDetailHandler = newShareDetailHandler()
	api.MetadataListMembershipsHandler = newListMembershipsHandler()
	api.MetadataListOrgMembersHandler = newListOrgMembersHandler()
	api.MetadataOrgAccountOverviewHandler = newOrgAccountOverviewHandler()
	api.MetadataOverviewHandler = newOverviewHandler()
	api.MetadataVersionHandler = metadata.VersionHandlerFunc(versionHandler)
	api.MetadataVersionInventoryHandler = metadata.VersionInventoryHandlerFunc(versionInventoryHandler)
	api.ShareAccessHandler = newAccessHandler()
	api.ShareShareHandler = newShareHandler()
	api.ShareUnaccessHandler = newUnaccessHandler()
	api.ShareUnshareHandler = newUnshareHandler()
	api.ShareUpdateAccessHandler = newUpdateAccessHandler()
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

	if cfg.Metrics != nil && cfg.Metrics.Agent != nil && cfg.Metrics.Influx != nil {
		ma, err := metrics.NewAgent(cfg.Metrics.Agent, str, cfg.Metrics.Influx)
		if err != nil {
			return errors.Wrap(err, "error creating metrics agent")
		}
		if err := ma.Start(); err != nil {
			return errors.Wrap(err, "error starting metrics agent")
		}
		defer func() { ma.Stop() }()

		if cfg.Limits != nil && cfg.Limits.Enforcing {
			limitsAgent, err = limits.NewAgent(cfg.Limits, cfg.Metrics.Influx, cfg.Ziti, cfg.Email, str)
			if err != nil {
				return errors.Wrap(err, "error creating limits agent")
			}
			ma.AddUsageSink(limitsAgent)
			limitsAgent.Start()
			defer func() { limitsAgent.Stop() }()
		}
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
	if cfg.Tls != nil {
		server.TLSHost = cfg.Endpoint.Host
		server.TLSPort = cfg.Endpoint.Port
		server.TLSCertificate = flags.Filename(cfg.Tls.CertPath)
		server.TLSCertificateKey = flags.Filename(cfg.Tls.KeyPath)
		server.EnabledListeners = []string{"https"}
	} else {
		server.Host = cfg.Endpoint.Host
		server.Port = cfg.Endpoint.Port
	}
	rest_server_zrok.HealthCheck = HealthCheckHTTP
	server.ConfigureAPI()
	if err := server.Serve(); err != nil {
		return errors.Wrap(err, "api server error")
	}

	return nil
}
