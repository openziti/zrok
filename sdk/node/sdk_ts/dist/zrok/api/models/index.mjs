var __defProp = Object.defineProperty;
var __defProps = Object.defineProperties;
var __getOwnPropDescs = Object.getOwnPropertyDescriptors;
var __getOwnPropSymbols = Object.getOwnPropertySymbols;
var __hasOwnProp = Object.prototype.hasOwnProperty;
var __propIsEnum = Object.prototype.propertyIsEnumerable;
var __defNormalProp = (obj, key, value) => key in obj ? __defProp(obj, key, { enumerable: true, configurable: true, writable: true, value }) : obj[key] = value;
var __spreadValues = (a, b) => {
  for (var prop in b || (b = {}))
    if (__hasOwnProp.call(b, prop))
      __defNormalProp(a, prop, b[prop]);
  if (__getOwnPropSymbols)
    for (var prop of __getOwnPropSymbols(b)) {
      if (__propIsEnum.call(b, prop))
        __defNormalProp(a, prop, b[prop]);
    }
  return a;
};
var __spreadProps = (a, b) => __defProps(a, __getOwnPropDescs(b));
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

// src/zrok/api/runtime.ts
var BASE_PATH = "/api/v1".replace(/\/+$/, "");
var Configuration = class {
  constructor(configuration = {}) {
    this.configuration = configuration;
  }
  set config(configuration) {
    this.configuration = configuration;
  }
  get basePath() {
    return this.configuration.basePath != null ? this.configuration.basePath : BASE_PATH;
  }
  get fetchApi() {
    return this.configuration.fetchApi;
  }
  get middleware() {
    return this.configuration.middleware || [];
  }
  get queryParamsStringify() {
    return this.configuration.queryParamsStringify || querystring;
  }
  get username() {
    return this.configuration.username;
  }
  get password() {
    return this.configuration.password;
  }
  get apiKey() {
    const apiKey = this.configuration.apiKey;
    if (apiKey) {
      return typeof apiKey === "function" ? apiKey : () => apiKey;
    }
    return void 0;
  }
  get accessToken() {
    const accessToken = this.configuration.accessToken;
    if (accessToken) {
      return typeof accessToken === "function" ? accessToken : () => __async(this, null, function* () {
        return accessToken;
      });
    }
    return void 0;
  }
  get headers() {
    return this.configuration.headers;
  }
  get credentials() {
    return this.configuration.credentials;
  }
};
var DefaultConfig = new Configuration();
var _BaseAPI = class _BaseAPI {
  constructor(configuration = DefaultConfig) {
    this.configuration = configuration;
    this.fetchApi = (url, init) => __async(this, null, function* () {
      let fetchParams = { url, init };
      for (const middleware of this.middleware) {
        if (middleware.pre) {
          fetchParams = (yield middleware.pre(__spreadValues({
            fetch: this.fetchApi
          }, fetchParams))) || fetchParams;
        }
      }
      let response = void 0;
      try {
        response = yield (this.configuration.fetchApi || fetch)(fetchParams.url, fetchParams.init);
      } catch (e) {
        for (const middleware of this.middleware) {
          if (middleware.onError) {
            response = (yield middleware.onError({
              fetch: this.fetchApi,
              url: fetchParams.url,
              init: fetchParams.init,
              error: e,
              response: response ? response.clone() : void 0
            })) || response;
          }
        }
        if (response === void 0) {
          if (e instanceof Error) {
            throw new FetchError(e, "The request failed and the interceptors did not return an alternative response");
          } else {
            throw e;
          }
        }
      }
      for (const middleware of this.middleware) {
        if (middleware.post) {
          response = (yield middleware.post({
            fetch: this.fetchApi,
            url: fetchParams.url,
            init: fetchParams.init,
            response: response.clone()
          })) || response;
        }
      }
      return response;
    });
    this.middleware = configuration.middleware;
  }
  withMiddleware(...middlewares) {
    const next = this.clone();
    next.middleware = next.middleware.concat(...middlewares);
    return next;
  }
  withPreMiddleware(...preMiddlewares) {
    const middlewares = preMiddlewares.map((pre) => ({ pre }));
    return this.withMiddleware(...middlewares);
  }
  withPostMiddleware(...postMiddlewares) {
    const middlewares = postMiddlewares.map((post) => ({ post }));
    return this.withMiddleware(...middlewares);
  }
  /**
   * Check if the given MIME is a JSON MIME.
   * JSON MIME examples:
   *   application/json
   *   application/json; charset=UTF8
   *   APPLICATION/JSON
   *   application/vnd.company+json
   * @param mime - MIME (Multipurpose Internet Mail Extensions)
   * @return True if the given MIME is JSON, false otherwise.
   */
  isJsonMime(mime) {
    if (!mime) {
      return false;
    }
    return _BaseAPI.jsonRegex.test(mime);
  }
  request(context, initOverrides) {
    return __async(this, null, function* () {
      const { url, init } = yield this.createFetchParams(context, initOverrides);
      const response = yield this.fetchApi(url, init);
      if (response && (response.status >= 200 && response.status < 300)) {
        return response;
      }
      throw new ResponseError(response, "Response returned an error code");
    });
  }
  createFetchParams(context, initOverrides) {
    return __async(this, null, function* () {
      let url = this.configuration.basePath + context.path;
      if (context.query !== void 0 && Object.keys(context.query).length !== 0) {
        url += "?" + this.configuration.queryParamsStringify(context.query);
      }
      const headers = Object.assign({}, this.configuration.headers, context.headers);
      Object.keys(headers).forEach((key) => headers[key] === void 0 ? delete headers[key] : {});
      const initOverrideFn = typeof initOverrides === "function" ? initOverrides : () => __async(this, null, function* () {
        return initOverrides;
      });
      const initParams = {
        method: context.method,
        headers,
        body: context.body,
        credentials: this.configuration.credentials
      };
      const overriddenInit = __spreadValues(__spreadValues({}, initParams), yield initOverrideFn({
        init: initParams,
        context
      }));
      let body;
      if (isFormData(overriddenInit.body) || overriddenInit.body instanceof URLSearchParams || isBlob(overriddenInit.body)) {
        body = overriddenInit.body;
      } else if (this.isJsonMime(headers["Content-Type"])) {
        body = JSON.stringify(overriddenInit.body);
      } else {
        body = overriddenInit.body;
      }
      const init = __spreadProps(__spreadValues({}, overriddenInit), {
        body
      });
      return { url, init };
    });
  }
  /**
   * Create a shallow clone of `this` by constructing a new instance
   * and then shallow cloning data members.
   */
  clone() {
    const constructor = this.constructor;
    const next = new constructor(this.configuration);
    next.middleware = this.middleware.slice();
    return next;
  }
};
_BaseAPI.jsonRegex = new RegExp("^(:?application/json|[^;/ 	]+/[^;/ 	]+[+]json)[ 	]*(:?;.*)?$", "i");
var BaseAPI = _BaseAPI;
function isBlob(value) {
  return typeof Blob !== "undefined" && value instanceof Blob;
}
function isFormData(value) {
  return typeof FormData !== "undefined" && value instanceof FormData;
}
var ResponseError = class extends Error {
  constructor(response, msg) {
    super(msg);
    this.response = response;
    this.name = "ResponseError";
  }
};
var FetchError = class extends Error {
  constructor(cause, msg) {
    super(msg);
    this.cause = cause;
    this.name = "FetchError";
  }
};
function exists(json, key) {
  const value = json[key];
  return value !== null && value !== void 0;
}
function querystring(params, prefix = "") {
  return Object.keys(params).map((key) => querystringSingleKey(key, params[key], prefix)).filter((part) => part.length > 0).join("&");
}
function querystringSingleKey(key, value, keyPrefix = "") {
  const fullKey = keyPrefix + (keyPrefix.length ? `[${key}]` : key);
  if (value instanceof Array) {
    const multiValue = value.map((singleValue) => encodeURIComponent(String(singleValue))).join(`&${encodeURIComponent(fullKey)}=`);
    return `${encodeURIComponent(fullKey)}=${multiValue}`;
  }
  if (value instanceof Set) {
    const valueAsArray = Array.from(value);
    return querystringSingleKey(key, valueAsArray, keyPrefix);
  }
  if (value instanceof Date) {
    return `${encodeURIComponent(fullKey)}=${encodeURIComponent(value.toISOString())}`;
  }
  if (value instanceof Object) {
    return querystring(value, fullKey);
  }
  return `${encodeURIComponent(fullKey)}=${encodeURIComponent(String(value))}`;
}

