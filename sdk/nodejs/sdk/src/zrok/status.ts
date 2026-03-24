import {Root} from "./environment";
import {Status} from "./model";

export const status = (root: Root): Status => {
    const apiEndpoint = root.apiEndpoint();

    return {
        enabled: root.isEnabled(),
        apiEndpoint: apiEndpoint.endpoint,
        apiEndpointSource: apiEndpoint.from,
        token: root.isEnabled() ? root.environment!.accountToken : "",
        zitiIdentity: root.isEnabled() ? root.environment!.zId : "",
    };
}
