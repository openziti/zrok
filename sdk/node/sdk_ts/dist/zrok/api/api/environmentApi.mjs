var __async = (__this, __arguments, generator) => {
  return new Promise((resolve, reject) => {
    var fulfilled = (value) => {
      try {
        step(generator.next(value));
      } catch (e) {
        reject(e);
      }
    };
    var rejected = (value) => {
      try {
        step(generator.throw(value));
      } catch (e) {
        reject(e);
      }
    };
    var step = (x) => x.done ? resolve(x.value) : Promise.resolve(x.value).then(fulfilled, rejected);
    step((generator = generator.apply(__this, __arguments)).next());
  });
};

// src/zrok/api/api/environmentApi.ts
import localVarRequest5 from "request";

// src/zrok/api/model/accessRequest.ts
var _AccessRequest = class _AccessRequest {
  static getAttributeTypeMap() {
    return _AccessRequest.attributeTypeMap;
  }
};
_AccessRequest.discriminator = void 0;
_AccessRequest.attributeTypeMap = [
  {
    "name": "envZId",
    "baseName": "envZId",
    "type": "string"
  },
  {
    "name": "shrToken",
    "baseName": "shrToken",
    "type": "string"
  }
];
var AccessRequest = _AccessRequest;

// src/zrok/api/model/accessResponse.ts
var _AccessResponse = class _AccessResponse {
  static getAttributeTypeMap() {
    return _AccessResponse.attributeTypeMap;
  }
};
_AccessResponse.discriminator = void 0;
_AccessResponse.attributeTypeMap = [
  {
    "name": "frontendToken",
    "baseName": "frontendToken",
    "type": "string"
  },
  {
    "name": "backendMode",
    "baseName": "backendMode",
    "type": "string"
  }
];
var AccessResponse = _AccessResponse;

// src/zrok/api/model/authUser.ts
var _AuthUser = class _AuthUser {
  static getAttributeTypeMap() {
    return _AuthUser.attributeTypeMap;
  }
};
_AuthUser.discriminator = void 0;
_AuthUser.attributeTypeMap = [
  {
    "name": "username",
    "baseName": "username",
    "type": "string"
  },
  {
    "name": "password",
    "baseName": "password",
    "type": "string"
  }
];
var AuthUser = _AuthUser;

// src/zrok/api/model/configuration.ts
var _Configuration = class _Configuration {
  static getAttributeTypeMap() {
    return _Configuration.attributeTypeMap;
  }
};
_Configuration.discriminator = void 0;
_Configuration.attributeTypeMap = [
  {
    "name": "version",
    "baseName": "version",
    "type": "string"
  },
  {
    "name": "touLink",
    "baseName": "touLink",
    "type": "string"
  },
  {
    "name": "invitesOpen",
    "baseName": "invitesOpen",
    "type": "boolean"
  },
  {
    "name": "requiresInviteToken",
    "baseName": "requiresInviteToken",
    "type": "boolean"
  },
  {
    "name": "inviteTokenContact",
    "baseName": "inviteTokenContact",
    "type": "string"
  },
  {
    "name": "passwordRequirements",
    "baseName": "passwordRequirements",
    "type": "PasswordRequirements"
  }
];
var Configuration = _Configuration;

// src/zrok/api/model/createFrontendRequest.ts
var _CreateFrontendRequest = class _CreateFrontendRequest {
  static getAttributeTypeMap() {
    return _CreateFrontendRequest.attributeTypeMap;
  }
};
_CreateFrontendRequest.discriminator = void 0;
_CreateFrontendRequest.attributeTypeMap = [
  {
    "name": "zId",
    "baseName": "zId",
    "type": "string"
  },
  {
    "name": "urlTemplate",
    "baseName": "url_template",
    "type": "string"
  },
  {
    "name": "publicName",
    "baseName": "public_name",
    "type": "string"
  }
];
var CreateFrontendRequest = _CreateFrontendRequest;

// src/zrok/api/model/createFrontendResponse.ts
var _CreateFrontendResponse = class _CreateFrontendResponse {
  static getAttributeTypeMap() {
    return _CreateFrontendResponse.attributeTypeMap;
  }
};
_CreateFrontendResponse.discriminator = void 0;
_CreateFrontendResponse.attributeTypeMap = [
  {
    "name": "token",
    "baseName": "token",
    "type": "string"
  }
];
var CreateFrontendResponse = _CreateFrontendResponse;

