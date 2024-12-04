import "@xyflow/react/dist/style.css";
import "./react-flow.css";
import {Background, Controls, ReactFlow, ReactFlowProvider, useEdgesState, useNodesState} from "@xyflow/react";
import {VisualOverview} from "./model/visualizer.ts";
import {useEffect} from "react";
import {stratify, tree} from "d3-hierarchy";

interface VisualizerProps {
    overview: VisualOverview;
}

const Visualizer = ({ overview }: VisualizerProps) => {
    const [nodes, setNodes, onNodesChange] = useNodesState([]);
    const [edges, setEdges, onEdgesChange] = useEdgesState([]);

    const layout = (nodes, edges): VisualOverview => {
        if(!nodes) {
            return { nodes: [], edges: [] };
        }
        let g = tree();
        if(nodes.length === 0) return { nodes, edges };
        // const { width, height } = document.querySelector(`[data-id="$nodes[0].id"]`).getBoundingClientRect();
        const width = 100;
        const height = 40;
        const hierarchy = stratify()
            .id((node) => node.id)
            .parentId((node) => edges.find((edge) => edge.target === node.id)?.source);
        const root = hierarchy(nodes);
        const layout = g.nodeSize([width * 2, height * 2])(root);
        return {
            nodes: layout
                .descendants()
                .map((node) => ({ ...node.data, position: { x: node.x, y: node.y }})),
            edges,
        }
    }

    useEffect(() => {
        let layouted = layout(overview.nodes, overview.edges);
        setNodes(layouted.nodes);
        setEdges(layouted.edges);
    }, [overview]);

    return (
        <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            fitView
        >
            <Background/>
            <Controls />
        </ReactFlow>
    );
}

export default ({ overview }: VisualizerProps) => {
    return (
        <div style={{ height: "400px" }}>
            <ReactFlowProvider>
                <Visualizer overview={overview}/>
            </ReactFlowProvider>
        </div>
    );
}