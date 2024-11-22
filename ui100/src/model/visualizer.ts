import {Overview} from "../api";

const buildVisualizerGraph = (overview: Overview) => {
    let nodes = [
        { id: "1", label: "michael@quigley.com" }
    ]
    let edges = [];
    overview.environments?.forEach((env, i) => {
        nodes.push({ id: (i+1).toString(), label: env.environment?.description! });
        edges.push({ source: (i+1).toString(), target: "1", id: (i+1)+"-1" });
    })
    return {
        nodes: nodes,
        edges: edges,
    }
}

export default buildVisualizerGraph;