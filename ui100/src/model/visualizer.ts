import {Overview} from "../api";
import {Edge, Node} from "@xyflow/react";
import {User} from "./user.ts";

export class VisualOverview {
    nodes: Node[];
    edges: Edge[];
}

export const mergeVisualOverview = (oldVov: VisualOverview, u: User, limited: boolean, newOv: Overview): VisualOverview => {
    let newVov = new VisualOverview();

    let accountNode = {
        id: u.token,
        data: {
            label: u.email,
            limited: !!limited
        },
        type: "account",
        position: { x: 0, y: 0 },
    }
    newVov.nodes = [ accountNode ];
    newVov.edges = [];

    if(newOv) {
        let allShares = {};
        let allFrontends = [];
        newOv.environments?.forEach(env => {
            let envNode = {
                id: env.environment?.zId!,
                data: {
                    label: env.environment?.description,
                    envZId: env.environment?.zId!,
                    limited: !!limited,
                    empty: true
                },
                type: "environment",
                position: { x: 0, y: 0 },
            };
            newVov.nodes.push(envNode);
            newVov.edges.push({
                id: accountNode.id + "-" + envNode.id,
                source: accountNode.id!,
                target: envNode.id!
            });
            if(env.shares) {
                env.shares.forEach(shr => {
                    let shrLabel = shr.token!;
                    if(shr.backendProxyEndpoint !== "") {
                        shrLabel = shr.backendProxyEndpoint!;
                    }
                    let shrNode = {
                        id: shr.token!,
                        data: {
                            label: shrLabel,
                            shrToken: shr.token!,
                            envZId: env.environment?.zId!,
                            limited: !!limited,
                            accessed: false,
                        },
                        type: "share",
                        position: { x: 0, y: 0 }
                    }
                    allShares[shr.token!] = shrNode;
                    newVov.nodes.push(shrNode);
                    newVov.edges.push({
                        id: envNode.id + "-" + shrNode.id,
                        source: envNode.id!,
                        target: shrNode.id!
                    });
                });
            }
            if(env.frontends) {
                env.frontends.forEach(fe => {
                    let feNode = {
                        id: fe.token!,
                        data: {
                            label: fe.token!,
                            feId: fe.id,
                            target: fe.shrToken,
                        },
                        type: "access",
                        position: { x: 0, y: 0 }
                    }
                    allFrontends.push(feNode);
                    newVov.nodes.push(feNode);
                    newVov.edges.push({
                        id: envNode.id + "-" + feNode.id,
                        source: envNode.id!,
                        target: feNode.id!
                    });
                });
            }
        });
        allFrontends.forEach(fe => {
            let target = allShares[fe.data.target];
            if(target) {
                newVov.edges.push({
                    id: target.id + "-" + fe.id,
                    source: target.id!,
                    target: fe.id!
                });
            }
        });
    }
    newVov.nodes = sortNodes(newVov.nodes);

    if(nodesEqual(oldVov.nodes, newVov.nodes)) {
        // if the list of nodes is equal, the graph hasn't changed; we can just return the oldGraph and save the
        // physics headaches in the visualizer.
        return oldVov;
    }

    let outNodes = oldVov.nodes.filter(oldNode => newVov.nodes.find(newNode => newNode.id === oldNode.id && newNode.data.limited == oldNode.data.limited && newNode.data.label === oldNode.data.label));
    let outEdges = oldVov.edges.filter(oldEdge => newVov.edges.find(newEdge => newEdge.target === oldEdge.target && newEdge.source === oldEdge.source));

    outNodes.push(...newVov.nodes.filter(newNode => !outNodes.find(oldNode => oldNode.id === newNode.id && oldNode.data.limited === newNode.data.limited && oldNode.data.label === newNode.data.label)));
    outEdges.push(...newVov.edges.filter(newEdge => !outEdges.find(oldEdge => oldEdge.target === newEdge.target && oldEdge.source === newEdge.source)));

    newVov.nodes = outNodes;
    newVov.edges = outEdges;
    return newVov;
}

const sortNodes = (nodes) => {
    return nodes.sort((a, b) => {
        if(a.id > b.id) {
            return 1;
        }
        if(a.id < b.id) {
            return -1;
        }
        return 0;
    });
}

const nodesEqual = (a, b) => {
    if(a.length !== b.length) return false;
    return a.every((e, i) => e.id === b[i].id && e.data.limited === b[i].data.limited && e.data.label === b[i].data.label);
}

export const buildVisualOverview = (overview: Overview): VisualOverview => {
    let out = new VisualOverview();
    out.nodes = [
        { id: "0", position: { x: 0, y: 0 }, data: { label: "michael@quigley.com" }, type: "account" }
    ];
    out.edges = [];

    let ownedShares: { [key: string]: Node } = {};

    overview.environments?.forEach(env => {
        if(env.environment && env.environment.zId) {
            let envNode = {
                id: env.environment.zId,
                position: { x: 0, y: 0 },
                data: { label: env.environment?.description!, empty: true },
                type: "environment",
            }
            out.nodes.push(envNode);
            out.edges.push({
                id: env.environment.zId + "-0",
                source: "0",
                target: env.environment.zId
            });

            if(env.shares) {
                envNode.data.empty = false;
                env.shares.forEach(shr => {
                    let shareNode = {
                        id: shr.token!,
                        position: { x: 0, y: 0 },
                        data: { label: shr.token!, accessed: false },
                        type: "share",
                    };
                    out.nodes.push(shareNode);
                    ownedShares[shr.token!] = shareNode;
                    out.edges.push({
                        id: env.environment?.zId + "-" + shr.token!,
                        source: env.environment?.zId!,
                        target: shr.token!
                    });
                });
            }
            if(env.frontends) {
                envNode.data.empty = false;
                env.frontends.forEach(acc => {
                    let accNode = {
                        id: acc.token!,
                        position: { x: 0, y: 0 },
                        data: { label: acc.token!, ownedShare: false, shrToken: acc.shrToken },
                        type: "access",
                    }
                    out.nodes.push(accNode);
                    out.edges.push({
                        id: env.environment?.zId + "-" + acc.token!,
                        source: env.environment?.zId!,
                        target: acc.token!
                    });
                    let ownedShare = ownedShares[acc.shrToken];

                });
            }
        }
    });
    out.nodes.forEach(n => {
        if(n.type == "access") {
            let ownedShare = ownedShares[n.data.shrToken];
            if(ownedShare) {
                console.log("linking owned share", n)
                n.data.ownedShare = true;
                ownedShare.data.accessed = true;
                out.edges.push({
                    id: n.id + "-" + n.data.shrToken,
                    source: n.id,
                    target: n.data.shrToken as string,
                    targetHandle: "access",
                    animated: true
                });
            }
        }
    })

    return out;
}

export const visualOverviewsEqual = (a: VisualOverview, b: VisualOverview): boolean => {
    return false;
}