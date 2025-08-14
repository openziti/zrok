package config

import (
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/michaelquigley/cf"
	"github.com/openziti/zrok/controller/agentController"
	"github.com/openziti/zrok/controller/emailUi"
	"github.com/openziti/zrok/controller/env"
	"github.com/openziti/zrok/controller/limits"
	"github.com/openziti/zrok/controller/metrics"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/pkg/errors"
)

const ConfigVersion = 4

type Config struct {
	V               int
	Admin           *AdminConfig
	AgentController *agentController.Config
	Bridge          *metrics.BridgeConfig
	Compatibility   *CompatibilityConfig
	Endpoint        *EndpointConfig
	Email           *emailUi.Config
	Invites         *InvitesConfig
	Limits          *limits.Config
	Maintenance     *MaintenanceConfig
	Metrics         *metrics.Config
	Registration    *RegistrationConfig
	ResetPassword   *ResetPasswordConfig
	Store           *store.Config
	Ziti            *zrokEdgeSdk.Config
	Tls             *TlsConfig
}

type AdminConfig struct {
	Secrets         []string `cf:"+secret"`
	TouLink         string
	NewAccountLink  string
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

type CompatibilityConfig struct {
	LogRequests      bool
	VersionPatterns  []string
	compiledPatterns []*regexp.Regexp
}

func DefaultConfig() *Config {
	cfg := &Config{
		Compatibility: &CompatibilityConfig{
			VersionPatterns: []string{
				`^(refs/(heads|tags)/)?v1\.1`,
				`^v1\.0`,
			},
		},
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
	// compile default patterns
	if err := cfg.compileCompatibilityPatterns(); err != nil {
		panic(errors.Wrap(err, "error compiling default compatibility patterns"))
	}
	return cfg
}

func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()
	if err := cf.BindYaml(cfg, path, env.GetCfOptions()); err != nil {
		return nil, errors.Wrapf(err, "error loading controller config '%v'", path)
	}
	if !envVersionOk() && cfg.V != ConfigVersion {
		return nil, errors.Errorf("expecting configuration version '%v', your configuration is version '%v'; please see zrok.io for changelog and configuration documentation", ConfigVersion, cfg.V)
	}
	if err := cfg.compileCompatibilityPatterns(); err != nil {
		return nil, errors.Wrap(err, "error compiling compatibility patterns")
	}
	return cfg, nil
}

func (cfg *Config) compileCompatibilityPatterns() error {
	if cfg.Compatibility == nil {
		return nil
	}

	compiled := make([]*regexp.Regexp, len(cfg.Compatibility.VersionPatterns))
	for i, pattern := range cfg.Compatibility.VersionPatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return errors.Wrapf(err, "invalid regex pattern '%v' at index %d", pattern, i)
		}
		compiled[i] = re
	}
	cfg.Compatibility.compiledPatterns = compiled
	return nil
}

func (cfg *CompatibilityConfig) GetCompiledPatterns() []*regexp.Regexp {
	return cfg.compiledPatterns
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