// src/zrok/api/model/createIdentity201Response.ts
var _CreateIdentity201Response = class _CreateIdentity201Response {
  static getAttributeTypeMap() {
    return _CreateIdentity201Response.attributeTypeMap;
  }
};
_CreateIdentity201Response.discriminator = void 0;
_CreateIdentity201Response.attributeTypeMap = [
  {
    "name": "identity",
    "baseName": "identity",
    "type": "string"
  },
  {
    "name": "cfg",
    "baseName": "cfg",
    "type": "string"
  }
];
var CreateIdentity201Response = _CreateIdentity201Response;

// src/zrok/api/model/createIdentityRequest.ts
var _CreateIdentityRequest = class _CreateIdentityRequest {
  static getAttributeTypeMap() {
    return _CreateIdentityRequest.attributeTypeMap;
  }
};
_CreateIdentityRequest.discriminator = void 0;
_CreateIdentityRequest.attributeTypeMap = [
  {
    "name": "name",
    "baseName": "name",
    "type": "string"
  }
];
var CreateIdentityRequest = _CreateIdentityRequest;

// src/zrok/api/model/deleteFrontendRequest.ts
var _DeleteFrontendRequest = class _DeleteFrontendRequest {
  static getAttributeTypeMap() {
    return _DeleteFrontendRequest.attributeTypeMap;
  }
};
_DeleteFrontendRequest.discriminator = void 0;
_DeleteFrontendRequest.attributeTypeMap = [
  {
    "name": "frontendToken",
    "baseName": "frontendToken",
    "type": "string"
  }
];
var DeleteFrontendRequest = _DeleteFrontendRequest;

// src/zrok/api/model/disableRequest.ts
var _DisableRequest = class _DisableRequest {
  static getAttributeTypeMap() {
    return _DisableRequest.attributeTypeMap;
  }
};
_DisableRequest.discriminator = void 0;
_DisableRequest.attributeTypeMap = [
  {
    "name": "identity",
    "baseName": "identity",
    "type": "string"
  }
];
var DisableRequest = _DisableRequest;

// src/zrok/api/model/enableRequest.ts
var _EnableRequest = class _EnableRequest {
  static getAttributeTypeMap() {
    return _EnableRequest.attributeTypeMap;
  }
};
_EnableRequest.discriminator = void 0;
_EnableRequest.attributeTypeMap = [
  {
    "name": "description",
    "baseName": "description",
    "type": "string"
  },
  {
    "name": "host",
    "baseName": "host",
    "type": "string"
  }
];
var EnableRequest = _EnableRequest;

// src/zrok/api/model/enableResponse.ts
var _EnableResponse = class _EnableResponse {
  static getAttributeTypeMap() {
    return _EnableResponse.attributeTypeMap;
  }
};
_EnableResponse.discriminator = void 0;
_EnableResponse.attributeTypeMap = [
  {
    "name": "identity",
    "baseName": "identity",
    "type": "string"
  },
  {
    "name": "cfg",
    "baseName": "cfg",
    "type": "string"
  }
];
var EnableResponse = _EnableResponse;

// src/zrok/api/model/environment.ts
var _Environment = class _Environment {
  static getAttributeTypeMap() {
    return _Environment.attributeTypeMap;
  }
};
_Environment.discriminator = void 0;
_Environment.attributeTypeMap = [
  {
    "name": "description",
    "baseName": "description",
    "type": "string"
  },
  {
    "name": "host",
    "baseName": "host",
    "type": "string"
  },
  {
    "name": "address",
    "baseName": "address",
    "type": "string"
  },
  {
    "name": "zId",
    "baseName": "zId",
    "type": "string"
  },
  {
    "name": "activity",
    "baseName": "activity",
    "type": "Array<SparkDataSample>"
  },
  {
    "name": "limited",
    "baseName": "limited",
    "type": "boolean"
  },
  {
    "name": "createdAt",
    "baseName": "createdAt",
    "type": "number"
  },
  {
    "name": "updatedAt",
    "baseName": "updatedAt",
    "type": "number"
  }
];
var Environment = _Environment;

// src/zrok/api/model/environmentAndResources.ts
var _EnvironmentAndResources = class _EnvironmentAndResources {
  static getAttributeTypeMap() {
    return _EnvironmentAndResources.attributeTypeMap;
  }
};
_EnvironmentAndResources.discriminator = void 0;
_EnvironmentAndResources.attributeTypeMap = [
  {
    "name": "environment",
    "baseName": "environment",
    "type": "Environment"
  },
  {
    "name": "frontends",
    "baseName": "frontends",
    "type": "Array<Frontend>"
  },
  {
    "name": "shares",
    "baseName": "shares",
    "type": "Array<Share>"
  }
];
var EnvironmentAndResources = _EnvironmentAndResources;

