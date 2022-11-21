package zrokdir

import (
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/openziti-test-kitchen/zrok/build"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"net/url"
	"os"
	"strings"
)

func AddZrokApiEndpointFlag(v *string, flags *pflag.FlagSet) {
	defaultEndpoint := os.Getenv("ZROK_API_ENDPOINT")
	if defaultEndpoint == "" {
		defaultEndpoint = "https://api.zrok.io"
	}
	flags.StringVarP(v, "endpoint", "e", defaultEndpoint, "zrok API endpoint address")
}

func ZrokClient(endpoint string) (*rest_client_zrok.Zrok, error) {
	apiUrl, err := url.Parse(endpoint)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing api endpoint '%v'", endpoint)
	}
	transport := httptransport.New(apiUrl.Host, "/api/v1", []string{apiUrl.Scheme})
	transport.Producers["application/zrok.v1+json"] = runtime.JSONProducer()
	transport.Consumers["application/zrok.v1+json"] = runtime.JSONConsumer()

	zrok := rest_client_zrok.New(transport, strfmt.Default)
	v, err := zrok.Metadata.Version(nil)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting version from api endpoint '%v': %v", endpoint, err)
	}
	if !strings.HasPrefix(string(v.Payload), build.Series) {
		return nil, errors.Errorf("expected a '%v' version, received: '%v'", build.Series, v.Payload)
	}

	return zrok, nil
}
