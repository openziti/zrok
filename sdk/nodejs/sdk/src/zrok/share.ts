import {Root} from "../environment/root"
import {
    Share,
    ShareApi,
    ShareRequest,
    ShareResponse,
    AuthUser,
    UnshareRequest} from "./api/api"
import * as model from "./model"

export async function CreateShare(root: Root, request: model.ShareRequest): Promise<model.Share> {
    if (!root.IsEnabled()){
        throw new Error("environment is not enabled; enable with 'zrok enable' first!")
    }
    let out: ShareRequest

    switch(request.ShareMode) {
        case ShareRequest.ShareModeEnum.Private:
            out = newPrivateShare(root, request)
            break
        case ShareRequest.ShareModeEnum.Public:
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

    let conf = await root.Client()
    let client = new ShareApi(conf)
    let shr = await client.share({body: out})
        .catch(resp => {
            throw new Error("unable to create share " + resp)
        })
    return new model.Share(shr.shareToken||"", shr.frontendProxyEndpoints||[])
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
        frontendSelection: request.Frontends,
        backendMode: model.zrokBackendModeToOpenApi(request.BackendMode),
        backendProxyEndpoint: request.Target,
        authScheme: model.AUTH_SCHEME_NONE,
        oauthProvider: model.zrokOauthProviderToOpenApi(request.OauthProvider),
        oauthEmailDomains: request.OauthEmailDomains,
        oauthAuthorizationCheckInterval: request.OauthAuthorizationCheckInterval}
}

export async function DeleteShare(root: Root, shr: model.Share): Promise<void> {
    let conf = await root.Client()
    let client = new ShareApi(conf)
    let req: UnshareRequest = {
        envZId: root.env.ZitiIdentity,
        shareToken: shr.Token,
    }
    req.envZId = root.env.ZitiIdentity
    return client.unshare({body: req})
        .catch(resp => {
            throw new Error("error deleting share " + resp)
        })
}