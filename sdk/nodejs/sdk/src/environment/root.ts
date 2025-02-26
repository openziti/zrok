import { configFile, environmentFile, identityFile, metadataFile, rootDir } from "./dirs";
import fs from "node:fs"
import {Configuration, MetadataApi, MetadataApiApiKeys} from "../zrok/api/api"

const V = "v0.4"

export class Metadata {
    V: string;
    RootPath: string;

    constructor(V: string, RootPath: string = "") {
        this.V = V
        this.RootPath = RootPath
    }
}

export class ApiEndpoint {
    endpoint: string
    frm: string

    constructor(endpoint: string, frm: string) {
        this.endpoint = endpoint
        this.frm = frm
    }
}

export class Config {
    ApiEndpoint: string

    constructor(ApiEndpoint: string) {
        this.ApiEndpoint = ApiEndpoint
    }
}

export class Environment {
    Token: string
    ZitiIdentity: string
    ApiEndpoint: string

    constructor(Token: string, ZitiIdentity: string, ApiEndpoint: string) {
        this.Token = Token
        this.ZitiIdentity = ZitiIdentity
        this.ApiEndpoint = ApiEndpoint
    }
}

export class Root {
    meta: Metadata
    cfg: Config
    env: Environment

    constructor(meta: Metadata = new Metadata(V, rootDir()), cfg: Config, env: Environment) {
        this.meta = meta
        this.cfg = cfg
        this.env = env
    }

    public HasConfig(): boolean {
        return this.cfg !== undefined && Object.keys(this.cfg).length === 0
    }

    public async Client(): Promise<Configuration> {
        let apiEndpoint = this.ApiEndpoint()
       let mapi = new MetadataApi({basePath: apiEndpoint.endpoint + "/api/v1"})
        mapi.setApiKey(MetadataApiApiKeys, this.env.Token);
       let ver: Promise<string> = mapi.version()

       const regex : RegExp = new RegExp("^(refs/(heads|tags)/)?" + V);
       await ver.then(v => {
        if(!regex.test(v)) {
            throw new Error("Expected a '" + V + "' version, received: '" + v+ "'")
        }
       })

       return conf
    }

    public ApiEndpoint(): ApiEndpoint {
        let apiEndpoint = "https://api-v1.zrok.io"
        let frm = "binary"

        if (this.cfg.ApiEndpoint != "") {
            apiEndpoint = this.cfg.ApiEndpoint
            frm = "config"
        }

        let env = process.env.ZROK_API_ENDPOINT
        if (env != null) {
            apiEndpoint = env
            frm = "ZROK_API_ENDPOINT"
        }

        if (this.IsEnabled()) {
            apiEndpoint = this.env.ApiEndpoint
            frm = "env"
        }

        return new ApiEndpoint(apiEndpoint.replace(/\/+$/, ""), frm)
    }

    public IsEnabled(): boolean {
        return this.env !== undefined && Object.keys(this.env).length > 0
    }

    private PublicIdentityName(): string {
        return "public"
    }

    public EnvironmentIdentityName(): string {
        return "environment"
    }

    public ZitiIdentityNamed(name: string): string {
        return identityFile(name)
    }
}

export function Assert(): boolean {
    if (rootExists()) {
        let meta = loadMetadata()
        return meta.V == V
    }
    return false
}

export function Load(): Root {
    if (rootExists()) {
        return new Root(loadMetadata(), loadConfig(), loadEnvironment())
    }
    throw new Error("unable to load root. Does not exist")
}

function rootExists(): boolean {
    return fs.existsSync(metadataFile())
}

function loadMetadata(): Metadata {
    let mf = metadataFile()
    let data = fs.readFileSync(mf)
    let serial = JSON.parse(data.toString())
    return new Metadata(serial.v)
}

function loadConfig(): Config {
    let cf = configFile()
    if (fs.existsSync(cf)) {    // the config.json file may not be present
        let data = fs.readFileSync(cf)
        let serial = JSON.parse(data.toString())
        return new Config(serial.api_endpoint)    
    } else {
        return new Config('')    
    }
}

function loadEnvironment(): Environment {
    let ef = environmentFile()
    let data = fs.readFileSync(ef)
    let serial = JSON.parse(data.toString())
    return new Environment(serial.zrok_token, serial.ziti_identity, serial.api_endpoint)
}