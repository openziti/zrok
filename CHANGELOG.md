# CHANGELOG

## v0.4.30

FIX: Fix to the Node.js release process to properly support releasing on a tag.

## v0.4.29

FIX: Backed out an incorrect change to support a FreeBSD port in progress.

## v0.4.28

FEATURE: Node.js support for the zrok SDK (https://github.com/openziti/zrok/issues/400)

FEATURE: A Docker Compose project for self-hosting a zrok instance and [accompanying Docker guide](https://docs.zrok.io/docs/guides/self-hosting/docker) for more information.

CHANGE: the container images run as "ziggy" (UID 2171) instead of the generic restricted user "nobody" (UID 65534). This reduces the risk of unexpected file permissions when binding the Docker host's filesystem to a zrok container.

CHANGE: the Docker sharing guides were simplified and expanded

## v0.4.27

FEATURE: New `vpn` backend mode. Use `sudo zrok share private --backend-mode vpn` on the _VPN server_ host, then `sudo zrok access private <token>` on _VPN client_ machine. Works with reserved shares using `zrok reserve private --backend-mode vpn`. Use `<target>` parameter to override default VPN network settings `zrok share private -b vpn 192.168.255.42/24` -- server IP is `192.168.255.42` and VPN netmask will be `192.168.255.0/24`. Client IPs are assigned automatically from netmask range.

CHANGE: Update to OpenZiti SDK (`github.com/openziti/sdk-golang`) at `v0.23.22`.

CHANGE: Added indexes to `environments`, `shares`, and `frontends` tables to improve overall query performance on both PostgreSQL and Sqlite.

FIX: Also update the Python SDK to include the permission mode and access grants fields on the `ShareRequest` (https://github.com/openziti/zrok/issues/432)

FIX: Add a way to find the username on Linux when /etc/passwd and stdlib can't resolve the UID (https://github.com/openziti/zrok/issues/454)

## v0.4.26

FEATURE: New _permission modes_ available for shares. _Open permission mode_ retains the behavior of previous zrok releases and is the default setting. _Closed permission mode_ (`--closed`) only allows a share to be accessed (`zrok access`) by users who have been granted access with the `--access-grant` flag. See the documentation at (https://docs.zrok.io/docs/guides/permission-modes/) (https://github.com/openziti/zrok/issues/432)

CHANGE: The target for a `socks` share is automatically set to `socks` to improve web console display.

CHANGE: Enhancements to the look and feel of the account actions tab in the web console. Textual improvements.

FIX: The regenerate account token dialog incorrectly specified the path `${HOME}/.zrok/environments.yml`. This, was corrected to be `${HOME}/.zrok/environments.json`.

FIX: Align zrok frontdoor examples and Linux package (`zrok-share`) with the new OAuth email flag `--oauth-email-address-patterns` introduced in v0.4.25.

FIX: Reloading the web console when logged in no longer provokes the user to the login page.

## v0.4.25

FEATURE: New action in the web console that allows changing the password of the logged-in account (https://github.com/openziti/zrok/issues/148)

FEATURE: The web console now supports revoking your current account token and generating a new one (https://github.com/openziti/zrok/issues/191)

CHANGE: When specifying OAuth configuration for public shares from the `zrok share public` or `zrok reserve` public commands, the flags and functionality for restricting the allowed email addresses of the authenticating users has changed. The old flag was `--oauth-email-domains`, which took a string value that needed to be contained in the user's email address. The new flag is `--oauth-email-address-patterns`, which accepts a glob-style filter, using https://github.com/gobwas/glob (https://github.com/openziti/zrok/issues/413)

CHANGE: Creating a reserved share checks for token collision and returns a more appropriate error message (https://github.com/openziti/zrok/issues/531)

CHANGE: Update UI to add a 'true' value on `reserved` boolean (https://github.com/openziti/zrok/issues/443)

CHANGE: OpenZiti SDK (github.com/openziti/sdk-golang) updated to version `v0.22.29`, which introduces changes to OpenZiti API session handling

FIX: Fixed bug where a second password reset request would for any account would fail (https://github.com/openziti/zrok/issues/452)

## v0.4.24

FEATURE: New `socks` backend mode for use with private sharing. Use `zrok share private --backend-mode socks` and then `zrok access private` that share from somewhere else... very lightweight VPN-like functionality (https://github.com/openziti/zrok/issues/558)

FEATURE: New `zrok admin create account` command that allows populating accounts directly into the underlying controller database (https://github.com/openziti/zrok/issues/551)

CHANGE: The `zrok test loopback public` utility to report non-`200` errors and also ensure that the listening side of the test is fully established before starting loopback testing.

CHANGE: The OpenZiti SDK for golang (https://github.com/openziti/sdk-golang) has been updated to version `v0.22.28`

## v0.4.23

FEATURE: New CLI commands have been implemented for working with the `drive` share backend mode (part of the "zrok Drives" functionality). These commands include `zrok cp`, `zrok mkdir` `zrok mv`, `zrok ls`, and `zrok rm`. These are initial, minimal versions of these commands and very likely contain bugs and ergonomic annoyances. There is a guide available at (`docs/guides/drives/cli.md`) that explains how to work with these tools in detail (https://github.com/openziti/zrok/issues/438)

FEATURE: Python SDK now has a decorator for integrating with various server side frameworks. See the `http-server` example.

FEATURE: Python SDK share and access handling now supports context management.

FEATURE: TLS for `zrok` controller and frontends. Add the `tls:` stanza to your controller configuration (see `etc/ctrl.yml`) to enable TLS support for the controller API. Add the `tls:` stanza to your frontend configuration (see `etc/frontend.yml`) to enable TLS support for frontends (be sure to check your `public` frontend template) (#24)(https://github.com/openziti/zrok/issues/24)

CHANGE: Improved OpenZiti resource cleanup resilience. Previous resource cleanup would stop when an error was encountered at any stage of the cleanup process (serps, sps, config, service). New cleanup implementation logs errors but continues to clean up anything that it can (https://github.com/openziti/zrok/issues/533)

CHANGE: Instead of setting the `ListenOptions.MaxConnections` property to `64`, use the default value of `3`. This property actually controls the number of terminators created on the underlying OpenZiti network. This property is actually getting renamed to `ListenOptions.MaxTerminators` in an upcoming release of `github.com/openziti/sdk-golang` (https://github.com/openziti/zrok/issues/535)

CHANGE: Versioning for the Python SDK has been updated to use versioneer for management.

CHANGE: Python SDK package name has been renamed to `zrok`, dropping the `-sdk` postfix. [pypi](https://pypi.org/project/zrok).

## v0.4.22

FIX: The goreleaser action is not updated to work with the latest golang build. Modifed `go.mod` to comply with what goreleaser expects

## v0.4.21

FEATURE: The web console now supports deleting `zrok access` frontends (https://github.com/openziti/zrok/issues/504)

CHANGE: The web console now displays the frontend token as the label for any `zrok access` frontends throughout the user interface (https://github.com/openziti/zrok/issues/504)

CHANGE: Updated `github.com/rubenv/sql-migrate` to `v1.6.0`

CHANGE: Updated `github.com/openziti/sdk-golang` to `v0.22.6`

FIX: The migration `sqlite3/015_v0_4_19_share_unique_name_constraint.sql` has been adjusted to delete the old `shares_old` table as the last step of the migration process. Not sure exactly why, but SQLite is unhappy otherwise (https://github.com/openziti/zrok/issues/504)

FIX: Email addresses have been made case-insensitive. Please note that there is a migration included in this release (`016_v0_4_21_lowercase_email.sql`) which will attempt to ensure that all email addresses in your existing database are stored in lowercase; **if this migration fails you will need to manually remediate the duplicate account entries** (https://github.com/openziti/zrok/issues/517)

FIX: Stop sending authentication cookies to non-authenticated shares (https://github.com/openziti/zrok/issues/512)

## v0.4.20

CHANGE: OpenZiti SDK updated to `v0.21.2`. All `ziti.ListenOptions` listener options configured to use `WaitForNEstablishedListeners: 1`. When a `zrok share` client or an `sdk.Share` client are connected to an OpenZiti router that supports "listener established" events, then listen calls will not return until the listener is fully established on the OpenZiti network. Previously a `zrok share` client could report that it is fully operational and listening before the listener is fully established on the OpenZiti network; in practice this produced a very small window of time when the share would not be ready to accept requests. This change eliminates this window of time (https://github.com/openziti/zrok/issues/490)

FIX: Require the JWT in a zrok OAuth cookie to have an audience claim that matches the public share hostname. This prevents a cookie from one share from being use to log in to another share.

## v0.4.19

FEATURE: Reserved shares now support unique names ("vanity tokens"). This allows for the creation of reserved shares with identifiable names rather than generated share tokens. Includes basic support for profanity checking (https://github.com/openziti/zrok/issues/401)

CHANGE: The `publicProxy` endpoint implementation used in the `zrok access public` frontend has been updated to use the new `RefreshService(serviceName)` call instead of `RefreshServices()`. This should greatly improve the performance of requests against missing or non-responsive zrok shares (https://github.com/openziti/zrok/issues/487)

CHANGE: The Python SDK has been updated to properly support the "reserved" flag on the `ShareRequest` passed to `CreateShare`

CHANGE: Dependency updates; `github.com/openziti/sdk-golang@v0.20.145`; `github.com/caddyserver/caddy/v2@2.7.6`; indirect dependencies

## v0.4.18

FEATURE: Python SDK added. Can be found on [pypi](https://test.pypi.org/project/zrok-sdk). `pastebin` example illustrates basic SDK usage (see `sdk/python/examples/README.md` for details) (https://github.com/openziti/zrok/issues/401)

CHANGE: Moved the golang zrok sdk into `sdk/golang/sdk` to normalize location for future SDK's. 

CHANGE: add restart policies to docker compose samples used by the guide docs, e.g., reserved public share should auto-start on boot, temp public share should not.

## v0.4.17

CHANGE: Replaced most in-line shell scripts in Docker Compose projects with installed scripts that are shared between the Docker and Linux service. This normalizes the operational configuration of both Docker shares and Linux service, i.e., to use the same env vars.

CHANGE: Upgrade to Docusaurus v3 for documentation.

FIX: Some Docker shares had broken env mountpoints

## v0.4.16

FEATURE: Publish Linux packages for `zrok` CLI and a systemd service for running a reserved public share (`zrok-share`).

## v0.4.15

CHANGE: Updated the code signing and notarization process for macos binaries. The previous release process used the `gon` utility to handle both code signing and notarization. Apple changed the requirements and the `gon` utility no longer properly functions as of 2023-11-01. The `goreleaser` process has been adjusted to use the `notarytool` utility that ships with XCode to sign and notarize the binary (https://github.com/openziti/zrok/issues/435)

## v0.4.14

FEATURE: `zrok` Drives "Phase 1" (`p1`) functionality included in this release. This includes new `--backend-mode drive`, which accepts a folder path as a target. A `drive` share can be mounted as a network drive on Windows, macOS, and Linux, allowing full read/write access from all applications on those systems (https://github.com/openziti/zrok/issues/218) Subsequent releases will address CLI use cases and provide further refinements to the overall approach.

FEATURE: Docker Compose project for a reserved public share in docker/compose/zrok-public-reserved/compose.yml is described in the [public share guide](https://docs.zrok.io/docs/guides/docker-share/docker_public_share_guide/).

## v0.4.13

FIX: Update to Homebrew automation to properly integrate with the latest version of the Homebrew release process.

## v0.4.12

FIX: The `zrok reserve` command was not properly recording the reserved share status of the shares that it created, preventing the `zrok release` command from properly releasing them (https://github.com/openziti/zrok/issues/427) If a user encounters reserved shares that cannot be released with the `zrok release` command, they can be deleted through the web console.

## v0.4.11

FEATURE: The `zrok reserve` command now incorporates the `--json-output|-j` flag, which outputs the reservation details as JSON, rather than as human-consumable log messages. Other commands will produce similar output in the future (https://github.com/openziti/zrok/issues/422)

FIX: Include `--oauth-provider` and associated flags for the `zrok reserve` command, allowing reserved shares to specify OAuth authentication (https://github.com/openziti/zrok/issues/421)

## v0.4.10

CHANGE: The public frontend configuration has been bumped from `v: 2` to `v: 3`. The `redirect_host`, `redirect_port` and `redirect_http_only` parameters have been removed. These three configuration options have been replaced with `bind_address`, `redirect_url` and `cookie_domain`. See the OAuth configuration guide at `docs/guides/self-hosting/oauth/configuring-oauth.md` for more details (https://github.com/openziti/zrok/issues/411)

## v0.4.9

FIX: Remove extraneous share token prepended to OAuth frontend redirect.

## v0.4.8

FEATURE: The `sdk` package now includes a `sdk.Overview` function, which returns a complete description of the account attached to the enabled environment. Useful for inventorying the deployed shares and environments (https://github.com/openziti/zrok/issues/407)

CHANGE: The `zrok access public` frontend configuration format has changed and now requires that the configuration document include a `v: 2` declaration. This frontend configuration format is now versioned and when the code updates the configuration structure, you will receive an error message at startup, provoking you to look into updating your configuration (https://github.com/openziti/zrok/issues/406)

CHANGE: The title color of the header was changed from white to flourescent green, to better match the overall branding

CHANGE: Tweaks to build and release process for logging and deprecations. Pin golang version at 1.21.3+ and node version at 18.x across all platforms

CHANGE: Improvements to email invitation sent in response to `zrok invite` to correct broken links, some minor HTML issues and improve overall deliverability (https://github.com/openziti/zrok/issues/405)

CHANGE: Added warning message after `zrok invite` submit directing the user to check their "spam" folder if they do not receive the invite message.

## v0.4.7

FEATURE: OAuth authentication with the ability to restrict authenticated users to specified domains for `zrok share public`. Supports both Google and GitHub authentication in this version. More authentication providers, and extensibility to come in future `zrok` releases. See the OAuth configuration guide at `docs/guides/self-hosting/oauth/configuring-oauth.md` for details (https://github.com/openziti/zrok/issues/45, https://github.com/openziti/zrok/issues/404)

CHANGE: `--basic-auth` realm now presented as the share token rather than as `zrok` in `publicProxy` frontend implementation

## v0.4.6

FEATURE: New `--backend-mode caddy`, which pre-processes a `Caddyfile` allowing a `bind` statement to work like this: `bind {{ .ZrokBindAddress }}`. Allows development of complicated API gateways and multi-backend shares, while maintaining the simple, ephemeral sharing model provided by `zrok` (https://github.com/openziti/zrok/issues/391)

CHANGE: `--backend-mode web` has been refactored to utilize Caddy as the integrated web server. This provides for a much nicer web-based file browsing experience, while maintaining the existing web server facilities (https://github.com/openziti/zrok/issues/392)

CHANGE: Updated the golang version for release builds to `1.21.0` and the node version to `18.x`

CHANGE: Added `FrontendEndponts` to `sdk.Share`, returning selected frontend URLs to callers of `sdk.CreateShare`

CHANGE: Added a short alias `-b` for `--backend-mode` to improve CLI ergonomics (https://github.com/openziti/zrok/issues/397)

## v0.4.5

FEATURE: New health check endpoint (`/health`), which verifies that the underlying SQL store and metrics repository (InfluxDB, if configured) are operating correctly (https://github.com/openziti/zrok/issues/372)

CHANGE: Updated to golang v1.21.0 and node v18.x

FIX: `zrok admin bootstrap` and `zrok enable` both broken with latest OpenZiti releases (tested with `v0.30.0`); updated to latest OpenZiti golang SDK (https://github.com/openziti/zrok/issues/389)

## v0.4.4

FIX: `zrok status`, `zrok enable`, `zrok config`, etc. were all causing a panic when used on systems that had no previous `~/.zrok` directory (https://github.com/openziti/zrok/issues/383)

## v0.4.3

FEATURE: New `zrok overview` command, which returns all of the account details as a single JSON structure. See the OpenAPI spec at `specs/zrok.yml` for more details of the `/api/v1/overview` endpoint (https://github.com/openziti/zrok/issues/374)

FEATURE: New `zrok` SDK (https://github.com/openziti/zrok/issues/34). `pastebin` example illustrates basic SDK usage (see `sdk/examples/pastebin/README.md` for details) ((https://github.com/openziti/zrok/issues/379)

## v0.4.2

Some days are just like this. `v0.4.2` is a re-do of `v0.4.1`. Trying to get Homebrew working and had a bad release. Hopefully this is the one.

## v0.4.1

FEATURE: New `zrok console` command to open the currently configured web console in the local web browser (https://github.com/openziti/zrok/issues/170)

CHANGE: Further tweaks to the release process to automatically get the latest release into Homebrew (https://github.com/openziti/zrok/issues/264)

## v0.4.0

FEATURE: New `tcpTunnel` backend mode allowing for private sharing of local TCP sockets with other `zrok` users (https://github.com/openziti/zrok/issues/170)

FEATURE: New `udpTunnel` backend mode allowing for private sharing of local UDP sockets with other `zrok` users (https://github.com/openziti/zrok/issues/306)

FEATURE: New metrics infrastructure based on OpenZiti usage events (https://github.com/openziti/zrok/issues/128). See the [v0.4 Metrics Guide](docs/guides/metrics-and-limits/configuring-metrics.md) for more information.

FEATURE: New limits implementation based on the new metrics infrastructure (https://github.com/openziti/zrok/issues/235). See the [v0.4 Limits Guide](docs/guides/metrics-and-limits/configuring-limits.md) for more information.

FEATURE: The invite mechanism has been reworked to improve user experience. The configuration has been updated to include a new `invite` stanza, and now includes a boolean flag indicating whether or not the instance allows new invitations to be created, and also includes contact details for requesting a new invite. These values are used by the `zrok invite` command to provide a smoother end-user invite experience https://github.com/openziti/zrok/issues/229)

FEATURE: New password strength checking rules and configuration. See the example configuration file (`etc/ctrl.yml`) for details about how to configure the strength checking rules (https://github.com/openziti/zrok/issues/167)

FEATURE: A new `admin/profile_endpoint` configuration option is available to start a `net/http/pprof` listener. See `etc/ctrl.yml` for details.

CHANGE: The controller configuration version bumps from `v: 2` to `v: 3` to support all of the new `v0.4` functionality. See the [example ctrl.yml](etc/ctrl.yml) for details on the new configuration.

CHANGE: The underlying database store now utilizes a `deleted` flag on all tables to implement "soft deletes". This was necessary for the new metrics infrastructure, where we need to account for metrics data that arrived after the lifetime of a share or environment; and also we're going to need this for limits, where we need to see historical information about activity in the past (https://github.com/openziti/zrok/issues/262)

CHANGE: Updated to latest `github.com/openziti/sdk-golang` (https://github.com/openziti/zrok/issues/335)

FIX: `zrok share reserved --override-endpoint` now works correctly; `--override-endpoint` was being incorrectly ignore previously (https://github.com/openziti/zrok/pull/348)

## v0.3.7

FIX: Improved TUI word-wrapping (https://github.com/openziti/zrok/issues/180)

## v0.3.6

CHANGE: Additional change to support branch builds (for CI purposes) and additional containerization efforts around k8s.

## v0.3.5

CHANGE: `zrok config set apiEndpoint` now validates that the new API endpoint correctly starts with `http://` or `https://` (https://github.com/openziti/zrok/issues/258)

CHANGE: Additional linting to support homebrew (https://github.com/openziti/zrok/issues/264)

## v0.3.4

CHANGE: `zrok test endpoint` incorporates `--ziti` mode (and related flags) to allow direct endpoint listening on a Ziti service

CHANGE: `zrok test websocket` command to test websockets, whether over TCP or over Ziti

FIX: Websocket support now functional

## v0.3.3

CHANGE: `zrok test loop` has been moved to `zrok test loop public`, making way for additional types of loopback testing tools. The `zrok test endpoint` server now includes an `/echo` endpoint, which provides a simple echo websocket (https://github.com/openziti/zrok/issues/237)

## v0.3.2

FEATURE: New docker infrastructure, including `docker-compose.yml` examples (and documentation) illustrating how to deploy `zrok` in `docker`-based environments

CHANGE: Include missing `--headless` flag for `zrok enable` and `zrok access private` (https://github.com/openziti/zrok/issues/246)

CHANGE: Fix for `zrok enable` error path handling (https://github.com/openziti/zrok/issues/244)

FEATURE: `zrok controller validate` and `zrok access public validate` will both perform a quick syntax validation on controller and public frontend configuration documents (https://github.com/openziti/zrok/issues/238)

	$ zrok controller validate etc/dev.yml 
	[ERROR]: controller config validation failed (error loading controller config 'etc/dev.yml': field 'maintenance': field 'registration': field 'expiration_timeout': got [bool], expected [time.Duration])

CHANGE: `zrok status` no longer shows secrets (secret token, ziti identity) unless the `--secrets` flag is passed (https://github.com/openziti/zrok/issues/243)

## v0.3.1

CHANGE: Incorporate initial docker image build (https://github.com/openziti/zrok/issues/217)

CHANGE: Improve target URL parsing for `zrok share` when using `--backend-mode` proxy (https://github.com/openziti/zrok/issues/211)

	New and improved URL handling for proxy backends:

	9090 -> http://127.0.0.1:9090
	localhost:9090 -> http://127.0.0.1:9090
	https://localhost:9090 -> https://localhost:9090

CHANGE: Improve usability of `zrok invite` TUI in low-color environments (https://github.com/openziti/zrok/issues/206)

CHANGE: Better error responses when `zrok invite` fails due to missing token (https://github.com/openziti/zrok/issues/207)

## v0.3.0

CHANGE: Removed some minor web console lint and warnings (https://github.com/openziti/zrok/issues/205)

## v0.3.0-rc6

CHANGE: Better error message when `zrok admin create frontend` runs into a duplicate name collision (https://github.com/openziti/zrok/issues/168)

CHANGE: Gentler CLI error messages by default (https://github.com/openziti/zrok/issues/203)

CHANGE: Add favicon to web console (https://github.com/openziti/zrok/issues/198)

CHANGE: Add configurable "terms of use" link in the controller configuration, and optionally display the link on the login form and registration forms (https://github.com/openziti/zrok/issues/184)

CHANGE: Prevent multiple `zrok enable` commands from succeeding (https://github.com/openziti/zrok/issues/190)

CHANGE: New `--insecure` flag for `share <public|private|reserved>` commands (https://github.com/openziti/zrok/issues/195)

## v0.3.0-rc5

CHANGE: Improvements to controller log messages to assist in operations (https://github.com/openziti/zrok/issues/186)

CHANGE: `armv7` builds for Linux are now shipped with releases; these builds were tested against a Raspberry Pi 4 (https://github.com/openziti/zrok/issues/93)

CHANGE: `zrok config set` now includes a warning when the `apiEndpoint` config is changed and an environment is already enabled; the user will not see the change until `zrok disable` is run. The CLI now includes a `zrok config unset` command (https://github.com/openziti/zrok/issues/188)

## v0.3.0-rc4

CHANGE: Enable notarization for macos binaries (https://github.com/openziti/zrok/issues/92)

## v0.3.0-rc3

> This release increments the configuration version from `1` to `2`. See the note below.

CHANGE: The email "from" configuration moved from `registration/email_from` to `email/from`. **NOTE: This change increments the configuration `V` from `1` to `2`.**

CHANGE: Replaced un-salted sha512 password hashing with salted hashing based on Argon2 **NOTE: This version will _invalidate_ all account passwords, and will require all users to use the 'Forgot Password?' function to reset their password.** (https://github.com/openziti/zrok/issues/156)

CHANGE: Switched from `ubuntu-latest` (`22.04`) for the Linux builds to `ubuntu-20.04`. Should improve `glibc` compatibility with older Linux distributions (https://github.com/openziti/zrok/issues/179)

CHANGE: `zrok admin generate` now outputs the generated tokens to `stdout` after successfully provisioning the tokens (https://github.com/openziti/zrok/issues/181)

FIX: Fixed log message in `resetPasswordRequest.go` (https://github.com/openziti/zrok/issues/175)

FIX: Fixed `-v` (verbose mode) on in TUI-based `zrok share` and `zrok access` (https://github.com/openziti/zrok/issues/174)

## v0.3.0-rc2

FEATURE: Allow users to reset their password (https://github.com/openziti/zrok/issues/65)

CHANGE: Improved email styling for new user invite emails (https://github.com/openziti/zrok/issues/157)

CHANGE: Migrated from `openziti-test-kitchen` to `openziti` (https://github.com/openziti/zrok/issues/158).

CHANGE: Show a hint when `zrok invite` fails, indicating that the user should check to see if they need to be using the `--token` flag and token-based invites (https://github.com/openziti/zrok/issues/172).

FIX: Fixed PostgreSQL migration issue where sequences got reset and resulted in primary key collisions on a couple of tables (https://github.com/openziti/zrok/issues/160).

FIX: Remove `frontend` instances when `zrok disable`-ing an environment containing them (https://github.com/openziti/zrok/issues/171)

## v0.3.x Series

The `v0.2` series was a _proof-of-concept_ implementation for the overall `zrok` architecture and the concept.

`v0.3` is a massive elaboration of the concept, pivoting it from being a simple ephemeral reverse proxy solution, to being the beginnings of a comprehensive sharing platform, complete with public and private sharing (built on top of OpenZiti).

`v0.3.0` includes the minimal functionality required to produce an early, preview version of the elaborated `zrok` concept, suitable for both production use at `zrok.io`, and also suitable for private self-hosting.

From `v0.3.0` forward, we will begin tracking notable changes in this document.

## v0.2.18

* DEFECT: Token generation has been improved to use an alphabet consisting of `[a-zA-Z0-9]`. Service token generation continues to use a case-insensitive alphabet consisting of `[a-z0-9]` to be DNS-safe.
