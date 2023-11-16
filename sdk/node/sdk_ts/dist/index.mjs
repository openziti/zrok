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

// src/zrok/api/models/AccessRequest.ts
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

// src/zrok/api/models/AuthUser.ts
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

// src/zrok/api/models/UnaccessRequest.ts
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

// src/zrok/api/models/UpdateShareRequest.ts
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

// src/zrok/api/apis/ShareApi.ts
var ShareApi = class extends BaseAPI {
  /**
   */
  accessRaw(requestParameters, initOverrides) {
    return __async(this, null, function* () {
      const queryParameters = {};
      const headerParameters = {};
      headerParameters["Content-Type"] = "application/zrok.v1+json";
      if (this.configuration && this.configuration.apiKey) {
        headerParameters["x-token"] = this.configuration.apiKey("x-token");
      }
      const response = yield this.request({
        path: `/access`,
        method: "POST",
        headers: headerParameters,
        query: queryParameters,
        body: AccessRequestToJSON(requestParameters.body)
      }, initOverrides);
      return new JSONApiResponse(response, (jsonValue) => AccessResponseFromJSON(jsonValue));
    });
  }
  /**
   */
  access() {
    return __async(this, arguments, function* (requestParameters = {}, initOverrides) {
      const response = yield this.accessRaw(requestParameters, initOverrides);
      return yield response.value();
    });
  }
  /**
   */
  shareRaw(requestParameters, initOverrides) {
    return __async(this, null, function* () {
      const queryParameters = {};
      const headerParameters = {};
      headerParameters["Content-Type"] = "application/zrok.v1+json";
      if (this.configuration && this.configuration.apiKey) {
        headerParameters["x-token"] = this.configuration.apiKey("x-token");
      }
      const response = yield this.request({
        path: `/share`,
        method: "POST",
        headers: headerParameters,
        query: queryParameters,
        body: ShareRequestToJSON(requestParameters.body)
      }, initOverrides);
      return new JSONApiResponse(response, (jsonValue) => ShareResponseFromJSON(jsonValue));
    });
  }
  /**
   */
  share() {
    return __async(this, arguments, function* (requestParameters = {}, initOverrides) {
      const response = yield this.shareRaw(requestParameters, initOverrides);
      return yield response.value();
    });
  }
  /**
   */
  unaccessRaw(requestParameters, initOverrides) {
    return __async(this, null, function* () {
      const queryParameters = {};
      const headerParameters = {};
      headerParameters["Content-Type"] = "application/zrok.v1+json";
      if (this.configuration && this.configuration.apiKey) {
        headerParameters["x-token"] = this.configuration.apiKey("x-token");
      }
      const response = yield this.request({
        path: `/unaccess`,
        method: "DELETE",
        headers: headerParameters,
        query: queryParameters,
        body: UnaccessRequestToJSON(requestParameters.body)
      }, initOverrides);
      return new VoidApiResponse(response);
    });
  }
  /**
   */
  unaccess() {
    return __async(this, arguments, function* (requestParameters = {}, initOverrides) {
      yield this.unaccessRaw(requestParameters, initOverrides);
    });
  }
  /**
   */
  unshareRaw(requestParameters, initOverrides) {
    return __async(this, null, function* () {
      const queryParameters = {};
      const headerParameters = {};
      headerParameters["Content-Type"] = "application/zrok.v1+json";
      if (this.configuration && this.configuration.apiKey) {
        headerParameters["x-token"] = this.configuration.apiKey("x-token");
      }
      const response = yield this.request({
        path: `/unshare`,
        method: "DELETE",
        headers: headerParameters,
        query: queryParameters,
        body: UnshareRequestToJSON(requestParameters.body)
      }, initOverrides);
      return new VoidApiResponse(response);
    });
  }
  /**
   */
  unshare() {
    return __async(this, arguments, function* (requestParameters = {}, initOverrides) {
      yield this.unshareRaw(requestParameters, initOverrides);
    });
  }
  /**
   */
  updateShareRaw(requestParameters, initOverrides) {
    return __async(this, null, function* () {
      const queryParameters = {};
      const headerParameters = {};
      headerParameters["Content-Type"] = "application/zrok.v1+json";
      if (this.configuration && this.configuration.apiKey) {
        headerParameters["x-token"] = this.configuration.apiKey("x-token");
      }
      const response = yield this.request({
        path: `/share`,
        method: "PATCH",
        headers: headerParameters,
        query: queryParameters,
        body: UpdateShareRequestToJSON(requestParameters.body)
      }, initOverrides);
      return new VoidApiResponse(response);
    });
  }
  /**
   */
  updateShare() {
    return __async(this, arguments, function* (requestParameters = {}, initOverrides) {
      yield this.updateShareRaw(requestParameters, initOverrides);
    });
  }
};