// src/zrok/api/models/AccessRequest.ts
function instanceOfAccessRequest(value) {
  let isInstance = true;
  return isInstance;
}
function AccessRequestFromJSON(json) {
  return AccessRequestFromJSONTyped(json, false);
}
function AccessRequestFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "envZId": !exists(json, "envZId") ? void 0 : json["envZId"],
    "shrToken": !exists(json, "shrToken") ? void 0 : json["shrToken"]
  };
}
function AccessRequestToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "envZId": value.envZId,
    "shrToken": value.shrToken
  };
}

// src/zrok/api/models/AccessResponse.ts
function instanceOfAccessResponse(value) {
  let isInstance = true;
  return isInstance;
}
function AccessResponseFromJSON(json) {
  return AccessResponseFromJSONTyped(json, false);
}
function AccessResponseFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "frontendToken": !exists(json, "frontendToken") ? void 0 : json["frontendToken"],
    "backendMode": !exists(json, "backendMode") ? void 0 : json["backendMode"]
  };
}
function AccessResponseToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "frontendToken": value.frontendToken,
    "backendMode": value.backendMode
  };
}

// src/zrok/api/models/AuthUser.ts
function instanceOfAuthUser(value) {
  let isInstance = true;
  return isInstance;
}
function AuthUserFromJSON(json) {
  return AuthUserFromJSONTyped(json, false);
}
function AuthUserFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "username": !exists(json, "username") ? void 0 : json["username"],
    "password": !exists(json, "password") ? void 0 : json["password"]
  };
}
function AuthUserToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "username": value.username,
    "password": value.password
  };
}

// src/zrok/api/models/CreateFrontendRequest.ts
function instanceOfCreateFrontendRequest(value) {
  let isInstance = true;
  return isInstance;
}
function CreateFrontendRequestFromJSON(json) {
  return CreateFrontendRequestFromJSONTyped(json, false);
}
function CreateFrontendRequestFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "zId": !exists(json, "zId") ? void 0 : json["zId"],
    "urlTemplate": !exists(json, "url_template") ? void 0 : json["url_template"],
    "publicName": !exists(json, "public_name") ? void 0 : json["public_name"]
  };
}
function CreateFrontendRequestToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "zId": value.zId,
    "url_template": value.urlTemplate,
    "public_name": value.publicName
  };
}

