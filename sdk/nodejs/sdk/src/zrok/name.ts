import {Root} from "./environment";
import {ShareApi} from "../api";
import {NameEntry, Namespace} from "./model";

export const createName = async (root: Root, name: string, namespaceToken?: string): Promise<NameEntry> => {
    if (!root.isEnabled()) {
        throw new Error("environment is not enabled; enable with 'zrok2 enable' first!");
    }

    const cfg = await root.client();
    await new ShareApi(cfg).createShareName({body: {namespaceToken, name}})
        .catch(err => {
            throw new Error("unable to create name '" + name + "': " + err);
        });

    return {
        namespaceToken: namespaceToken || "",
        namespaceName: "",
        name,
        shareToken: "",
        reserved: false,
        createdAt: 0,
    };
}

export const deleteName = async (root: Root, name: string, namespaceToken?: string): Promise<void> => {
    if (!root.isEnabled()) {
        throw new Error("environment is not enabled; enable with 'zrok2 enable' first!");
    }

    const cfg = await root.client();
    await new ShareApi(cfg).deleteShareName({body: {namespaceToken, name}})
        .catch(err => {
            throw new Error("unable to delete name '" + name + "': " + err);
        });
}

export const listNames = async (root: Root, namespaceToken?: string): Promise<NameEntry[]> => {
    if (!root.isEnabled()) {
        throw new Error("environment is not enabled; enable with 'zrok2 enable' first!");
    }

    const cfg = await root.client();
    const shareApi = new ShareApi(cfg);

    let names;
    try {
        if (namespaceToken) {
            names = await shareApi.listNamesForNamespace({namespaceToken});
        } else {
            names = await shareApi.listAllNames();
        }
    } catch (err) {
        throw new Error("unable to list names: " + err);
    }

    return names.map(n => ({
        namespaceToken: n.namespaceToken || "",
        namespaceName: n.namespaceName || "",
        name: n.name || "",
        shareToken: n.shareToken || "",
        reserved: n.reserved || false,
        createdAt: n.createdAt || 0,
    }));
}

export const listNamespaces = async (root: Root): Promise<Namespace[]> => {
    if (!root.isEnabled()) {
        throw new Error("environment is not enabled; enable with 'zrok2 enable' first!");
    }

    const cfg = await root.client();
    const res = await new ShareApi(cfg).listShareNamespaces()
        .catch(err => {
            throw new Error("unable to list namespaces: " + err);
        });

    return res.map(ns => ({
        namespaceToken: ns.namespaceToken || "",
        name: ns.name || "",
        description: ns.description || "",
    }));
}
