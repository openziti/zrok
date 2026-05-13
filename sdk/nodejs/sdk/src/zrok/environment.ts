import {configFile, environmentFile, identitiesDir, identityFile, metadataFile, rootDir} from "./paths";
import * as fs from "node:fs";
import * as path from "node:path";
import {Configuration, MetadataApi} from "../api";

const ENVIRONMENT_V = "v0.4";
const CLIENT_V = "v2.0";

export class Root {
    metadata: Metadata;
    config: Config | undefined;
    environment: Environment | undefined;

    constructor(metadata: Metadata, environment: Environment | undefined, config?: Config) {
        this.metadata = metadata;
        this.config = config;
        this.environment = environment;
    }

    async client(): Promise<Configuration> {
        const apiEndpoint = this.apiEndpoint();
        let cfg: Configuration;
        if (this.isEnabled()) {
            cfg = new Configuration({basePath: apiEndpoint.endpoint + "/api/v2", apiKey: this.environment?.accountToken});
        } else {
            cfg = new Configuration({basePath: apiEndpoint.endpoint + "/api/v2"});
        }

        const api = new MetadataApi(cfg);
        await api.clientVersionCheck({body: {clientVersion: CLIENT_V}})
            .catch(err => {
                throw new Error("client version check failed: " + err);
            });

        return cfg;
    }

    apiEndpoint(): ApiEndpoint {
        let endpoint = "https://api-v2.zrok.io";
        let from = "binary";

        if (this.config?.apiEndpoint && this.config.apiEndpoint !== "") {
            endpoint = this.config.apiEndpoint;
            from = "config";
        }

        const env2 = process.env.ZROK2_API_ENDPOINT;
        if (env2 != null && env2 !== "") {
            endpoint = env2;
            from = "ZROK2_API_ENDPOINT";
        } else {
            const env = process.env.ZROK_API_ENDPOINT;
            if (env != null && env !== "") {
                endpoint = env;
                from = "ZROK_API_ENDPOINT";
                console.warn("WARNING: ZROK_API_ENDPOINT is deprecated, use ZROK2_API_ENDPOINT instead");
            }
        }

        if (this.isEnabled()) {
            endpoint = this.environment?.apiEndpoint!;
            from = "env";
        }

        endpoint = endpoint.replace(/\/+$/, "");

        return new ApiEndpoint(endpoint, from);
    }

    hasConfig(): boolean {
        return this.config !== undefined;
    }

    isEnabled(): boolean {
        return this.environment !== undefined;
    }

    environmentIdentityName(): string {
        return "environment";
    }

    zitiIdentityName(name: string): string {
        return identityFile(name);
    }

    setEnvironment(env: Environment): void {
        const ef = environmentFile();
        fs.mkdirSync(path.dirname(ef), {recursive: true});
        const data = {
            zrok_token: env.accountToken,
            ziti_identity: env.zId,
            api_endpoint: env.apiEndpoint,
        };
        fs.writeFileSync(ef, JSON.stringify(data, null, 2));
        this.environment = env;
    }

    saveZitiIdentityNamed(name: string, cfg: string): void {
        const idd = identitiesDir();
        fs.mkdirSync(idd, {recursive: true});
        const idf = identityFile(name);
        fs.writeFileSync(idf, cfg);
    }

    deleteEnvironment(): void {
        const ef = environmentFile();
        if (fs.existsSync(ef)) {
            fs.unlinkSync(ef);
        }
        const idf = identityFile(this.environmentIdentityName());
        if (fs.existsSync(idf)) {
            fs.unlinkSync(idf);
        }
        this.environment = undefined;
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
    defaultFrontend: string;

    constructor(apiEndpoint: string, defaultFrontend: string = "") {
        this.apiEndpoint = apiEndpoint;
        this.defaultFrontend = defaultFrontend;
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

export const defaultRoot = (): Root => {
    const root = rootDir();
    const metadata = new Metadata(ENVIRONMENT_V, root);
    return new Root(metadata, undefined);
}

export const assertRoot = (): boolean => {
    if (rootExists()) {
        const metadata = loadMetadata();
        return metadata.v === ENVIRONMENT_V;
    }
    return false;
}

export const loadRoot = (): Root => {
    if (rootExists()) {
        const metadata = loadMetadata();
        const config = loadConfig();
        const environment = loadEnvironment();
        return new Root(metadata, environment, config);
    }
    return defaultRoot();
};

export const rootExists = (): boolean => {
    return fs.existsSync(metadataFile());
}

const loadMetadata = (): Metadata => {
    const f = metadataFile();
    const data = fs.readFileSync(f);
    const obj = JSON.parse(data.toString());
    if (obj.v != ENVIRONMENT_V) {
        throw new Error("invalid environment version! got version '" + obj.v + "' expected '" + ENVIRONMENT_V + "'");
    }
    return new Metadata(obj.v, f);
}

const loadConfig = (): Config | undefined => {
    const f = configFile();
    try {
        const data = fs.readFileSync(f);
        const obj = JSON.parse(data.toString());
        return new Config(obj.api_endpoint || "", obj.default_frontend || "");
    } catch {
        return undefined;
    }
}

const loadEnvironment = (): Environment | undefined => {
    const f = environmentFile();
    if (!fs.existsSync(f)) {
        return undefined;
    }
    const data = fs.readFileSync(f);
    const obj = JSON.parse(data.toString());
    return new Environment(obj.zrok_token, obj.ziti_identity, obj.api_endpoint);
}
