"use strict";
var __defProp = Object.defineProperty;
var __defProps = Object.defineProperties;
var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
var __getOwnPropDescs = Object.getOwnPropertyDescriptors;
var __getOwnPropNames = Object.getOwnPropertyNames;
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

// src/zrok/api/apis/AdminApi.ts
var AdminApi_exports = {};
__export(AdminApi_exports, {
  AdminApi: () => AdminApi
});
module.exports = __toCommonJS(AdminApi_exports);

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
var JSONApiResponse = class {
  constructor(raw, transformer = (jsonValue) => jsonValue) {
    this.raw = raw;
    this.transformer = transformer;
  }
  value() {
    return __async(this, null, function* () {
      return this.transformer(yield this.raw.json());
    });
  }
};
var VoidApiResponse = class {
  constructor(raw) {
    this.raw = raw;
  }
  value() {
    return __async(this, null, function* () {
      return void 0;
    });
  }
};

// src/zrok/api/models/CreateFrontendRequest.ts
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

// src/zrok/api/models/CreateIdentity201Response.ts
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

// src/zrok/api/models/CreateIdentityRequest.ts
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

// src/zrok/api/models/InviteTokenGenerateRequest.ts
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

// src/zrok/api/models/PublicFrontend.ts
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

// src/zrok/api/models/UpdateFrontendRequest.ts
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

// src/zrok/api/apis/AdminApi.ts
var AdminApi = class extends BaseAPI {
  /**
   */
  createFrontendRaw(requestParameters, initOverrides) {
    return __async(this, null, function* () {
      const queryParameters = {};
      const headerParameters = {};
      headerParameters["Content-Type"] = "application/zrok.v1+json";
      if (this.configuration && this.configuration.apiKey) {
        headerParameters["x-token"] = this.configuration.apiKey("x-token");
      }
      const response = yield this.request({
        path: `/frontend`,
        method: "POST",
        headers: headerParameters,
        query: queryParameters,
        body: CreateFrontendRequestToJSON(requestParameters.body)
      }, initOverrides);
      return new JSONApiResponse(response, (jsonValue) => CreateFrontendResponseFromJSON(jsonValue));
    });
  }
  /**
   */
  createFrontend() {
    return __async(this, arguments, function* (requestParameters = {}, initOverrides) {
      const response = yield this.createFrontendRaw(requestParameters, initOverrides);
      return yield response.value();
    });
  }
  /**
   */
  createIdentityRaw(requestParameters, initOverrides) {
    return __async(this, null, function* () {
      const queryParameters = {};
      const headerParameters = {};
      headerParameters["Content-Type"] = "application/zrok.v1+json";
      if (this.configuration && this.configuration.apiKey) {
        headerParameters["x-token"] = this.configuration.apiKey("x-token");
      }
      const response = yield this.request({
        path: `/identity`,
        method: "POST",
        headers: headerParameters,
        query: queryParameters,
        body: CreateIdentityRequestToJSON(requestParameters.body)
      }, initOverrides);
      return new JSONApiResponse(response, (jsonValue) => CreateIdentity201ResponseFromJSON(jsonValue));
    });
  }
  /**
   */
  createIdentity() {
    return __async(this, arguments, function* (requestParameters = {}, initOverrides) {
      const response = yield this.createIdentityRaw(requestParameters, initOverrides);
      return yield response.value();
    });
  }
  /**
   */
  deleteFrontendRaw(requestParameters, initOverrides) {
    return __async(this, null, function* () {
      const queryParameters = {};
      const headerParameters = {};
      headerParameters["Content-Type"] = "application/zrok.v1+json";
      if (this.configuration && this.configuration.apiKey) {
        headerParameters["x-token"] = this.configuration.apiKey("x-token");
      }
      const response = yield this.request({
        path: `/frontend`,
        method: "DELETE",
        headers: headerParameters,
        query: queryParameters,
        body: DeleteFrontendRequestToJSON(requestParameters.body)
      }, initOverrides);
      return new VoidApiResponse(response);
    });
  }
  /**
   */
  deleteFrontend() {
    return __async(this, arguments, function* (requestParameters = {}, initOverrides) {
      yield this.deleteFrontendRaw(requestParameters, initOverrides);
    });
  }
  /**
   */
  inviteTokenGenerateRaw(requestParameters, initOverrides) {
    return __async(this, null, function* () {
      const queryParameters = {};
      const headerParameters = {};
      headerParameters["Content-Type"] = "application/zrok.v1+json";
      if (this.configuration && this.configuration.apiKey) {
        headerParameters["x-token"] = this.configuration.apiKey("x-token");
      }
      const response = yield this.request({
        path: `/invite/token/generate`,
        method: "POST",
        headers: headerParameters,
        query: queryParameters,
        body: InviteTokenGenerateRequestToJSON(requestParameters.body)
      }, initOverrides);
      return new VoidApiResponse(response);
    });
  }
  /**
   */
  inviteTokenGenerate() {
    return __async(this, arguments, function* (requestParameters = {}, initOverrides) {
      yield this.inviteTokenGenerateRaw(requestParameters, initOverrides);
    });
  }
  /**
   */
  listFrontendsRaw(initOverrides) {
    return __async(this, null, function* () {
      const queryParameters = {};
      const headerParameters = {};
      if (this.configuration && this.configuration.apiKey) {
        headerParameters["x-token"] = this.configuration.apiKey("x-token");
      }
      const response = yield this.request({
        path: `/frontends`,
        method: "GET",
        headers: headerParameters,
        query: queryParameters
      }, initOverrides);
      return new JSONApiResponse(response, (jsonValue) => jsonValue.map(PublicFrontendFromJSON));
    });
  }
  /**
   */
  listFrontends(initOverrides) {
    return __async(this, null, function* () {
      const response = yield this.listFrontendsRaw(initOverrides);
      return yield response.value();
    });
  }
  /**
   */
  updateFrontendRaw(requestParameters, initOverrides) {
    return __async(this, null, function* () {
      const queryParameters = {};
      const headerParameters = {};
      headerParameters["Content-Type"] = "application/zrok.v1+json";
      if (this.configuration && this.configuration.apiKey) {
        headerParameters["x-token"] = this.configuration.apiKey("x-token");
      }
      const response = yield this.request({
        path: `/frontend`,
        method: "PATCH",
        headers: headerParameters,
        query: queryParameters,
        body: UpdateFrontendRequestToJSON(requestParameters.body)
      }, initOverrides);
      return new VoidApiResponse(response);
    });
  }
  /**
   */
  updateFrontend() {
    return __async(this, arguments, function* (requestParameters = {}, initOverrides) {
      yield this.updateFrontendRaw(requestParameters, initOverrides);
    });
  }
};
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  AdminApi
});
//# sourceMappingURL=AdminApi.js.map