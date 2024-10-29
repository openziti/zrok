import {AgentApi, ApiClient} from "../api/src/index.js";

export const getAgentApi = () => {
    return new AgentApi(new ApiClient(window.location.origin));
}

export const createShare = (opts) => {
    switch(opts.shareMode) {
        case "public":
            getAgentApi().agentSharePublic(opts, (e, d) => {
                if(e) {
                    console.log("createShare", e, d);
                }
            });
            break;

        case "private":
            getAgentApi().agentSharePrivate(opts, (e, d) => {
                if(e) {
                    console.log("createShare", e, d);
                }
            });
            break;
    }
}

export const releaseShare = (opts) => {
    getAgentApi().agentReleaseShare(opts, (e, d) => {
        if(e) {
            console.log("releaseShare", e, d);
        }
    });
}

export const createAccess = (opts) => {
    getAgentApi().agentAccessPrivate(opts, (e, d) => {
        if(e) {
            console.log("createAccess", e, d);
        }
    });
}

export const releaseAccess = (opts) => {
    if(e) {
        getAgentApi().agentReleaseAccess(opts, (e, d) => {
            console.log("releaseAccess", e, d);
        });
    }
}