import {create} from "zustand";
import {Environment} from "../api";
import {VisualOverview} from "./visualizer.ts";
import {Node} from "@xyflow/react";
import {User} from "./user.ts";

type StoreState = {
    user: User;
    environments: Array<Environment>;
    overview: VisualOverview;
    selectedNode: Node;
    viewport: Array<Number>;
};

type StoreAction = {
    updateUser: (user: StoreState['user']) => void,
    updateOverview: (vov: StoreState['overview']) => void,
    updateEnvironments: (environments: StoreState['environments']) => void,
    updateSelectedNode: (selectedNode: StoreState['selectedNode']) => void,
    updateViewport: (viewport: StoreState['viewport']) => void,
};

const useStore = create<StoreState & StoreAction>((set) => ({
    user: null,
    overview: new VisualOverview(),
    environments: new Array<Environment>(),
    selectedNode: null,
    viewport: [0, 0, 1.5],
    updateUser: (user) => set({user: user}),
    updateOverview: (vov) => set({overview: vov}),
    updateEnvironments: (environments) => set({environments: environments}),
    updateSelectedNode: (selectedNode) => set({selectedNode: selectedNode}),
    updateViewport: (viewport) => set({viewport: viewport})
}));

export default useStore;
