import {Root} from "../environment/root"

function Overview(root) {
    if (!root.IsEnabled()){
        throw new Error("environment is not enabled; enable with 'zrok enable' first!")
    }
}