import {environmentFile, metadataFile} from "./paths";
import * as fs from "node:fs";

const ENVIRONMENT_V = "v0.4";

export class Root {
	metadata: Metadata;
    environment: Environment|undefined;

    constructor(metadata: Metadata, environment: Environment|undefined) {
        this.metadata = metadata;
        this.environment = environment;
    }

    public isEnabled = (): boolean => {
        return this.environment != undefined;
    }
}

export class Metadata {
    v: string;
    rootPath: string;

    constructor(v: string, rootPath: string = "") {
        this.v = v;
        this.rootPath = rootPath;
    }
}

export class Environment {
    accountToken: string;
    zId: string;
    apiEndpoint: string;

    constructor(accountToken: string, zId: string, apiEndpoint: string) {
        this.accountToken = accountToken;
        this.zId = zId;
        this.apiEndpoint = apiEndpoint;
    }
}

export const loadRoot = (): Root => {
	if(rootExists()) {
        let metadata = loadMetadata();
        let environment = loadEnvironment();
        return new Root(metadata, environment);
    }
    throw new Error("unable to load root; did you 'zrok enable'?");
};

export const rootExists = (): boolean => {
    return fs.existsSync(metadataFile());
}

const loadMetadata = (): Metadata => {
    let f = metadataFile();
    let data = fs.readFileSync(f);
    let obj = JSON.parse(data.toString());
    if(obj.v != ENVIRONMENT_V) {
        throw new Error("invalid environment version! got version '" + obj.v + "' expected '" + ENVIRONMENT_V + "'");
    }
    return new Metadata(obj.v, f);
}

const loadEnvironment = (): Environment|undefined => {
	let f = environmentFile();
    if(!fs.existsSync(f)) {
        return undefined;
    }
    let data = fs.readFileSync(f);
    let obj = JSON.parse(data.toString());
    return new Environment(obj.zrok_token, obj.ziti_identity, obj.api_endpoint);
}