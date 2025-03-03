import {Share} from "./api";
import Express from "express";
// @ts-ignore
import ziti from "@openziti/ziti-sdk-nodejs";
import {Root} from "./environment";

export const express = (share: Share): Express.Application => {
    return ziti.express(Express, share.shareToken);
}

export const init = (root: Root): Promise<any> => {
    return ziti.init(root.zitiIdentityName(root.environmentIdentityName()))
}