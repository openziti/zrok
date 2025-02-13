# Public/Private Sharing

In `v0.3` new sharing modes and new types of built-in services were introduced.

## Share Modes

_Note: In `v0.3`, the `tunnel` and `untunnel` concepts get renamed to `share` and `unshare`._

_Note: We're going to continue using `frontend` and `backend` as concepts, even though those words will be changing in the `zrok` CLI. A `frontend` will continue to describe an "ingress" into the `zrok`service, and is the tool that is used by the user "consuming" or `access`-ing the the `zrok` service. A `backend` will continue to describe the "binding" created by a user that wants to `share` a resource._

### Public Sharing

In `v0.2`, `zrok` only offered a "public" sharing mode. The public sharing mode will allow any configured `frontend` instances to send traffic to any `backend`. The policy and permission model was very simple and flat. A `v0.2` deployment considers any available `frontend` instance to be allowed to send traffic to configured services. The access for `frontend` instances is controlled by identity provisioning within the underlying OpenZiti network.

In `v0.3`, `zrok` will offer both a "public" and a "private" sharing mode. When `v0.3` configures the policies for a service, a publicly-shared service will have policies created that allow whichever selected public `frontend` instances to access the shared `backend`. A `v0.3` deployment will have a collection of multi-tenant, high-capacity `frontend` instances available to be selected from. The `zrok` CLI will default to selecting the `public` `frontend` instances.

The `frontend` selection approach also gives us a clean implementation for picking public `frontend` instances based on geography (either network or physical). The production `zrok.io` service could easily offer multiple different fleets of `frontend` instances, and this mechanism will allow `backend` users to choose where they want to offer access to their service.

### Private Sharing

`v0.3` introduced "private" sharing mode. When provisioning a service for private sharing, `zrok` will not create any policies for the service, until a request for a `frontend` binding is created for the service (through the `v0.3` `zrok access` command).

The `v0.3` `zrok` API will support creating `frontend` instances for both identified users (where the `zrok` user has a provisioned `environment`), as well as ephemeral users (the `zrok` controller will create a single-use "ephemeral environment" for these `frontend` instances).

## Backend Modes

In `v0.2`, the only possible `backend` "mode" was used for reverse proxying HTTP traffic to a local endpoint. The `v0.3` `zrok` client will support several different `backend` modes, providing a number of built-in conveniences.

### Web Mode

A user has a collection of files on disk. Sharing with a `backend` mode of "web", will create a `backend` that shares a file tree as if it were a local web server. This effectively allows a user to bind a web-server backend to a document root with a single CLI command.

### DAV Mode

A user wants to operate a read/write repository of files accessible through either conventional WebDAV clients (through `public` `frontend` instances), or through the `zrok` CLI (a convenience wrapper, embedding WebDAV capabilities).

This allows users to create read/write repositories of files that can be shared with multiple users, and also allows for the creation of write-only "drop boxes" for receiving files from another user (often a tricky thing to do well and securely on the public internet).

### Proxy Mode

`v0.3` will retain the classic reverse proxy mode, as well. Will continue to allow a user to expose a local HTTP endpoint through `zrok`.

## Entities (SQL)

`zrok` v0.3 introduced a new `frontends` table to allow the `zrok` controller to track the frontend instances that are available to any account or environment.

The following illustration shows the possibilities available.

![Frontend Selection](../../images/zrok_frontends_v0.3.png)

The `*.in.zrok.io` frontend is a "public" frontend, available to all `zrok` users. Most `zrok` installations will want to have at least one public, global frontend for all public, internet-facing ingress traffic for private backend instances. In the underlying data store, the public frontend will have a `name` set to `public` (or some other representative name), allowing users to reference that `frontend` using a friendly label.

The other two "private" frontends are configured with no `name` label (the lack of a `name` label signifies that these are "private" frontends). The ephemeral environment is allocated when a `zrok` frontend request is made without an account on behalf of a private share.
