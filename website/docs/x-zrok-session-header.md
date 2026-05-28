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
2. **API-friendly failure mode** — a per-share option to return `401
   Unauthorized` with a `Location` header pointing to the login URL instead of
   issuing a browser redirect.
3. **Token retrieval flow** — a way for a user to complete the OIDC login in a
   browser and then copy the resulting session token for use in API clients.
4. **Inline token refresh** — when an OIDC token is refreshed mid-session for a
   header-based request, the new token is returned in an `X-Zrok-Session`
   response header so the client can update its stored value.

---

## Decisions

| # | Question | Decision |
|---|---|---|
| 1 | Header name | `X-Zrok-Session` |
| 2 | Token format in header | Raw JWT string (no gzip compression, no cookie striping) |
| 3 | Login flow for non-browser clients | 401 + `Location` header; OIDC callback page shows the token for copying |
| 4 | 401 configuration scope | Per-share, via `no_redirect: true` in the share's oauth config element |
| 5 | Which proxies | Both `dynamicProxy` and `publicProxy` |
| 6 | Strip header before proxying to backend | Yes — same as session cookies are currently stripped |
| 7 | Token refresh | Inline OIDC refresh; new JWT echoed in `X-Zrok-Session` response header |

---

## Authorization logic

The existing flow redirects to the OIDC provider whenever no valid session is
found. The new flow adds header reading, a 401 path, and inline refresh. The
cookie path is unchanged when `no_redirect` is false (the default).

```
Request arrives
├── X-Zrok-Session header present?
│   ├── YES → parse raw JWT
│   │   ├── invalid or expired → 401 + Location (always; header implies API client)
│   │   └── valid; NextRefresh passed?
│   │       ├── OIDC provider → inline refresh → set X-Zrok-Session response header → continue
│   │       ├── GitHub or Google (no refresh support) → 401 + Location
│   │       └── NextRefresh not yet passed → continue
│   └── NO → read session cookie (existing behavior)
│       ├── cookie absent or invalid
│       │   ├── no_redirect: true  → 401 + Location
│       │   └── no_redirect: false → 302 redirect to OIDC login (unchanged)
│       └── cookie valid; NextRefresh passed?
│           ├── OIDC provider → 302 to refresh endpoint (unchanged)
│           ├── GitHub or Google + no_redirect: true  → 401 + Location
│           └── GitHub or Google + no_redirect: false → 302 to login (unchanged)
```

### Header precedence

When both `X-Zrok-Session` header and a session cookie are present, the header
takes precedence and the cookie is ignored.

### `no_redirect` behavior

`no_redirect` only affects the cookie path. When a request carries
`X-Zrok-Session`, the response is always `401` on failure — regardless of
`no_redirect` — because an API client cannot usefully follow a browser
redirect.

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

```
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

```
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

## Per-share configuration

`no_redirect` is a per-share setting stored in the share's oauth config map
(not in the proxy-level config). It travels from the CLI through the full stack
to the OpenZiti proxy config where the proxy reads it at request time.

### CLI

```
zrok2 share public --oauth-provider oidc --oauth-no-redirect <target>
```

Default: `false` (existing redirect behavior is preserved).

### Agent API (`agent.SharePublicRequest`)

```go
SharePublicRequest{
    OauthProvider:   "oidc",
    OauthNoRedirect: true,
    ...
}
```

### Wire format

The field flows through the stack as follows:

```
CLI flag --oauth-no-redirect
  → sdk.ShareRequest.OauthNoRedirect (bool)
  → REST body oauthNoRedirect (bool)
  → controller reads params.Body.OauthNoRedirect
  → sdk.OauthConfig{NoRedirect: true}
  → serialised as "no_redirect": true in the OpenZiti proxy config map
  → proxy reads oauthMap["no_redirect"].(bool) at request time
```

### Proxy config map (what the proxy reads)

```json
{
  "auth_scheme": "oauth",
  "oauth": {
    "provider": "oidc",
    "authorization_check_interval": "1h",
    "email_domains": ["*.example.com"],
    "no_redirect": true
  }
}
```

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
|---|---|
| `endpoints/proxyUi/token.go` | `WriteTokenDisplay(w, token, expiry)` — renders the token page |
| `endpoints/proxyUi/token.html` | HTML template: token text box, copy button, usage instructions, expiry |

`token.html` uses the same visual style as `template.html`: Poppins + JetBrains
Mono fonts from Google Fonts, dark purple (`#241775`) background, the zrok SVG
logo banner, and a centered white card for the content area.

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
    `proxyUi.WriteTokenDisplay(w, sTkn, expiry)` instead of `http.Redirect`.
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

- Reads `no_redirect` from `oauthMap["no_redirect"].(bool)`.
- Tries `GetSessionHeader` first; sets `fromHeader = true` if found.
- Falls back to `getSessionCookie` when no header is present.
- Calls the new `oauthUnauthorized` (or existing `oauthLoginRequired`)
  based on `fromHeader` and `noRedirect`.
