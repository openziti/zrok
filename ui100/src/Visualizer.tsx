import "@xyflow/react/dist/style.css";
import "./react-flow.css";
import {Background, Controls, MiniMap, ReactFlow, ReactFlowProvider, useEdgesState, useNodesState} from "@xyflow/react";
import {VisualOverview} from "./model/visualizer.ts";
import {useEffect} from "react";
import {stratify, tree} from "d3-hierarchy";
import ShareNode from "./ShareNode.tsx";
import EnvironmentNode from "./EnvironmentNode.tsx";
import AccountNode from "../AccountNode.tsx";
import AccessNode from "./AccessNode.tsx";
import {Minimize} from "@mui/icons-material";

interface VisualizerProps {
    vov: VisualOverview;
}

const nodeTypes = {
    access: AccessNode,
    account: AccountNode,
    environment: EnvironmentNode,
    share: ShareNode
};

const Visualizer = ({ vov }: VisualizerProps) => {
    const [nodes, setNodes, onNodesChange] = useNodesState([]);
    const [edges, setEdges, onEdgesChange] = useEdgesState([]);

    const layout = (nodes, edges): VisualOverview => {
        if(!nodes) {
            return { nodes: [], edges: [] };
        }
        let g = tree();
        if(nodes.length === 0) return { nodes, edges };
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
        let layouted = layout(vov.nodes, vov.edges);
        setNodes(layouted.nodes);
        setEdges(layouted.edges);
    }, [vov]);

    return (
        <ReactFlow
            nodeTypes={nodeTypes}
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            fitView
            nodesDraggable={false}
        >
            <Background/>
            <Controls position="bottom-right" orientation="horizontal" />
        </ReactFlow>
    );
}

export default ({ vov }: VisualizerProps) => {
    return (
        <div style={{ height: "400px" }}>
            <ReactFlowProvider>
                <Visualizer vov={vov}/>
                <MiniMap />
            </ReactFlowProvider>
        </div>
    );
}