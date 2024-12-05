import {Overview} from "../api";
import {Edge, Node} from "@xyflow/react";
import ShareIcon from "@mui/icons-material/Share";

export class VisualOverview {
    nodes: Node[];
    edges: Edge[];
}

const buildVisualizerGraph = (overview: Overview): VisualOverview => {
    let out = new VisualOverview();
    out.nodes = [
        { id: "0", position: { x: 0, y: 0 }, data: { label: "michael@quigley.com" }, type: "input" }
    ];
    out.edges = [];

    overview.environments?.forEach(env => {
        if(env.environment && env.environment.zId) {
            let envNode = {
                id: env.environment.zId,
                position: { x: 0, y: 0 },
                data: { label: env.environment?.description! },
                type: "output",
            }
            out.nodes.push(envNode);
            out.edges.push({
                id: env.environment.zId + "-0",
                source: "0",
                target: env.environment.zId
            });
            if(env.shares) {
                envNode.type = "default";
                env.shares.forEach(shr => {
                    out.nodes.push({
                        id: shr.token!,
                        position: { x: 0, y: 0 },
                        data: { label: shr.token! },
                        type: "output",
                    });
                    out.edges.push({
                        id: env.environment?.zId + "-" + shr.token!,
                        source: env.environment?.zId!,
                        target: shr.token!
                    });
                });
            }
        }
    });

    return out;
}

export default buildVisualizerGraph;