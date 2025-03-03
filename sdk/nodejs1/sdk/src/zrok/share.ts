import {Root} from "./environment";
import {
    ShareApi,
    ShareRequest as ApiShareRequest,
    ShareRequestBackendModeEnum,
    ShareRequestShareModeEnum,
    UnshareRequest
} from "../api";

export type ShareMode = string;
export const PRIVATE_SHARE_MODE: ShareMode = "private";
export const PUBLIC_SHARE_MODE: ShareMode = "public";

export type BackendMode = string;
export const PROXY_BACKEND_MODE: BackendMode = "proxy";
export const TCP_TUNNEL_BACKEND_MODE: BackendMode = "tcpTunnel";
export const UDP_TUNNEL_BACKEND_MODE: BackendMode = "udpTunnel";

export type PermissionMode = string;
export const OPEN_PERMISSION_MODE = "open";
export const CLOSED_PERMISSION_MODE = "closed";

export class ShareRequest {
    reserved: boolean;
    uniqueName: string|undefined;
    backendMode: BackendMode;
    shareMode: ShareMode;
    target: string;
    frontends: string[]|undefined;
    basicAuth: string[]|undefined;
    oauthProvider: string|undefined;
    oauthEmailAddressPatterns: string[]|undefined;
    oauthAuthorizationCheckInterval: string|undefined;
    permissionMode: PermissionMode;
    accessGrants: string[]|undefined;

    constructor(shareMode: ShareMode, backendMode: BackendMode, target: string) {
        this.reserved = false;
        this.uniqueName = undefined;
        this.backendMode = backendMode;
        this.shareMode = shareMode;
        this.target = target;
        this.frontends = undefined;
        this.basicAuth = undefined;
        this.oauthProvider = undefined;
        this.oauthEmailAddressPatterns = undefined;
        this.oauthAuthorizationCheckInterval = undefined;
        this.permissionMode = CLOSED_PERMISSION_MODE;
        this.accessGrants = undefined;
    }
}

export class Share {
    shareToken: string;
    frontendEndpoints: string[]|undefined;

    constructor(shareToken: string, frontendEndpoints: string[]|undefined) {
        this.shareToken = shareToken;
        this.frontendEndpoints = frontendEndpoints;
    }
}

export const createShare = async (root: Root, request: ShareRequest): Promise<Share> => {
    if(!root.isEnabled()) {
        throw new Error("environment is not enabled; enable with 'zrok enable' first!");
    }

    let req: ApiShareRequest;
    switch(request.shareMode) {
        case PRIVATE_SHARE_MODE:
            req = toPrivateApiShareRequest(root, request);
            break;
        case PUBLIC_SHARE_MODE:
            req = toPublicApiShareRequest(root, request);
            break;
        default:
            throw new Error("unknown share mode '" + request.shareMode + "'");
    }

    let shr = await new ShareApi(root.apiConfiguration()).share({body: req})
        .catch(resp => {
            throw new Error("unable to create share: " + resp);
        });

    return new Share(shr.shareToken!, shr.frontendProxyEndpoints);
}

export const deleteShare = (root: Root, shr: Share): void => {
    if(!root.isEnabled()) {
        throw new Error("environment is not enable; enable with 'zrok enable' first!");
    }
    let req: UnshareRequest = {
        envZId: root.environment?.zId!,
        shareToken: shr.shareToken
    };
    new ShareApi(root.apiConfiguration()).unshare({body: req})
        .catch(resp => {
            throw new Error("unable to delete share: " + resp);
        });
}

const toPrivateApiShareRequest = (root: Root, request: ShareRequest): ApiShareRequest => {
    return {
        envZId: root.environment?.zId,
        shareMode: ShareRequestShareModeEnum.Private,
        backendMode: toApiBackendMode(request.backendMode),
        backendProxyEndpoint: request.target,
    };
}

const toPublicApiShareRequest = (root: Root, request: ShareRequest): ApiShareRequest => {
    return {
        envZId: root.environment?.zId,
        shareMode: ShareRequestShareModeEnum.Public,
        frontendSelection: request.frontends,
        backendMode: toApiBackendMode(request.backendMode),
        backendProxyEndpoint: request.target,
        authScheme: "none",
    };
}

const toApiBackendMode = (mode: BackendMode): ShareRequestBackendModeEnum|undefined => {
    switch(mode) {
        case PROXY_BACKEND_MODE:
            return ShareRequestBackendModeEnum.Proxy;
        case TCP_TUNNEL_BACKEND_MODE:
            return ShareRequestBackendModeEnum.TcpTunnel;
        case UDP_TUNNEL_BACKEND_MODE:
            return ShareRequestBackendModeEnum.UdpTunnel;
        default:
            return undefined;
    }
}