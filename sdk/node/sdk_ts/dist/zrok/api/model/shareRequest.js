"use strict";
var __defProp = Object.defineProperty;
var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
var __getOwnPropNames = Object.getOwnPropertyNames;
var __hasOwnProp = Object.prototype.hasOwnProperty;
var __export = (target, all) => {
  for (var name in all)
    __defProp(target, name, { get: all[name], enumerable: true });
};
var __copyProps = (to, from, except, desc) => {
  if (from && typeof from === "object" || typeof from === "function") {
    for (let key of __getOwnPropNames(from))
      if (!__hasOwnProp.call(to, key) && key !== except)
        __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
  }
  return to;
};
var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);

// src/zrok/api/model/shareRequest.ts
var shareRequest_exports = {};
__export(shareRequest_exports, {
  ShareRequest: () => ShareRequest
});
module.exports = __toCommonJS(shareRequest_exports);
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
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  ShareRequest
});
//# sourceMappingURL=shareRequest.js.map