import {AgentApi, ApiClient} from "../api/src/index.js";

let api = new AgentApi(new ApiClient(window.location.protocol+'//'+window.location.host));

export const shareHandler = (values) => {
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
    api.agentAccessPrivate({
        token: values.token,
        bindAddress: values.bindAddress,
    }, (err, data) => {
        console.log(err, data);
    });
}

export const releaseShare = (opts) => {
    api.agentReleaseShare(opts, (err, data) => {
        console.log(data);
    });
}

export const releaseAccess = (opts) => {
    api.agentReleaseAccess(opts, (err, data) => {
        console.log(data);
    });
}