// src/zrok/model.ts
var PROXY_BACKEND_MODE = "proxy";
var WEB_BACKEND_MODE = "web";
var TCP_TUNNEL_BACKEND_MODE = "tcpTunnel";
var UDP_TUNNEL_BACKEND_MODE = "udpTunnel";
var CADDY_BACKEND_MODE = "caddy";
var PRIVATE_SHARE_MODE = "private";
var PUBLIC_SHARE_MODE = "public";
var ShareRequest = class {
  constructor(backendMode, shareMode, target, frontends = [], basicAuth = [], oauthProvider = "", oauthEmailDomains = [], oauthAuthorizationCheckInterval = "") {
    this.BackendMode = backendMode;
    this.ShareMode = shareMode;
    this.Target = target;
    this.Frontends = frontends;
    this.BasicAuth = basicAuth;
    this.OauthProvider = oauthProvider;
    this.OauthEmailDomains = oauthEmailDomains;
    this.OauthAuthorizationCheckInterval = oauthAuthorizationCheckInterval;
  }
};
var Share = class {
  constructor(Token, FrontendEndpoints) {
    this.Token = Token;
    this.FrontendEndpoints = FrontendEndpoints;
  }
};
var AccessRequest = class {
  constructor(ShareToken) {
    this.ShareToken = ShareToken;
  }
};
var Access = class {
  constructor(Token, ShareToken, BackendMode) {
    this.Token = Token;
    this.ShareToken = ShareToken;
    this.BackendMode = BackendMode;
  }
};
var SessionMetrics = class {
  constructor(BytesRead, BytesWrittern, LastUpdate) {
    this.BytesRead = BytesRead;
    this.BytesWritten = BytesWrittern;
    this.LastUpdate = LastUpdate;
  }
};
var Metrics = class {
  constructor(Namespace, Sessions) {
    this.Namespace = Namespace;
    this.Sessions = Sessions;
  }
};
var AUTH_SCHEME_NONE = "none";
var AUTH_SCHEME_BASIC = "basic";
var AUTH_SCHEME_OAUTH = "oauth";
function zrokBackendModeToOpenApi(z) {
  switch (z) {
    case PROXY_BACKEND_MODE:
      return ShareRequestBackendModeEnum.Proxy;
    case WEB_BACKEND_MODE:
      return ShareRequestBackendModeEnum.Web;
    case TCP_TUNNEL_BACKEND_MODE:
      return ShareRequestBackendModeEnum.TcpTunnel;
    case UDP_TUNNEL_BACKEND_MODE:
      return ShareRequestBackendModeEnum.UdpTunnel;
    case CADDY_BACKEND_MODE:
      return ShareRequestBackendModeEnum.Caddy;
    default:
      return void 0;
  }
}
function zrokShareModeToOpenApi(z) {
  switch (z) {
    case PRIVATE_SHARE_MODE:
      return ShareRequestShareModeEnum.Private;
    case PUBLIC_SHARE_MODE:
      return ShareRequestShareModeEnum.Public;
    default:
      return void 0;
  }
}
function zrokOauthProviderToOpenApi(z) {
  switch (z.toLowerCase()) {
    case ShareRequestOauthProviderEnum.Github.toString().toLowerCase():
      return ShareRequestOauthProviderEnum.Github;
    case ShareRequestOauthProviderEnum.Google.toString().toLowerCase():
      return ShareRequestOauthProviderEnum.Google;
    default:
      return void 0;
  }
}

