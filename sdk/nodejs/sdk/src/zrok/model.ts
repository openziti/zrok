export type ShareMode = string;
export const PRIVATE_SHARE_MODE: ShareMode = "private";
export const PUBLIC_SHARE_MODE: ShareMode = "public";

export type BackendMode = string;
export const PROXY_BACKEND_MODE: BackendMode = "proxy";
export const WEB_BACKEND_MODE: BackendMode = "web";
export const TCP_TUNNEL_BACKEND_MODE: BackendMode = "tcpTunnel";
export const UDP_TUNNEL_BACKEND_MODE: BackendMode = "udpTunnel";
export const CADDY_BACKEND_MODE: BackendMode = "caddy";
export const DRIVE_BACKEND_MODE: BackendMode = "drive";
export const SOCKS_BACKEND_MODE: BackendMode = "socks";

export type AuthScheme = string;
export const AUTH_SCHEME_NONE: AuthScheme = "none";
export const AUTH_SCHEME_BASIC: AuthScheme = "basic";
export const AUTH_SCHEME_OAUTH: AuthScheme = "oauth";

export type PermissionMode = string;
export const OPEN_PERMISSION_MODE: PermissionMode = "open";
export const CLOSED_PERMISSION_MODE: PermissionMode = "closed";

export class NameSelection {
    namespaceToken: string;
    name: string | undefined;

    constructor(namespaceToken: string, name?: string) {
        this.namespaceToken = namespaceToken;
        this.name = name;
    }
}

export class ShareRequest {
    reserved: boolean;
    uniqueName: string | undefined;
    backendMode: BackendMode;
    shareMode: ShareMode;
    target: string;
    frontends: string[] | undefined;
    nameSelections: NameSelection[] | undefined;
    basicAuth: string[] | undefined;
    oauthProvider: string | undefined;
    oauthEmailDomains: string[] | undefined;
    oauthRefreshInterval: string | undefined;
    permissionMode: PermissionMode;
    accessGrants: string[] | undefined;

    constructor(shareMode: ShareMode, backendMode: BackendMode, target: string) {
        this.reserved = false;
        this.uniqueName = undefined;
        this.backendMode = backendMode;
        this.shareMode = shareMode;
        this.target = target;
        this.frontends = undefined;
        this.nameSelections = shareMode === PUBLIC_SHARE_MODE ? [new NameSelection("public")] : undefined;
        this.basicAuth = undefined;
        this.oauthProvider = undefined;
        this.oauthEmailDomains = undefined;
        this.oauthRefreshInterval = undefined;
        this.permissionMode = CLOSED_PERMISSION_MODE;
        this.accessGrants = undefined;
    }
}

export class Share {
    shareToken: string;
    frontendEndpoints: string[] | undefined;

    constructor(shareToken: string, frontendEndpoints: string[] | undefined) {
        this.shareToken = shareToken;
        this.frontendEndpoints = frontendEndpoints;
    }
}

export class AccessRequest {
    shareToken: string;
    bindAddress: string | undefined;

    constructor(shareToken: string, bindAddress?: string) {
        this.shareToken = shareToken;
        this.bindAddress = bindAddress;
    }
}

export class Access {
    frontendToken: string;
    shareToken: string;
    backendMode: BackendMode;

    constructor(frontendToken: string, shareToken: string, backendMode: BackendMode) {
        this.frontendToken = frontendToken;
        this.shareToken = shareToken;
        this.backendMode = backendMode;
    }
}

export class EnableRequest {
    description: string;
    host: string;

    constructor(description: string = "", host: string = "") {
        this.description = description;
        this.host = host;
    }
}

export interface ShareDetail {
    token: string;
    zId: string;
    envZId: string;
    shareMode: string;
    backendMode: string;
    frontendEndpoints: string[];
    target: string;
    limited: boolean;
    createdAt: number;
    updatedAt: number;
}

export interface AccessDetail {
    id: number;
    frontendToken: string;
    envZId: string;
    shareToken: string;
    backendMode: string;
    bindAddress: string;
    description: string;
    limited: boolean;
    createdAt: number;
    updatedAt: number;
}

export interface NameEntry {
    namespaceToken: string;
    namespaceName: string;
    name: string;
    shareToken: string;
    reserved: boolean;
    createdAt: number;
}

export interface Namespace {
    namespaceToken: string;
    name: string;
    description: string;
}

export interface Status {
    enabled: boolean;
    apiEndpoint: string;
    apiEndpointSource: string;
    token: string;
    zitiIdentity: string;
}
