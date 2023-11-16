import {Root} from "../environment/root"
import {accessRequest, unaccessRequest, authUser} from "./api/types"
import {access, unaccess} from "./api/share"
import * as model from "./model"

export function CreateAccess(root, request) {
    if (!root.IsEnabled()){
        throw new Error("environment is not enabled; enable with 'zrok enable' first!")
    }

    out = accessRequest(request.ShareToken, root.env.ZitiIdentity)
    root.Client()
    access({body:out})
        .catch(resp => {
            throw new Error("unable to create access", resp)
        })
}

export function DeleteAccess(root, acc) {
    if (!root.IsEnabled()){
        throw new Error("environment is not enabled; enable with 'zrok enable' first!")
    }

    req = unaccessRequest(acc.Token,acc.ShareToken, root.env.ZitiIdentity)
    unaccess({body:req})
        .catch(resp => {
            throw new Error("error deleting access", resp)
        })
}