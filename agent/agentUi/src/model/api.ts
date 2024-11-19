import {AgentApi, Configuration} from "../api";

export const GetAgentApi = () => {
    return new AgentApi(new Configuration({basePath: window.location.origin}));
}
