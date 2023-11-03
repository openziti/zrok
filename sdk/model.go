package sdk

import "time"

type BackendMode string

const (
	ProxyBackendMode     BackendMode = "proxy"
	WebBackendMode       BackendMode = "web"
	TcpTunnelBackendMode BackendMode = "tcpTunnel"
	UdpTunnelBackendMode BackendMode = "udpTunnel"
	CaddyBackendMode     BackendMode = "caddy"
	DriveBackendMode     BackendMode = "drive"
)

type ShareMode string

const (
	PrivateShareMode ShareMode = "private"
	PublicShareMode  ShareMode = "public"
)

type ShareRequest struct {
	Reserved                        bool
	BackendMode                     BackendMode
	ShareMode                       ShareMode
	Target                          string
	Frontends                       []string
	BasicAuth                       []string
	OauthProvider                   string
	OauthEmailDomains               []string
	OauthAuthorizationCheckInterval time.Duration
}

type Share struct {
	Token             string        `json:"token"`
	FrontendEndpoints []string      `json:"frontend_endpoints"`
	Request           *ShareRequest `json:"request"`
}

type AccessRequest struct {
	ShareToken string
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
