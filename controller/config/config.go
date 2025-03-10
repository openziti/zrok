package config

import (
	"os"
	"strconv"
	"time"

	"github.com/openziti/zrok/controller/emailUi"
	"github.com/openziti/zrok/controller/env"
	"github.com/openziti/zrok/controller/limits"
	"github.com/openziti/zrok/controller/metrics"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"

	"github.com/michaelquigley/cf"
	"github.com/openziti/zrok/controller/store"
	"github.com/pkg/errors"
)

const ConfigVersion = 4

type Config struct {
	V             int
	Admin         *AdminConfig
	Bridge        *metrics.BridgeConfig
	Endpoint      *EndpointConfig
	Email         *emailUi.Config
	Invites       *InvitesConfig
	Limits        *limits.Config
	Maintenance   *MaintenanceConfig
	Metrics       *metrics.Config
	Registration  *RegistrationConfig
	ResetPassword *ResetPasswordConfig
	Store         *store.Config
	Ziti          *zrokEdgeSdk.Config
	Tls           *TlsConfig
}

type AdminConfig struct {
	Secrets         []string `cf:"+secret"`
	TouLink         string
	ProfileEndpoint string
}

type EndpointConfig struct {
	Host string
	Port int
}

type InvitesConfig struct {
	InvitesOpen   bool
	TokenStrategy string
	TokenContact  string
}

type MaintenanceConfig struct {
	ResetPassword *ResetPasswordMaintenanceConfig
	Registration  *RegistrationMaintenanceConfig
}

type RegistrationConfig struct {
	RegistrationUrlTemplate string
}

type ResetPasswordConfig struct {
	ResetUrlTemplate string
}

type RegistrationMaintenanceConfig struct {
	ExpirationTimeout time.Duration
	CheckFrequency    time.Duration
	BatchLimit        int
}

type ResetPasswordMaintenanceConfig struct {
	ExpirationTimeout time.Duration
	CheckFrequency    time.Duration
	BatchLimit        int
}

type TlsConfig struct {
	CertPath string
	KeyPath  string
}

func DefaultConfig() *Config {
	return &Config{
		Limits: limits.DefaultConfig(),
		Maintenance: &MaintenanceConfig{
			ResetPassword: &ResetPasswordMaintenanceConfig{
				ExpirationTimeout: time.Minute * 15,
				CheckFrequency:    time.Minute * 15,
				BatchLimit:        500,
			},
			Registration: &RegistrationMaintenanceConfig{
				ExpirationTimeout: time.Hour * 24,
				CheckFrequency:    time.Hour,
				BatchLimit:        500,
			},
		},
	}
}

func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()
	if err := cf.BindYaml(cfg, path, env.GetCfOptions()); err != nil {
		return nil, errors.Wrapf(err, "error loading controller config '%v'", path)
	}
	if !envVersionOk() && cfg.V != ConfigVersion {
		return nil, errors.Errorf("expecting configuration version '%v', your configuration is version '%v'; please see zrok.io for changelog and configuration documentation", ConfigVersion, cfg.V)
	}
	return cfg, nil
}

func envVersionOk() bool {
	vStr := os.Getenv("ZROK_CTRL_CONFIG_VERSION")
	if vStr != "" {
		envV, err := strconv.Atoi(vStr)
		if err != nil {
			return false
		}
		if envV == ConfigVersion {
			return true
		}
	}
	return false
}
