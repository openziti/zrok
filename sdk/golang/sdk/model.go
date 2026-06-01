package sdk

import (
	"strings"
	"time"

	"github.com/pkg/errors"
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

type NameSelection struct {
	NamespaceToken string
	Name           string
}

// ParseNameSelection converts a string in the format "<namespaceToken>[:<name>]"
// into a NameSelection struct. if no name is provided, the Name field will be empty.
func ParseNameSelection(input string) (NameSelection, error) {
	parts := strings.SplitN(input, ":", 2)
	if len(parts) > 2 {
		return NameSelection{}, errors.New("invalid namespace selection")
	}
	selection := NameSelection{
		NamespaceToken: parts[0],
	}
	if len(parts) == 2 {
		selection.Name = parts[1]
	}
	return selection, nil
}

type ShareRequest struct {
	Reserved                  bool
	UniqueName                string
	BackendMode               BackendMode
	ShareMode                 ShareMode
	Target                    string
	NameSelections            []NameSelection
	PrivateShareToken         string
	BasicAuth                 []string
	OauthProvider             string
	OauthEmailAddressPatterns []string
	OauthRefreshInterval      time.Duration
	PermissionMode            PermissionMode
	AccessGrants              []string
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
