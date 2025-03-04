import {Share} from "../api";
import Express from "express";
// @ts-ignore
import ziti from "@openziti/ziti-sdk-nodejs";

export const express = (share: Share): Express.Application => {
    return ziti.express(Express, share.shareToken);
}

