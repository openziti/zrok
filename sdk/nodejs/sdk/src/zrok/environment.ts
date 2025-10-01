import {environmentFile, identityFile, metadataFile} from "./paths";
import * as fs from "node:fs";
import {Configuration} from "../api";

const ENVIRONMENT_V = "v0.4";

export class Root {
	metadata: Metadata;
    config: Config|undefined;
    environment: Environment|undefined;

    constructor(metadata: Metadata, environment: Environment|undefined) {
        this.metadata = metadata;
        this.config = undefined;
        this.environment = environment;
    }

    public apiConfiguration = (): Configuration => {
        let apiEndpoint = this.apiEndpoint();
        if(this.isEnabled()) {
            return new Configuration({basePath: apiEndpoint.endpoint + "/api/v2", apiKey: this.environment?.accountToken})
        } else {
            return new Configuration({basePath: apiEndpoint.endpoint + "/api/v2"});
        }
    }

    public apiEndpoint = (): ApiEndpoint => {
        let endpoint = "https://api-v2.zrok.io";
        let from = "binary";

        if(this.config?.apiEndpoint !== "") {
            endpoint = this.config?.apiEndpoint!;
            from = "config";
        }

        let env = process.env.ZROK_API_ENDPOINT;
        if(env != null) {
            endpoint = env;
            from = "ZROK_API_ENDPOINT";
        }

        if(this.isEnabled()) {
            endpoint = this.environment?.apiEndpoint!;
            from = "env";
        }

        return new ApiEndpoint(endpoint, from);
    }

    public hasConfig = (): boolean => {
        return this.config !== undefined;
    }

    public isEnabled = (): boolean => {
        return this.environment !== undefined;
    }

    public environmentIdentityName = (): string => {
        return "environment";
    }

    public zitiIdentityName = (name: string): string => {
        return identityFile(name);
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

export class Config {
    apiEndpoint: string;

    constructor(apiEndpoint: string) {
        this.apiEndpoint = apiEndpoint;
    }
}

export class ApiEndpoint {
    endpoint: string;
    from: string;

    constructor(endpoint: string, from: string) {
        this.endpoint = endpoint;
        this.from = from;
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