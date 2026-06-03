# Design: `X-Zrok-Session` header authentication

## Overview

This document describes the design for allowing the zrok session token to be
presented as an HTTP request header (`X-Zrok-Session`) in addition to the
existing browser cookie. The goal is to give non-browser API clients a
first-class way to authenticate to OAuth-protected shares.

Applies to: `dynamicProxy`, `publicProxy`.

---

## Motivation

zrok shares protected by OAuth/OIDC currently authenticate via a session
cookie (`zrok_session` or the configured `CookieName`). Browser clients handle
cookies automatically, but programmatic clients — CLI tools, SDKs, server-side
applications — cannot easily follow OIDC redirect flows or store cookies. This
design adds:

1. **Header-based session presentation** — clients may send
   `X-Zrok-Session: <jwt>` instead of (or alongside) the cookie.
2. **API-friendly failure mode** — when `X-Zrok-Session` is present and
   authentication fails, the proxy returns `401 Unauthorized` with a `Location`
   header pointing to the login URL instead of issuing a browser redirect.
3. **Token retrieval flow** — a way for a user to complete the OIDC login in a
   browser and then copy the resulting session token for use in API clients.
4. **Inline token refresh** — when an OIDC token is refreshed mid-session for a
   header-based request, the new token is returned in an `X-Zrok-Session`
   response header so the client can update its stored value.

---

## Decisions

| # | Question | Decision |
| --- | --- | --- |
| 1 | Header name | `X-Zrok-Session` |
| 2 | Token format in header | Raw JWT string (no gzip compression, no cookie striping) |
| 3 | Login flow for non-browser clients | 401 + `Location` header; OIDC callback page shows the token for copying |
| 4 | Which proxies | Both `dynamicProxy` and `publicProxy` |
| 5 | Strip header before proxying to backend | Yes — same as session cookies are currently stripped |
| 6 | Token refresh | Inline OIDC refresh; new JWT echoed in `X-Zrok-Session` response header |
| 7 | CORS exposure | Proxy appends `X-Zrok-Session` to `Access-Control-Allow-Headers` and `Access-Control-Expose-Headers` on every upstream response |

---

## Authorization logic

The existing flow redirects to the OIDC provider whenever no valid session is
found. The new flow adds header reading, a 401 path, and inline refresh. The
cookie path is unchanged.

```text
Request arrives
├── X-Zrok-Session header present?
│   ├── YES → parse raw JWT
│   │   ├── invalid or expired → 401 + Location (always; header implies API client)
│   │   └── valid; NextRefresh passed?
│   │       ├── OIDC provider → inline refresh → set X-Zrok-Session response header → continue
│   │       ├── GitHub or Google (no refresh support) → 401 + Location
│   │       └── NextRefresh not yet passed → continue
│   └── NO → read session cookie (existing behavior)
│       ├── cookie absent or invalid → 302 redirect to OIDC login (unchanged)
│       └── cookie valid; NextRefresh passed?
│           ├── OIDC provider → 302 to refresh endpoint (unchanged)
│           └── GitHub or Google → 302 to login (unchanged)
```

### Header precedence

When both `X-Zrok-Session` header and a session cookie are present, the header
takes precedence and the cookie is ignored.

---

## 401 response format

When the server returns `401 Unauthorized`, it sets:

- `Location: <login-url>` — absolute URL to the provider's login endpoint,
  with `return_token=true` appended so the post-login page displays the token.
- `Content-Type: application/json`
- Body:

```json
{"login_url": "https://example.com/oidc/login?targetHost=...&refreshInterval=...&return_token=true"}
```

Clients can open the `login_url` in a browser, complete the OIDC flow, and
copy the token from the resulting page.

---

## Token retrieval flow (Option C)

The `return_token=true` query parameter propagates through the OIDC flow via
the intermediate JWT state claim:

```text
1. API client makes request → 401 + Location: .../oidc/login?...&return_token=true
2. User opens Location URL in a browser
3. Login handler reads return_token=true → encodes ReturnToken: true in IntermediateJWT
4. User authenticates with OIDC provider
5. Callback handler parses IntermediateJWT; sees ReturnToken == true
6. Instead of redirecting to targetHost, renders a "token display" page:
   - Shows raw JWT in a text box
   - "Copy to clipboard" button
   - Usage instruction: X-Zrok-Session: <token>
   - Token expiry time
   (The session cookie is also set, so the browser session still works.)
7. User copies token and pastes it into their API client config
```

---

## Inline token refresh (OIDC only)

