import {create} from "zustand";
import {Environment, Metrics} from "../api";

type State = {
    account: Metrics;
    environments: Array<Environment>;
};

type Action = {
    updateAccount: (metrics: State['account']) => void
    updateEnvironments: (environments: State['environments']) => void
};

const useMetricsStore = create<State & Action>((set) => ({
    account: null,
    environments: new Array<Environment>(),
    updateAccount: (metrics) => set({account: metrics}),
    updateEnvironments: (environments) => set({environments: environments}),
}));

export default useMetricsStore;