- Passes `jwtString string`, `fromHeader bool`, `noRedirect bool` to
  `validateOAuthToken`.

**`validateOAuthToken`** — signature change from `cookie *http.Cookie` to
`jwtString string, fromHeader bool, noRedirect bool`:

| Condition | Behavior |
|---|---|
| JWT parse/validation failure + `fromHeader` | `oauthUnauthorized` (401) |
| JWT parse/validation failure + `!fromHeader && noRedirect` | `oauthUnauthorized` (401) |
| JWT parse/validation failure + `!fromHeader && !noRedirect` | `oauthLoginRequired` (302) |
| `NextRefresh` passed + `fromHeader` + OIDC | inline refresh → `X-Zrok-Session` response header → continue |
| `NextRefresh` passed + `fromHeader` + GitHub/Google | `oauthUnauthorized` (401) |
| `NextRefresh` passed + `!fromHeader` + OIDC | `oauthRefreshRequired` (302, unchanged) |
| `NextRefresh` passed + `!fromHeader` + GitHub/Google + `noRedirect` | `oauthUnauthorized` (401) |
| `NextRefresh` passed + `!fromHeader` + GitHub/Google + `!noRedirect` | `oauthLoginRequired` (302, unchanged) |

**`validateEmailDomain`** — signature change from `cookie *http.Cookie` to
`jwtString string`.

**New `oauthUnauthorized`** alongside existing `oauthLoginRequired`:

```go
func oauthUnauthorized(w http.ResponseWriter, cfg *oauthConfig,
    provider, target string, refreshInterval time.Duration) {

    loginURL := fmt.Sprintf(
        "%s/%s/login?targetHost=%s&refreshInterval=%s&return_token=true",
        cfg.EndpointUrl, provider,
        url.QueryEscape(target), refreshInterval.String(),
    )
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Location", loginURL)
    w.WriteHeader(http.StatusUnauthorized)
    fmt.Fprintf(w, `{"login_url":%q}`, loginURL)
}
```

---

### Cookie filtering — strip `X-Zrok-Session` from proxied requests

**`dynamicProxy/cookies.go`** — `filterSessionCookies()`:
add `endpoints.StripSessionHeader(r)`.

**`publicProxy`** — add `endpoints.StripSessionHeader(r)` at the same call
site where session cookies are currently filtered.

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
|---|---|
| `endpoints/oauthCookies.go` | Add `SessionHeaderName`, `GetSessionHeader`, `StripSessionHeader` |
| `endpoints/proxyUi/token.go` | **New** — `WriteTokenDisplay` function |
| `endpoints/proxyUi/token.html` | **New** — token display page template (styled to match `template.html`) |
| `endpoints/dynamicProxy/auth.go` | Add `ReturnToken` to `IntermediateJWT`; add `sessionRefresher` interface |
| `endpoints/dynamicProxy/authOauth.go` | Rewrite `handleOAuth`, `validateOAuthToken`, `validateEmailDomain`; add `oauthUnauthorized`, `tryInlineRefresh` |
| `endpoints/dynamicProxy/authOauthRouter.go` | Add `GetProvider()` method |
| `endpoints/dynamicProxy/cookies.go` | Add `buildSessionJWT`; update `setSessionCookie` to return JWT string; update `filterSessionCookies` to strip header |
| `endpoints/dynamicProxy/http.go` | **Bug fix** — add `filterSessionCookies` call on OAuth auth-success path |
| `endpoints/dynamicProxy/providerOidc.go` | Add `ReturnToken` handling; add `RefreshSessionJWT` |
| `endpoints/dynamicProxy/providerGithub.go` | Add `ReturnToken` handling in `authHandler` / callback |
| `endpoints/dynamicProxy/providerGoogle.go` | Add `ReturnToken` handling in `authHandler` / callback |
| `endpoints/publicProxy/auth.go` | Add `ReturnToken` to `IntermediateJWT`; add `sessionRefresher` interface |
| `endpoints/publicProxy/authOAuth.go` | Same changes as `dynamicProxy/authOauth.go` |
| `endpoints/publicProxy/cookies.go` | Add `buildSessionJWT`; update `setSessionCookie` to return JWT string; strip `X-Zrok-Session` header in `filterSessionCookies` |
| `endpoints/publicProxy/http.go` | **Bug fix** — add `filterSessionCookies` call on OAuth auth-success path |
| `endpoints/publicProxy/oidcRegistry.go` | **New** — package-level `oidcProviderRegistry` map for inline refresh lookups |
| `endpoints/publicProxy/providerOidc.go` | Add `ReturnToken` handling; add `RefreshSessionJWT`; register in `oidcProviderRegistry` |
| `endpoints/publicProxy/providerGithub.go` | Add `ReturnToken` handling |
| `endpoints/publicProxy/providerGoogle.go` | Add `ReturnToken` handling |
| `sdk/golang/sdk/config.go` | Add `NoRedirect bool` to `OauthConfig`; add `"no_redirect"` case in `OauthConfigFromMap` |
| `sdk/golang/sdk/model.go` | Add `OauthNoRedirect bool` to `ShareRequest` |
| `sdk/golang/sdk/share.go` | Pass `OauthNoRedirect` to REST body in `newPublicShare` |
| `specs/src/definitions.yml` | Add `oauthNoRedirect: boolean` to `shareRequest` definition |
| `rest_model_zrok/share_request.go` | Add `OauthNoRedirect bool` field |
| `controller/share.go` | Read `OauthNoRedirect` in both create-share and update-share paths |
| `cmd/zrok2/sharePublic.go` | Add `--oauth-no-redirect` flag; pass through local SDK and agent gRPC paths |
| `agent/share.go` | Add `OauthNoRedirect bool` to `SharePublicRequest` |
| `agent/commandBuilder.go` | Add `OauthNoRedirect()` builder method |
| `agent/sharePublic.go` | Wire through command builder and gRPC→agent mapping |
| `agent/agentGrpc/agent.proto` | Add `bool oauthNoRedirect = 11` to `SharePublicRequest` |
| `agent/agentGrpc/agent.pb.go` | Regenerated — field 11 present with updated wire descriptor |