// src/zrok/api/models/CreateFrontendResponse.ts
function instanceOfCreateFrontendResponse(value) {
  let isInstance = true;
  return isInstance;
}
function CreateFrontendResponseFromJSON(json) {
  return CreateFrontendResponseFromJSONTyped(json, false);
}
function CreateFrontendResponseFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "token": !exists(json, "token") ? void 0 : json["token"]
  };
}
function CreateFrontendResponseToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "token": value.token
  };
}

// src/zrok/api/models/CreateIdentity201Response.ts
function instanceOfCreateIdentity201Response(value) {
  let isInstance = true;
  return isInstance;
}
function CreateIdentity201ResponseFromJSON(json) {
  return CreateIdentity201ResponseFromJSONTyped(json, false);
}
function CreateIdentity201ResponseFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "identity": !exists(json, "identity") ? void 0 : json["identity"],
    "cfg": !exists(json, "cfg") ? void 0 : json["cfg"]
  };
}
function CreateIdentity201ResponseToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "identity": value.identity,
    "cfg": value.cfg
  };
}

// src/zrok/api/models/CreateIdentityRequest.ts
function instanceOfCreateIdentityRequest(value) {
  let isInstance = true;
  return isInstance;
}
function CreateIdentityRequestFromJSON(json) {
  return CreateIdentityRequestFromJSONTyped(json, false);
}
function CreateIdentityRequestFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "name": !exists(json, "name") ? void 0 : json["name"]
  };
}
function CreateIdentityRequestToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "name": value.name
  };
}

// src/zrok/api/models/DeleteFrontendRequest.ts
function instanceOfDeleteFrontendRequest(value) {
  let isInstance = true;
  return isInstance;
}
function DeleteFrontendRequestFromJSON(json) {
  return DeleteFrontendRequestFromJSONTyped(json, false);
}
function DeleteFrontendRequestFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "frontendToken": !exists(json, "frontendToken") ? void 0 : json["frontendToken"]
  };
}
function DeleteFrontendRequestToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "frontendToken": value.frontendToken
  };
}

// src/zrok/api/models/DisableRequest.ts
function instanceOfDisableRequest(value) {
  let isInstance = true;
  return isInstance;
}
function DisableRequestFromJSON(json) {
  return DisableRequestFromJSONTyped(json, false);
}
function DisableRequestFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "identity": !exists(json, "identity") ? void 0 : json["identity"]
  };
}
function DisableRequestToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "identity": value.identity
  };
}

// src/zrok/api/models/EnableRequest.ts
function instanceOfEnableRequest(value) {
  let isInstance = true;
  return isInstance;
}
function EnableRequestFromJSON(json) {
  return EnableRequestFromJSONTyped(json, false);
}
function EnableRequestFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "description": !exists(json, "description") ? void 0 : json["description"],
    "host": !exists(json, "host") ? void 0 : json["host"]
  };
}
function EnableRequestToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "description": value.description,
    "host": value.host
  };
}

// src/zrok/api/models/EnableResponse.ts
function instanceOfEnableResponse(value) {
  let isInstance = true;
  return isInstance;
}
function EnableResponseFromJSON(json) {
  return EnableResponseFromJSONTyped(json, false);
}
function EnableResponseFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "identity": !exists(json, "identity") ? void 0 : json["identity"],
    "cfg": !exists(json, "cfg") ? void 0 : json["cfg"]
  };
}
function EnableResponseToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "identity": value.identity,
    "cfg": value.cfg
  };
}

// src/zrok/api/models/SparkDataSample.ts
function instanceOfSparkDataSample(value) {
  let isInstance = true;
  return isInstance;
}
function SparkDataSampleFromJSON(json) {
  return SparkDataSampleFromJSONTyped(json, false);
}
function SparkDataSampleFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "rx": !exists(json, "rx") ? void 0 : json["rx"],
    "tx": !exists(json, "tx") ? void 0 : json["tx"]
  };
}
function SparkDataSampleToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "rx": value.rx,
    "tx": value.tx
  };
}

// src/zrok/api/models/Environment.ts
function instanceOfEnvironment(value) {
  let isInstance = true;
  return isInstance;
}
function EnvironmentFromJSON(json) {
  return EnvironmentFromJSONTyped(json, false);
}
function EnvironmentFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "description": !exists(json, "description") ? void 0 : json["description"],
    "host": !exists(json, "host") ? void 0 : json["host"],
    "address": !exists(json, "address") ? void 0 : json["address"],
    "zId": !exists(json, "zId") ? void 0 : json["zId"],
    "activity": !exists(json, "activity") ? void 0 : json["activity"].map(SparkDataSampleFromJSON),
    "limited": !exists(json, "limited") ? void 0 : json["limited"],
    "createdAt": !exists(json, "createdAt") ? void 0 : json["createdAt"],
    "updatedAt": !exists(json, "updatedAt") ? void 0 : json["updatedAt"]
  };
}
function EnvironmentToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "description": value.description,
    "host": value.host,
    "address": value.address,
    "zId": value.zId,
    "activity": value.activity === void 0 ? void 0 : value.activity.map(SparkDataSampleToJSON),
    "limited": value.limited,
    "createdAt": value.createdAt,
    "updatedAt": value.updatedAt
  };
}

