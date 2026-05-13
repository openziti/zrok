# zrok2 Node.js SDK

The Node.js SDK for [zrok](https://zrok.io) -- zero-trust peer-to-peer application sharing built on [OpenZiti](https://openziti.io).

## Installation

```bash
npm install @openziti/zrok2
```

## Prerequisites

You need a zrok2 environment enabled before using the SDK:

```bash
zrok2 enable <your-account-token>
```

This creates `~/.zrok2/` with your environment configuration and identity files.

## Quick Start

### Share a service

```typescript
import {
    loadRoot,
    createShare,
    deleteShare,
    ShareRequest,
    PUBLIC_SHARE_MODE,
    PROXY_BACKEND_MODE,
} from "@openziti/zrok2";

const root = loadRoot();
const request = new ShareRequest(PUBLIC_SHARE_MODE, PROXY_BACKEND_MODE, "http://localhost:3000");
const share = await createShare(root, request);

console.log("shared at:", share.frontendEndpoints);

// clean up when done
await deleteShare(root, share);
```

### Auto-cleanup with wrappers

```typescript
import { loadRoot, withShare, ShareRequest, PUBLIC_SHARE_MODE, PROXY_BACKEND_MODE } from "@openziti/zrok2";

const root = loadRoot();
const request = new ShareRequest(PUBLIC_SHARE_MODE, PROXY_BACKEND_MODE, "http://localhost:3000");

// share is automatically deleted when the callback completes (or throws)
await withShare(root, request, async (share) => {
    console.log("shared at:", share.frontendEndpoints);
    // ... do work ...
});
```

### Check environment status

```typescript
import { loadRoot, status } from "@openziti/zrok2";

const root = loadRoot();
const s = status(root);
console.log("enabled:", s.enabled);
console.log("api endpoint:", s.apiEndpoint);
console.log("token:", s.token);
```

## API Reference

### Environment

| Function | Description |
|---|---|
| `loadRoot()` | Load the zrok2 environment from `~/.zrok2/`. Returns a `Root`. |
| `defaultRoot()` | Create a default (disabled) `Root` without reading from disk. |
| `rootExists()` | Check if `~/.zrok2/metadata.json` exists. |
| `assertRoot()` | Check if the root exists and has a compatible environment version. |
| `status(root)` | Return a `Status` with the current environment state. No API call. |

### Enable / Disable

| Function | Description |
|---|---|
| `enable(root, token, description?, host?)` | Enable an environment from an account token. Idempotent. |
| `disable(root)` | Disable the current environment. No-op if not enabled. |

### Shares

| Function | Description |
|---|---|
| `createShare(root, request)` | Create a share. Returns a `Share` with `shareToken` and `frontendEndpoints`. |
| `deleteShare(root, share)` | Delete a share. |
| `releaseReservedShare(root, share)` | Release a reserved share. Calls the same API as `deleteShare`. |
| `modifyShare(root, shareToken, addAccessGrants?, removeAccessGrants?)` | Update access grants on a share. |
| `getShareDetail(root, shareToken)` | Get detailed metadata for a share. Returns a `ShareDetail`. |
| `listShares(root, filters?)` | List shares with optional filters. Returns `ShareDetail[]`. |

### Access

| Function | Description |
|---|---|
| `createAccess(root, request)` | Create access to a share. Returns an `Access` with `frontendToken`. |
| `deleteAccess(root, access)` | Delete an access. |
| `listAccesses(root, filters?)` | List accesses with optional filters. Returns `AccessDetail[]`. |

### Names and Namespaces

| Function | Description |
|---|---|
| `createName(root, name, namespaceToken?)` | Create a name in a namespace. |
| `deleteName(root, name, namespaceToken?)` | Delete a name. |
| `listNames(root, namespaceToken?)` | List names, optionally filtered by namespace. |
| `listNamespaces(root)` | List all available namespaces. |

### Overview

| Function | Description |
|---|---|
| `getOverview(root)` | Get the full account overview (environments, shares, namespaces). |

### Convenience Wrappers

| Function | Description |
|---|---|
| `withShare(root, request, fn)` | Create a share, call `fn(share)`, delete on completion. Skips delete if `request.reserved` is true. |
| `withAccess(root, request, fn)` | Create access, call `fn(access)`, delete on completion. |
| `ProxyShare.create(root, target, options?)` | Create a managed proxy share with optional `uniqueName`, `frontends`, `shareMode`, and `verifySsl`. |

### OpenZiti Data Plane

These functions use the native `@openziti/ziti-sdk-nodejs` module to send traffic over the OpenZiti overlay network directly from Node.js:

| Function | Description |
|---|---|
| `init(root)` | Initialize the Ziti SDK with the environment identity. |
| `setLogLevel(level)` | Set the Ziti SDK log level. |
| `listener(share, onConnect, onData?, onListen?, onClient?)` | Listen for connections on a share. |
| `dialer(access, onConnect, onData)` | Dial a share to send/receive data. |
| `write(conn, buf, callback?)` | Write data to a Ziti connection. |
| `express(share)` | Create an Express app bound to a share's Ziti service. |

### Share Modes

| Constant | Value | Description |
|---|---|---|
| `PUBLIC_SHARE_MODE` | `"public"` | Publicly accessible via frontend endpoints |
| `PRIVATE_SHARE_MODE` | `"private"` | Accessible only via `createAccess` |

### Backend Modes

| Constant | Value |
|---|---|
| `PROXY_BACKEND_MODE` | `"proxy"` |
| `WEB_BACKEND_MODE` | `"web"` |
| `TCP_TUNNEL_BACKEND_MODE` | `"tcpTunnel"` |
| `UDP_TUNNEL_BACKEND_MODE` | `"udpTunnel"` |
| `CADDY_BACKEND_MODE` | `"caddy"` |
| `DRIVE_BACKEND_MODE` | `"drive"` |
| `SOCKS_BACKEND_MODE` | `"socks"` |

### Permission Modes

| Constant | Value |
|---|---|
| `OPEN_PERMISSION_MODE` | `"open"` |
| `CLOSED_PERMISSION_MODE` | `"closed"` |

## Examples

Two working examples are in `examples/`.

### Prerequisites

```bash
# build the SDK first
cd sdk && npm install && npm run build && cd ..
```

Node.js 20 or 22 is required for the native OpenZiti module.

### http-server

A public HTTP server shared via zrok2:

```bash
cd examples/http-server
npm install
npm run build
node dist/index.js http-server
```

This creates a public proxy share, starts an Express server on the Ziti overlay, and prints the frontend endpoint. Visit `http://<endpoint>:<port>/` to see "hello, world!".

### pastebin

A private peer-to-peer text transfer over a TCP tunnel:

```bash
cd examples/pastebin
npm install
npm run build

# terminal 1: serve text
node dist/index.js copyto
# enter text when prompted, then note the share token

# terminal 2: receive text
node dist/index.js pastefrom <shareToken>
```

## Development

### Build

```bash
cd sdk
npm install
npm run build
```

### Test

```bash
# unit tests (no network required)
npm test

# watch mode
npm run test:watch
```

### Integration Tests

Integration tests require a live zrok2 environment:

```bash
ZROK2_API_ENDPOINT=http://localhost:18080 \
ZROK2_ADMIN_TOKEN=<admin-secret> \
npm test -- test/integration/
```
