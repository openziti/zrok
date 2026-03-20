import {Root} from "./environment";
import {
    AuthUser,
    MetadataApi,
    ShareApi,
    ShareRequest as ApiShareRequest,
    ShareRequestBackendModeEnum,
    ShareRequestPermissionModeEnum,
    ShareRequestShareModeEnum,
    UnshareRequest,
    UpdateShareRequest as ApiUpdateShareRequest,
    NameSelection as ApiNameSelection,
} from "../api";
import {
    BackendMode,
    Share,
    ShareRequest,
    ShareDetail,
    PRIVATE_SHARE_MODE,
    PUBLIC_SHARE_MODE,
    PROXY_BACKEND_MODE,
    WEB_BACKEND_MODE,
    TCP_TUNNEL_BACKEND_MODE,
    UDP_TUNNEL_BACKEND_MODE,
    CADDY_BACKEND_MODE,
    DRIVE_BACKEND_MODE,
    SOCKS_BACKEND_MODE,
    AUTH_SCHEME_NONE,
    AUTH_SCHEME_BASIC,
    AUTH_SCHEME_OAUTH,
    CLOSED_PERMISSION_MODE,
} from "./model";

export const createShare = async (root: Root, request: ShareRequest): Promise<Share> => {
    if (!root.isEnabled()) {
        throw new Error("environment is not enabled; enable with 'zrok2 enable' first!");
    }

    let req: ApiShareRequest;
    switch (request.shareMode) {
        case PRIVATE_SHARE_MODE:
            req = toPrivateApiShareRequest(root, request);
            break;
        case PUBLIC_SHARE_MODE:
            req = toPublicApiShareRequest(root, request);
            break;
        default:
            throw new Error("unknown share mode '" + request.shareMode + "'");
    }

    const cfg = await root.client();
    const shr = await new ShareApi(cfg).share({body: req})
        .catch(err => {
            throw new Error("unable to create share: " + err);
        });

    return new Share(shr.shareToken!, shr.frontendProxyEndpoints);
}

export const deleteShare = async (root: Root, shr: Share): Promise<void> => {
    if (!root.isEnabled()) {
        throw new Error("environment is not enabled; enable with 'zrok2 enable' first!");
    }
    const req: UnshareRequest = {
        envZId: root.environment?.zId!,
        shareToken: shr.shareToken
    };
    const cfg = await root.client();
    await new ShareApi(cfg).unshare({body: req})
        .catch(err => {
            throw new Error("unable to delete share: " + err);
        });
}

export const releaseReservedShare = async (root: Root, shr: Share): Promise<void> => {
    if (!root.isEnabled()) {
        throw new Error("environment is not enabled; enable with 'zrok2 enable' first!");
    }
    const req: UnshareRequest = {
        envZId: root.environment?.zId!,
        shareToken: shr.shareToken,
    };
    const cfg = await root.client();
    await new ShareApi(cfg).unshare({body: req})
        .catch(err => {
            throw new Error("unable to release reserved share: " + err);
        });
}

export const modifyShare = async (
    root: Root,
    shareToken: string,
    addAccessGrants?: string[],
    removeAccessGrants?: string[],
): Promise<void> => {
    if (!root.isEnabled()) {
        throw new Error("environment is not enabled; enable with 'zrok2 enable' first!");
    }
    const req: ApiUpdateShareRequest = {
        shareToken: shareToken,
        addAccessGrants: addAccessGrants || [],
        removeAccessGrants: removeAccessGrants || [],
    };
    const cfg = await root.client();
    await new ShareApi(cfg).updateShare({body: req})
        .catch(err => {
            throw new Error("unable to modify share: " + err);
        });
}

export const getShareDetail = async (root: Root, shareToken: string): Promise<ShareDetail> => {
    if (!root.isEnabled()) {
        throw new Error("environment is not enabled; enable with 'zrok2 enable' first!");
    }
    const cfg = await root.client();
    const res = await new MetadataApi(cfg).getShareDetail({shareToken})
        .catch(err => {
            throw new Error("unable to get share detail: " + err);
        });

    return {
        token: res.shareToken || "",
        zId: res.zId || "",
        envZId: res.envZId || "",
        shareMode: res.shareMode || "",
        backendMode: res.backendMode || "",
        frontendEndpoints: res.frontendEndpoints || [],
        target: res.target || "",
        limited: res.limited || false,
        createdAt: res.createdAt || 0,
        updatedAt: res.updatedAt || 0,
    };
}

const toPrivateApiShareRequest = (root: Root, request: ShareRequest): ApiShareRequest => {
    const out: ApiShareRequest = {
        envZId: root.environment?.zId,
        shareMode: ShareRequestShareModeEnum.Private,
        backendMode: toApiBackendMode(request.backendMode),
        target: request.target,
        authScheme: AUTH_SCHEME_NONE,
        permissionMode: (request.permissionMode || CLOSED_PERMISSION_MODE) as ShareRequestPermissionModeEnum,
    };
    if (request.nameSelections) {
        out.nameSelections = request.nameSelections.map(n => ({
            namespaceToken: n.namespaceToken,
            name: n.name,
        } as ApiNameSelection));
    }
    if (request.accessGrants) {
        out.accessGrants = request.accessGrants;
    }
    return out;
}

const toPublicApiShareRequest = (root: Root, request: ShareRequest): ApiShareRequest => {
    const out: ApiShareRequest = {
        envZId: root.environment?.zId,
        shareMode: ShareRequestShareModeEnum.Public,
        backendMode: toApiBackendMode(request.backendMode),
        target: request.target,
        authScheme: AUTH_SCHEME_NONE,
        permissionMode: request.permissionMode as ShareRequestPermissionModeEnum,
    };
    if (request.nameSelections) {
        out.nameSelections = request.nameSelections.map(n => ({
            namespaceToken: n.namespaceToken,
            name: n.name,
        } as ApiNameSelection));
    }
    if (request.accessGrants) {
        out.accessGrants = request.accessGrants;
    }

    if (request.oauthProvider !== undefined) {
        out.authScheme = AUTH_SCHEME_OAUTH;
        out.oauthProvider = request.oauthProvider;
        out.oauthEmailDomains = request.oauthEmailDomains;
        out.oauthRefreshInterval = request.oauthRefreshInterval;

    } else if (request.basicAuth && request.basicAuth.length > 0) {
        out.authScheme = AUTH_SCHEME_BASIC;
        out.basicAuthUsers = new Array<AuthUser>();
        for (const pair of request.basicAuth) {
            const tokens = pair.split(":");
            if (tokens.length === 2) {
                out.basicAuthUsers.push({username: tokens[0].trim(), password: tokens[1].trim()});
            }
        }
    }

    return out;
}

export const toApiBackendMode = (mode: BackendMode): ShareRequestBackendModeEnum | undefined => {
    switch (mode) {
        case PROXY_BACKEND_MODE:
            return ShareRequestBackendModeEnum.Proxy;
        case WEB_BACKEND_MODE:
            return ShareRequestBackendModeEnum.Web;
        case TCP_TUNNEL_BACKEND_MODE:
            return ShareRequestBackendModeEnum.TcpTunnel;
        case UDP_TUNNEL_BACKEND_MODE:
            return ShareRequestBackendModeEnum.UdpTunnel;
        case CADDY_BACKEND_MODE:
            return ShareRequestBackendModeEnum.Caddy;
        case DRIVE_BACKEND_MODE:
            return ShareRequestBackendModeEnum.Drive;
        case SOCKS_BACKEND_MODE:
            return ShareRequestBackendModeEnum.Socks;
        default:
            return undefined;
    }
}