// src/zrok/api/models/Frontend.ts
function instanceOfFrontend(value) {
  let isInstance = true;
  return isInstance;
}
function FrontendFromJSON(json) {
  return FrontendFromJSONTyped(json, false);
}
function FrontendFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "id": !exists(json, "id") ? void 0 : json["id"],
    "shrToken": !exists(json, "shrToken") ? void 0 : json["shrToken"],
    "zId": !exists(json, "zId") ? void 0 : json["zId"],
    "createdAt": !exists(json, "createdAt") ? void 0 : json["createdAt"],
    "updatedAt": !exists(json, "updatedAt") ? void 0 : json["updatedAt"]
  };
}
function FrontendToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "id": value.id,
    "shrToken": value.shrToken,
    "zId": value.zId,
    "createdAt": value.createdAt,
    "updatedAt": value.updatedAt
  };
}

// src/zrok/api/models/Share.ts
function instanceOfShare(value) {
  let isInstance = true;
  return isInstance;
}
function ShareFromJSON(json) {
  return ShareFromJSONTyped(json, false);
}
function ShareFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "token": !exists(json, "token") ? void 0 : json["token"],
    "zId": !exists(json, "zId") ? void 0 : json["zId"],
    "shareMode": !exists(json, "shareMode") ? void 0 : json["shareMode"],
    "backendMode": !exists(json, "backendMode") ? void 0 : json["backendMode"],
    "frontendSelection": !exists(json, "frontendSelection") ? void 0 : json["frontendSelection"],
    "frontendEndpoint": !exists(json, "frontendEndpoint") ? void 0 : json["frontendEndpoint"],
    "backendProxyEndpoint": !exists(json, "backendProxyEndpoint") ? void 0 : json["backendProxyEndpoint"],
    "reserved": !exists(json, "reserved") ? void 0 : json["reserved"],
    "activity": !exists(json, "activity") ? void 0 : json["activity"].map(SparkDataSampleFromJSON),
    "limited": !exists(json, "limited") ? void 0 : json["limited"],
    "createdAt": !exists(json, "createdAt") ? void 0 : json["createdAt"],
    "updatedAt": !exists(json, "updatedAt") ? void 0 : json["updatedAt"]
  };
}
function ShareToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "token": value.token,
    "zId": value.zId,
    "shareMode": value.shareMode,
    "backendMode": value.backendMode,
    "frontendSelection": value.frontendSelection,
    "frontendEndpoint": value.frontendEndpoint,
    "backendProxyEndpoint": value.backendProxyEndpoint,
    "reserved": value.reserved,
    "activity": value.activity === void 0 ? void 0 : value.activity.map(SparkDataSampleToJSON),
    "limited": value.limited,
    "createdAt": value.createdAt,
    "updatedAt": value.updatedAt
  };
}

// src/zrok/api/models/EnvironmentAndResources.ts
function instanceOfEnvironmentAndResources(value) {
  let isInstance = true;
  return isInstance;
}
function EnvironmentAndResourcesFromJSON(json) {
  return EnvironmentAndResourcesFromJSONTyped(json, false);
}
function EnvironmentAndResourcesFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "environment": !exists(json, "environment") ? void 0 : EnvironmentFromJSON(json["environment"]),
    "frontends": !exists(json, "frontends") ? void 0 : json["frontends"].map(FrontendFromJSON),
    "shares": !exists(json, "shares") ? void 0 : json["shares"].map(ShareFromJSON)
  };
}
function EnvironmentAndResourcesToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "environment": EnvironmentToJSON(value.environment),
    "frontends": value.frontends === void 0 ? void 0 : value.frontends.map(FrontendToJSON),
    "shares": value.shares === void 0 ? void 0 : value.shares.map(ShareToJSON)
  };
}

// src/zrok/api/models/InviteRequest.ts
function instanceOfInviteRequest(value) {
  let isInstance = true;
  return isInstance;
}
function InviteRequestFromJSON(json) {
  return InviteRequestFromJSONTyped(json, false);
}
function InviteRequestFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "email": !exists(json, "email") ? void 0 : json["email"],
    "token": !exists(json, "token") ? void 0 : json["token"]
  };
}
function InviteRequestToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "email": value.email,
    "token": value.token
  };
}

// src/zrok/api/models/InviteTokenGenerateRequest.ts
function instanceOfInviteTokenGenerateRequest(value) {
  let isInstance = true;
  return isInstance;
}
function InviteTokenGenerateRequestFromJSON(json) {
  return InviteTokenGenerateRequestFromJSONTyped(json, false);
}
function InviteTokenGenerateRequestFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "tokens": !exists(json, "tokens") ? void 0 : json["tokens"]
  };
}
function InviteTokenGenerateRequestToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "tokens": value.tokens
  };
}

// src/zrok/api/models/LoginRequest.ts
function instanceOfLoginRequest(value) {
  let isInstance = true;
  return isInstance;
}
function LoginRequestFromJSON(json) {
  return LoginRequestFromJSONTyped(json, false);
}
function LoginRequestFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "email": !exists(json, "email") ? void 0 : json["email"],
    "password": !exists(json, "password") ? void 0 : json["password"]
  };
}
function LoginRequestToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "email": value.email,
    "password": value.password
  };
}

