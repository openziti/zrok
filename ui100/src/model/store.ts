import {create} from "zustand";
import {Environment} from "../api";
import {VisualOverview} from "./visualizer.ts";

type StoreState = {
    overview: VisualOverview;
    environments: Array<Environment>;
};

type StoreAction = {
    updateOverview: (vov: StoreState['overview']) => void,
    updateEnvironments: (environments: StoreState['environments']) => void
};

const useStore = create<StoreState & StoreAction>((set) => ({
    overview: new VisualOverview(),
    environments: new Array<Environment>(),
    updateOverview: (vov) => set({overview: vov}),
    updateEnvironments: (environments) => set({environments: environments}),
}));

export default useStore;
