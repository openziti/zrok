import {Overview} from "../api";
import {Edge, Node} from "@xyflow/react";
import {User} from "./user.ts";
import {stratify, tree} from "d3-hierarchy";

export class Graph {
    nodes: Node[];
    edges: Edge[];
}

export const mergeGraph = (oldVov: Graph, u: User, limited: boolean, newOv: Overview): Graph => {
    let newVov = new Graph();

    let accountNode = {
        id: u.email,
        data: {
            label: u.email,
            limited: limited
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
                    limited: limited,
                    empty: true
                },
                type: "environment",
                position: { x: 0, y: 0 },
            };
            newVov.nodes.push(envNode);
            newVov.edges.push({
                id: accountNode.id + "-" + envNode.id,
                source: accountNode.id!,
                target: envNode.id!,
                type: "hierarchy"
            });
            if(env.shares) {
                envNode.data.empty = false;
                env.shares.forEach(shr => {
                    let shrLabel = shr.shareToken!;
                    if(shr.target !== "") {
                        shrLabel = shr.target!;
                    }
                    let shrNode = {
                        id: shr.shareToken!,
                        data: {
                            label: shrLabel,
                            shareToken: shr.shareToken!,
                            envZId: env.environment?.zId!,
                            limited: limited,
                            accessed: false,
                        },
                        type: "share",
                        position: { x: 0, y: 0 }
                    }
                    allShares[shr.shareToken!] = shrNode;
                    newVov.nodes.push(shrNode);
                    newVov.edges.push({
                        id: envNode.id + "-" + shrNode.id,
                        source: envNode.id!,
                        target: shrNode.id!,
                        type: "hierarchy"
                    });
                });
            }
            if(env.frontends) {
                envNode.data.empty = false;
                env.frontends.forEach(fe => {
                    let feNode = {
                        id: fe.frontendToken!,
                        data: {
                            label: fe.frontendToken!,
                            feId: fe.id,
                            target: fe.shareToken,
                            bindAddress: fe.bindAddress,
                            backendMode: fe.backendMode,
                            envZId: fe.zId,
                        },
                        type: "access",
                        position: { x: 0, y: 0 }
                    }
                    allFrontends.push(feNode);
                    newVov.nodes.push(feNode);
                    newVov.edges.push({
                        id: envNode.id + "-" + feNode.id,
                        source: envNode.id!,
                        target: feNode.id!,
                        type: "hierarchy"
                    });
                });
            }
        });
        allFrontends.forEach(fe => {
            let target = allShares[fe.data.target];
            if(target) {
                target.data.accessed = true;
                fe.data.ownedShare = true;
                let edge: Edge = {
                    id: target.id + "-" + fe.id,
                    source: fe.id!,
                    sourceHandle: "share",
                    target: target.id!,
                    targetHandle: "access",
                    type: "access",
                    animated: true
                }
                newVov.edges.push(edge);
            }
        });
    }
    newVov.nodes = sortNodes(newVov.nodes);

    if(nodesEqual(oldVov.nodes, newVov.nodes)) {
        // if the list of nodes is equal, the graph hasn't changed; we can just return the oldGraph and save the
        // physics headaches in the visualizer.
        return oldVov;
    }

    let outNodes = [];
    if(oldVov.nodes) {
        outNodes = oldVov.nodes.filter(oldNode => newVov.nodes.find(newNode => newNode.id === oldNode.id
            && newNode.data.accessed === oldNode.data.accessed
            && newNode.data.ownedShare === oldNode.data.ownedShare
            && newNode.data.limited === oldNode.data.limited
            && newNode.data.label === oldNode.data.label));
    }
    let outEdges = [];
    if(oldVov.edges) {
        outEdges = oldVov.edges.filter(oldEdge => newVov.edges.find(newEdge => newEdge.target === oldEdge.target
            && newEdge.source === oldEdge.source));
    }

    // and then do the opposite; add any nodes that are in the new overview, but missing from the old overview.
    outNodes.push(...newVov.nodes.filter(newNode => !outNodes.find(oldNode => oldNode.id === newNode.id
        && oldNode.data.accessed == newNode.data.accessed
        && oldNode.data.ownedShare === newNode.data.ownedShare
        && oldNode.data.limited === newNode.data.limited
        && oldNode.data.label === newNode.data.label)));
    outEdges.push(...newVov.edges.filter(newEdge => !outEdges.find(oldEdge => oldEdge.target === newEdge.target
        && oldEdge.source === newEdge.source)));

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

