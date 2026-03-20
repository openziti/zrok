import {Root} from "./environment";
import {MetadataApi, Overview} from "../api";

export const getOverview = async (root: Root): Promise<Overview> => {
    if (!root.isEnabled()) {
        throw new Error("environment is not enabled; enable with 'zrok2 enable' first!");
    }

    const cfg = await root.client();
    return new MetadataApi(cfg).overview()
        .catch(err => {
            throw new Error("unable to get account overview: " + err);
        });
}
