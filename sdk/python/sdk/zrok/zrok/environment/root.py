from dataclasses import dataclass, field
from typing import NamedTuple
from .dirs import identityFile, rootDir, configFile, environmentFile, metadataFile
import os
import json
import zrok_api as zrok
from zrok_api.configuration import Configuration
import re

V = "v0.4"


@dataclass
class Metadata:
    V: str = ""
    RootPath: str = ""


@dataclass
class Config:
    ApiEndpoint: str = ""


@dataclass
class Environment:
    Token: str = ""
    ZitiIdentity: str = ""
    ApiEndpoint: str = ""


class ApiEndpoint(NamedTuple):
    endpoint: str
    frm: str


@dataclass
class Root:
    meta: Metadata = field(default_factory=Metadata)
    cfg: Config = field(default_factory=Config)
    env: Environment = field(default_factory=Environment)

    def HasConfig(self) -> bool:
        return self.cfg != Config()

    def Client(self) -> zrok.ApiClient:
        apiEndpoint = self.ApiEndpoint()

        cfg = Configuration()
        cfg.host = apiEndpoint[0] + "/api/v1"
        cfg.api_key["x-token"] = self.env.Token
        cfg.api_key_prefix['Authorization'] = 'Bearer'

        zrock_client = zrok.ApiClient(configuration=cfg)
        v = zrok.MetadataApi(zrock_client).version()
        # allow reported version string to be optionally prefixed with
        # "refs/heads/" or "refs/tags/"
        rxp = re.compile("^(refs/(heads|tags)/)?" + V)
        if not rxp.match(v):
            raise Exception("expected a '" + V + "' version, received: '" + v + "'")
        return zrock_client

    def ApiEndpoint(self) -> ApiEndpoint:
        apiEndpoint = "https://api.zrok.io"
        frm = "binary"

        if self.cfg.ApiEndpoint != "":
            apiEndpoint = self.cfg.ApiEndpoint
            frm = "config"

        env = os.getenv("ZROK_API_ENDPOINT")
        if env != "":
            apiEndpoint = env
            frm = "ZROK_API_ENDPOINT"

        if self.IsEnabled():
            apiEndpoint = self.env.ApiEndpoint
            frm = "env"

        return ApiEndpoint(apiEndpoint.rstrip("/"), frm)

    def IsEnabled(self) -> bool:
        return self.env != Environment()

    def PublicIdentityName(self) -> str:
        return "public"

    def EnvironmentIdentityName(self) -> str:
        return "environment"

    def ZitiIdentityNamed(self, name: str) -> str:
        return identityFile(name)


def Default() -> Root:
    r = Root()
    root = rootDir()
    r.meta = Metadata(V=V, RootPath=root)
    return r


def Assert() -> bool:
    exists = __rootExists()
    if exists:
        meta = __loadMetadata()
        return meta.V == V
    return False


def Load() -> Root:
    r = Root()
    if __rootExists():
        r.meta = __loadMetadata()
        r.cfg = __loadConfig()
        r.env = __loadEnvironment()
    else:
        r = Default()
    return r


def __rootExists() -> bool:
    mf = metadataFile()
    return os.path.isfile(mf)


def __assertMetadata():
    pass


def __loadMetadata() -> Metadata:
    mf = metadataFile()
    with open(mf) as f:
        data = json.load(f)
        return Metadata(V=data["v"])


def __loadConfig() -> Config:
    cf = configFile()
    with open(cf) as f:
        data = json.load(f)
        return Config(ApiEndpoint=data["api_endpoint"])


def isEnabled() -> bool:
    ef = environmentFile()
    return os.path.isfile(ef)


def __loadEnvironment() -> Environment:
    ef = environmentFile()
    with open(ef) as f:
        data = json.load(f)
        return Environment(
            Token=data["zrok_token"],
            ZitiIdentity=data["ziti_identity"],
            ApiEndpoint=data["api_endpoint"])