When a request carries `X-Zrok-Session` and `NextRefresh` has passed:

```text
handleOAuth detects header-based request + NextRefresh elapsed
  → looks up OIDC provider (via globalOAuthRouter.GetProvider in dynamicProxy,
    or providerRegistry in publicProxy)
  → calls provider.RefreshSessionJWT(claims):
      1. decrypt claims.AccessToken with provider's encryptionKey
      2. call rp.RefreshTokens against the OIDC issuer
      3. re-encrypt new access token
      4. build new zrokClaims (updated NextRefresh, same Email/TargetHost/etc.)
      5. sign with provider's signingKey
      6. return new JWT string
  → on success: w.Header().Set("X-Zrok-Session", newJWT); continue request
  → on failure: 401 + Location
```

GitHub and Google providers do not support token refresh (`supportsRefresh:
false`). When `NextRefresh` passes for a header-based request on those
providers, the client receives `401 + Location` and must re-authenticate.

---

## Changes required

### New shared utilities — `endpoints/oauthCookies.go`

- `const SessionHeaderName = "X-Zrok-Session"`
- `func GetSessionHeader(r *http.Request) string`
  — returns the `X-Zrok-Session` header value, or `""` if absent.
- `func StripSessionHeader(r *http.Request)`
  — removes `X-Zrok-Session` from the request before it is proxied to the
  backend, mirroring how session cookies are currently stripped.

---

### New token display page — `endpoints/proxyUi/`

| File | Purpose |
| --- | --- |
| `endpoints/proxyUi/token.go` | `WriteTokenDisplay(w, token, expiry, targetHost)` — renders the token page |
| `endpoints/proxyUi/token.html` | HTML template: token text box, copy button, usage instructions, expiry |

`token.html` uses the same visual style as `template.html`: Poppins + JetBrains
Mono fonts from Google Fonts, dark purple (`#241775`) background, the zrok SVG
logo banner, and a centered white card for the content area.

When `targetHost` is non-empty, the page emits a small script that calls
`window.opener.postMessage({ xZrokSession: <token> }, '*')`
so that browser-based clients that open the login in a popup can receive the
token automatically. The script is suppressed entirely when `targetHost` is
empty. The `'*'` target origin is used because the opener's origin is unknown
at the time the page is rendered; the token itself is the credential and is
already visible on the page, so the broadcast does not increase exposure.

---

### `IntermediateJWT` — new `ReturnToken` field

Both `endpoints/dynamicProxy/auth.go` and `endpoints/publicProxy/auth.go`:

```go
type IntermediateJWT struct {
    State           string `json:"st"`
    TargetHost      string `json:"th"`
    RefreshInterval string `json:"rfi"`
    ReturnToken     bool   `json:"rt,omitempty"`   // new
    jwt.RegisteredClaims
}
```

---

### Provider login handlers — read `?return_token=true`

All six `authHandler()` functions (three providers × two proxies):

- Read `r.URL.Query().Get("return_token")`; if `"true"`, set
  `ReturnToken: true` in the `IntermediateJWT`.

---

### Provider callback handlers — render token display

All six callback / `loginHandler()` functions:

- After calling `setSessionCookie()` (which produces the signed JWT string):
  - If `IntermediateJWT.ReturnToken == true` → call
    `proxyUi.WriteTokenDisplay(w, sTkn, expiry, intermediateJWT.TargetHost)`
    instead of `http.Redirect`.
  - Otherwise → existing redirect to `targetHost`.
- The cookie is set in both cases so the browser session also works.

---

### Inline refresh infrastructure

**`dynamicProxy/authOauthRouter.go`**

Add a provider lookup method:

```go
func (r *oauthRouter) GetProvider(name string) (oauthProvider, bool)
```

**`dynamicProxy/cookies.go`**

Factor the JWT signing logic out of `setSessionCookie` into a reusable helper:

```go
func buildSessionJWT(req sessionCookieRequest) (string, error)
```

`setSessionCookie` calls `buildSessionJWT` and then writes the cookie. The
inline refresh path calls `buildSessionJWT` and writes a response header
instead.

**`dynamicProxy/providerOidc.go`**

New method on `oidcProvider`:

```go
func (p *oidcProvider) RefreshSessionJWT(claims *zrokClaims) (string, error)
```

**`publicProxy/config.go`** (or `publicProxy/auth.go`)

Add a package-level provider registry (mirrors `globalOAuthRouter` in
`dynamicProxy`):

```go
var providerRegistry = map[string]oauthProvider{}
```

Each provider's `configure()` registers itself here. `authHandler` uses the
registry for inline refresh lookups.

