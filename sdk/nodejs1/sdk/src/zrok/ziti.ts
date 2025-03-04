import {Root} from "./environment";
// @ts-ignore
import ziti from "@openziti/ziti-sdk-nodejs";
import {Share} from "../api";

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

export const write = (conn: any, buf: any, writeCallback: any = () => {}) => {
    ziti.write(conn, buf, writeCallback);
}