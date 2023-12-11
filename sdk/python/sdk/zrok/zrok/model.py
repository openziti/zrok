from dataclasses import dataclass, field

BackendMode = str

PROXY_BACKEND_MODE: BackendMode = "proxy"
WEB_BACKEND_MODE: BackendMode = "web"
TCP_TUNNEL_BACKEND_MODE: BackendMode = "tcpTunnel"
UDP_TUNNEL_BACKEND_MODE: BackendMode = "udpTunnel"
CADDY_BACKEND_MODE: BackendMode = "caddy"

ShareMode = str

PRIVATE_SHARE_MODE: ShareMode = "private"
PUBLIC_SHARE_MODE: ShareMode = "public"

@dataclass
class ShareRequest:
    BackendMode: BackendMode
    ShareMode: ShareMode
    Target: str
    Frontends: list[str] = field(default_factory=list[str])
    BasicAuth: list[str] = field(default_factory=list[str])
    OauthProvider: str = ""
    OauthEmailDomains: list[str] = field(default_factory=list[str])
    OauthAuthorizationCheckInterval: str = ""
    Reserved: bool = False
    UniqueName: str = ""

@dataclass
class Share:
    Token: str
    FrontendEndpoints: list[str]

@dataclass
class AccessRequest:
    ShareToken: str

@dataclass
class Access:
    Token: str
    ShareToken: str
    BackendMode: BackendMode

@dataclass
class SessionMetrics:
    BytesRead: int
    BytesWritten: int
    LastUpdate: int

@dataclass
class Metrics:
    Namespace: str
    Sessions: dict[str, SessionMetrics]

AuthScheme = str

AUTH_SCHEME_NONE: AuthScheme = "none"
AUTH_SCHEME_BASIC: AuthScheme = "basic"
AUTH_SCHEME_OAUTH: AuthScheme = "oauth"