import {create} from "zustand";
import {Environment} from "../api";
import {VisualOverview} from "./visualizer.ts";

type MetricsStoreState = {
    overview: VisualOverview;
    environments: Array<Environment>;
};

type MetricsStoreAction = {
    updateOverview: (vov: MetricsStoreState['overview']) => void,
    updateEnvironments: (environments: MetricsStoreState['environments']) => void
};

const useMetricsStore = create<MetricsStoreState & MetricsStoreAction>((set) => ({
    overview: new VisualOverview(),
    environments: new Array<Environment>(),
    updateOverview: (vov) => set({overview: vov}),
    updateEnvironments: (environments) => set({environments: environments}),
}));

export default useMetricsStore;