// src/zrok/api/model/frontend.ts
var _Frontend = class _Frontend {
  static getAttributeTypeMap() {
    return _Frontend.attributeTypeMap;
  }
};
_Frontend.discriminator = void 0;
_Frontend.attributeTypeMap = [
  {
    "name": "id",
    "baseName": "id",
    "type": "number"
  },
  {
    "name": "shrToken",
    "baseName": "shrToken",
    "type": "string"
  },
  {
    "name": "zId",
    "baseName": "zId",
    "type": "string"
  },
  {
    "name": "createdAt",
    "baseName": "createdAt",
    "type": "number"
  },
  {
    "name": "updatedAt",
    "baseName": "updatedAt",
    "type": "number"
  }
];
var Frontend = _Frontend;

// src/zrok/api/model/inviteRequest.ts
var _InviteRequest = class _InviteRequest {
  static getAttributeTypeMap() {
    return _InviteRequest.attributeTypeMap;
  }
};
_InviteRequest.discriminator = void 0;
_InviteRequest.attributeTypeMap = [
  {
    "name": "email",
    "baseName": "email",
    "type": "string"
  },
  {
    "name": "token",
    "baseName": "token",
    "type": "string"
  }
];
var InviteRequest = _InviteRequest;

// src/zrok/api/model/inviteTokenGenerateRequest.ts
var _InviteTokenGenerateRequest = class _InviteTokenGenerateRequest {
  static getAttributeTypeMap() {
    return _InviteTokenGenerateRequest.attributeTypeMap;
  }
};
_InviteTokenGenerateRequest.discriminator = void 0;
_InviteTokenGenerateRequest.attributeTypeMap = [
  {
    "name": "tokens",
    "baseName": "tokens",
    "type": "Array<string>"
  }
];
var InviteTokenGenerateRequest = _InviteTokenGenerateRequest;

// src/zrok/api/model/loginRequest.ts
var _LoginRequest = class _LoginRequest {
  static getAttributeTypeMap() {
    return _LoginRequest.attributeTypeMap;
  }
};
_LoginRequest.discriminator = void 0;
_LoginRequest.attributeTypeMap = [
  {
    "name": "email",
    "baseName": "email",
    "type": "string"
  },
  {
    "name": "password",
    "baseName": "password",
    "type": "string"
  }
];
var LoginRequest = _LoginRequest;

// src/zrok/api/model/metrics.ts
var _Metrics = class _Metrics {
  static getAttributeTypeMap() {
    return _Metrics.attributeTypeMap;
  }
};
_Metrics.discriminator = void 0;
_Metrics.attributeTypeMap = [
  {
    "name": "scope",
    "baseName": "scope",
    "type": "string"
  },
  {
    "name": "id",
    "baseName": "id",
    "type": "string"
  },
  {
    "name": "period",
    "baseName": "period",
    "type": "number"
  },
  {
    "name": "samples",
    "baseName": "samples",
    "type": "Array<MetricsSample>"
  }
];
var Metrics = _Metrics;

// src/zrok/api/model/metricsSample.ts
var _MetricsSample = class _MetricsSample {
  static getAttributeTypeMap() {
    return _MetricsSample.attributeTypeMap;
  }
};
_MetricsSample.discriminator = void 0;
_MetricsSample.attributeTypeMap = [
  {
    "name": "rx",
    "baseName": "rx",
    "type": "number"
  },
  {
    "name": "tx",
    "baseName": "tx",
    "type": "number"
  },
  {
    "name": "timestamp",
    "baseName": "timestamp",
    "type": "number"
  }
];
var MetricsSample = _MetricsSample;

// src/zrok/api/model/overview.ts
var _Overview = class _Overview {
  static getAttributeTypeMap() {
    return _Overview.attributeTypeMap;
  }
};
_Overview.discriminator = void 0;
_Overview.attributeTypeMap = [
  {
    "name": "accountLimited",
    "baseName": "accountLimited",
    "type": "boolean"
  },
  {
    "name": "environments",
    "baseName": "environments",
    "type": "Array<EnvironmentAndResources>"
  }
];
var Overview = _Overview;

// src/zrok/api/model/passwordRequirements.ts
var _PasswordRequirements = class _PasswordRequirements {
  static getAttributeTypeMap() {
    return _PasswordRequirements.attributeTypeMap;
  }
};
_PasswordRequirements.discriminator = void 0;
_PasswordRequirements.attributeTypeMap = [
  {
    "name": "length",
    "baseName": "length",
    "type": "number"
  },
  {
    "name": "requireCapital",
    "baseName": "requireCapital",
    "type": "boolean"
  },
  {
    "name": "requireNumeric",
    "baseName": "requireNumeric",
    "type": "boolean"
  },
  {
    "name": "requireSpecial",
    "baseName": "requireSpecial",
    "type": "boolean"
  },
  {
    "name": "validSpecialCharacters",
    "baseName": "validSpecialCharacters",
    "type": "string"
  }
];
var PasswordRequirements = _PasswordRequirements;

