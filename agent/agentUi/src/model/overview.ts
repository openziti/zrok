import {AccessDetail, ShareDetail, StatusResponse} from "../api";

export class AgentObject {
    type: string;
    id: string;
    v: (ShareDetail|AccessDetail);
}

export function buildOverview(status: StatusResponse): Array<AgentObject> {
    let out = new Array<AgentObject>();
    if(status) {
        if(status.accesses) {
            status.accesses.forEach(acc => {
                let accObj = new AgentObject();
                accObj.type = "access";
                accObj.id = acc.frontendToken!;
                accObj.v = acc;
                out.push(accObj);
            });
        }
        if(status.shares) {
            status.shares.forEach(shr => {
               let shrObj = new AgentObject();
               shrObj.type = "share";
               shrObj.id = shr.token!;
               shrObj.v = shr;
               out.push(shrObj);
            });
        }
        out.sort((a, b) => {
            if(a.id < b.id) return -1;
            if(a.id > b.id) return 1;
        });
    }
    return out;
}