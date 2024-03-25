import Express from "express"
import ziti from "@openziti/ziti-sdk-nodejs"
import { Root } from "../environment/root"

export function express(shrToken: string): Express.Application {
    return ziti.express(Express, shrToken)
}

export async function init(root: Root): Promise<any> {
    return ziti.init(root.ZitiIdentityNamed(root.EnvironmentIdentityName()))
}

export function dialer(root: Root, shrToken: string, connectCallback: any, dataCallback: any): ziti.dialer {
    ziti.dial(shrToken, false, connectCallback, dataCallback)
}

export function setLogLevel(level:number) {
    ziti.setLogLevel(level)
}