// src/zrok/api/model/principal.ts
var _Principal = class _Principal {
  static getAttributeTypeMap() {
    return _Principal.attributeTypeMap;
  }
};
_Principal.discriminator = void 0;
_Principal.attributeTypeMap = [
  {
    "name": "id",
    "baseName": "id",
    "type": "number"
  },
  {
    "name": "email",
    "baseName": "email",
    "type": "string"
  },
  {
    "name": "token",
    "baseName": "token",
    "type": "string"
  },
  {
    "name": "limitless",
    "baseName": "limitless",
    "type": "boolean"
  },
  {
    "name": "admin",
    "baseName": "admin",
    "type": "boolean"
  }
];
var Principal = _Principal;

// src/zrok/api/model/publicFrontend.ts
var _PublicFrontend = class _PublicFrontend {
  static getAttributeTypeMap() {
    return _PublicFrontend.attributeTypeMap;
  }
};
_PublicFrontend.discriminator = void 0;
_PublicFrontend.attributeTypeMap = [
  {
    "name": "token",
    "baseName": "token",
    "type": "string"
  },
  {
    "name": "zId",
    "baseName": "zId",
    "type": "string"
  },
  {
    "name": "urlTemplate",
    "baseName": "urlTemplate",
    "type": "string"
  },
  {
    "name": "publicName",
    "baseName": "publicName",
    "type": "string"
  },
  {
    "name": "createdAt",
    "baseName": "createdAt",
    "type": "number"
  },
  {
    "name": "updatedAt",
    "baseName": "updatedAt",
    "type": "number"
  }
];
var PublicFrontend = _PublicFrontend;

// src/zrok/api/model/registerRequest.ts
var _RegisterRequest = class _RegisterRequest {
  static getAttributeTypeMap() {
    return _RegisterRequest.attributeTypeMap;
  }
};
_RegisterRequest.discriminator = void 0;
_RegisterRequest.attributeTypeMap = [
  {
    "name": "token",
    "baseName": "token",
    "type": "string"
  },
  {
    "name": "password",
    "baseName": "password",
    "type": "string"
  }
];
var RegisterRequest = _RegisterRequest;

// src/zrok/api/model/registerResponse.ts
var _RegisterResponse = class _RegisterResponse {
  static getAttributeTypeMap() {
    return _RegisterResponse.attributeTypeMap;
  }
};
_RegisterResponse.discriminator = void 0;
_RegisterResponse.attributeTypeMap = [
  {
    "name": "token",
    "baseName": "token",
    "type": "string"
  }
];
var RegisterResponse = _RegisterResponse;

// src/zrok/api/model/resetPasswordRequest.ts
var _ResetPasswordRequest = class _ResetPasswordRequest {
  static getAttributeTypeMap() {
    return _ResetPasswordRequest.attributeTypeMap;
  }
};
_ResetPasswordRequest.discriminator = void 0;
_ResetPasswordRequest.attributeTypeMap = [
  {
    "name": "token",
    "baseName": "token",
    "type": "string"
  },
  {
    "name": "password",
    "baseName": "password",
    "type": "string"
  }
];
var ResetPasswordRequest = _ResetPasswordRequest;

// src/zrok/api/model/resetPasswordRequestRequest.ts
var _ResetPasswordRequestRequest = class _ResetPasswordRequestRequest {
  static getAttributeTypeMap() {
    return _ResetPasswordRequestRequest.attributeTypeMap;
  }
};
_ResetPasswordRequestRequest.discriminator = void 0;
_ResetPasswordRequestRequest.attributeTypeMap = [
  {
    "name": "emailAddress",
    "baseName": "emailAddress",
    "type": "string"
  }
];
var ResetPasswordRequestRequest = _ResetPasswordRequestRequest;