// src/zrok/share.ts
function CreateShare(root, request) {
  if (!root.IsEnabled()) {
    throw new Error("environment is not enabled; enable with 'zrok enable' first!");
  }
  let out;
  switch (request.ShareMode) {
    case ShareRequestShareModeEnum.Private:
      out = newPrivateShare(root, request);
      break;
    case ShareRequestShareModeEnum.Public:
      out = newPublicShare(root, request);
      break;
    default:
      throw new Error("unknown share mode " + request.ShareMode);
  }
  if (request.BasicAuth.length > 0) {
    out.authScheme = AUTH_SCHEME_BASIC;
    for (let pair in request.BasicAuth) {
      let tokens = pair.split(":");
      if (tokens.length === 2) {
        if (out.authUsers === void 0) {
          out.authUsers = new Array();
        }
        out.authUsers.push({ username: tokens[0].trim(), password: tokens[1].trim() });
      } else {
        throw new Error("invalid username:password pair: " + pair);
      }
    }
  }
  if (request.OauthProvider !== void 0) {
    out.authScheme = AUTH_SCHEME_OAUTH;
  }
  let client = new ShareApi(root.Client());
  let shr = null;
  client.share({ body: out }).then((resp) => {
    console.log("creating shr ret");
    shr = new Share(resp.shrToken || "", resp.frontendProxyEndpoints || []);
    console.log(shr);
  }).catch((resp) => {
    console.log("unable to create share");
    throw new Error("unable to create share " + resp);
  });
  console.log("wat");
  console.log(shr);
  return shr;
}
function newPrivateShare(root, request) {
  return {
    envZId: root.env.ZitiIdentity,
    shareMode: zrokShareModeToOpenApi(request.ShareMode),
    backendMode: zrokBackendModeToOpenApi(request.BackendMode),
    backendProxyEndpoint: request.Target,
    authScheme: AUTH_SCHEME_NONE
  };
}
function newPublicShare(root, request) {
  return {
    envZId: root.env.ZitiIdentity,
    shareMode: zrokShareModeToOpenApi(request.ShareMode),
    backendMode: zrokBackendModeToOpenApi(request.BackendMode),
    backendProxyEndpoint: request.Target,
    authScheme: AUTH_SCHEME_NONE,
    oauthProvider: zrokOauthProviderToOpenApi(request.OauthProvider),
    oauthEmailDomains: request.OauthEmailDomains,
    oauthAuthorizationCheckInterval: request.OauthAuthorizationCheckInterval
  };
}
function DeleteShare(root, shr) {
  let client = new ShareApi(root.Client());
  let req = {
    envZId: root.env.ZitiIdentity,
    shrToken: shr.Token
  };
  req.envZId = root.env.ZitiIdentity;
  client.unshare({ body: {} }).catch((resp) => {
    throw new Error("error deleting share " + resp);
  });
}

// src/environment/dirs.ts
import { homedir } from "os";
import { join } from "path";
function rootDir() {
  return join(homedir(), ".zrok");
}
function metadataFile() {
  return join(rootDir(), "metadata.json");
}
function configFile() {
  return join(rootDir(), "config.json");
}
function environmentFile() {
  return join(rootDir(), "environment.json");
}
function identitiesDir() {
  return join(rootDir(), "identities");
}
function identityFile(name) {
  return join(identitiesDir(), name + ".json");
}

