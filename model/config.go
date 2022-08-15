package model

const ZrokProxyConfig = "zrok.proxy.v1"

type AuthScheme string

const (
	None  AuthScheme = "none"
	Basic            = "basic"
)

type ProxyConfig struct {
	AuthScheme AuthScheme `json:"auth_scheme"`
}
