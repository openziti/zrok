import {Overview} from "../api";
import {Edge, Node} from "@xyflow/react";

export class VisualOverview {
    nodes: Node[];
    edges: Edge[];
}

const buildVisualizerGraph = (overview: Overview): VisualOverview => {
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

export default buildVisualizerGraph;