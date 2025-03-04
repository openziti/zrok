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
import ShareNode from "./ShareNode.tsx";
import EnvironmentNode from "./EnvironmentNode.tsx";
import AccountNode from "./AccountNode.tsx";
import AccessNode from "./AccessNode.tsx";
import {Box} from "@mui/material";
import useApiConsoleStore from "./model/store.ts";
import AccessEdge from "./AccessEdge.tsx";

const edgeTypes = {
    access: AccessEdge
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
    const nodes = useApiConsoleStore((state) => state.nodes);
    const updateNodes = useApiConsoleStore((state) => state.updateNodes);
    const edges = useApiConsoleStore((state) => state.edges);

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

    const nodeColor = (node) => {
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