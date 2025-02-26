package env_v0_4

import (
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/openziti/zrok/build"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/rest_client_zrok"
	metadata2 "github.com/openziti/zrok/rest_client_zrok/metadata"
	"github.com/pkg/errors"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

func (r *Root) Metadata() *env_core.Metadata {
	return r.meta
}

func (r *Root) HasConfig() (bool, error) {
	return r.cfg != nil, nil
}

func (r *Root) Config() *env_core.Config {
	return r.cfg
}

func (r *Root) SetConfig(cfg *env_core.Config) error {
	if err := assertMetadata(); err != nil {
		return err
	}
	if err := saveConfig(cfg); err != nil {
		return err
	}
	r.cfg = cfg
	return nil
}

func (r *Root) Client() (*rest_client_zrok.Zrok, error) {
	apiEndpoint, _ := r.ApiEndpoint()
	apiUrl, err := url.Parse(apiEndpoint)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing api endpoint '%v'", r)
	}
	transport := httptransport.New(apiUrl.Host, "/api/v1", []string{apiUrl.Scheme})
	transport.Producers["application/zrok.v1+json"] = runtime.JSONProducer()
	transport.Consumers["application/zrok.v1+json"] = runtime.JSONConsumer()

	zrok := rest_client_zrok.New(transport, strfmt.Default)
	_, err = zrok.Metadata.ClientVersionCheck(&metadata2.ClientVersionCheckParams{
		Body: metadata2.ClientVersionCheckBody{
			ClientVersion: build.String(),
		},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "client version error accessing api endpoint '%v': %v", apiEndpoint, err)
	}
	return zrok, nil
}

func (r *Root) ApiEndpoint() (string, string) {
	apiEndpoint := "https://api-v1.zrok.io"
	from := "binary"

	if r.Config() != nil && r.Config().ApiEndpoint != "" {
		apiEndpoint = r.Config().ApiEndpoint
		from = "config"
	}

	env := os.Getenv("ZROK_API_ENDPOINT")
	if env != "" {
		apiEndpoint = env
		from = "ZROK_API_ENDPOINT"
	}

	if r.IsEnabled() {
		apiEndpoint = r.Environment().ApiEndpoint
		from = "env"
	}

	return apiEndpoint, from
}

func (r *Root) DefaultFrontend() (string, string) {
	defaultFrontend := "public"
	from := "binary"

	if r.Config() != nil && r.Config().DefaultFrontend != "" {
		defaultFrontend = r.Config().DefaultFrontend
		from = "config"
	}

	env := os.Getenv("ZROK_DEFAULT_FRONTEND")
	if env != "" {
		defaultFrontend = env
		from = "ZROK_DEFAULT_FRONTEND"
	}

	return defaultFrontend, from
}

func (r *Root) Headless() (bool, string) {
	headless := false
	from := "binary"

	if r.Config() != nil {
		headless = r.Config().Headless
		from = "config"
	}

	env := os.Getenv("ZROK_HEADLESS")
	if env != "" {
		if v, err := strconv.ParseBool(env); err == nil {
			headless = v
			from = "ZROK_HEADLESS"
		}
	}

	return headless, from
}

func (r *Root) Environment() *env_core.Environment {
	return r.env
}

func (r *Root) SetEnvironment(env *env_core.Environment) error {
	if err := assertMetadata(); err != nil {
		return err
	}
	if err := saveEnvironment(env); err != nil {
		return err
	}
	r.env = env
	return nil
}

func (r *Root) DeleteEnvironment() error {
	ef, err := environmentFile()
	if err != nil {
		return errors.Wrap(err, "error getting environment file")
	}
	if err := os.Remove(ef); err != nil {
		return errors.Wrap(err, "error removing environment file")
	}
	r.env = nil
	return nil
}

func (r *Root) IsEnabled() bool {
	return r.env != nil
}

func (r *Root) PublicIdentityName() string {
	return "public"
}

func (r *Root) EnvironmentIdentityName() string {
	return "environment"
}

func (r *Root) ZitiIdentityNamed(name string) (string, error) {
	return identityFile(name)
}

func (r *Root) SaveZitiIdentityNamed(name, data string) error {
	if err := assertMetadata(); err != nil {
		return err
	}
	zif, err := r.ZitiIdentityNamed(name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(zif), os.FileMode(0700)); err != nil {
		return errors.Wrapf(err, "error creating environment path '%v'", filepath.Dir(zif))
	}
	if err := os.WriteFile(zif, []byte(data), os.FileMode(0600)); err != nil {
		return errors.Wrapf(err, "error writing ziti identity file '%v'", zif)
	}
	return nil
}

func (r *Root) DeleteZitiIdentityNamed(name string) error {
	zif, err := r.ZitiIdentityNamed(name)
	if err != nil {
		return errors.Wrapf(err, "error getting ziti identity file path for '%v'", name)
	}
	if err := os.Remove(zif); err != nil {
		return errors.Wrapf(err, "error removing ziti identity file '%v'", zif)
	}
	return nil
}

func (r *Root) AgentSocket() (string, error) {
	return agentSocket()
}

func (r *Root) Obliterate() error {
	zrd, err := rootDir()
	if err != nil {
		return err
	}
	if err := os.RemoveAll(zrd); err != nil {
		return err
	}
	return nil
}