// src/zrok/api/model/share.ts
var _Share = class _Share {
  static getAttributeTypeMap() {
    return _Share.attributeTypeMap;
  }
};
_Share.discriminator = void 0;
_Share.attributeTypeMap = [
  {
    "name": "token",
    "baseName": "token",
    "type": "string"
  },
  {
    "name": "zId",
    "baseName": "zId",
    "type": "string"
  },
  {
    "name": "shareMode",
    "baseName": "shareMode",
    "type": "string"
  },
  {
    "name": "backendMode",
    "baseName": "backendMode",
    "type": "string"
  },
  {
    "name": "frontendSelection",
    "baseName": "frontendSelection",
    "type": "string"
  },
  {
    "name": "frontendEndpoint",
    "baseName": "frontendEndpoint",
    "type": "string"
  },
  {
    "name": "backendProxyEndpoint",
    "baseName": "backendProxyEndpoint",
    "type": "string"
  },
  {
    "name": "reserved",
    "baseName": "reserved",
    "type": "boolean"
  },
  {
    "name": "activity",
    "baseName": "activity",
    "type": "Array<SparkDataSample>"
  },
  {
    "name": "limited",
    "baseName": "limited",
    "type": "boolean"
  },
  {
    "name": "createdAt",
    "baseName": "createdAt",
    "type": "number"
  },
  {
    "name": "updatedAt",
    "baseName": "updatedAt",
    "type": "number"
  }
];
var Share = _Share;

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

// src/zrok/api/model/shareResponse.ts
var _ShareResponse = class _ShareResponse {
  static getAttributeTypeMap() {
    return _ShareResponse.attributeTypeMap;
  }
};
_ShareResponse.discriminator = void 0;
_ShareResponse.attributeTypeMap = [
  {
    "name": "frontendProxyEndpoints",
    "baseName": "frontendProxyEndpoints",
    "type": "Array<string>"
  },
  {
    "name": "shrToken",
    "baseName": "shrToken",
    "type": "string"
  }
];
var ShareResponse = _ShareResponse;

// src/zrok/api/model/sparkDataSample.ts
var _SparkDataSample = class _SparkDataSample {
  static getAttributeTypeMap() {
    return _SparkDataSample.attributeTypeMap;
  }
};
_SparkDataSample.discriminator = void 0;
_SparkDataSample.attributeTypeMap = [
  {
    "name": "rx",
    "baseName": "rx",
    "type": "number"
  },
  {
    "name": "tx",
    "baseName": "tx",
    "type": "number"
  }
];
var SparkDataSample = _SparkDataSample;

// src/zrok/api/model/unaccessRequest.ts
var _UnaccessRequest = class _UnaccessRequest {
  static getAttributeTypeMap() {
    return _UnaccessRequest.attributeTypeMap;
  }
};
_UnaccessRequest.discriminator = void 0;
_UnaccessRequest.attributeTypeMap = [
  {
    "name": "frontendToken",
    "baseName": "frontendToken",
    "type": "string"
  },
  {
    "name": "envZId",
    "baseName": "envZId",
    "type": "string"
  },
  {
    "name": "shrToken",
    "baseName": "shrToken",
    "type": "string"
  }
];
var UnaccessRequest = _UnaccessRequest;

// src/zrok/api/model/unshareRequest.ts
var _UnshareRequest = class _UnshareRequest {
  static getAttributeTypeMap() {
    return _UnshareRequest.attributeTypeMap;
  }
};
_UnshareRequest.discriminator = void 0;
_UnshareRequest.attributeTypeMap = [
  {
    "name": "envZId",
    "baseName": "envZId",
    "type": "string"
  },
  {
    "name": "shrToken",
    "baseName": "shrToken",
    "type": "string"
  },
  {
    "name": "reserved",
    "baseName": "reserved",
    "type": "boolean"
  }
];
var UnshareRequest = _UnshareRequest;

// src/zrok/api/model/updateFrontendRequest.ts
var _UpdateFrontendRequest = class _UpdateFrontendRequest {
  static getAttributeTypeMap() {
    return _UpdateFrontendRequest.attributeTypeMap;
  }
};
_UpdateFrontendRequest.discriminator = void 0;
_UpdateFrontendRequest.attributeTypeMap = [
  {
    "name": "frontendToken",
    "baseName": "frontendToken",
    "type": "string"
  },
  {
    "name": "publicName",
    "baseName": "publicName",
    "type": "string"
  },
  {
    "name": "urlTemplate",
    "baseName": "urlTemplate",
    "type": "string"
  }
];
var UpdateFrontendRequest = _UpdateFrontendRequest;

// src/zrok/api/model/updateShareRequest.ts
var _UpdateShareRequest = class _UpdateShareRequest {
  static getAttributeTypeMap() {
    return _UpdateShareRequest.attributeTypeMap;
  }
};
_UpdateShareRequest.discriminator = void 0;
_UpdateShareRequest.attributeTypeMap = [
  {
    "name": "shrToken",
    "baseName": "shrToken",
    "type": "string"
  },
  {
    "name": "backendProxyEndpoint",
    "baseName": "backendProxyEndpoint",
    "type": "string"
  }
];
var UpdateShareRequest = _UpdateShareRequest;