**`publicProxy/providerOidc.go`**

New method `RefreshSessionJWT` (same contract as the `dynamicProxy` version).

---

### `handleOAuth` and `validateOAuthToken` — both proxies

**`handleOAuth`** (in `dynamicProxy/authOauth.go` and
`publicProxy/authOAuth.go`):

- Tries `GetSessionHeader` first; sets `fromHeader = true` if found.
- Falls back to `getSessionCookie` when no header is present.
- Calls `oauthUnauthorized` (401) when `fromHeader` is true and auth fails,
  or `oauthLoginRequired` (302) when falling back to cookie-based flow.
- Passes `jwtString string`, `fromHeader bool` to `validateOAuthToken`.

**`validateOAuthToken`** — signature change from `cookie *http.Cookie` to
`jwtString string, fromHeader bool`:

| Condition | Behavior |
| --- | --- |
| JWT parse/validation failure + `fromHeader` | `oauthUnauthorized` (401) |
| JWT parse/validation failure + `!fromHeader` | `oauthLoginRequired` (302) |
| `NextRefresh` passed + `fromHeader` + OIDC | inline refresh → `X-Zrok-Session` response header → continue |
| `NextRefresh` passed + `fromHeader` + GitHub/Google | `oauthUnauthorized` (401) |
| `NextRefresh` passed + `!fromHeader` + OIDC | `oauthRefreshRequired` (302, unchanged) |
| `NextRefresh` passed + `!fromHeader` + GitHub/Google | `oauthLoginRequired` (302, unchanged) |

**`validateEmailDomain`** — signature change from `cookie *http.Cookie` to
`jwtString string`.

**New `oauthUnauthorized`** alongside existing `oauthLoginRequired`:

```go
func oauthUnauthorized(w http.ResponseWriter, r *http.Request, cfg *oauthConfig,
    provider, target string, refreshInterval time.Duration) {

    loginURL := fmt.Sprintf(
        "%s/%s/login?targetHost=%s&refreshInterval=%s&return_token=true",
        cfg.EndpointUrl, provider,
        url.QueryEscape(target), refreshInterval.String(),
    )
    if origin := r.Header.Get("Origin"); origin != "" {
        w.Header().Set("Access-Control-Allow-Origin", origin)
        w.Header().Set("Access-Control-Allow-Credentials", "true")
        w.Header().Add("Vary", "Origin")
    }
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Location", loginURL)
    w.WriteHeader(http.StatusUnauthorized)
    fmt.Fprintf(w, `{"login_url":%q}`, loginURL)
}
```

The `r *http.Request` parameter is used to echo the request's `Origin` back as
`Access-Control-Allow-Origin` so that browser-based clients using the
`X-Zrok-Session` header can read the 401 JSON body across origins. Without this,
the browser's CORS policy would block the response entirely.

---

### Cookie filtering — strip `X-Zrok-Session` from proxied requests

**`dynamicProxy/cookies.go`** and **`publicProxy/cookies.go`** —
`filterSessionCookies()`:

- `endpoints.StripSessionHeader(r)` — removes the `X-Zrok-Session` request
  header before the request is forwarded to the backend, so the backend never
  sees the zrok session token.
- `endpoints.StripSessionFromACRH(r)` — removes `X-Zrok-Session` from the
  `Access-Control-Request-Headers` field of CORS preflights before forwarding
  to the backend. Backends that do not recognise this header would otherwise
  reject the preflight with a 4xx response. `AppendSessionCORSHeaders` in
  `ModifyResponse` adds `X-Zrok-Session` back to `Access-Control-Allow-Headers`
  in the response, so the browser still knows it may send the header on the
  real request.

### CORS headers — expose `X-Zrok-Session` to browser clients

Browser-based API clients that open the OIDC login in a popup and read the
refreshed JWT from the response header require two CORS changes:

- **`Access-Control-Allow-Headers: X-Zrok-Session`** — without this, the
  browser's preflight check blocks cross-origin requests that carry the header.
- **`Access-Control-Expose-Headers: X-Zrok-Session`** — without this, the
  browser hides the response header from JavaScript even when the server sets
  it (e.g. during inline OIDC refresh).

Both are injected via `ModifyResponse` on the reverse proxy so they are added
to every upstream response. The proxy appends to any values the upstream
already set rather than overwriting them, and the append is idempotent.

**`endpoints/oauthCookies.go`** — new helper:

