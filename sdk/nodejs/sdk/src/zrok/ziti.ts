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

export const dialer = async (acc: Access, connectCallback: any, dataCallback: any): ziti.dialer => {
    ziti.ziti_services_refresh();
    await awaitDialPermissions(acc.shareToken);
    ziti.dial(acc.shareToken, false, connectCallback, dataCallback);
}

export const write = (conn: any, buf: any, writeCallback: any = () => {}) => {
    ziti.write(conn, buf, writeCallback);
}

export enum ServicePermissions {
    None = 0,
    Bind = 2,
    Dial = 3,
}

export class ServiceStatus {
    status: number | undefined;
    permissions: ServicePermissions | undefined;

    constructor(status: number | undefined, permissions: ServicePermissions | undefined) {
        this.status = status;
        this.permissions = permissions;
    }
}

const awaitDialPermissions = async (shareToken: string) => {
    let serviceStatus: ServiceStatus;
    const startTime = process.hrtime();
    return new Promise((resolve: (value: any) => void, reject: (reason?: any) => void) => {
        (async function waitForPermission() {
            serviceStatus = await serviceAvailable(shareToken);
            if(serviceStatus.status !== 0) {
                throw new Error("shareToken '" + shareToken + "' not present in network");
            }
            if(serviceStatus.permissions !== ServicePermissions.Dial) {
                const now = process.hrtime(startTime);
                const elapsedTimeInMs = now[0] * 1000 + now[1] / 1000000;
                if(elapsedTimeInMs > 30 * 1000) {
                    return reject("timeout waiting for shareToken '" + shareToken + "' to acquire dial permissions");
                }
                setTimeout(waitForPermission, 100);
            } else {
                return resolve(0);
            }
        })();
    });
}

const serviceAvailable = async (serviceName: string): Promise<ServiceStatus> => {
    return new Promise((resolve) => {
        ziti.ziti_service_available(serviceName, (status: any) => {
            const serviceStatus = new ServiceStatus(status.status, status.permissions);
            resolve(serviceStatus);
        })
    });
}