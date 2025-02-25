import {Root} from "../environment/root"
import {
    ShareApi,
    AccessRequest,
    Access201Response,
    UnaccessRequest} from "./api/api"
import * as model from "./model"

export async function CreateAccess(root: Root, request: model.AccessRequest): Promise<model.Access> {
    if (!root.IsEnabled()){
        throw new Error("environment is not enabled; enable with 'zrok enable' first!")
    }

    let out: AccessRequest = {
        envZId: root.env.ZitiIdentity,
        shareToken: request.ShareToken
    }
    
    let conf = await root.Client()
    let client = new ShareApi(conf)
    let shr = await client.access({body: out})
        .catch(resp => {
            throw new Error("unable to create access " + resp)
        })

    if (shr.frontendToken == undefined) {
        throw new Error("expected frontend token from access. Got none")
    }
    if (shr.backendMode == undefined) {
        throw new Error("expected backend mode from access. Got none")
    }

    return new model.Access(shr.frontendToken, request.ShareToken, shr.backendMode)
}

export async function DeleteAccess(root: Root, acc: model.Access): Promise<void> {
    let out: UnaccessRequest = {
        frontendToken: acc.Token,
        shareToken: acc.ShareToken,
        envZId: root.env.ZitiIdentity
    }
    let conf = await root.Client()
    let client = new ShareApi(conf)

    return client.unaccess({body:out})
        .catch(resp => {
            throw new Error("error deleting access " + resp)
        })
}