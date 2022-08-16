package model

const ZrokProxyConfig = "zrok.proxy.v1"

type AuthScheme string

const (
	None  AuthScheme = "none"
	Basic            = "basic"
)

type ProxyConfig struct {
	AuthScheme AuthScheme `json:"auth_scheme"`
	BasicAuth  BasicAuth  `json:"basic_auth"`
}

type BasicAuth struct {
	Users []*AuthUser `json:"users"`
}

type AuthUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
