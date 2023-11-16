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
export {
  AUTH_SCHEME_BASIC,
  AUTH_SCHEME_NONE,
  AUTH_SCHEME_OAUTH,
  Access,
  AccessRequest,
  CADDY_BACKEND_MODE,
  Metrics,
  PRIVATE_SHARE_MODE,
  PROXY_BACKEND_MODE,
  PUBLIC_SHARE_MODE,
  SessionMetrics,
  Share,
  ShareRequest,
  TCP_TUNNEL_BACKEND_MODE,
  UDP_TUNNEL_BACKEND_MODE,
  WEB_BACKEND_MODE,
  zrokBackendModeToOpenApi,
  zrokOauthProviderToOpenApi,
  zrokShareModeToOpenApi
};
//# sourceMappingURL=model.mjs.map