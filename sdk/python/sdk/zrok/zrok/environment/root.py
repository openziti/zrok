from dataclasses import dataclass, field
from typing import NamedTuple
from .dirs import identityFile, rootDir, configFile, environmentFile, metadataFile
import os
import json
import zrok_api as zrok
from zrok_api.configuration import Configuration
from ..api_client_wrapper import ZrokApiClient
import re

V = "v1.0"

@dataclass
class Metadata:
    V: str = ""
    RootPath: str = ""


@dataclass
class Config:
    ApiEndpoint: str = ""
    DefaultFrontend: str = ""


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

    def Client(self) -> ZrokApiClient:
        apiEndpoint = self.ApiEndpoint()

        cfg = Configuration()
        cfg.host = apiEndpoint[0] + "/api/v1"
        cfg.api_key["x-token"] = self.env.Token
        cfg.api_key_prefix['Authorization'] = 'Bearer'

        version_check_client = ZrokApiClient(configuration=cfg)
        # Perform version check without authentication
        self.client_version_check(version_check_client)
        
        # Create a new client with the same configuration for authenticated requests
        auth_client = ZrokApiClient(configuration=cfg)
        # Explicitly set the x-token header in default_headers to ensure it's used
        auth_client.set_default_header('x-token', self.env.Token)
        
        return auth_client

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

    def client_version_check(self, zrock_client):
        """Check if the client version is compatible with the API."""
        metadata_api = zrok.MetadataApi(zrock_client)
        try:
            # Create a request with NO authentication for version check
            # We'll remove any authentication headers for this initial check
            # The client's default_headers might already have x-token, so let's ensure this call doesn't use it
            custom_headers = {}
            for key, value in zrock_client.default_headers.items():
                if key.lower() != 'x-token':  # Skip the auth token header
                    custom_headers[key] = value
                    
            response = metadata_api.client_version_check_with_http_info(
                body={"clientVersion": V},
                _headers=custom_headers  # Pass custom headers without auth token
            )

            # Check if the response status code is 200 OK
            if response.status_code != 200:
                raise Exception(f"Client version check failed: Unexpected status code {response.status_code}")

            # Success case - status code is 200 and empty response body is expected
            return

        except Exception as e:
            raise Exception(f"Client version check failed: {str(e)}")

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
    try:
        with open(cf) as f:
            data = json.load(f)
            return Config(
                ApiEndpoint=data.get("api_endpoint", ""),
                DefaultFrontend=data.get("default_frontend", "")
            )
    except (FileNotFoundError, json.JSONDecodeError):
        return Config(ApiEndpoint="", DefaultFrontend="")


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
