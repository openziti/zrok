import {Root} from "./environment";
import {ShareApi} from "../api";
import {Access, AccessRequest} from "./model";

export const createAccess = async (root: Root, request: AccessRequest): Promise<Access> => {
    if (!root.isEnabled()) {
        throw new Error("environment is not enabled; enable with 'zrok2 enable' first!");
    }

    const cfg = await root.client();
    const acc = await new ShareApi(cfg).access({body: {
            envZId: root.environment?.zId,
            shareToken: request.shareToken,
            bindAddress: request.bindAddress,
        }});

    return new Access(acc.frontendToken!, request.shareToken, acc.backendMode!);
}

export const deleteAccess = async (root: Root, acc: Access): Promise<void> => {
    if (!root.isEnabled()) {
        throw new Error("environment is not enabled; enable with 'zrok2 enable' first!");
    }
    const cfg = await root.client();
    await new ShareApi(cfg).unaccess({body: {
            envZId: root.environment?.zId,
            shareToken: acc.shareToken,
            frontendToken: acc.frontendToken
        }});
}
