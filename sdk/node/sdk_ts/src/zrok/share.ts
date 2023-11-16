import {Root} from "../environment/root"
import {
    Share,
    ShareApi,
    ShareRequest,
    ShareResponse,
    AuthUser,
    ShareRequestOauthProviderEnum,
    ShareRequestShareModeEnum,
    UnshareRequest} from "./api"
import * as model from "./model"

export function CreateShare(root: Root, request: model.ShareRequest): model.Share | null | undefined {
    if (!root.IsEnabled()){
        throw new Error("environment is not enabled; enable with 'zrok enable' first!")
    }
    let out: ShareRequest

    switch(request.ShareMode) {
        case ShareRequestShareModeEnum.Private:
            out = newPrivateShare(root, request)
            break
        case ShareRequestShareModeEnum.Public:
            out = newPublicShare(root, request)
            break
        default:
            throw new Error("unknown share mode " + request.ShareMode)
    }

    if (request.BasicAuth.length > 0) {
        out.authScheme = model.AUTH_SCHEME_BASIC
        for(let pair in request.BasicAuth) {
            let tokens = pair.split(":")
            if (tokens.length === 2) {
                if (out.authUsers === undefined) {
                    out.authUsers = new Array<AuthUser>
                }
                out.authUsers.push({username: tokens[0].trim(), password: tokens[1].trim()})
            }
            else {
                throw new Error("invalid username:password pair: " + pair)
            }
        }
    }

    if (request.OauthProvider !== undefined) {
        out.authScheme = model.AUTH_SCHEME_OAUTH
    }

    let client = new ShareApi(root.Client())
    let shr: model.Share | null = null
    client.share({body: out})
        .then(resp => {
            console.log("creating shr ret")
            shr = new model.Share(resp.shrToken||"", resp.frontendProxyEndpoints||[])
            console.log(shr)
        })
        .catch(resp => {
            console.log("unable to create share")
            throw new Error("unable to create share " + resp)
        })
    console.log("wat")
    console.log(shr)
    return shr
}

function newPrivateShare(root: Root, request: model.ShareRequest): ShareRequest {
    return {envZId: root.env.ZitiIdentity,
        shareMode: model.zrokShareModeToOpenApi(request.ShareMode),
        backendMode: model.zrokBackendModeToOpenApi(request.BackendMode),
        backendProxyEndpoint: request.Target,
        authScheme: model.AUTH_SCHEME_NONE}
}

function newPublicShare(root: Root, request: model.ShareRequest): ShareRequest {
    return {envZId: root.env.ZitiIdentity,
        shareMode: model.zrokShareModeToOpenApi(request.ShareMode),
        backendMode: model.zrokBackendModeToOpenApi(request.BackendMode),
        backendProxyEndpoint: request.Target,
        authScheme: model.AUTH_SCHEME_NONE,
        oauthProvider: model.zrokOauthProviderToOpenApi(request.OauthProvider),
        oauthEmailDomains: request.OauthEmailDomains,
        oauthAuthorizationCheckInterval: request.OauthAuthorizationCheckInterval}
}

export function DeleteShare(root: Root, shr: model.Share) {
    let client = new ShareApi(root.Client())
    let req: UnshareRequest = {
        envZId: root.env.ZitiIdentity,
        shrToken: shr.Token,
    }
    req.envZId = root.env.ZitiIdentity
    client.unshare({body: {}})
        .catch(resp => {
            throw new Error("error deleting share " + resp)
        })
}