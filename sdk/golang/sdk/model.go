package sdk

import (
	"time"
	"github.com/openziti/zrok/rest_model_zrok"
)

type EnableRequest struct {
	Host        string
	Description string
}

type Environment struct {
	Host         string
	Description  string
	ZitiIdentity string
	ZitiConfig   string
}

type BackendMode string

const (
	ProxyBackendMode     BackendMode = "proxy"
	WebBackendMode       BackendMode = "web"
	TcpTunnelBackendMode BackendMode = "tcpTunnel"
	UdpTunnelBackendMode BackendMode = "udpTunnel"
	CaddyBackendMode     BackendMode = "caddy"
	DriveBackendMode     BackendMode = "drive"
	SocksBackendMode     BackendMode = "socks"
	VpnBackendMode       BackendMode = "vpn"
)

type ShareMode string

const (
	PrivateShareMode ShareMode = "private"
	PublicShareMode  ShareMode = "public"
)

type PermissionMode string

const (
	OpenPermissionMode   PermissionMode = "open"
	ClosedPermissionMode PermissionMode = "closed"
)

type ShareRequest struct {
	Reserved                        bool
	UniqueName                      string
	BackendMode                     BackendMode
	ShareMode                       ShareMode
	Target                          string
	Frontends                       []string
	BasicAuth                       []string
	OauthProvider                   string
	OauthEmailAddressPatterns       []string
	OauthAuthorizationCheckInterval time.Duration
	PermissionMode                  PermissionMode
	AccessGrants                    []string
}

type Share struct {
	Token             string   `json:"token"`
	FrontendEndpoints []string `json:"frontend_endpoints"`
}

type AccessRequest struct {
	ShareToken  string
	BindAddress string
}

type Access struct {
	Token       string
	ShareToken  string
	BackendMode BackendMode
}

type Metrics struct {
	Namespace string
	Sessions  map[string]SessionMetrics
}

type SessionMetrics struct {
	BytesRead    int64
	BytesWritten int64
	LastUpdate   int64
}

type AuthScheme string

const (
	None  AuthScheme = "none"
	Basic AuthScheme = "basic"
	Oauth AuthScheme = "oauth"
)

type Share12Request struct {
	EnvZId                   string
	ShareMode                string
	Target                   string
	BackendMode              string
	PermissionMode           PermissionMode
	AccessGrants             []string
	BasicAuthUsers           []string
	OauthProvider            string
	OauthEmailDomains        []string
	OauthRefreshInterval     string
	NamespaceSelections      []*rest_model_zrok.NamespaceSelection
}

type Share12Response struct {
	ShareToken             string   `json:"shareToken"`
	FrontendProxyEndpoints []string `json:"frontendProxyEndpoints"`
}