// src/zrok/api/models/MetricsSample.ts
function instanceOfMetricsSample(value) {
  let isInstance = true;
  return isInstance;
}
function MetricsSampleFromJSON(json) {
  return MetricsSampleFromJSONTyped(json, false);
}
function MetricsSampleFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "rx": !exists(json, "rx") ? void 0 : json["rx"],
    "tx": !exists(json, "tx") ? void 0 : json["tx"],
    "timestamp": !exists(json, "timestamp") ? void 0 : json["timestamp"]
  };
}
function MetricsSampleToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "rx": value.rx,
    "tx": value.tx,
    "timestamp": value.timestamp
  };
}

// src/zrok/api/models/Metrics.ts
function instanceOfMetrics(value) {
  let isInstance = true;
  return isInstance;
}
function MetricsFromJSON(json) {
  return MetricsFromJSONTyped(json, false);
}
function MetricsFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "scope": !exists(json, "scope") ? void 0 : json["scope"],
    "id": !exists(json, "id") ? void 0 : json["id"],
    "period": !exists(json, "period") ? void 0 : json["period"],
    "samples": !exists(json, "samples") ? void 0 : json["samples"].map(MetricsSampleFromJSON)
  };
}
function MetricsToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "scope": value.scope,
    "id": value.id,
    "period": value.period,
    "samples": value.samples === void 0 ? void 0 : value.samples.map(MetricsSampleToJSON)
  };
}

// src/zrok/api/models/PasswordRequirements.ts
function instanceOfPasswordRequirements(value) {
  let isInstance = true;
  return isInstance;
}
function PasswordRequirementsFromJSON(json) {
  return PasswordRequirementsFromJSONTyped(json, false);
}
function PasswordRequirementsFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "length": !exists(json, "length") ? void 0 : json["length"],
    "requireCapital": !exists(json, "requireCapital") ? void 0 : json["requireCapital"],
    "requireNumeric": !exists(json, "requireNumeric") ? void 0 : json["requireNumeric"],
    "requireSpecial": !exists(json, "requireSpecial") ? void 0 : json["requireSpecial"],
    "validSpecialCharacters": !exists(json, "validSpecialCharacters") ? void 0 : json["validSpecialCharacters"]
  };
}
function PasswordRequirementsToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "length": value.length,
    "requireCapital": value.requireCapital,
    "requireNumeric": value.requireNumeric,
    "requireSpecial": value.requireSpecial,
    "validSpecialCharacters": value.validSpecialCharacters
  };
}

// src/zrok/api/models/ModelConfiguration.ts
function instanceOfModelConfiguration(value) {
  let isInstance = true;
  return isInstance;
}
function ModelConfigurationFromJSON(json) {
  return ModelConfigurationFromJSONTyped(json, false);
}
function ModelConfigurationFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "version": !exists(json, "version") ? void 0 : json["version"],
    "touLink": !exists(json, "touLink") ? void 0 : json["touLink"],
    "invitesOpen": !exists(json, "invitesOpen") ? void 0 : json["invitesOpen"],
    "requiresInviteToken": !exists(json, "requiresInviteToken") ? void 0 : json["requiresInviteToken"],
    "inviteTokenContact": !exists(json, "inviteTokenContact") ? void 0 : json["inviteTokenContact"],
    "passwordRequirements": !exists(json, "passwordRequirements") ? void 0 : PasswordRequirementsFromJSON(json["passwordRequirements"])
  };
}
function ModelConfigurationToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "version": value.version,
    "touLink": value.touLink,
    "invitesOpen": value.invitesOpen,
    "requiresInviteToken": value.requiresInviteToken,
    "inviteTokenContact": value.inviteTokenContact,
    "passwordRequirements": PasswordRequirementsToJSON(value.passwordRequirements)
  };
}

// src/zrok/api/models/Overview.ts
function instanceOfOverview(value) {
  let isInstance = true;
  return isInstance;
}
function OverviewFromJSON(json) {
  return OverviewFromJSONTyped(json, false);
}
function OverviewFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "accountLimited": !exists(json, "accountLimited") ? void 0 : json["accountLimited"],
    "environments": !exists(json, "environments") ? void 0 : json["environments"].map(EnvironmentAndResourcesFromJSON)
  };
}
function OverviewToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "accountLimited": value.accountLimited,
    "environments": value.environments === void 0 ? void 0 : value.environments.map(EnvironmentAndResourcesToJSON)
  };
}

// src/zrok/api/models/Principal.ts
function instanceOfPrincipal(value) {
  let isInstance = true;
  return isInstance;
}
function PrincipalFromJSON(json) {
  return PrincipalFromJSONTyped(json, false);
}
function PrincipalFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "id": !exists(json, "id") ? void 0 : json["id"],
    "email": !exists(json, "email") ? void 0 : json["email"],
    "token": !exists(json, "token") ? void 0 : json["token"],
    "limitless": !exists(json, "limitless") ? void 0 : json["limitless"],
    "admin": !exists(json, "admin") ? void 0 : json["admin"]
  };
}
function PrincipalToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "id": value.id,
    "email": value.email,
    "token": value.token,
    "limitless": value.limitless,
    "admin": value.admin
  };
}

