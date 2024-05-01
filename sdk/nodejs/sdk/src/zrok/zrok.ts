import Express from "express"
import ziti from "@openziti/ziti-sdk-nodejs"
import { Root } from "../environment/root"
import { ServiceStatus, Permissions } from "./servicestatus"

export function express(shrToken: string): Express.Application {
    return ziti.express(Express, shrToken)
}

export async function init(root: Root): Promise<any> {
    return ziti.init(root.ZitiIdentityNamed(root.EnvironmentIdentityName()))
}

async function service_available(service: string): Promise<any>  {
    return new Promise((resolve) => {
        ziti.ziti_service_available(service, (status: any) => {
            const serviceStatus = new ServiceStatus(status);
            resolve(serviceStatus);
        });
    });
};

/**
 * Remain in lazy-sleepy loop until the specified zrok share (a.k.a. Ziti Service) has Dial permissions.
 * 
 */
async function awaitDialPermissionPresent(shrToken: string) {
    let serviceStatus: ServiceStatus;
    const startTime = process.hrtime();
    return new Promise((resolve: (value: any) => void, reject: (reason?: any) => void) => {
        (async function waitForDialPermissionPresent() {
            serviceStatus = await service_available(shrToken)
            if (serviceStatus.getStatus() !== 0) {
                throw new Error('shrToken [' + shrToken + '] does not appear to be present in the network');
            }
            if (serviceStatus.getPermissions() !== Permissions.Dial) {
                const now = process.hrtime(startTime);
                const elapsedTimeInMs = now[0] * 1000 + now[1] / 1000000;
                if (elapsedTimeInMs > 30 * 1000) {  // 30-sec timeout
                    return reject('timeout waiting for shrToken [' + shrToken + '] to acquire Dial permissions');
                }
                setTimeout(waitForDialPermissionPresent, 100);
            } else {
                return resolve(0); 
            }
        })();
    });
}

export async function dialer(root: Root, shrToken: string, connectCallback: any, dataCallback: any): ziti.dialer {
    ziti.ziti_services_refresh()
    await awaitDialPermissionPresent(shrToken).catch(error => {throw new Error(error)});
    ziti.dial(shrToken, false, connectCallback, dataCallback)
}

export function listener(shrToken: string, listenCallback: any, listenClientCallback: any, clientConnectCallback: any, clientDataCallback: any): ziti.listener {
    ziti.listen(shrToken, 0, listenCallback, listenClientCallback, clientConnectCallback, clientDataCallback)
}

export function write(conn: any, buf: any, writeCallback: any ){
    ziti.write(conn, buf, writeCallback)
}


export function setLogLevel(level:number) {
    ziti.setLogLevel(level)
}