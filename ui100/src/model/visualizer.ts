import {Overview} from "../api";
import {Edge, Node} from "@xyflow/react";

export class VisualOverview {
    nodes: Node[];
    edges: Edge[];
}

const buildVisualizerGraph = (overview: Overview): VisualOverview => {
    let out = new VisualOverview();
    out.nodes = [
        { id: "1", position: { x: 0, y: 0 }, data: { label: "michael@quigley.com" } }
    ];
    out.edges = [];

    overview.environments?.forEach((env, i) => {
        out.nodes.push({
            id: (i+2).toString(),
            position: { x: 0, y: 0 },
            data: { label: env.environment?.description! },
        });
        out.edges.push({
            id: "e" + (i+2) + "-1",
            source: "1",
            target: (i+2).toString()
        });
    })

    console.log(out);

    return out;
}

export default buildVisualizerGraph;