// src/zrok/api/model/verifyRequest.ts
var _VerifyRequest = class _VerifyRequest {
  static getAttributeTypeMap() {
    return _VerifyRequest.attributeTypeMap;
  }
};
_VerifyRequest.discriminator = void 0;
_VerifyRequest.attributeTypeMap = [
  {
    "name": "token",
    "baseName": "token",
    "type": "string"
  }
];
var VerifyRequest = _VerifyRequest;

// src/zrok/api/model/verifyResponse.ts
var _VerifyResponse = class _VerifyResponse {
  static getAttributeTypeMap() {
    return _VerifyResponse.attributeTypeMap;
  }
};
_VerifyResponse.discriminator = void 0;
_VerifyResponse.attributeTypeMap = [
  {
    "name": "email",
    "baseName": "email",
    "type": "string"
  }
];
var VerifyResponse = _VerifyResponse;

// src/zrok/api/model/models.ts
var primitives = [
  "string",
  "boolean",
  "double",
  "integer",
  "long",
  "float",
  "number",
  "any"
];
var enumsMap = {
  "ShareRequest.ShareModeEnum": ShareRequest.ShareModeEnum,
  "ShareRequest.BackendModeEnum": ShareRequest.BackendModeEnum,
  "ShareRequest.OauthProviderEnum": ShareRequest.OauthProviderEnum
};
var typeMap = {
  "AccessRequest": AccessRequest,
  "AccessResponse": AccessResponse,
  "AuthUser": AuthUser,
  "Configuration": Configuration,
  "CreateFrontendRequest": CreateFrontendRequest,
  "CreateFrontendResponse": CreateFrontendResponse,
  "CreateIdentity201Response": CreateIdentity201Response,
  "CreateIdentityRequest": CreateIdentityRequest,
  "DeleteFrontendRequest": DeleteFrontendRequest,
  "DisableRequest": DisableRequest,
  "EnableRequest": EnableRequest,
  "EnableResponse": EnableResponse,
  "Environment": Environment,
  "EnvironmentAndResources": EnvironmentAndResources,
  "Frontend": Frontend,
  "InviteRequest": InviteRequest,
  "InviteTokenGenerateRequest": InviteTokenGenerateRequest,
  "LoginRequest": LoginRequest,
  "Metrics": Metrics,
  "MetricsSample": MetricsSample,
  "Overview": Overview,
  "PasswordRequirements": PasswordRequirements,
  "Principal": Principal,
  "PublicFrontend": PublicFrontend,
  "RegisterRequest": RegisterRequest,
  "RegisterResponse": RegisterResponse,
  "ResetPasswordRequest": ResetPasswordRequest,
  "ResetPasswordRequestRequest": ResetPasswordRequestRequest,
  "Share": Share,
  "ShareRequest": ShareRequest,
  "ShareResponse": ShareResponse,
  "SparkDataSample": SparkDataSample,
  "UnaccessRequest": UnaccessRequest,
  "UnshareRequest": UnshareRequest,
  "UpdateFrontendRequest": UpdateFrontendRequest,
  "UpdateShareRequest": UpdateShareRequest,
  "VerifyRequest": VerifyRequest,
  "VerifyResponse": VerifyResponse
};
var ObjectSerializer = class _ObjectSerializer {
  static findCorrectType(data, expectedType) {
    if (data == void 0) {
      return expectedType;
    } else if (primitives.indexOf(expectedType.toLowerCase()) !== -1) {
      return expectedType;
    } else if (expectedType === "Date") {
      return expectedType;
    } else {
      if (enumsMap[expectedType]) {
        return expectedType;
      }
      if (!typeMap[expectedType]) {
        return expectedType;
      }
      let discriminatorProperty = typeMap[expectedType].discriminator;
      if (discriminatorProperty == null) {
        return expectedType;
      } else {
        if (data[discriminatorProperty]) {
          var discriminatorType = data[discriminatorProperty];
          if (typeMap[discriminatorType]) {
            return discriminatorType;
          } else {
            return expectedType;
          }
        } else {
          return expectedType;
        }
      }
    }
  }
  static serialize(data, type) {
    if (data == void 0) {
      return data;
    } else if (primitives.indexOf(type.toLowerCase()) !== -1) {
      return data;
    } else if (type.lastIndexOf("Array<", 0) === 0) {
      let subType = type.replace("Array<", "");
      subType = subType.substring(0, subType.length - 1);
      let transformedData = [];
      for (let index = 0; index < data.length; index++) {
        let datum = data[index];
        transformedData.push(_ObjectSerializer.serialize(datum, subType));
      }
      return transformedData;
    } else if (type === "Date") {
      return data.toISOString();
    } else {
      if (enumsMap[type]) {
        return data;
      }
      if (!typeMap[type]) {
        return data;
      }
      type = this.findCorrectType(data, type);
      let attributeTypes = typeMap[type].getAttributeTypeMap();
      let instance = {};
      for (let index = 0; index < attributeTypes.length; index++) {
        let attributeType = attributeTypes[index];
        instance[attributeType.baseName] = _ObjectSerializer.serialize(data[attributeType.name], attributeType.type);
      }
      return instance;
    }
  }
  static deserialize(data, type) {
    type = _ObjectSerializer.findCorrectType(data, type);
    if (data == void 0) {
      return data;
    } else if (primitives.indexOf(type.toLowerCase()) !== -1) {
      return data;
    } else if (type.lastIndexOf("Array<", 0) === 0) {
      let subType = type.replace("Array<", "");
      subType = subType.substring(0, subType.length - 1);
      let transformedData = [];
      for (let index = 0; index < data.length; index++) {
        let datum = data[index];
        transformedData.push(_ObjectSerializer.deserialize(datum, subType));
      }
      return transformedData;
    } else if (type === "Date") {
      return new Date(data);
    } else {
      if (enumsMap[type]) {
        return data;
      }
      if (!typeMap[type]) {
        return data;
      }
      let instance = new typeMap[type]();
      let attributeTypes = typeMap[type].getAttributeTypeMap();
      for (let index = 0; index < attributeTypes.length; index++) {
        let attributeType = attributeTypes[index];
        instance[attributeType.name] = _ObjectSerializer.deserialize(data[attributeType.baseName], attributeType.type);
      }
      return instance;
    }
  }
};
var ApiKeyAuth = class {
  constructor(location, paramName) {
    this.location = location;
    this.paramName = paramName;
    this.apiKey = "";
  }
  applyToRequest(requestOptions) {
    if (this.location == "query") {
      requestOptions.qs[this.paramName] = this.apiKey;
    } else if (this.location == "header" && requestOptions && requestOptions.headers) {
      requestOptions.headers[this.paramName] = this.apiKey;
    } else if (this.location == "cookie" && requestOptions && requestOptions.headers) {
      if (requestOptions.headers["Cookie"]) {
        requestOptions.headers["Cookie"] += "; " + this.paramName + "=" + encodeURIComponent(this.apiKey);
      } else {
        requestOptions.headers["Cookie"] = this.paramName + "=" + encodeURIComponent(this.apiKey);
      }
    }
  }
};
var VoidAuth = class {
  constructor() {
    this.username = "";
    this.password = "";
  }
  applyToRequest(_) {
  }
};

