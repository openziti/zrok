package shareOidc

import (
	"strings"

	"github.com/openziti/zrok/controller/secretsGrpc"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/pkg/errors"
)

type Secrets struct {
	ClientId              string
	ClientSecret          string
	Scopes                []string
	Issuer                string
	AuthorizationEndpoint string
	TokenEndpoint         string
	UserinfoEndpoint      string
	JwksUri               string
}

func NewSecrets(clientId, clientSecret string, meta *IssuerMetadata) Secrets {
	secrets := Secrets{
		ClientId:              clientId,
		ClientSecret:          clientSecret,
		Scopes:                meta.ScopesSupported,
		Issuer:                meta.Issuer,
		AuthorizationEndpoint: meta.AuthorizationEndpoint,
		TokenEndpoint:         meta.TokenEndpoint,
		UserinfoEndpoint:      meta.UserinfoEndpoint,
		JwksUri:               meta.JwksURI,
	}
	return secrets
}

func FromGrpc(in *secretsGrpc.SecretsResponse) (Secrets, error) {
	secrets := Secrets{}
	for _, secret := range in.Secrets {
		switch secret.Key {
		case "auth_scheme":
			if secret.Value != string(sdk.Oidc) {
				return Secrets{}, errors.Errorf("expected 'oidc' auth_scheme, got '%v'", secret.Value)
			}
		case "oidc_client_id":
			secrets.ClientId = secret.Value
		case "oidc_client_secret":
			secrets.ClientSecret = secret.Value
		case "oidc_scopes":
			secrets.Scopes = strings.Split(secret.Value, ",")
		case "oidc_issuer":
			secrets.Issuer = secret.Value
		case "oidc_authorization_endpoint":
			secrets.AuthorizationEndpoint = secret.Value
		case "oidc_token_endpoint":
			secrets.TokenEndpoint = secret.Value
		case "oidc_userinfo_endpoint":
			secrets.UserinfoEndpoint = secret.Value
		case "oidc_jwks_uri":
			secrets.JwksUri = secret.Value
		}
	}
	return secrets, nil
}

func (s Secrets) ToStore(shareId int) store.Secrets {
	var secrets []store.Secret
	secrets = append(secrets, store.Secret{Key: "auth_scheme", Value: "oidc"})
	secrets = append(secrets, store.Secret{Key: "oidc_client_id", Value: s.ClientId})
	secrets = append(secrets, store.Secret{Key: "oidc_client_secret", Value: s.ClientSecret})
	secrets = append(secrets, store.Secret{Key: "oidc_scopes", Value: strings.Join(s.Scopes, ",")})
	secrets = append(secrets, store.Secret{Key: "oidc_issuer", Value: s.Issuer})
	secrets = append(secrets, store.Secret{Key: "oidc_authorization_endpoint", Value: s.AuthorizationEndpoint})
	secrets = append(secrets, store.Secret{Key: "oidc_token_endpoint", Value: s.TokenEndpoint})
	secrets = append(secrets, store.Secret{Key: "oidc_userinfo_endpoint", Value: s.UserinfoEndpoint})
	secrets = append(secrets, store.Secret{Key: "oidc_jwks_uri", Value: s.JwksUri})
	return store.Secrets{ShareId: shareId, Secrets: secrets}
}