export const nodesEqual = (a: Node[], b: Node[]) => {
    if(!a && !b) return true;
    if(a && !b) return false;
    if(b && !a) return false;
    if(a.length !== b.length) return false;
    return a.every((e, i) => e.id === b[i].id && e.data.limited === b[i].data.limited && e.data.label === b[i].data.label);
}

export const focusGraph = (graph: Graph, focusNodeId: string): Graph => {
    let nodeMap = new Map<string, Node>();
    for(let n of graph.nodes) {
        nodeMap.set(n.id, n);
    }

    let parentOf = new Map<string, string>();
    for(let e of graph.edges) {
        if(e.type === "hierarchy") {
            parentOf.set(e.target, e.source);
        }
    }

    let childrenOf = new Map<string, string[]>();
    for(let e of graph.edges) {
        if(e.type === "hierarchy") {
            let list = childrenOf.get(e.source) || [];
            list.push(e.target);
            childrenOf.set(e.source, list);
        }
    }

    let focusNode = nodeMap.get(focusNodeId);
    if(!focusNode) return graph;

    let included = new Set<string>();

    let addWithParents = (id: string) => {
        let cur = id;
        while(cur) {
            included.add(cur);
            cur = parentOf.get(cur);
        }
    };

    if(focusNode.type === "account") {
        return graph;
    } else if(focusNode.type === "environment") {
        addWithParents(focusNodeId);
        let children = childrenOf.get(focusNodeId) || [];
        for(let childId of children) {
            included.add(childId);
            let child = nodeMap.get(childId);
            if(child?.type === "access") {
                for(let e of graph.edges) {
                    if(e.type === "access" && e.source === childId) {
                        included.add(e.target);
                        addWithParents(e.target);
                    }
                }
            }
        }
    } else if(focusNode.type === "share") {
        addWithParents(focusNodeId);
        for(let e of graph.edges) {
            if(e.type === "access" && e.target === focusNodeId) {
                included.add(e.source);
                addWithParents(e.source);
            }
        }
    } else if(focusNode.type === "access") {
        addWithParents(focusNodeId);
        for(let e of graph.edges) {
            if(e.type === "access" && e.source === focusNodeId) {
                included.add(e.target);
                addWithParents(e.target);
            }
        }
    }

    let out = new Graph();
    out.nodes = graph.nodes.filter(n => included.has(n.id));
    out.edges = graph.edges.filter(e => included.has(e.source) && included.has(e.target));
    return out;
}

export const layout = (nodes, edges): Graph => {
    if(!nodes || nodes.length === 0) {
        return { nodes: nodes || [], edges };
    }

    const hierarchyEdges = edges.filter((edge) => edge.type !== "access");

    // compute spacing from measured node dimensions if available
    let maxWidth = 0;
    let maxHeight = 0;
    let hasMeasurements = false;
    for(const node of nodes) {
        if(node.measured?.width && node.measured?.height) {
            hasMeasurements = true;
            maxWidth = Math.max(maxWidth, node.measured.width);
            maxHeight = Math.max(maxHeight, node.measured.height);
        }
    }

    // use actual node sizes + padding, or compact defaults before first measurement
    const nodeWidth = hasMeasurements ? maxWidth + 10 : 120;
    const nodeHeight = hasMeasurements ? (maxHeight + 60) * 2 : 260;

    const hierarchy = stratify()
        .id((node) => node.id)
        .parentId((node) => hierarchyEdges.find((edge) => edge.target === node.id)?.source);

    const root = hierarchy(nodes);

    // sort children: shares left, accesses right
    root.each((node) => {
        if(node.children) {
            node.children.sort((a, b) => {
                const aType = a.data.type || "";
                const bType = b.data.type || "";
                if(aType === "share" && bType === "access") return -1;
                if(aType === "access" && bType === "share") return 1;
                return (a.data.id || "").localeCompare(b.data.id || "");
            });
        }
    });

    const g = tree()
        .nodeSize([nodeWidth, nodeHeight])
        .separation(() => 1);
    const laid = g(root);

    // assign lane indices to access edges for distinct routing
    let accessEdgeIndex = 0;
    const laneCount = edges.filter(e => e.type === "access").length;
    const indexedEdges = edges.map((edge) => {
        if(edge.type === "access") {
            return { ...edge, data: { ...edge.data, laneIndex: accessEdgeIndex++, laneCount } };
        }
        return edge;
    });

    return {
        nodes: laid.descendants()
            .map((node) => ({...node.data, position: {x: node.x, y: node.y}})),
        edges: indexedEdges,
    } as Graph;
}