// src/zrok/api/api/accountApi.ts
import localVarRequest from "request";

// src/zrok/api/api/adminApi.ts
import localVarRequest2 from "request";

// src/zrok/api/api/metadataApi.ts
import localVarRequest3 from "request";

// src/zrok/api/api/shareApi.ts
import localVarRequest4 from "request";

// src/zrok/api/api/apis.ts
var HttpError = class extends Error {
  constructor(response, body, statusCode) {
    super("HTTP request failed");
    this.response = response;
    this.body = body;
    this.statusCode = statusCode;
    this.name = "HttpError";
  }
};

// src/zrok/api/api/environmentApi.ts
var defaultBasePath = "/api/v1";
var EnvironmentApiApiKeys = /* @__PURE__ */ ((EnvironmentApiApiKeys2) => {
  EnvironmentApiApiKeys2[EnvironmentApiApiKeys2["key"] = 0] = "key";
  return EnvironmentApiApiKeys2;
})(EnvironmentApiApiKeys || {});
var EnvironmentApi = class {
  constructor(basePathOrUsername, password, basePath) {
    this._basePath = defaultBasePath;
    this._defaultHeaders = {};
    this._useQuerystring = false;
    this.authentications = {
      "default": new VoidAuth(),
      "key": new ApiKeyAuth("header", "x-token")
    };
    this.interceptors = [];
    if (password) {
      if (basePath) {
        this.basePath = basePath;
      }
    } else {
      if (basePathOrUsername) {
        this.basePath = basePathOrUsername;
      }
    }
  }
  set useQuerystring(value) {
    this._useQuerystring = value;
  }
  set basePath(basePath) {
    this._basePath = basePath;
  }
  set defaultHeaders(defaultHeaders) {
    this._defaultHeaders = defaultHeaders;
  }
  get defaultHeaders() {
    return this._defaultHeaders;
  }
  get basePath() {
    return this._basePath;
  }
  setDefaultAuthentication(auth) {
    this.authentications.default = auth;
  }
  setApiKey(key, value) {
    this.authentications[EnvironmentApiApiKeys[key]].apiKey = value;
  }
  addInterceptor(interceptor) {
    this.interceptors.push(interceptor);
  }
  /**
   * 
   * @param body 
   */
  disable(_0) {
    return __async(this, arguments, function* (body, options = { headers: {} }) {
      const localVarPath = this.basePath + "/disable";
      let localVarQueryParameters = {};
      let localVarHeaderParams = Object.assign({}, this._defaultHeaders);
      let localVarFormParams = {};
      Object.assign(localVarHeaderParams, options.headers);
      let localVarUseFormData = false;
      let localVarRequestOptions = {
        method: "POST",
        qs: localVarQueryParameters,
        headers: localVarHeaderParams,
        uri: localVarPath,
        useQuerystring: this._useQuerystring,
        json: true,
        body: ObjectSerializer.serialize(body, "DisableRequest")
      };
      let authenticationPromise = Promise.resolve();
      if (this.authentications.key.apiKey) {
        authenticationPromise = authenticationPromise.then(() => this.authentications.key.applyToRequest(localVarRequestOptions));
      }
      authenticationPromise = authenticationPromise.then(() => this.authentications.default.applyToRequest(localVarRequestOptions));
      let interceptorPromise = authenticationPromise;
      for (const interceptor of this.interceptors) {
        interceptorPromise = interceptorPromise.then(() => interceptor(localVarRequestOptions));
      }
      return interceptorPromise.then(() => {
        if (Object.keys(localVarFormParams).length) {
          if (localVarUseFormData) {
            localVarRequestOptions.formData = localVarFormParams;
          } else {
            localVarRequestOptions.form = localVarFormParams;
          }
        }
        return new Promise((resolve, reject) => {
          localVarRequest5(localVarRequestOptions, (error, response, body2) => {
            if (error) {
              reject(error);
            } else {
              if (response.statusCode && response.statusCode >= 200 && response.statusCode <= 299) {
                resolve({ response, body: body2 });
              } else {
                reject(new HttpError(response, body2, response.statusCode));
              }
            }
          });
        });
      });
    });
  }
  /**
   * 
   * @param body 
   */
  enable(_0) {
    return __async(this, arguments, function* (body, options = { headers: {} }) {
      const localVarPath = this.basePath + "/enable";
      let localVarQueryParameters = {};
      let localVarHeaderParams = Object.assign({}, this._defaultHeaders);
      const produces = ["application/zrok.v1+json"];
      if (produces.indexOf("application/json") >= 0) {
        localVarHeaderParams.Accept = "application/json";
      } else {
        localVarHeaderParams.Accept = produces.join(",");
      }
      let localVarFormParams = {};
      Object.assign(localVarHeaderParams, options.headers);
      let localVarUseFormData = false;
      let localVarRequestOptions = {
        method: "POST",
        qs: localVarQueryParameters,
        headers: localVarHeaderParams,
        uri: localVarPath,
        useQuerystring: this._useQuerystring,
        json: true,
        body: ObjectSerializer.serialize(body, "EnableRequest")
      };
      let authenticationPromise = Promise.resolve();
      if (this.authentications.key.apiKey) {
        authenticationPromise = authenticationPromise.then(() => this.authentications.key.applyToRequest(localVarRequestOptions));
      }
      authenticationPromise = authenticationPromise.then(() => this.authentications.default.applyToRequest(localVarRequestOptions));
      let interceptorPromise = authenticationPromise;
      for (const interceptor of this.interceptors) {
        interceptorPromise = interceptorPromise.then(() => interceptor(localVarRequestOptions));
      }
      return interceptorPromise.then(() => {
        if (Object.keys(localVarFormParams).length) {
          if (localVarUseFormData) {
            localVarRequestOptions.formData = localVarFormParams;
          } else {
            localVarRequestOptions.form = localVarFormParams;
          }
        }
        return new Promise((resolve, reject) => {
          localVarRequest5(localVarRequestOptions, (error, response, body2) => {
            if (error) {
              reject(error);
            } else {
              if (response.statusCode && response.statusCode >= 200 && response.statusCode <= 299) {
                body2 = ObjectSerializer.deserialize(body2, "EnableResponse");
                resolve({ response, body: body2 });
              } else {
                reject(new HttpError(response, body2, response.statusCode));
              }
            }
          });
        });
      });
    });
  }
};
export {
  EnvironmentApi,
  EnvironmentApiApiKeys
};
//# sourceMappingURL=environmentApi.mjs.map