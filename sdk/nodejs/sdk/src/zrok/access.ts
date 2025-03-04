import {BackendMode} from "./share";
import {Root} from "./environment";
import {instanceOfCreateFrontend201Response, ShareApi} from "../api";

export class AccessRequest {
    shareToken: string;

    constructor(shareToken: string) {
        this.shareToken = shareToken;
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

export const createAccess = async (root: Root, request: AccessRequest): Promise<Access> => {
    if(!root.isEnabled()) {
        throw new Error("environment is not enabled; enable with 'zrok enable' first!");
    }

    let acc = await new ShareApi(root.apiConfiguration()).access({body: {
            envZId: root.environment?.zId,
            shareToken: request.shareToken
        }});

    return new Access(acc.frontendToken!, request.shareToken, acc.backendMode!);
}

export const deleteAccess = async (root: Root, acc: Access): Promise<any> => {
    if(!root.isEnabled()) {
        throw new Error("environment is not enabled; enable with 'zrok enable' first!");
    }
    return new ShareApi(root.apiConfiguration()).unaccess({body: {
            envZId: root.environment?.zId,
            shareToken: acc.shareToken,
            frontendToken: acc.frontendToken
        }});
}