import {create} from "zustand";
import {Environment} from "../api";
import {Graph} from "./graph.ts";
import {Edge, Node, Viewport} from "@xyflow/react";
import {User} from "./user.ts";
import {MRT_PaginationState, MRT_SortingState} from "material-react-table";

type StoreState = {
    user: User;
    limited: boolean;
    graph: Graph;
    environments: Array<Environment>;
    sparkdata: Map<string, Number[]>;
    nodes: Node[];
    edges: Edge[];
    selectedNode: Node;
    viewport: Viewport;
    pagination: MRT_PaginationState;
    sorting: MRT_SortingState;
};

type StoreAction = {
    updateUser: (user: StoreState['user']) => void,
    updateLimited: (limited: StoreState['limited']) => void,
    updateGraph: (vov: StoreState['graph']) => void,
    updateEnvironments: (environments: StoreState['environments']) => void,
    updateSparkdata: (sparkdata: StoreState['sparkdata']) => void,
    updateNodes: (nodes: StoreState['nodes']) => void,
    updateEdges: (edges: StoreState['edges']) => void,
    updateSelectedNode: (selectedNode: StoreState['selectedNode']) => void,
    updateViewport: (viewport: StoreState['viewport']) => void,
    updatePagination: (pagination: StoreState['pagination']) => void,
    updateSorting: (sorting: StoreState['sorting']) => void,
 };

const useApiConsoleStore = create<StoreState & StoreAction>((set) => ({
    user: null,
    limited: false,
    graph: new Graph(),
    environments: new Array<Environment>(),
    sparkdata: new Map<string, Number[]>(),
    nodes: [],
    edges: [],
    selectedNode: null,
    viewport: {x: 0, y: 0, zoom: 1},
    pagination: {pageIndex: 0, pageSize: 15},
    sorting: [{id: "data.label", desc: false}] as MRT_SortingState,
    updateUser: (user) => set({user: user}),
    updateLimited: (limited) => set({limited: limited}),
    updateGraph: (vov) => set({overview: vov}),
    updateEnvironments: (environments) => set({environments: environments}),
    updateSparkdata: (sparkdata) => set({sparkdata: sparkdata}),
    updateNodes: (nodes) => set({nodes: nodes}),
    updateEdges: (edges) => set({edges: edges}),
    updateSelectedNode: (selectedNode) => set({selectedNode: selectedNode}),
    updateViewport: (viewport) => set({viewport: viewport}),
    updatePagination: (pagination) => set({pagination: pagination}),
    updateSorting: (sorting) => set({sorting: sorting})
}));

export default useApiConsoleStore;