---

## Tests implemented

### `endpoints/oauthCookies_test.go` — extended

| Test | Behavior |
|---|---|
| `TestGetSessionHeaderPresent` | Header present → returns value |
| `TestGetSessionHeaderAbsent` | No header → returns `""` |
| `TestStripSessionHeaderRemoves` | Header present → removed from request |
| `TestStripSessionHeaderNoOp` | No header → no panic |

---

### `endpoints/dynamicProxy/http_test.go` — extended

A `mintSessionJWT` helper signs a `zrokClaims` JWT with a caller-supplied
`[]byte` key. The key must match the `signingKey` argument passed to
`shareHandler`.

New standalone tests:

| Test | Behavior |
|---|---|
| `TestShareHandlerGetWithValidHeaderBypassesRedirect` | Valid `X-Zrok-Session` JWT → auth passes, request reaches upstream |
| `TestShareHandlerXZrokSessionStrippedBeforeProxy` | Valid header + auth passes → `X-Zrok-Session` absent in proxied request |
| `TestShareHandlerNoRedirectReturns401` | No session, `no_redirect: true` → `401`, `application/json`, `Location` header, `{"login_url":...}` body |
| `TestShareHandlerOptionsStripsSessionHeaderBeforeProxy` | OPTIONS with `X-Zrok-Session` → header stripped before upstream, no auth challenge |

---

### `endpoints/publicProxy/http_test.go` — extended

Mirror of the `dynamicProxy` tests above, adjusted for `publicProxy` types
(`Config`, `OauthConfig`, `stubPublicProxyService`). Same four tests, same
`mintSessionJWT` helper.

---

### `endpoints/proxyUi/token_test.go` — new file

| Test | Behavior |
|---|---|
| `TestWriteTokenDisplayStatusOK` | Returns HTTP 200 |
| `TestWriteTokenDisplayContainsToken` | Token value appears in response body |
| `TestWriteTokenDisplayEscapesHTML` | Token containing `<script>` is HTML-escaped (XSS) |
| `TestWriteTokenDisplayContainsHeaderName` | Body references `X-Zrok-Session` |

---

### `endpoints/dynamicProxy/cookies_test.go` — new file

| Test | Behavior |
|---|---|
| `TestBuildSessionJWTRejectsEmptyTargetHost` | Empty `targetHost` → error returned |
| `TestBuildSessionJWTRoundTrip` | Valid inputs → non-empty signed JWT string returned |
| `TestBuildSessionJWTStripsPathFromTargetHost` | `host/path` input → path component stripped from `TargetHost` claim |

---

### Test file summary

| File | New / Extended | Covers |
|---|---|---|
| `endpoints/oauthCookies_test.go` | Extended | `GetSessionHeader`, `StripSessionHeader` |
| `endpoints/dynamicProxy/http_test.go` | Extended | Header auth, 401 path, header stripping before proxy, OPTIONS + header stripping |
| `endpoints/publicProxy/http_test.go` | Extended | Same for `publicProxy` |
| `endpoints/proxyUi/token_test.go` | **New** | `WriteTokenDisplay` rendering and XSS escaping |
| `endpoints/dynamicProxy/cookies_test.go` | **New** | `buildSessionJWT` error and round-trip cases |

---

## Out of scope

- The `refreshHandler()` HTTP endpoint (used by browser-based OIDC refresh
  redirects) is not modified. Header-based refresh is handled entirely inline
  in `validateOAuthToken`; the browser redirect path is unchanged.
- No changes to GitHub or Google provider refresh behavior (neither supports
  token refresh; clients must re-authenticate when `NextRefresh` passes).
- `no_redirect` is not currently exposed in the zrok web console UI; it is
  CLI/API only. It should be added to the console for consistency with the other
  oauth settings (`email_domains`, `authorization_check_interval`), ideally in
  an "advanced" collapsible section to avoid cluttering the default share
  creation form.
