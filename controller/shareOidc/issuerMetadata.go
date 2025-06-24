package shareOidc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

type IssuerMetadata struct {
	Issuer                string   `json:"issuer"`
	AuthorizationEndpoint string   `json:"authorization_endpoint"`
	TokenEndpoint         string   `json:"token_endpoint"`
	UserinfoEndpoint      string   `json:"userinfo_endpoint"`
	JwksURI               string   `json:"jwks_uri"`
	ScopesSupported       []string `json:"scopes_supported"`
}

func (m *IssuerMetadata) String() string {
	jsonBytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Sprintf("error marshaling OIDC metadata: %v", err)
	}
	return string(jsonBytes)
}

func FetchAndValidateIssuerMetadata(issuerUrl string) (*IssuerMetadata, error) {
	// fetch the OIDC configuration from the well-known endpoint
	wellKnownURL := issuerUrl
	if !strings.HasSuffix(wellKnownURL, "/") {
		wellKnownURL += "/"
	}
	wellKnownURL += ".well-known/openid-configuration"

	resp, err := http.Get(wellKnownURL)
	if err != nil {
		return nil, errors.Wrapf(err, "error fetching OIDC configuration from %s", wellKnownURL)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("received non-200 status code (%d) from OIDC configuration endpoint '%v'", resp.StatusCode, wellKnownURL)
	}

	metadata := &IssuerMetadata{}
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, errors.Wrap(err, "error decoding OIDC configuration")
	}

	if metadata.Issuer == "" {
		return nil, errors.New("issuer not found in OIDC configuration")
	}
	if metadata.AuthorizationEndpoint == "" {
		return nil, errors.New("authorization_endpoint not found in OIDC configuration")
	}
	if metadata.TokenEndpoint == "" {
		return nil, errors.New("token_endpoint not found in OIDC configuration")
	}
	if metadata.UserinfoEndpoint == "" {
		return nil, errors.New("userinfo_endpoint not found in OIDC configuration")
	}

	return metadata, nil
}
