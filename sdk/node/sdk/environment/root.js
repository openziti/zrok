import {environmntFile, configFile, metadataFile, identityFile} from "./dirs"
import fs from "node:fs"
import JSON from "JSON"
import * as gateway from "../zrok/api/gateway"

const V = "v0.4" 

class Metadata {
    constructor(V, RootPath) {
        this.V = V
        this.RootPath = RootPath
    }
}

class Config {
    constructor(ApiEndpoint) {
        this.ApiEndpoint = ApiEndpoint
    }
}

class Environment {
    constructor(Token, ZitiIdentity, ApiEndpoint) {
        this.Token = Token
        this.ZitiIdentity = ZitiIdentity
        this.ApiEndpoint = ApiEndpoint
    }
}

class ApiEndpoint {
    constructor(endpoint, frm) {
        this.endpoint = endpoint
        this.frm = frm
    }
}

class Root {
    constructor(meta, cfg=Config(), env=Environment()){
        this.meta = meta
        this.cfg = cfg
        this.env = env
        if (meta === undefined) {
            let root = rootDir()
            this.meta = Metadata(V, root)
        }
    }

    HasConfig() {
        return this.cfg !== undefined && Object.keys(this.cfg).length === 0
    }

    Client() {
        let apiEndpoint = this.ApiEndpoint()

        function getAuthorization(security) {
            switch(security.id) {
                case 'key': return "why"
                //case 'key': return getApiKey();
                default: console.log('default');
            }
        }

       gateway.init({
        url: apiEndpoint + '/api/v1',
        getAuthorization
       })
    }

    ApiEndpoint() {
        let apiEndpoint = "https://api.zrok.io"
        let frm = "binary"

        if (this.cfg.ApiEndpoint != "") {
            apiEndpoint = this.cfg.ApiEndpoint
            frm = "config"
        }

        env = process.env.ZROK_API_ENDPOINT
        if (env != "") {
            apiEndpoint = env
            frm = "ZROK_API_ENDPOINT"
        }

        if (this.IsEnabled()) {
            apiEndpoint = this.env.ApiEndpoint
            frm = "env"
        }

        return ApiEndpoint(apiEndpoint.replace(/\/+$/, ""), frm)
    }

    IsEnabled() {
        return this.env !== undefined && Object.keys(this.env).length === 0
    }

    PublicIdentityName() {
        return "public"
    }

    EnvironmentIdentityName() {
        return "environment"
    }

    ZitiIdentityName(name) {
        return identityFile(name)
    }

}

function Assert() {
    if (rootExists()){
        meta = loadMetadata()
        return meta.V == V
    }
    return false
}

function Load() {
    if (rootExists()) {
        return Root(loadMetadata(), loadConfig(), loadEnvironment())
    }
    return Root()
}

function rootExists() {
    mf = metadataFile()
    return fs.existsSync(mf)
}

function loadMetadata() {
    mf = metadataFile()
    data = fs.readFileSync(mf)
    serialized = JSON.parse(data)
    return Metadata(serialized.v)
}

function loadConfig() {
    cf = configFile()
    data = fs.readFileSync(cf)
    serialized = JSON.parse(data)
    return Config(serialized.api_endpoint)

}

function isEnabled() {
    ef = environmntFile()
    return fs.existsSync(ef)
}

function loadEnvironment() {
    ef = environmntFile()
    data = fs.readFileSync(ef)
    serialized = JSON.parse(data)
    return Environment(serialized.zrok_token, serialized.ziti_identity, serialized.api_endpoint)
}