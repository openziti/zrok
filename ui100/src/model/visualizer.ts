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
                    out.nodes.push({
                        id: shr.token!,
                        position: { x: 0, y: 0 },
                        data: { label: shr.token! },
                        type: "share",
                    });
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
                    out.nodes.push({
                        id: acc.token!,
                        position: { x: 0, y: 0 },
                        data: { label: acc.token! },
                        type: "access",
                    });
                    out.edges.push({
                        id: env.environment?.zId + "-" + acc.token!,
                        source: env.environment?.zId!,
                        target: acc.token!
                    });
                });
            }
        }
    });

    return out;
}

export default buildVisualizerGraph;