```go
func AppendSessionCORSHeaders(h http.Header) {
    for _, key := range []string{
        "Access-Control-Allow-Headers",
        "Access-Control-Expose-Headers",
    } {
        existing := h.Get(key)
        if existing == "" {
            h.Set(key, SessionHeaderName)
        } else if !strings.Contains(existing, SessionHeaderName) {
            h.Set(key, existing+", "+SessionHeaderName)
        }
    }
}
```

**`dynamicProxy/http.go`** and **`publicProxy/http.go`** — `ModifyResponse`:

```go
proxy.ModifyResponse = func(resp *http.Response) error {
    endpoints.AppendSessionCORSHeaders(resp.Header)
    return nil
}
```

### Bug fix — `filterSessionCookies` missing from OAuth auth-success path

**`dynamicProxy/http.go`** and **`publicProxy/http.go`**: `filterSessionCookies`
was already called on the `sdk.None` and `sdk.Basic` auth-success paths, but
was absent from the `sdk.Oauth` path — meaning session cookies and the
`X-Zrok-Session` header were never stripped for OAuth-authenticated requests.
This was discovered during testing and corrected:

```go
case string(sdk.Oauth):
    if auth.handleOAuth(w, r, svcCfg, shrToken) {
        filterSessionCookies(w, r, cfg)   // added
        handler.ServeHTTP(w, r)
    }
```

---

## File change summary

| File | Type of change |
| --- | --- |
| `endpoints/oauthCookies.go` | Add `SessionHeaderName`, `GetSessionHeader`, `StripSessionHeader`, `StripSessionFromACRH`, `AppendSessionCORSHeaders` |
| `endpoints/proxyUi/token.go` | **New** — `WriteTokenDisplay(w, token, expiry, targetHost)` function |
| `endpoints/proxyUi/token.html` | **New** — token display page template (styled to match `template.html`); conditional `postMessage` script using `'*'` target origin, suppressed when `targetHost` is empty |
| `endpoints/dynamicProxy/auth.go` | Add `ReturnToken` to `IntermediateJWT`; add `sessionRefresher` interface |
| `endpoints/dynamicProxy/authOauth.go` | Rewrite `handleOAuth`, `validateOAuthToken`, `validateEmailDomain`; add `oauthUnauthorized`, `tryInlineRefresh` |
| `endpoints/dynamicProxy/authOauthRouter.go` | Add `GetProvider()` method |
| `endpoints/dynamicProxy/cookies.go` | Add `buildSessionJWT`; update `setSessionCookie` to return JWT string; update `filterSessionCookies` to strip header and ACRH |
| `endpoints/dynamicProxy/http.go` | **Bug fix** — add `filterSessionCookies` call on OAuth auth-success path; add OPTIONS bypass; inject CORS headers in `ModifyResponse` |
| `endpoints/dynamicProxy/providerOidc.go` | Add `ReturnToken` handling; add `RefreshSessionJWT`; pass `targetHost` to `WriteTokenDisplay` |
| `endpoints/dynamicProxy/providerGithub.go` | Add `ReturnToken` handling; pass `targetHost` to `WriteTokenDisplay` |
| `endpoints/dynamicProxy/providerGoogle.go` | Add `ReturnToken` handling; pass `targetHost` to `WriteTokenDisplay` |
| `endpoints/publicProxy/auth.go` | Add `ReturnToken` to `IntermediateJWT`; add `sessionRefresher` interface |
| `endpoints/publicProxy/authOAuth.go` | Same changes as `dynamicProxy/authOauth.go` |
| `endpoints/publicProxy/cookies.go` | Add `buildSessionJWT`; update `setSessionCookie` to return JWT string; strip `X-Zrok-Session` header and ACRH in `filterSessionCookies` |
| `endpoints/publicProxy/http.go` | **Bug fix** — add `filterSessionCookies` call on OAuth auth-success path; add OPTIONS bypass; inject CORS headers in `ModifyResponse` |
| `endpoints/publicProxy/oidcRegistry.go` | **New** — package-level `oidcProviderRegistry` map for inline refresh lookups |
| `endpoints/publicProxy/providerOidc.go` | Add `ReturnToken` handling; add `RefreshSessionJWT`; register in `oidcProviderRegistry`; pass `targetHost` to `WriteTokenDisplay` |
| `endpoints/publicProxy/providerGithub.go` | Add `ReturnToken` handling; pass `targetHost` to `WriteTokenDisplay` |
| `endpoints/publicProxy/providerGoogle.go` | Add `ReturnToken` handling; pass `targetHost` to `WriteTokenDisplay` |

---

## Tests implemented

