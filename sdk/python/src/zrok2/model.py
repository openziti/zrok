from dataclasses import dataclass, field

BackendMode = str

PROXY_BACKEND_MODE: BackendMode = "proxy"
WEB_BACKEND_MODE: BackendMode = "web"
TCP_TUNNEL_BACKEND_MODE: BackendMode = "tcpTunnel"
UDP_TUNNEL_BACKEND_MODE: BackendMode = "udpTunnel"
CADDY_BACKEND_MODE: BackendMode = "caddy"
DRIVE_BACKEND_MODE: BackendMode = "drive"
SOCKS_BACKEND_MODE: BackendMode = "socks"

ShareMode = str

PRIVATE_SHARE_MODE: ShareMode = "private"
PUBLIC_SHARE_MODE: ShareMode = "public"

PermissionMode = str

OPEN_PERMISSION_MODE: PermissionMode = "open"
CLOSED_PERMISSION_MODE: PermissionMode = "closed"


@dataclass
class ShareRequest:
    BackendMode: BackendMode
    ShareMode: ShareMode
    Target: str
    Frontends: list[str] = field(default_factory=list)
    BasicAuth: list[str] = field(default_factory=list)
    OauthProvider: str = ""
    OauthEmailAddressPatterns: list[str] = field(default_factory=list)
    OauthAuthorizationCheckInterval: str = ""
    Reserved: bool = False
    UniqueName: str = ""
    PermissionMode: PermissionMode = OPEN_PERMISSION_MODE
    AccessGrants: list[str] = field(default_factory=list)
    NameSelections: list = field(default_factory=list)


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


@dataclass
class EnableRequest:
    Description: str = ""
    Host: str = ""


@dataclass
class NameSelection:
    NamespaceToken: str = ""
    Name: str = ""


@dataclass
class ShareDetail:
    Token: str = ""
    ZId: str = ""
    EnvZId: str = ""
    ShareMode: str = ""
    BackendMode: str = ""
    FrontendEndpoints: list[str] = field(default_factory=list)
    Target: str = ""
    Limited: bool = False
    CreatedAt: int = 0
    UpdatedAt: int = 0


@dataclass
class AccessDetail:
    Id: int = 0
    FrontendToken: str = ""
    EnvZId: str = ""
    ShareToken: str = ""
    BackendMode: str = ""
    BindAddress: str = ""
    Description: str = ""
    Limited: bool = False
    CreatedAt: int = 0
    UpdatedAt: int = 0


@dataclass
class NameEntry:
    NamespaceToken: str = ""
    NamespaceName: str = ""
    Name: str = ""
    ShareToken: str = ""
    Reserved: bool = False
    CreatedAt: int = 0


@dataclass
class Namespace:
    NamespaceToken: str = ""
    Name: str = ""
    Description: str = ""


@dataclass
class Status:
    Enabled: bool = False
    ApiEndpoint: str = ""
    ApiEndpointSource: str = ""
    Token: str = ""
    ZitiIdentity: str = ""