// src/zrok/api/models/PublicFrontend.ts
function instanceOfPublicFrontend(value) {
  let isInstance = true;
  return isInstance;
}
function PublicFrontendFromJSON(json) {
  return PublicFrontendFromJSONTyped(json, false);
}
function PublicFrontendFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "token": !exists(json, "token") ? void 0 : json["token"],
    "zId": !exists(json, "zId") ? void 0 : json["zId"],
    "urlTemplate": !exists(json, "urlTemplate") ? void 0 : json["urlTemplate"],
    "publicName": !exists(json, "publicName") ? void 0 : json["publicName"],
    "createdAt": !exists(json, "createdAt") ? void 0 : json["createdAt"],
    "updatedAt": !exists(json, "updatedAt") ? void 0 : json["updatedAt"]
  };
}
function PublicFrontendToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "token": value.token,
    "zId": value.zId,
    "urlTemplate": value.urlTemplate,
    "publicName": value.publicName,
    "createdAt": value.createdAt,
    "updatedAt": value.updatedAt
  };
}

// src/zrok/api/models/RegisterRequest.ts
function instanceOfRegisterRequest(value) {
  let isInstance = true;
  return isInstance;
}
function RegisterRequestFromJSON(json) {
  return RegisterRequestFromJSONTyped(json, false);
}
function RegisterRequestFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "token": !exists(json, "token") ? void 0 : json["token"],
    "password": !exists(json, "password") ? void 0 : json["password"]
  };
}
function RegisterRequestToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "token": value.token,
    "password": value.password
  };
}

// src/zrok/api/models/RegisterResponse.ts
function instanceOfRegisterResponse(value) {
  let isInstance = true;
  return isInstance;
}
function RegisterResponseFromJSON(json) {
  return RegisterResponseFromJSONTyped(json, false);
}
function RegisterResponseFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "token": !exists(json, "token") ? void 0 : json["token"]
  };
}
function RegisterResponseToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "token": value.token
  };
}

// src/zrok/api/models/ResetPasswordRequest.ts
function instanceOfResetPasswordRequest(value) {
  let isInstance = true;
  return isInstance;
}
function ResetPasswordRequestFromJSON(json) {
  return ResetPasswordRequestFromJSONTyped(json, false);
}
function ResetPasswordRequestFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "token": !exists(json, "token") ? void 0 : json["token"],
    "password": !exists(json, "password") ? void 0 : json["password"]
  };
}
function ResetPasswordRequestToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "token": value.token,
    "password": value.password
  };
}

// src/zrok/api/models/ResetPasswordRequestRequest.ts
function instanceOfResetPasswordRequestRequest(value) {
  let isInstance = true;
  return isInstance;
}
function ResetPasswordRequestRequestFromJSON(json) {
  return ResetPasswordRequestRequestFromJSONTyped(json, false);
}
function ResetPasswordRequestRequestFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "emailAddress": !exists(json, "emailAddress") ? void 0 : json["emailAddress"]
  };
}
function ResetPasswordRequestRequestToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "emailAddress": value.emailAddress
  };
}

// src/zrok/api/models/ShareRequest.ts
var ShareRequestShareModeEnum = {
  Public: "public",
  Private: "private"
};
var ShareRequestBackendModeEnum = {
  Proxy: "proxy",
  Web: "web",
  TcpTunnel: "tcpTunnel",
  UdpTunnel: "udpTunnel",
  Caddy: "caddy"
};
var ShareRequestOauthProviderEnum = {
  Github: "github",
  Google: "google"
};
function instanceOfShareRequest(value) {
  let isInstance = true;
  return isInstance;
}
function ShareRequestFromJSON(json) {
  return ShareRequestFromJSONTyped(json, false);
}
function ShareRequestFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "envZId": !exists(json, "envZId") ? void 0 : json["envZId"],
    "shareMode": !exists(json, "shareMode") ? void 0 : json["shareMode"],
    "frontendSelection": !exists(json, "frontendSelection") ? void 0 : json["frontendSelection"],
    "backendMode": !exists(json, "backendMode") ? void 0 : json["backendMode"],
    "backendProxyEndpoint": !exists(json, "backendProxyEndpoint") ? void 0 : json["backendProxyEndpoint"],
    "authScheme": !exists(json, "authScheme") ? void 0 : json["authScheme"],
    "authUsers": !exists(json, "authUsers") ? void 0 : json["authUsers"].map(AuthUserFromJSON),
    "oauthProvider": !exists(json, "oauthProvider") ? void 0 : json["oauthProvider"],
    "oauthEmailDomains": !exists(json, "oauthEmailDomains") ? void 0 : json["oauthEmailDomains"],
    "oauthAuthorizationCheckInterval": !exists(json, "oauthAuthorizationCheckInterval") ? void 0 : json["oauthAuthorizationCheckInterval"],
    "reserved": !exists(json, "reserved") ? void 0 : json["reserved"]
  };
}
function ShareRequestToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "envZId": value.envZId,
    "shareMode": value.shareMode,
    "frontendSelection": value.frontendSelection,
    "backendMode": value.backendMode,
    "backendProxyEndpoint": value.backendProxyEndpoint,
    "authScheme": value.authScheme,
    "authUsers": value.authUsers === void 0 ? void 0 : value.authUsers.map(AuthUserToJSON),
    "oauthProvider": value.oauthProvider,
    "oauthEmailDomains": value.oauthEmailDomains,
    "oauthAuthorizationCheckInterval": value.oauthAuthorizationCheckInterval,
    "reserved": value.reserved
  };
}

