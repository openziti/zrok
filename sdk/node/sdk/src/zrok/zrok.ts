import Express from "express"
import ziti from "@openziti/ziti-sdk-nodejs"
import { Root } from "../environment/root"

export function express(shrToken: string): Express.Application {
    return ziti.express(Express, shrToken)
}

export async function init(root: Root): Promise<any> {
    return ziti.init(root.env.ZitiIdentity)
}