package config

import (
	"github.com/openziti/zrok/controller/env"
	"github.com/openziti/zrok/controller/limits"
	"github.com/openziti/zrok/controller/metrics"
	"github.com/openziti/zrok/controller/metrics2"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"time"

	"github.com/michaelquigley/cf"
	"github.com/openziti/zrok/controller/store"
	"github.com/pkg/errors"
)

const ConfigVersion = 2

type Config struct {
	V             int
	Admin         *AdminConfig
	Bridge        *metrics2.BridgeConfig
	Endpoint      *EndpointConfig
	Email         *EmailConfig
	Limits        *limits.Config
	Maintenance   *MaintenanceConfig
	Metrics       *metrics.Config
	Registration  *RegistrationConfig
	ResetPassword *ResetPasswordConfig
	Store         *store.Config
	Ziti          *zrokEdgeSdk.Config
}

type AdminConfig struct {
	Secrets []string `cf:"+secret"`
	TouLink string
}

type EndpointConfig struct {
	Host string
	Port int
}

type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string `cf:"+secret"`
	From     string
}

type RegistrationConfig struct {
	RegistrationUrlTemplate string
	TokenStrategy           string
}

type ResetPasswordConfig struct {
	ResetUrlTemplate string
}

type MaintenanceConfig struct {
	ResetPassword *ResetPasswordMaintenanceConfig
	Registration  *RegistrationMaintenanceConfig
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
	if cfg.V != ConfigVersion {
		return nil, errors.Errorf("expecting configuration version '%v', your configuration is version '%v'; please see zrok.io for changelog and configuration documentation", ConfigVersion, cfg.V)
	}
	return cfg, nil
}
