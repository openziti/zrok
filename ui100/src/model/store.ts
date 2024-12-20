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
};

type StoreAction = {
    updateUser: (user: StoreState['user']) => void,
    updateOverview: (vov: StoreState['overview']) => void,
    updateEnvironments: (environments: StoreState['environments']) => void,
    updateSelectedNode: (selectedNode: StoreState['selectedNode']) => void
};

const useStore = create<StoreState & StoreAction>((set) => ({
    user: null,
    overview: new VisualOverview(),
    environments: new Array<Environment>(),
    selectedNode: null,
    updateUser: (user) => set({user: user}),
    updateOverview: (vov) => set({overview: vov}),
    updateEnvironments: (environments) => set({environments: environments}),
    updateSelectedNode: (selectedNode) => set({selectedNode: selectedNode})
}));

export default useStore;
