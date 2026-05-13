import {hostname} from "os";
import {Root, Environment} from "./environment";
import {EnvironmentApi} from "../api";

export const enable = async (
    root: Root,
    token: string,
    description?: string,
    host?: string,
): Promise<Environment> => {
    if (root.isEnabled()) {
        return root.environment!;
    }

    const hostName = host || hostname();
    const desc = description || "";

    const apiEndpoint = root.apiEndpoint();
    root.environment = new Environment(token, "", apiEndpoint.endpoint);

    let cfg;
    try {
        cfg = await root.client();
    } catch (err) {
        root.environment = undefined;
        throw new Error("error getting zrok client: " + err);
    }

    let res;
    try {
        res = await new EnvironmentApi(cfg).enable({body: {description: desc, host: hostName}});
    } catch (err) {
        root.environment = undefined;
        throw new Error("unable to enable environment: " + err);
    }

    const env = new Environment(token, res.identity!, apiEndpoint.endpoint);
    root.setEnvironment(env);
    root.saveZitiIdentityNamed(root.environmentIdentityName(), res.cfg!);

    return env;
}

export const disable = async (root: Root): Promise<void> => {
    if (!root.isEnabled()) {
        return;
    }

    const cfg = await root.client();
    await new EnvironmentApi(cfg).disable({body: {identity: root.environment?.zId}})
        .catch(err => {
            throw new Error("unable to disable environment: " + err);
        });

    root.deleteEnvironment();
}
