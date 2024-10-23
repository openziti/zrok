import {AgentApi, ApiClient} from "../api/src/index.js";

export const getAgentApi = () => {
    return new AgentApi(new ApiClient(window.location.protocol+'//'+window.location.host));
}

export const shareHandler = (values) => {
    let api = getAgentApi();
    switch(values.shareMode) {
        case "public":
            api.agentSharePublic({
                target: values.target,
                backendMode: values.backendMode,
            }, (err, data) => {
                console.log(err, data);
            });
            break;

        case "private":
            api.agentSharePrivate({
                target: values.target,
                backendMode: values.backendMode,
            }, (err, data) => {
                console.log(err, data);
            });
            break;
    }
}

export const accessHandler = (values) => {
    getAgentApi().agentAccessPrivate({
        token: values.token,
        bindAddress: values.bindAddress,
    }, (err, data) => {
        console.log(err, data);
    });
}

export const releaseShare = (opts) => {
    getAgentApi().agentReleaseShare(opts, (err, data) => {
        console.log(data);
    });
}

export const releaseAccess = (opts) => {
    getAgentApi().agentReleaseAccess(opts, (err, data) => {
        console.log(data);
    });
}