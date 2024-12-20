import "@xyflow/react/dist/style.css";
import "./styling/react-flow.css";
import {
    Background,
    Controls,
    MiniMap,
    Node,
    ReactFlow,
    ReactFlowProvider,
    useEdgesState,
    useNodesState
} from "@xyflow/react";
import {VisualOverview} from "./model/visualizer.ts";
import {useEffect} from "react";
import {stratify, tree} from "d3-hierarchy";
import ShareNode from "./ShareNode.tsx";
import EnvironmentNode from "./EnvironmentNode.tsx";
import AccountNode from "./AccountNode.tsx";
import AccessNode from "./AccessNode.tsx";
import {Box} from "@mui/material";

interface VisualizerProps {
    vov: VisualOverview;
    onSelectionChanged: (node: Node) => void;
}

const nodeTypes = {
    access: AccessNode,
    account: AccountNode,
    environment: EnvironmentNode,
    share: ShareNode
};

const Visualizer = ({ vov, onSelectionChanged }: VisualizerProps) => {
    const [nodes, setNodes, onNodesChange] = useNodesState([]);
    const [edges, setEdges, onEdgesChange] = useEdgesState([]);

    const onSelectionChange = ({ nodes }) => {
        if(nodes.length > 0) {
            onSelectionChanged(nodes[0]);
        } else {
            onSelectionChanged(null as Node);
        }
    };

    const layout = (nodes, edges): VisualOverview => {
        if(!nodes) {
            return { nodes: [], edges: [] };
        }
        let g = tree();
        if(nodes.length === 0) return { nodes, edges };
        const width = 100;
        const height = 75;
        const hierarchy = stratify()
            .id((node) => node.id)
            .parentId((node) => edges.find((edge) => edge.target === node.id)?.source);
        const root = hierarchy(nodes);
        const layout = g.nodeSize([width * 2, height * 2])(root);
        return {
            nodes: layout
                .descendants()
                .map((node) => ({...node.data, position: {x: node.x, y: node.y}})),
            edges,
        } as VisualOverview
    }

    const nodeColor = (node) => {
        if(node.selected) {
            return "#9bf316";
        }
        return "#241775";
    }

    useEffect(() => {
        if(vov) {
            let laidOut = layout(vov.nodes, vov.edges);
            setNodes(laidOut.nodes);
            setEdges(laidOut.edges);
        }
    }, [vov]);

    return (
        <ReactFlow
            nodeTypes={nodeTypes}
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onSelectionChange={onSelectionChange}
            nodesDraggable={false}
            fitView
        >
            <Background  />
            <Controls position="bottom-left" orientation="horizontal" showInteractive={false} />
            <MiniMap
                nodeColor={nodeColor}
                maskColor="rgb(36, 23, 117, 0.5)"
                pannable={true}
            />
        </ReactFlow>
    );
}

export default ({ vov, onSelectionChanged }: VisualizerProps) => {
    return (
        <Box sx={{ width: "100%", mt: 2 }} height={{ xs: 400, sm: 600, md: 800 }}>
            <ReactFlowProvider>
                <Visualizer vov={vov} onSelectionChanged={onSelectionChanged} />
            </ReactFlowProvider>
        </Box>
    );
}