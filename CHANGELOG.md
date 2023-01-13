# v0.3.0 (WiP)

## CLI/zrok Client Changes

### Versioning 

The `zrok` client now checks the version of the configured API endpoint before attempting to connect. This means a `zrok` client will only work with the same major/minor versions.

This means that a `v0.3` client will NOT work with a `v0.2` service. This also means that breaking API changes will require a minor revision change. A breaking change made in `v0.3` will provoke a new `v0.4` series to begin.

## API Changes

Naming has been streamlined:

* The `tunnel` operations are all tagged with `service`.
* `tunnel.Tunnel` becomes `service.Share`
* `tunnel.Untunnel` becomes `service.Unshare`
* `TunnelRequest` and `TunnelResponse` become `ShareRequest` and `ShareResponse` 
* `UntunnelRequest` becomes `UnshareRequest`.

Sharing now includes the new mode options:

* `ShareRequest` now includes a `ShareMode` enum which includes `public` and `private` values
* `ShareRequest` now includes a `BackendMode` enum which includes `proxy`, `web`, and `dav` values

## Frontend Selection; Private Shares

The `zrok` model has been extended to include support for both a "public share" (exposing a backend through the globally-available `frontend` instances), and also a "private share" (exposing a backend service to a user who instantiates a private, local `frontend`).

### Underlying Schema Changes

* Added new `frontends` table
* Added new `availability_type` enumeration for use in the new `frontends` table
* Made the `account_id` column of the `environments` table `NULL`-able; a `NULL` value in the `account_id` column signifies an "ephemeral" environment

## Loop Test Shutdown Hook

The `zrok test loop` command now includes a shutdown hook to allow premature cancellation of a running test.

# v0.2.18

* DEFECT: Token generation has been improved to use an alphabet consisting of `[a-zA-Z0-9]`. Service token generation continues to use a case-insensitive alphabet consisting of `[a-z0-9]` to be DNS-safe.
