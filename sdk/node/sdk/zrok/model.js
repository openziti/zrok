PROXY_BACKEND_MODE = "proxy"
WEB_BACKEND_MODE = "web"
TCP_TUNNEL_BACKEND_MODE = "tcpTunnel"
UDP_TUNNEL_BACKEND_MODE = "udpTunnel"
CADDY_BACKEND_MODE = "caddy"

PRIVATE_SHARE_MODE = "private"
PUBLIC_SHARE_MODE = "public"

export class ShareRequest {
    constructor(backendMode, shareMode, target, frontends, basicAuth, oauthProvider, oauthEmailDomains, oauthAuthorizationCheckInterval) {
        this.backendMode = backendMode
        this.shareMode = shareMode
        this.target = target
        this.frontends = frontends
        this.basicAuth = basicAuth
        this.oauthProvider = oauthProvider
        this.oauthEmailDomains = oauthEmailDomains
        this.oauthAuthorizationCheckInterval = oauthAuthorizationCheckInterval
    }
}

export class Share {
    constructor(token, frontendEndpoints) {
        this.token = token
        this.frontendEndpoints = frontendEndpoints
    }
}

export class AccessRequest {
    constructor(shareToken) {
        this.shareToken = shareToken
    }
}

export class Access {
    constructor(token, shareToken, backendMode) {
        this.token = token
        this.shareToken = shareToken
        this.backendMode = backendMode
    }
}

export class SessionMetrics {
    constructor(bytesRead, bytesWritten, lastUpdate) {
        this.bytesRead = bytesRead
        this.bytesWritten = bytesWritten
        this.lastUpdate = lastUpdate
    }
}

export class Metrics {
    constructor(namespace, sessions) {
        this.namespace = namespace
        this.sessions = sessions
    }
}

AUTH_SCHEME_NONE = "none"
AUTH_SCHEME_BASIC = "basic"
AUTH_SCHEME_OAUTH = "oauth"