package env_v0_3

import (
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/openziti/zrok/build"
	"github.com/openziti/zrok/rest_client_zrok"
	"github.com/pkg/errors"
	"net/url"
	"os"
	"regexp"
)

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
	v, err := zrok.Metadata.Version(nil)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting version from api endpoint '%v': %v", apiEndpoint, err)
	}
	// allow reported version string to be optionally prefixed with
	// "refs/heads/" or "refs/tags/"
	re := regexp.MustCompile(`^(refs/(heads|tags)/)?` + build.Series)
	if !re.MatchString(string(v.Payload)) {
		return nil, errors.Errorf("expected a '%v' version, received: '%v'", build.Series, v.Payload)
	}

	return zrok, nil
}

func (r *Root) ApiEndpoint() (apiEndpoint string, from string) {
	apiEndpoint = "https://api.zrok.io"
	from = "binary"

	if r.Cfg != nil && r.Cfg.ApiEndpoint != "" {
		apiEndpoint = r.Cfg.ApiEndpoint
		from = "config"
	}

	env := os.Getenv("ZROK_API_ENDPOINT")
	if env != "" {
		apiEndpoint = env
		from = "ZROK_API_ENDPOINT"
	}

	if r.Env != nil && r.Env.ApiEndpoint != "" {
		apiEndpoint = r.Env.ApiEndpoint
		from = "env"
	}

	return apiEndpoint, from
}
