import {Root} from "./environment";
// @ts-ignore
import ziti from "@openziti/ziti-sdk-nodejs";
import {Share} from "../api";
import {Access} from "./access";

export const init = (root: Root): Promise<any> => {
    return ziti.init(root.zitiIdentityName(root.environmentIdentityName()))
}

export const setLogLevel = (level: number) => {
    ziti.setLogLevel(level);
}

export const listener = (shr: Share,
                         clientConnectCallback: any,
                         clientDataCallback: any = (data: any) => {},
                         listenCallback: any = (data: any) => {},
                         listenClientCallback: any = (data: any) => {}): ziti.listener => {
    ziti.listen(shr.shareToken, 0, listenCallback, listenClientCallback, clientConnectCallback, clientDataCallback);
}

export const dialer = (acc: Access, connectCallback: any, dataCallback: any): ziti.dialer => {
    ziti.ziti_services_refresh();
    ziti.dial(acc.shareToken, false, connectCallback, dataCallback);
}

export const write = (conn: any, buf: any, writeCallback: any = () => {}) => {
    ziti.write(conn, buf, writeCallback);
}
