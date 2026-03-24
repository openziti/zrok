import {Root} from "./environment";
import {MetadataApi} from "../api";
import {ShareDetail, AccessDetail} from "./model";

export interface ListSharesFilters {
    envZId?: string;
    shareMode?: string;
    backendMode?: string;
    shareToken?: string;
    target?: string;
    permissionMode?: string;
    hasActivity?: boolean;
    idle?: boolean;
    activityDuration?: string;
    createdAfter?: string;
    createdBefore?: string;
    updatedAfter?: string;
    updatedBefore?: string;
}

export interface ListAccessesFilters {
    envZId?: string;
    shareToken?: string;
    bindAddress?: string;
    description?: string;
    createdAfter?: string;
    createdBefore?: string;
    updatedAfter?: string;
    updatedBefore?: string;
}

export const listShares = async (root: Root, filters?: ListSharesFilters): Promise<ShareDetail[]> => {
    if (!root.isEnabled()) {
        throw new Error("environment is not enabled; enable with 'zrok2 enable' first!");
    }

    const cfg = await root.client();
    const res = await new MetadataApi(cfg).listShares(filters || {})
        .catch(err => {
            throw new Error("unable to list shares: " + err);
        });

    const shares = res.shares || [];
    return shares.map(s => ({
        token: s.shareToken || "",
        zId: s.zId || "",
        envZId: s.envZId || "",
        shareMode: s.shareMode || "",
        backendMode: s.backendMode || "",
        frontendEndpoints: s.frontendEndpoints || [],
        target: s.target || "",
        limited: s.limited || false,
        createdAt: s.createdAt || 0,
        updatedAt: s.updatedAt || 0,
    }));
}

export const listAccesses = async (root: Root, filters?: ListAccessesFilters): Promise<AccessDetail[]> => {
    if (!root.isEnabled()) {
        throw new Error("environment is not enabled; enable with 'zrok2 enable' first!");
    }

    const cfg = await root.client();
    const res = await new MetadataApi(cfg).listAccesses(filters || {})
        .catch(err => {
            throw new Error("unable to list accesses: " + err);
        });

    const accesses = res.accesses || [];
    return accesses.map(a => ({
        id: a.id || 0,
        frontendToken: a.frontendToken || "",
        envZId: a.envZId || "",
        shareToken: a.shareToken || "",
        backendMode: a.backendMode || "",
        bindAddress: a.bindAddress || "",
        description: a.description || "",
        limited: a.limited || false,
        createdAt: a.createdAt || 0,
        updatedAt: a.updatedAt || 0,
    }));
}
