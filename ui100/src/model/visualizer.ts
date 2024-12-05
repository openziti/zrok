import {Overview} from "../api";
import {Edge, Node} from "@xyflow/react";

export class VisualOverview {
    nodes: Node[];
    edges: Edge[];
}

const buildVisualizerGraph = (overview: Overview): VisualOverview => {
    let out = new VisualOverview();
    out.nodes = [
        { id: "0", position: { x: 0, y: 0 }, data: { label: "michael@quigley.com" } }
    ];
    out.edges = [];

    overview.environments?.forEach((env, i) => {
        if(env.environment && env.environment.zId) {
            out.nodes.push({
                id: env.environment.zId,
                position: { x: 0, y: 0 },
                data: { label: env.environment?.description! },
            });
            out.edges.push({
                id: env.environment.zId + "-0",
                source: "0",
                target: env.environment.zId
            });

        }
    })

    console.log(out);

    return out;
}

export default buildVisualizerGraph;