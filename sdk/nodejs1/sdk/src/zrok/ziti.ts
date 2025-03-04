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
                         callbacks?: {
                            listenCallback?: any,
                            listenClientCallback?: any,
                            clientConnectCallback: any,
                            clientDataCallback?: any
                         }): ziti.listener => {
    let listenCallback = callbacks?.listenCallback ? callbacks.listenClientCallback : (data: any) => {};
    let listenClientCallback = callbacks?.listenClientCallback ? callbacks.listenClientCallback : (data: any) => {};
    let clientConnectCallback = callbacks?.clientConnectCallback ? callbacks.clientConnectCallback : (data: any) => {};
    let clientDataCallback = callbacks?.clientDataCallback ? callbacks.clientDataCallback : (data: any) => {};
    console.log("client connect callback", clientConnectCallback);
    ziti.listen(shr.shareToken, 0, listenCallback, listenClientCallback, clientConnectCallback, clientDataCallback);
}

export const write = (conn: any, buf: any, writeCallback: any = () => {}) => {
    ziti.write(conn, buf, writeCallback);
}