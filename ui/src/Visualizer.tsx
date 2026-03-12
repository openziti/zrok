import "@xyflow/react/dist/style.css";
import "./styling/react-flow.css";
import {
    applyNodeChanges,
    Background,
    Controls,
    MiniMap,
    Node,
    NodeChange,
    ReactFlow,
    ReactFlowProvider,
    useNodesInitialized,
    useOnViewportChange,
    Viewport
} from "@xyflow/react";
import ShareNode from "./ShareNode.tsx";
import EnvironmentNode from "./EnvironmentNode.tsx";
import AccountNode from "./AccountNode.tsx";
import AccessNode from "./AccessNode.tsx";
import {Box} from "@mui/material";
import {useEffect, useRef} from "react";
import useApiConsoleStore from "./model/store.ts";
import {layout} from "./model/graph.ts";
import AccessEdge from "./AccessEdge.tsx";
import HierarchyEdge from "./HierarchyEdge.tsx";

const edgeTypes = {
    access: AccessEdge,
    hierarchy: HierarchyEdge,
};

const nodeTypes = {
    access: AccessNode,
    account: AccountNode,
    environment: EnvironmentNode,
    share: ShareNode
};

const Visualizer = () => {
    const updateSelectedNode = useApiConsoleStore((state) => state.updateSelectedNode);
    const viewport = useApiConsoleStore((state) => state.viewport);
    const updateViewport = useApiConsoleStore((state) => state.updateViewport);
    const selectedNode = useApiConsoleStore((state) => state.selectedNode);
    const focusNodeId = useApiConsoleStore((state) => state.focusNodeId);
    const nodes = useApiConsoleStore((state) => state.nodes);
    const updateNodes = useApiConsoleStore((state) => state.updateNodes);
    const edges = useApiConsoleStore((state) => state.edges);
    const updateEdges = useApiConsoleStore((state) => state.updateEdges);
    const nodesInitialized = useNodesInitialized();
    const prevInitialized = useRef(false);

    // re-layout once after React Flow measures node dimensions for tighter spacing
    useEffect(() => {
        if(nodesInitialized && !prevInitialized.current && nodes.length > 0) {
            const laidOut = layout(nodes, edges);
            updateNodes(laidOut.nodes.map((n) => ({
                ...n,
                selected: selectedNode ? selectedNode.id === n.id : false,
            })));
            updateEdges(laidOut.edges);
        }
        prevInitialized.current = nodesInitialized;
    }, [nodesInitialized]);

    const onNodesChange = (changes: NodeChange[]) => {
        updateNodes(applyNodeChanges(changes, nodes));
    }

    useOnViewportChange({
        onEnd: (viewport: Viewport) => {
            updateViewport(viewport);
        }
    });

    const onSelectionChange = ({ nodes }: { nodes: Node[] }) => {
        if(nodes.length > 0) {
            updateSelectedNode(nodes[0]);
        } else {
            updateSelectedNode(null);
        }
    };

    const nodeColor = (node: Node) => {
        if(node.selected) {
            return "#9bf316";
        }
        return "#241775";
    }

    let fitView = false;
    if(viewport.x === 0 && viewport.y === 0 && viewport.zoom === 1) {
        fitView = true;
    }

    return (
        <ReactFlow
            edgeTypes={edgeTypes}
            nodeTypes={nodeTypes}
            nodes={nodes}
            onNodesChange={onNodesChange}
            edges={edges}
            onSelectionChange={onSelectionChange}
            nodesDraggable={false}
            nodesConnectable={false}
            defaultViewport={viewport}
            fitView={fitView}
        >
            {focusNodeId && (
                <div style={{
                    position: "absolute",
                    top: 10,
                    left: "50%",
                    transform: "translateX(-50%)",
                    background: "rgba(36, 23, 117, 0.85)",
                    color: "#fff",
                    padding: "4px 14px",
                    borderRadius: 6,
                    fontFamily: "Poppins",
                    fontSize: 13,
                    zIndex: 5,
                    pointerEvents: "none",
                    whiteSpace: "nowrap",
                }}>
                    Focus mode — press f or Esc to exit
                </div>
            )}
            <Background  />
            <Controls position="bottom-left" orientation="horizontal" showInteractive={false} />
            <MiniMap
                nodeColor={nodeColor}
                maskColor="rgb(36, 23, 117, 0.5)"
                pannable={true}
                position="bottom-right"
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