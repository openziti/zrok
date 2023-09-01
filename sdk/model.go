package sdk

type BackendMode string

const (
	ProxyBackendMode     BackendMode = "proxy"
	WebBackendMode       BackendMode = "web"
	TcpTunnelBackendMode BackendMode = "tcpTunnel"
	UdpTunnelBackendMode BackendMode = "udpTunnel"
	CaddyBackendMode     BackendMode = "caddy"
)

type ShareMode string

const (
	PrivateShareMode ShareMode = "private"
	PublicShareMode  ShareMode = "public"
)

type ShareRequest struct {
	BackendMode BackendMode
	ShareMode   ShareMode
	Frontends   []string
	Auth        []string
	Target      string
}

type Share struct {
	Token             string
	FrontendEndpoints []string
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
)