### `endpoints/oauthCookies_test.go` — extended

| Test | Behavior |
| --- | --- |
| `TestGetSessionHeaderPresent` | Header present → returns value |
| `TestGetSessionHeaderAbsent` | No header → returns `""` |
| `TestStripSessionHeaderRemoves` | Header present → removed from request |
| `TestStripSessionHeaderNoOp` | No header → no panic |
| `TestAppendSessionCORSHeadersSetsWhenAbsent` | No existing CORS headers → both set to `X-Zrok-Session` |
| `TestAppendSessionCORSHeadersAppendsWhenPresent` | Existing values → `X-Zrok-Session` appended with comma-space separator |
| `TestAppendSessionCORSHeadersIdempotent` | Called twice → header value unchanged on second call |
| `TestAppendSessionCORSHeadersIdempotentLowercase` | Lowercase variant already present → no duplicate appended |
| `TestStripSessionFromACRHRemovesSessionHeader` | `x-zrok-session,zt-session` → `zt-session` after stripping |
| `TestStripSessionFromACRHRemovesOnlyEntry` | `X-Zrok-Session` only → header deleted |
| `TestStripSessionFromACRHNoopWhenAbsent` | No `X-Zrok-Session` in list → list unchanged |

---

### `endpoints/dynamicProxy/http_test.go` — new file

A `stubDynamicProxyService` helper replaces the `getRefreshedService` package
variable for the duration of each test, returning a synthetic `ServiceDetail`
with the caller-supplied proxy config map.

| Test | Behavior |
| --- | --- |
| `TestShareHandlerOptionsBypassesFrontendAuth` | OPTIONS request → reaches upstream without auth challenge; session cookie stripped for OAuth scheme |
| `TestShareHandlerGetStillRedirectsForOAuth` | GET without session, OAuth scheme → 302 redirect to provider login |
| `TestShareHandlerGetStillChallengesBasicAuth` | GET without credentials, Basic scheme → 401 |

---

### `endpoints/publicProxy/http_test.go` — new file

Mirror of the `dynamicProxy` tests above, adjusted for `publicProxy` types
(`Config`, `OauthConfig`, `stubPublicProxyService`). Same three tests.

---

### `endpoints/proxyUi/token_test.go` — new file

| Test | Behavior |
| --- | --- |
| `TestWriteTokenDisplayStatusOK` | Returns HTTP 200 |
| `TestWriteTokenDisplayContainsToken` | Token value appears in response body |
| `TestWriteTokenDisplayEscapesHTML` | Token containing `<script>` is HTML-escaped (XSS) |
| `TestWriteTokenDisplayContainsHeaderName` | Body references `X-Zrok-Session` |
| `TestWriteTokenDisplayPostMessageScriptPresentWithTargetHost` | Non-empty `targetHost` → `postMessage` script present in body |
| `TestWriteTokenDisplayNoPostMessageScriptWithoutTargetHost` | Empty `targetHost` → `postMessage` script absent from body |

---

### `endpoints/dynamicProxy/cookies_test.go` — new file

| Test | Behavior |
| --- | --- |
| `TestBuildSessionJWTRejectsEmptyTargetHost` | Empty `targetHost` → error returned |
| `TestBuildSessionJWTRoundTrip` | Valid inputs → non-empty signed JWT string returned |
| `TestBuildSessionJWTStripsPathFromTargetHost` | `host/path` input → path component stripped from `TargetHost` claim |

---

### Test file summary

| File | New / Extended | Covers |
| --- | --- | --- |
| `endpoints/oauthCookies_test.go` | Extended | `GetSessionHeader`, `StripSessionHeader`, `AppendSessionCORSHeaders` |
| `endpoints/dynamicProxy/http_test.go` | **New** | OPTIONS bypass, OAuth redirect, Basic auth challenge |
| `endpoints/publicProxy/http_test.go` | **New** | Same for `publicProxy` |
| `endpoints/proxyUi/token_test.go` | **New** | `WriteTokenDisplay` rendering, XSS escaping, `postMessage` script conditional |
| `endpoints/dynamicProxy/cookies_test.go` | **New** | `buildSessionJWT` error and round-trip cases |

---

## Out of scope

- The `refreshHandler()` HTTP endpoint (used by browser-based OIDC refresh
  redirects) is not modified. Header-based refresh is handled entirely inline
  in `validateOAuthToken`; the browser redirect path is unchanged.
- No changes to GitHub or Google provider refresh behavior (neither supports
  token refresh; clients must re-authenticate when `NextRefresh` passes).