// src/environment/root.ts
import fs from "fs";
var V = "v0.4";
var Metadata = class {
  constructor(V2, RootPath = "") {
    this.V = V2;
    this.RootPath = RootPath;
  }
};
var ApiEndpoint = class {
  constructor(endpoint, frm) {
    this.endpoint = endpoint;
    this.frm = frm;
  }
};
var Config = class {
  constructor(ApiEndpoint2) {
    this.ApiEndpoint = ApiEndpoint2;
  }
};
var Environment = class {
  constructor(Token, ZitiIdentity, ApiEndpoint2) {
    this.Token = Token;
    this.ZitiIdentity = ZitiIdentity;
    this.ApiEndpoint = ApiEndpoint2;
  }
};
var Root = class {
  constructor(meta = new Metadata(V, rootDir()), cfg, env) {
    this.meta = meta;
    this.cfg = cfg;
    this.env = env;
  }
  HasConfig() {
    return this.cfg !== void 0 && Object.keys(this.cfg).length === 0;
  }
  Client() {
    let apiEndpoint = this.ApiEndpoint();
    let conf = new Configuration({
      basePath: apiEndpoint.endpoint + "/api/v1",
      accessToken: this.env.Token
    });
    let mapi = new MetadataApi(conf);
    let ver = mapi.version();
    const regex = new RegExp("^(refs/(heads|tags)/)?" + V);
    ver.then((v) => {
      console.log("got version " + v);
      if (!regex.test(v)) {
        throw new Error("Expected a '" + V + "' version, received: '" + v + "'");
      }
    });
    return conf;
  }
  ApiEndpoint() {
    let apiEndpoint = "https://api.zrok.io";
    let frm = "binary";
    if (this.cfg.ApiEndpoint != "") {
      apiEndpoint = this.cfg.ApiEndpoint;
      frm = "config";
    }
    let env = process.env.ZROK_API_ENDPOINT;
    if (env != null) {
      apiEndpoint = env;
      frm = "ZROK_API_ENDPOINT";
    }
    if (this.IsEnabled()) {
      apiEndpoint = this.env.ApiEndpoint;
      frm = "env";
    }
    return new ApiEndpoint(apiEndpoint.replace(/\/+$/, ""), frm);
  }
  IsEnabled() {
    return this.env !== void 0 && Object.keys(this.env).length > 0;
  }
  PublicIdentityName() {
    return "public";
  }
  EnvironmentIdentityName() {
    return "environment";
  }
  ZitiIdentityName(name) {
    return identityFile(name);
  }
};
function Assert() {
  if (rootExists()) {
    let meta = loadMetadata();
    return meta.V == V;
  }
  return false;
}
function Load() {
  if (rootExists()) {
    return new Root(loadMetadata(), loadConfig(), loadEnvironment());
  }
  throw new Error("unable to load root. Does not exist");
}
function rootExists() {
  return fs.existsSync(metadataFile());
}
function loadMetadata() {
  let mf = metadataFile();
  let data = fs.readFileSync(mf);
  let serial = JSON.parse(data.toString());
  return new Metadata(serial.v);
}
function loadConfig() {
  let cf = configFile();
  let data = fs.readFileSync(cf);
  let serial = JSON.parse(data.toString());
  return new Config(serial.api_endpoint);
}
function loadEnvironment() {
  let ef = environmentFile();
  let data = fs.readFileSync(ef);
  let serial = JSON.parse(data.toString());
  return new Environment(serial.zrok_token, serial.ziti_identity, serial.api_endpoint);
}
export {
  AUTH_SCHEME_BASIC,
  AUTH_SCHEME_NONE,
  AUTH_SCHEME_OAUTH,
  Access,
  AccessRequest,
  ApiEndpoint,
  Assert,
  CADDY_BACKEND_MODE,
  Config,
  CreateShare,
  DeleteShare,
  Environment,
  Load,
  Metadata,
  Metrics,
  PRIVATE_SHARE_MODE,
  PROXY_BACKEND_MODE,
  PUBLIC_SHARE_MODE,
  Root,
  SessionMetrics,
  Share,
  ShareRequest,
  TCP_TUNNEL_BACKEND_MODE,
  UDP_TUNNEL_BACKEND_MODE,
  WEB_BACKEND_MODE,
  configFile,
  environmentFile,
  identitiesDir,
  identityFile,
  metadataFile,
  rootDir,
  zrokBackendModeToOpenApi,
  zrokOauthProviderToOpenApi,
  zrokShareModeToOpenApi
};
//# sourceMappingURL=index.mjs.map