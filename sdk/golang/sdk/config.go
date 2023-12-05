package sdk

import "github.com/pkg/errors"

const ZrokProxyConfig = "zrok.proxy.v1"

type FrontendConfig struct {
	AuthScheme AuthScheme       `json:"auth_scheme"`
	BasicAuth  *BasicAuthConfig `json:"basic_auth"`
	OauthAuth  *OauthConfig     `json:"oauth"`
}

type BasicAuthConfig struct {
	Users []*AuthUserConfig `json:"users"`
}

type AuthUserConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type OauthConfig struct {
	Provider                   string   `json:"provider"`
	EmailDomains               []string `json:"email_domains"`
	AuthorizationCheckInterval string   `json:"authorization_check_interval"`
}

func ParseAuthScheme(authScheme string) (AuthScheme, error) {
	switch authScheme {
	case string(None):
		return None, nil
	case string(Basic):
		return Basic, nil
	case string(Oauth):
		return Oauth, nil
	default:
		return None, errors.Errorf("unknown auth scheme '%v'", authScheme)
	}
}
