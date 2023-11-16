import {Root} from "../environment/root"
import {shareRequest, unshareRequest, authUser} from "./api/types"
import {share, unshare} from "./api/share"
import * as model from "./model"

function CreateShare(root, request) {
    if (!root.IsEnabled()){
        throw new Error("environment is not enabled; enable with 'zrok enable' first!")
    }
    switch(request.shareMode) {
        case model.PRIVATE_SHARE_MODE:
            out = newPrivateShare(root, request)
            break
        case model.PUBLIC_SHARE_MODE:
            out = newPublicShare(root, request)
            break
        default:
            throw new Error("unknown share mode " + request.shareMode)
    }

    if (request.basicAuth.length > 0) {
        out.auth_scheme = model.AUTH_SCHEME_BASIC
        for(pair in request.basicAuth) {
            tokens = pair.split(":")
            if (tokens.length === 2) {
                out.auth_users.push(authUser(tokens[0].strip(), tokens[1].strip()))
            }
            else {
                throw new Error("invalid username:password pair: " + pair)
            }
        }
    }

    if (request.oauthProvider !== "") {
        out.auth_scheme = model.AUTH_SCHEME_OAUTH
    }
    console.log(out)
    root.Client()
    share({body: out})
        .catch(resp => {
            throw new Error("unable tp create share", resp)
        })
}

function newPrivateShare(root, request) {
    return shareRequest(root.env.ZitiIdentity,
            request.shareMode,
            request.backendMode,
            request.target,
            model.AUTH_SCHEME_NONE)
}

function newPublicShare(root, request) {
    return shareRequest(root.env.ZitiIdentity,
        request.shareMode,
        request.backendMode,
        request.target,
        model.AUTH_SCHEME_NONE,
        request.oauthEmailDomains,
        request.oauthProvider,
        request.oauthAuthroizationCheckInterval)
}

function DeleteShare(root, shr) {
    req = unshareRequest(root.env.ZitiIdentity, shr.Token)
    root.Client()
    unshare({body:req})
        .catch(resp => {
            throw new Error("error deleting share", resp)
        })
}