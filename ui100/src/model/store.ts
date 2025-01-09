import {create} from "zustand";
import {Environment} from "../api";
import {Graph} from "./graph.ts";
import {Edge, Node, Viewport} from "@xyflow/react";
import {User} from "./user.ts";

type StoreState = {
    user: User;
    graph: Graph;
    environments: Array<Environment>;
    nodes: Node[];
    edges: Edge[];
    selectedNode: Node;
    viewport: Viewport;
};

type StoreAction = {
    updateUser: (user: StoreState['user']) => void,
    updateGraph: (vov: StoreState['graph']) => void,
    updateEnvironments: (environments: StoreState['environments']) => void,
    updateSelectedNode: (selectedNode: StoreState['selectedNode']) => void,
    updateNodes: (nodes: StoreState['nodes']) => void,
    updateEdges: (edges: StoreState['edges']) => void,
    updateViewport: (viewport: StoreState['viewport']) => void,
};

const useStore = create<StoreState & StoreAction>((set) => ({
    user: null,
    graph: new Graph(),
    environments: new Array<Environment>(),
    nodes: [],
    edges: [],
    selectedNode: null,
    viewport: {x: 0, y: 0, zoom: 1},
    updateUser: (user) => set({user: user}),
    updateGraph: (vov) => set({overview: vov}),
    updateEnvironments: (environments) => set({environments: environments}),
    updateNodes: (nodes) => set({nodes: nodes}),
    updateEdges: (edges) => set({edges: edges}),
    updateSelectedNode: (selectedNode) => set({selectedNode: selectedNode}),
    updateViewport: (viewport) => set({viewport: viewport})
}));

export default useStore;
