import {AccessDetail, ShareDetail, StatusResponse} from "../api";

export class AgentObject {
    type: string;
    id: string;
    v: (ShareDetail|AccessDetail);
    displayToken: string;
    status: string;
}

export function buildOverview(status: StatusResponse): Array<AgentObject> {
    const out = new Array<AgentObject>();
    if(status) {
        if(status.accesses) {
            status.accesses.forEach(acc => {
                const accObj = new AgentObject();
                accObj.type = "access";
                // use failure ID in token column if token is empty (failed items)
                const displayToken = acc.frontendToken || (acc.failure?.id || "");
                accObj.id = displayToken;
                accObj.displayToken = displayToken;
                accObj.status = acc.status || "unknown";
                accObj.v = acc;
                out.push(accObj);
            });
        }
        if(status.shares) {
            status.shares.forEach(shr => {
               const shrObj = new AgentObject();
               shrObj.type = "share";
               // use failure ID in token column if token is empty (failed items)
               const displayToken = shr.token || (shr.failure?.id || "");
               shrObj.id = displayToken;
               shrObj.displayToken = displayToken;
               shrObj.status = shr.status || "unknown";
               shrObj.v = shr;
               out.push(shrObj);
            });
        }
        out.sort((a, b) => {
            if(a.id < b.id) return -1;
            if(a.id > b.id) return 1;
            return 0;
        });
    }
    return out;
}

export interface StatusCounts {
    active: number;
    retrying: number;
    failed: number;
}

export function categorizeShares(shares: ShareDetail[]): StatusCounts {
    const counts = { active: 0, retrying: 0, failed: 0 };
    shares.forEach(share => {
        switch (share.status) {
            case "active":
                counts.active++;
                break;
            case "retrying":
                counts.retrying++;
                break;
            case "failed":
                counts.failed++;
                break;
        }
    });
    return counts;
}

export function categorizeAccesses(accesses: AccessDetail[]): StatusCounts {
    const counts = { active: 0, retrying: 0, failed: 0 };
    accesses.forEach(access => {
        switch (access.status) {
            case "active":
                counts.active++;
                break;
            case "retrying":
                counts.retrying++;
                break;
            case "failed":
                counts.failed++;
                break;
        }
    });
    return counts;
}