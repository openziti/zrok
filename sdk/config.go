package sdk

import "github.com/pkg/errors"

const ZrokProxyConfig = "zrok.proxy.v1"

type ProxyConfig struct {
	AuthScheme AuthScheme `json:"auth_scheme"`
	BasicAuth  *BasicAuth `json:"basic_auth"`
}

type BasicAuth struct {
	Users []*AuthUser `json:"users"`
}

type AuthUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func ParseAuthScheme(authScheme string) (AuthScheme, error) {
	switch authScheme {
	case string(None):
		return None, nil
	case string(Basic):
		return Basic, nil
	default:
		return None, errors.Errorf("unknown auth scheme '%v'", authScheme)
	}
}
