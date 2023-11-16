// src/zrok/api/model/shareRequest.ts
var _ShareRequest = class _ShareRequest {
  static getAttributeTypeMap() {
    return _ShareRequest.attributeTypeMap;
  }
};
_ShareRequest.discriminator = void 0;
_ShareRequest.attributeTypeMap = [
  {
    "name": "envZId",
    "baseName": "envZId",
    "type": "string"
  },
  {
    "name": "shareMode",
    "baseName": "shareMode",
    "type": "ShareRequest.ShareModeEnum"
  },
  {
    "name": "frontendSelection",
    "baseName": "frontendSelection",
    "type": "Array<string>"
  },
  {
    "name": "backendMode",
    "baseName": "backendMode",
    "type": "ShareRequest.BackendModeEnum"
  },
  {
    "name": "backendProxyEndpoint",
    "baseName": "backendProxyEndpoint",
    "type": "string"
  },
  {
    "name": "authScheme",
    "baseName": "authScheme",
    "type": "string"
  },
  {
    "name": "authUsers",
    "baseName": "authUsers",
    "type": "Array<AuthUser>"
  },
  {
    "name": "oauthProvider",
    "baseName": "oauthProvider",
    "type": "ShareRequest.OauthProviderEnum"
  },
  {
    "name": "oauthEmailDomains",
    "baseName": "oauthEmailDomains",
    "type": "Array<string>"
  },
  {
    "name": "oauthAuthorizationCheckInterval",
    "baseName": "oauthAuthorizationCheckInterval",
    "type": "string"
  },
  {
    "name": "reserved",
    "baseName": "reserved",
    "type": "boolean"
  }
];
var ShareRequest = _ShareRequest;
((ShareRequest2) => {
  let ShareModeEnum;
  ((ShareModeEnum2) => {
    ShareModeEnum2["Public"] = "public";
    ShareModeEnum2["Private"] = "private";
  })(ShareModeEnum = ShareRequest2.ShareModeEnum || (ShareRequest2.ShareModeEnum = {}));
  let BackendModeEnum;
  ((BackendModeEnum2) => {
    BackendModeEnum2["Proxy"] = "proxy";
    BackendModeEnum2["Web"] = "web";
    BackendModeEnum2["TcpTunnel"] = "tcpTunnel";
    BackendModeEnum2["UdpTunnel"] = "udpTunnel";
    BackendModeEnum2["Caddy"] = "caddy";
  })(BackendModeEnum = ShareRequest2.BackendModeEnum || (ShareRequest2.BackendModeEnum = {}));
  let OauthProviderEnum;
  ((OauthProviderEnum2) => {
    OauthProviderEnum2["Github"] = "github";
    OauthProviderEnum2["Google"] = "google";
  })(OauthProviderEnum = ShareRequest2.OauthProviderEnum || (ShareRequest2.OauthProviderEnum = {}));
})(ShareRequest || (ShareRequest = {}));
export {
  ShareRequest
};
//# sourceMappingURL=shareRequest.mjs.map