// src/zrok/api/models/ShareResponse.ts
function instanceOfShareResponse(value) {
  let isInstance = true;
  return isInstance;
}
function ShareResponseFromJSON(json) {
  return ShareResponseFromJSONTyped(json, false);
}
function ShareResponseFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "frontendProxyEndpoints": !exists(json, "frontendProxyEndpoints") ? void 0 : json["frontendProxyEndpoints"],
    "shrToken": !exists(json, "shrToken") ? void 0 : json["shrToken"]
  };
}
function ShareResponseToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "frontendProxyEndpoints": value.frontendProxyEndpoints,
    "shrToken": value.shrToken
  };
}

// src/zrok/api/models/UnaccessRequest.ts
function instanceOfUnaccessRequest(value) {
  let isInstance = true;
  return isInstance;
}
function UnaccessRequestFromJSON(json) {
  return UnaccessRequestFromJSONTyped(json, false);
}
function UnaccessRequestFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "frontendToken": !exists(json, "frontendToken") ? void 0 : json["frontendToken"],
    "envZId": !exists(json, "envZId") ? void 0 : json["envZId"],
    "shrToken": !exists(json, "shrToken") ? void 0 : json["shrToken"]
  };
}
function UnaccessRequestToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "frontendToken": value.frontendToken,
    "envZId": value.envZId,
    "shrToken": value.shrToken
  };
}

// src/zrok/api/models/UnshareRequest.ts
function instanceOfUnshareRequest(value) {
  let isInstance = true;
  return isInstance;
}
function UnshareRequestFromJSON(json) {
  return UnshareRequestFromJSONTyped(json, false);
}
function UnshareRequestFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "envZId": !exists(json, "envZId") ? void 0 : json["envZId"],
    "shrToken": !exists(json, "shrToken") ? void 0 : json["shrToken"],
    "reserved": !exists(json, "reserved") ? void 0 : json["reserved"]
  };
}
function UnshareRequestToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "envZId": value.envZId,
    "shrToken": value.shrToken,
    "reserved": value.reserved
  };
}

// src/zrok/api/models/UpdateFrontendRequest.ts
function instanceOfUpdateFrontendRequest(value) {
  let isInstance = true;
  return isInstance;
}
function UpdateFrontendRequestFromJSON(json) {
  return UpdateFrontendRequestFromJSONTyped(json, false);
}
function UpdateFrontendRequestFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "frontendToken": !exists(json, "frontendToken") ? void 0 : json["frontendToken"],
    "publicName": !exists(json, "publicName") ? void 0 : json["publicName"],
    "urlTemplate": !exists(json, "urlTemplate") ? void 0 : json["urlTemplate"]
  };
}
function UpdateFrontendRequestToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "frontendToken": value.frontendToken,
    "publicName": value.publicName,
    "urlTemplate": value.urlTemplate
  };
}

// src/zrok/api/models/UpdateShareRequest.ts
function instanceOfUpdateShareRequest(value) {
  let isInstance = true;
  return isInstance;
}
function UpdateShareRequestFromJSON(json) {
  return UpdateShareRequestFromJSONTyped(json, false);
}
function UpdateShareRequestFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "shrToken": !exists(json, "shrToken") ? void 0 : json["shrToken"],
    "backendProxyEndpoint": !exists(json, "backendProxyEndpoint") ? void 0 : json["backendProxyEndpoint"]
  };
}
function UpdateShareRequestToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "shrToken": value.shrToken,
    "backendProxyEndpoint": value.backendProxyEndpoint
  };
}

// src/zrok/api/models/VerifyRequest.ts
function instanceOfVerifyRequest(value) {
  let isInstance = true;
  return isInstance;
}
function VerifyRequestFromJSON(json) {
  return VerifyRequestFromJSONTyped(json, false);
}
function VerifyRequestFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "token": !exists(json, "token") ? void 0 : json["token"]
  };
}
function VerifyRequestToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "token": value.token
  };
}

