import {create} from "zustand";
import {Environment} from "../api";

type State = {
    environments: Array<Environment>;
};

type Action = {
    updateEnvironments: (environments: State['environments']) => void
};

const useMetricsStore = create<State & Action>((set) => ({
    environments: new Array<Environment>(),
    updateEnvironments: (environments) => set({environments: environments}),
}));

export default useMetricsStore;
