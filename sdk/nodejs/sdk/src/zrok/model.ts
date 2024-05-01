import {ShareRequestBackendModeEnum, ShareRequestShareModeEnum, ShareRequestOauthProviderEnum} from "./api"

export type BackendMode = string

export const PROXY_BACKEND_MODE: BackendMode = "proxy"
export const WEB_BACKEND_MODE: BackendMode = "web"
export const TCP_TUNNEL_BACKEND_MODE: BackendMode = "tcpTunnel"
export const UDP_TUNNEL_BACKEND_MODE: BackendMode = "udpTunnel"
export const CADDY_BACKEND_MODE: BackendMode = "caddy"

export type ShareMode = string

export const PRIVATE_SHARE_MODE: ShareMode = "private"
export const PUBLIC_SHARE_MODE: ShareMode = "public"

export class ShareRequest {
    BackendMode: BackendMode
    ShareMode: ShareMode
    Target: string
    Frontends: string[]
    BasicAuth: string[]
    OauthProvider: string
    OauthEmailDomains: string[]
    OauthAuthorizationCheckInterval: string

    constructor(backendMode: BackendMode,
                shareMode: ShareMode,
                target: string,
                frontends: string[] = [],
                basicAuth: string[] = [],
                oauthProvider: string = "",
                oauthEmailDomains: string[] = [],
                oauthAuthorizationCheckInterval: string = "") {
        this.BackendMode = backendMode
        this.ShareMode = shareMode
        this.Target = target
        this.Frontends = frontends
        this.BasicAuth = basicAuth
        this.OauthProvider = oauthProvider
        this.OauthEmailDomains = oauthEmailDomains
        this.OauthAuthorizationCheckInterval = oauthAuthorizationCheckInterval
    }
}

export class Share {
    Token: string
    FrontendEndpoints: string[]

    constructor(Token: string, FrontendEndpoints: string[]) {
        this.Token = Token
        this.FrontendEndpoints = FrontendEndpoints
    }
}

export class AccessRequest {
    ShareToken: string

    constructor(ShareToken: string) {
        this.ShareToken = ShareToken
    }
}

export class Access {
    Token: string
    ShareToken: string
    BackendMode: BackendMode

    constructor(Token: string, ShareToken: string, BackendMode: BackendMode) {
        this.Token = Token
        this.ShareToken = ShareToken
        this.BackendMode = BackendMode
    }
}

export class SessionMetrics {
    BytesRead: number
    BytesWritten: number
    LastUpdate: number

    constructor(BytesRead: number, BytesWrittern: number, LastUpdate: number) {
        this.BytesRead = BytesRead
        this.BytesWritten = BytesWrittern
        this.LastUpdate = LastUpdate
    }
}

export class Metrics {
    Namespace: string
    Sessions: Record<string, SessionMetrics>

    constructor(Namespace: string, Sessions: Record<string, SessionMetrics>) {
        this.Namespace = Namespace
        this.Sessions = Sessions
    }
}

export type AuthScheme = string

export const AUTH_SCHEME_NONE: AuthScheme = "none"
export const AUTH_SCHEME_BASIC: AuthScheme = "basic"
export const AUTH_SCHEME_OAUTH: AuthScheme = "oauth"

export function zrokBackendModeToOpenApi(z: BackendMode): ShareRequestBackendModeEnum | undefined{
    switch(z){
        case PROXY_BACKEND_MODE:
            return ShareRequestBackendModeEnum.Proxy
        case WEB_BACKEND_MODE:
            return ShareRequestBackendModeEnum.Web
        case TCP_TUNNEL_BACKEND_MODE:
            return ShareRequestBackendModeEnum.TcpTunnel
        case UDP_TUNNEL_BACKEND_MODE:
            return ShareRequestBackendModeEnum.UdpTunnel
        case CADDY_BACKEND_MODE:
            return ShareRequestBackendModeEnum.Caddy
        default:
            return undefined
    }
}

export function zrokShareModeToOpenApi(z: ShareMode): ShareRequestShareModeEnum | undefined {
    switch(z) {
        case PRIVATE_SHARE_MODE:
            return ShareRequestShareModeEnum.Private
        case PUBLIC_SHARE_MODE:
            return ShareRequestShareModeEnum.Public
        default:
            return undefined
    }
}

export function zrokOauthProviderToOpenApi(z: string): ShareRequestOauthProviderEnum | undefined {
    switch(z.toLowerCase()){
        case ShareRequestOauthProviderEnum.Github.toString().toLowerCase():
            return ShareRequestOauthProviderEnum.Github
        case ShareRequestOauthProviderEnum.Google.toString().toLowerCase():
            return ShareRequestOauthProviderEnum.Google 
        default:
            return undefined 
    }
}