// src/zrok/api/models/VerifyResponse.ts
function instanceOfVerifyResponse(value) {
  let isInstance = true;
  return isInstance;
}
function VerifyResponseFromJSON(json) {
  return VerifyResponseFromJSONTyped(json, false);
}
function VerifyResponseFromJSONTyped(json, ignoreDiscriminator) {
  if (json === void 0 || json === null) {
    return json;
  }
  return {
    "email": !exists(json, "email") ? void 0 : json["email"]
  };
}
function VerifyResponseToJSON(value) {
  if (value === void 0) {
    return void 0;
  }
  if (value === null) {
    return null;
  }
  return {
    "email": value.email
  };
}
export {
  AccessRequestFromJSON,
  AccessRequestFromJSONTyped,
  AccessRequestToJSON,
  AccessResponseFromJSON,
  AccessResponseFromJSONTyped,
  AccessResponseToJSON,
  AuthUserFromJSON,
  AuthUserFromJSONTyped,
  AuthUserToJSON,
  CreateFrontendRequestFromJSON,
  CreateFrontendRequestFromJSONTyped,
  CreateFrontendRequestToJSON,
  CreateFrontendResponseFromJSON,
  CreateFrontendResponseFromJSONTyped,
  CreateFrontendResponseToJSON,
  CreateIdentity201ResponseFromJSON,
  CreateIdentity201ResponseFromJSONTyped,
  CreateIdentity201ResponseToJSON,
  CreateIdentityRequestFromJSON,
  CreateIdentityRequestFromJSONTyped,
  CreateIdentityRequestToJSON,
  DeleteFrontendRequestFromJSON,
  DeleteFrontendRequestFromJSONTyped,
  DeleteFrontendRequestToJSON,
  DisableRequestFromJSON,
  DisableRequestFromJSONTyped,
  DisableRequestToJSON,
  EnableRequestFromJSON,
  EnableRequestFromJSONTyped,
  EnableRequestToJSON,
  EnableResponseFromJSON,
  EnableResponseFromJSONTyped,
  EnableResponseToJSON,
  EnvironmentAndResourcesFromJSON,
  EnvironmentAndResourcesFromJSONTyped,
  EnvironmentAndResourcesToJSON,
  EnvironmentFromJSON,
  EnvironmentFromJSONTyped,
  EnvironmentToJSON,
  FrontendFromJSON,
  FrontendFromJSONTyped,
  FrontendToJSON,
  InviteRequestFromJSON,
  InviteRequestFromJSONTyped,
  InviteRequestToJSON,
  InviteTokenGenerateRequestFromJSON,
  InviteTokenGenerateRequestFromJSONTyped,
  InviteTokenGenerateRequestToJSON,
  LoginRequestFromJSON,
  LoginRequestFromJSONTyped,
  LoginRequestToJSON,
  MetricsFromJSON,
  MetricsFromJSONTyped,
  MetricsSampleFromJSON,
  MetricsSampleFromJSONTyped,
  MetricsSampleToJSON,
  MetricsToJSON,
  ModelConfigurationFromJSON,
  ModelConfigurationFromJSONTyped,
  ModelConfigurationToJSON,
  OverviewFromJSON,
  OverviewFromJSONTyped,
  OverviewToJSON,
  PasswordRequirementsFromJSON,
  PasswordRequirementsFromJSONTyped,
  PasswordRequirementsToJSON,
  PrincipalFromJSON,
  PrincipalFromJSONTyped,
  PrincipalToJSON,
  PublicFrontendFromJSON,
  PublicFrontendFromJSONTyped,
  PublicFrontendToJSON,
  RegisterRequestFromJSON,
  RegisterRequestFromJSONTyped,
  RegisterRequestToJSON,
  RegisterResponseFromJSON,
  RegisterResponseFromJSONTyped,
  RegisterResponseToJSON,
  ResetPasswordRequestFromJSON,
  ResetPasswordRequestFromJSONTyped,
  ResetPasswordRequestRequestFromJSON,
  ResetPasswordRequestRequestFromJSONTyped,
  ResetPasswordRequestRequestToJSON,
  ResetPasswordRequestToJSON,
  ShareFromJSON,
  ShareFromJSONTyped,
  ShareRequestBackendModeEnum,
  ShareRequestFromJSON,
  ShareRequestFromJSONTyped,
  ShareRequestOauthProviderEnum,
  ShareRequestShareModeEnum,
  ShareRequestToJSON,
  ShareResponseFromJSON,
  ShareResponseFromJSONTyped,
  ShareResponseToJSON,
  ShareToJSON,
  SparkDataSampleFromJSON,
  SparkDataSampleFromJSONTyped,
  SparkDataSampleToJSON,
  UnaccessRequestFromJSON,
  UnaccessRequestFromJSONTyped,
  UnaccessRequestToJSON,
  UnshareRequestFromJSON,
  UnshareRequestFromJSONTyped,
  UnshareRequestToJSON,
  UpdateFrontendRequestFromJSON,
  UpdateFrontendRequestFromJSONTyped,
  UpdateFrontendRequestToJSON,
  UpdateShareRequestFromJSON,
  UpdateShareRequestFromJSONTyped,
  UpdateShareRequestToJSON,
  VerifyRequestFromJSON,
  VerifyRequestFromJSONTyped,
  VerifyRequestToJSON,
  VerifyResponseFromJSON,
  VerifyResponseFromJSONTyped,
  VerifyResponseToJSON,
  instanceOfAccessRequest,
  instanceOfAccessResponse,
  instanceOfAuthUser,
  instanceOfCreateFrontendRequest,
  instanceOfCreateFrontendResponse,
  instanceOfCreateIdentity201Response,
  instanceOfCreateIdentityRequest,
  instanceOfDeleteFrontendRequest,
  instanceOfDisableRequest,
  instanceOfEnableRequest,
  instanceOfEnableResponse,
  instanceOfEnvironment,
  instanceOfEnvironmentAndResources,
  instanceOfFrontend,
  instanceOfInviteRequest,
  instanceOfInviteTokenGenerateRequest,
  instanceOfLoginRequest,
  instanceOfMetrics,
  instanceOfMetricsSample,
  instanceOfModelConfiguration,
  instanceOfOverview,
  instanceOfPasswordRequirements,
  instanceOfPrincipal,
  instanceOfPublicFrontend,
  instanceOfRegisterRequest,
  instanceOfRegisterResponse,
  instanceOfResetPasswordRequest,
  instanceOfResetPasswordRequestRequest,
  instanceOfShare,
  instanceOfShareRequest,
  instanceOfShareResponse,
  instanceOfSparkDataSample,
  instanceOfUnaccessRequest,
  instanceOfUnshareRequest,
  instanceOfUpdateFrontendRequest,
  instanceOfUpdateShareRequest,
  instanceOfVerifyRequest,
  instanceOfVerifyResponse
};
//# sourceMappingURL=index.mjs.map