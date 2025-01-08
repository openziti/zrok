import "@xyflow/react/dist/style.css";
import "./styling/react-flow.css";
import {
    applyNodeChanges,
    Background,
    Controls,
    MiniMap,
    Node,
    ReactFlow,
    ReactFlowProvider,
    useOnViewportChange,
    Viewport
} from "@xyflow/react";
import {VisualOverview} from "./model/visualizer.ts";
import {useEffect} from "react";
import {stratify, tree} from "d3-hierarchy";
import ShareNode from "./ShareNode.tsx";
import EnvironmentNode from "./EnvironmentNode.tsx";
import AccountNode from "./AccountNode.tsx";
import AccessNode from "./AccessNode.tsx";
import {Box} from "@mui/material";
import useStore from "./model/store.ts";

const nodeTypes = {
    access: AccessNode,
    account: AccountNode,
    environment: EnvironmentNode,
    share: ShareNode
};

const Visualizer = () => {
    const overview = useStore((state) => state.overview);
    const selectedNode = useStore((state) => state.selectedNode);
    const updateSelectedNode = useStore((state) => state.updateSelectedNode);
    const viewport = useStore((state) => state.viewport);
    const updateViewport = useStore((state) => state.updateViewport);
    const nodes = useStore((state) => state.nodes);
    const updateNodes = useStore((state) => state.updateNodes);
    const edges = useStore((state) => state.edges);
    const updateEdges = useStore((state) => state.updateEdges);

    const onNodesChange = (changes) => {
        updateNodes(applyNodeChanges(changes, nodes));
    }

    useOnViewportChange({
        onEnd: (viewport: Viewport) => {
            updateViewport(viewport);
        }
    });

    const onSelectionChange = ({ nodes }) => {
        if(nodes.length > 0) {
            updateSelectedNode(nodes[0]);
        } else {
            updateSelectedNode(null as Node);
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
        if(overview) {
            let laidOut = layout(overview.nodes, overview.edges);
            let selected = laidOut.nodes.map((n) => ({
                ...n,
                selected: selectedNode ? selectedNode.id === n.id : false,
            }));
            updateNodes(selected);
            updateEdges(laidOut.edges);
        }
    }, [overview]);

    let fitView = false;
    if(viewport.x === 0 && viewport.y === 0 && viewport.zoom === 1) {
        fitView = true;
    }

    return (
        <ReactFlow
            nodeTypes={nodeTypes}
            nodes={nodes}
            onNodesChange={onNodesChange}
            edges={edges}
            onSelectionChange={onSelectionChange}
            nodesDraggable={false}
            defaultViewport={viewport}
            fitView={fitView}
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

export default () => {
    return (
        <Box sx={{ width: "100%", mt: 2 }} height={{ xs: 400, sm: 600, md: 800 }}>
            <ReactFlowProvider>
                <Visualizer />
            </ReactFlowProvider>
        </Box>
    );
}