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

// src/zrok/api/apis/MetadataApi.ts
var MetadataApi_exports = {};
__export(MetadataApi_exports, {
  MetadataApi: () => MetadataApi
});
module.exports = __toCommonJS(MetadataApi_exports);

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
var RequiredError = class extends Error {
  constructor(field, msg) {
    super(msg);
    this.field = field;
    this.name = "RequiredError";
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
var TextApiResponse = class {
  constructor(raw) {
    this.raw = raw;
  }
  value() {
    return __async(this, null, function* () {
      return yield this.raw.text();
    });
  }
};

// src/zrok/api/models/SparkDataSample.ts
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

// src/zrok/api/models/Environment.ts
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

// src/zrok/api/models/Frontend.ts
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

// src/zrok/api/models/Share.ts
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

// src/zrok/api/models/EnvironmentAndResources.ts
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

// src/zrok/api/models/MetricsSample.ts
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

// src/zrok/api/models/Metrics.ts
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

// src/zrok/api/models/PasswordRequirements.ts
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

// src/zrok/api/models/ModelConfiguration.ts
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

// src/zrok/api/models/Overview.ts
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

// src/zrok/api/apis/MetadataApi.ts
var MetadataApi = class extends BaseAPI {
  /**
   */
  _configurationRaw(initOverrides) {
    return __async(this, null, function* () {
      const queryParameters = {};
      const headerParameters = {};
      const response = yield this.request({
        path: `/configuration`,
        method: "GET",
        headers: headerParameters,
        query: queryParameters
      }, initOverrides);
      return new JSONApiResponse(response, (jsonValue) => ModelConfigurationFromJSON(jsonValue));
    });
  }
  /**
   */
  _configuration(initOverrides) {
    return __async(this, null, function* () {
      const response = yield this._configurationRaw(initOverrides);
      return yield response.value();
    });
  }
  /**
   */
  getAccountDetailRaw(initOverrides) {
    return __async(this, null, function* () {
      const queryParameters = {};
      const headerParameters = {};
      if (this.configuration && this.configuration.apiKey) {
        headerParameters["x-token"] = this.configuration.apiKey("x-token");
      }
      const response = yield this.request({
        path: `/detail/account`,
        method: "GET",
        headers: headerParameters,
        query: queryParameters
      }, initOverrides);
      return new JSONApiResponse(response, (jsonValue) => jsonValue.map(EnvironmentFromJSON));
    });
  }
  /**
   */
  getAccountDetail(initOverrides) {
    return __async(this, null, function* () {
      const response = yield this.getAccountDetailRaw(initOverrides);
      return yield response.value();
    });
  }
  /**
   */
  getAccountMetricsRaw(requestParameters, initOverrides) {
    return __async(this, null, function* () {
      const queryParameters = {};
      if (requestParameters.duration !== void 0) {
        queryParameters["duration"] = requestParameters.duration;
      }
      const headerParameters = {};
      if (this.configuration && this.configuration.apiKey) {
        headerParameters["x-token"] = this.configuration.apiKey("x-token");
      }
      const response = yield this.request({
        path: `/metrics/account`,
        method: "GET",
        headers: headerParameters,
        query: queryParameters
      }, initOverrides);
      return new JSONApiResponse(response, (jsonValue) => MetricsFromJSON(jsonValue));
    });
  }
  /**
   */
  getAccountMetrics() {
    return __async(this, arguments, function* (requestParameters = {}, initOverrides) {
      const response = yield this.getAccountMetricsRaw(requestParameters, initOverrides);
      return yield response.value();
    });
  }
  /**
   */
  getEnvironmentDetailRaw(requestParameters, initOverrides) {
    return __async(this, null, function* () {
      if (requestParameters.envZId === null || requestParameters.envZId === void 0) {
        throw new RequiredError("envZId", "Required parameter requestParameters.envZId was null or undefined when calling getEnvironmentDetail.");
      }
      const queryParameters = {};
      const headerParameters = {};
      if (this.configuration && this.configuration.apiKey) {
        headerParameters["x-token"] = this.configuration.apiKey("x-token");
      }
      const response = yield this.request({
        path: `/detail/environment/{envZId}`.replace(`{${"envZId"}}`, encodeURIComponent(String(requestParameters.envZId))),
        method: "GET",
        headers: headerParameters,
        query: queryParameters
      }, initOverrides);
      return new JSONApiResponse(response, (jsonValue) => EnvironmentAndResourcesFromJSON(jsonValue));
    });
  }
  /**
   */
  getEnvironmentDetail(requestParameters, initOverrides) {
    return __async(this, null, function* () {
      const response = yield this.getEnvironmentDetailRaw(requestParameters, initOverrides);
      return yield response.value();
    });
  }
  /**
   */
  getEnvironmentMetricsRaw(requestParameters, initOverrides) {
    return __async(this, null, function* () {
      if (requestParameters.envId === null || requestParameters.envId === void 0) {
        throw new RequiredError("envId", "Required parameter requestParameters.envId was null or undefined when calling getEnvironmentMetrics.");
      }
      const queryParameters = {};
      if (requestParameters.duration !== void 0) {
        queryParameters["duration"] = requestParameters.duration;
      }
      const headerParameters = {};
      if (this.configuration && this.configuration.apiKey) {
        headerParameters["x-token"] = this.configuration.apiKey("x-token");
      }
      const response = yield this.request({
        path: `/metrics/environment/{envId}`.replace(`{${"envId"}}`, encodeURIComponent(String(requestParameters.envId))),
        method: "GET",
        headers: headerParameters,
        query: queryParameters
      }, initOverrides);
      return new JSONApiResponse(response, (jsonValue) => MetricsFromJSON(jsonValue));
    });
  }
  /**
   */
  getEnvironmentMetrics(requestParameters, initOverrides) {
    return __async(this, null, function* () {
      const response = yield this.getEnvironmentMetricsRaw(requestParameters, initOverrides);
      return yield response.value();
    });
  }
  /**
   */
  getFrontendDetailRaw(requestParameters, initOverrides) {
    return __async(this, null, function* () {
      if (requestParameters.feId === null || requestParameters.feId === void 0) {
        throw new RequiredError("feId", "Required parameter requestParameters.feId was null or undefined when calling getFrontendDetail.");
      }
      const queryParameters = {};
      const headerParameters = {};
      if (this.configuration && this.configuration.apiKey) {
        headerParameters["x-token"] = this.configuration.apiKey("x-token");
      }
      const response = yield this.request({
        path: `/detail/frontend/{feId}`.replace(`{${"feId"}}`, encodeURIComponent(String(requestParameters.feId))),
        method: "GET",
        headers: headerParameters,
        query: queryParameters
      }, initOverrides);
      return new JSONApiResponse(response, (jsonValue) => FrontendFromJSON(jsonValue));
    });
  }
  /**
   */
  getFrontendDetail(requestParameters, initOverrides) {
    return __async(this, null, function* () {
      const response = yield this.getFrontendDetailRaw(requestParameters, initOverrides);
      return yield response.value();
    });
  }
  /**
   */
  getShareDetailRaw(requestParameters, initOverrides) {
    return __async(this, null, function* () {
      if (requestParameters.shrToken === null || requestParameters.shrToken === void 0) {
        throw new RequiredError("shrToken", "Required parameter requestParameters.shrToken was null or undefined when calling getShareDetail.");
      }
      const queryParameters = {};
      const headerParameters = {};
      if (this.configuration && this.configuration.apiKey) {
        headerParameters["x-token"] = this.configuration.apiKey("x-token");
      }
      const response = yield this.request({
        path: `/detail/share/{shrToken}`.replace(`{${"shrToken"}}`, encodeURIComponent(String(requestParameters.shrToken))),
        method: "GET",
        headers: headerParameters,
        query: queryParameters
      }, initOverrides);
      return new JSONApiResponse(response, (jsonValue) => ShareFromJSON(jsonValue));
    });
  }
  /**
   */
  getShareDetail(requestParameters, initOverrides) {
    return __async(this, null, function* () {
      const response = yield this.getShareDetailRaw(requestParameters, initOverrides);
      return yield response.value();
    });
  }
  /**
   */
  getShareMetricsRaw(requestParameters, initOverrides) {
    return __async(this, null, function* () {
      if (requestParameters.shrToken === null || requestParameters.shrToken === void 0) {
        throw new RequiredError("shrToken", "Required parameter requestParameters.shrToken was null or undefined when calling getShareMetrics.");
      }
      const queryParameters = {};
      if (requestParameters.duration !== void 0) {
        queryParameters["duration"] = requestParameters.duration;
      }
      const headerParameters = {};
      if (this.configuration && this.configuration.apiKey) {
        headerParameters["x-token"] = this.configuration.apiKey("x-token");
      }
      const response = yield this.request({
        path: `/metrics/share/{shrToken}`.replace(`{${"shrToken"}}`, encodeURIComponent(String(requestParameters.shrToken))),
        method: "GET",
        headers: headerParameters,
        query: queryParameters
      }, initOverrides);
      return new JSONApiResponse(response, (jsonValue) => MetricsFromJSON(jsonValue));
    });
  }
  /**
   */
  getShareMetrics(requestParameters, initOverrides) {
    return __async(this, null, function* () {
      const response = yield this.getShareMetricsRaw(requestParameters, initOverrides);
      return yield response.value();
    });
  }
  /**
   */
  overviewRaw(initOverrides) {
    return __async(this, null, function* () {
      const queryParameters = {};
      const headerParameters = {};
      if (this.configuration && this.configuration.apiKey) {
        headerParameters["x-token"] = this.configuration.apiKey("x-token");
      }
      const response = yield this.request({
        path: `/overview`,
        method: "GET",
        headers: headerParameters,
        query: queryParameters
      }, initOverrides);
      return new JSONApiResponse(response, (jsonValue) => OverviewFromJSON(jsonValue));
    });
  }
  /**
   */
  overview(initOverrides) {
    return __async(this, null, function* () {
      const response = yield this.overviewRaw(initOverrides);
      return yield response.value();
    });
  }
  /**
   */
  versionRaw(initOverrides) {
    return __async(this, null, function* () {
      const queryParameters = {};
      const headerParameters = {};
      const response = yield this.request({
        path: `/version`,
        method: "GET",
        headers: headerParameters,
        query: queryParameters
      }, initOverrides);
      if (this.isJsonMime(response.headers.get("content-type"))) {
        return new JSONApiResponse(response);
      } else {
        return new TextApiResponse(response);
      }
    });
  }
  /**
   */
  version(initOverrides) {
    return __async(this, null, function* () {
      const response = yield this.versionRaw(initOverrides);
      return yield response.value();
    });
  }
};
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  MetadataApi
});
//# sourceMappingURL=MetadataApi.js.map