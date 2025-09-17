# CHANGELOG

## v2.0.0

FEATURE: `zrok admin migrate` now supports a `--down <n>` flag, which allows for reverse-migration by a specified number of migrations

CHANGE: Completely overhauled the core ziti automation logic. The legacy `controller/zrokEdgeSdk` package has been replaced with a much more streamlined, clearer package `controller/automation`. This makes comprehending the controller code a lot simpler. (https://github.com/openziti/zrok/issues/1054)

CHANGE: Updated `github.com/openziti/sdk-golang` to `v1.2.4`.

## v1.1.4

CHANGE: Update `github.com/caddyserver/caddy/v2` to `v2.9.1`; fixes CVE-2024-53259 (would only potentially effect users using the QUIC protocol, very atypical) (https://github.com/openziti/zrok/issues/1047)

## v1.1.3

FEATURE: A new `compatibility` > `version_patterns` array is included in the controller configuration, allowing for dynamic adjustment of allowed client version strings (https://github.com/openziti/zrok/issues/1030)

FEATURE: A new `compatibility` > `log_version` boolean is included in the controller configuration. When this boolean is set to `true`, the controller will log all client versions provided for compatibility checking.

CHANGE: Update `github.com/openziti/sdk-golang` to `v1.2.3`

CHANGE: Minor vulnerability packages updated in `ui` and `agent/agentUi`

FIX: The `scope` field of the metrics returned from `/metrics/environment/...` is now properly set as `environment` and the from `/metrics/share/...` is now properly set as `share` (https://github.com/openziti/zrok/issues/1031)

## v1.1.2

FIX: A panic happened in the `publicProxy` implementation when no `oauth` config block is present (https://github.com/openziti/zrok/issues/1032)

## v1.1.1

FIX: Masquerade as `v1.0-v1.1.1 [gitHash]` when performing client version checks. Will be replaced with the usual client identifier in `v1.1.2` when the regular expressions for controlling client compatibility are externalized in the controller config (https://github.com/openziti/zrok/issues/1028)

## v1.1.0

FEATURE: Rewritten and improved `publicProxy` package (`zrok access public`), with support for extensible OAuth-compliant identity providers. The `publicProxy` configuration now supports any number of configured OAuth-compliant providers (rather than just a single `google` provider and/or a single `github` provider). Also includes a new OIDC-compliant generic IDP provider integration. Improvements to authentication flows and security all around. See the [updated guide](https://docs.zrok.io/docs/guides/self-hosting/oauth/configuring-oauth/) on using OAuth-based identity providers with the zrok public frontend (https://github.com/openziti/zrok/issues/968)

FEATURE: Templatized and improved static pages (not found/404, unauthorized/401, health check, etc.) used by the public frontend. Consolidated variable data using golang `text/template` so that static `proxyUi` package can display additional error information and provide extension points for replacing all of the templated content with external files. See the [error pages guide](https://docs.zrok.io/docs/guides/self-hosting/error-pages/) for more information on customizing the built-in template (https://github.com/openziti/zrok/issues/1012)

FEATURE: `zrok access private` now includes a `--template-path` allowing the embedded `proxyUi` template to be replaced with an external HTML file (https://github.com/openziti/zrok/issues/1012)

FIX: Invoking `/agent/*` endpoints to remotely manage agents with remoting was causing a new API session to be allocated in the ziti controller for each request. A slightly different strategy was employed for embedding the ziti SDK into the zrok controller that should mitigate this (https://github.com/openziti/zrok/issues/1023)

## v1.0.8

FEATURE: New opt-in configuration item `superNetwork` which enables multiple data plane connections to the OpenZiti underlay, a separate control plane connection, enabling SDK-based flow control. To opt-in use `zrok config set superNetwork true` in each environment, or set the `ZROK_SUPER_NETWORK` environment variable to `true` (https://github.com/openziti/zrok/issues/1010)

CHANGE: Updated `github.com/openziti/sdk-golang` to `v1.2.1` (https://github.com/openziti/zrok/issues/1010)

## v1.0.7

FEATURE: zrok Agent now supports health checks (against the target endpoint) for `proxy` backend shares using the `zrok agent share http-healthcheck` command. The zrok API now includes an `/agent/share/http-healthcheck` endpoint for remotely performing these checks against remoted Agents. See the guide for using the feature at https://docs.zrok.io/guides/agent/http-healthcheck/ (https://github.com/openziti/zrok/issues/1002)

FEATURE: `/overview`, `/detail/share`, `/detail/environment`, and `/overview/{organizationToken}/{accountEmail}` all adjusted to include `envZId` in share detail output (https://github.com/openziti/zrok/issues/998)

FEATURE: New add and delete API endpoints for frontend grants. New `zrok admin create frontend-grant` and `zrok admin delete frontend-grant` CLI for invoking these API endpoints from the command line (https://github.com/openziti/zrok/issues/992)

FEATURE: New admin endpoint for deleting accounts. New `zrok admin delete account` CLI for invoking the API endpoint from the command line (https://github.com/openziti/zrok/issues/993)

FEATURE: New admin endpoint for deleting identities. New `zrok admin delete identity` CLI for invoking the API endpoint from the command line (https://github.com/openziti/zrok/issues/800)

FEATURE: New API endpoint (`/overview/public-frontends`) that returns the public frontends available to authenticated account. The public frontends include those marked with the `open` permission mode, and those marked `closed` where the user has a frontend grant allowing them to access the frontend. New CLI command `zrok overview public-frontends` to allow end users to list the public frontends their account can use (https://github.com/openziti/zrok/issues/996)

CHANGE: Updated `openapi-generator-cli` from `7.12.0` to `7.14.0`

## v1.0.6

CHANGE: The `/overview` endpoint has been adjusted to include a new `remoteAgent` `boolean` on the `environment` instances, indicating whether or not the environment has an enrolled remote agent (https://github.com/openziti/zrok/issues/977)

CHANGE: Adjusted core framework entry points to support changing zrokdir, and host interrogation functions to better support embedded zrok functionality (https://github.com/openziti/zrok/issues/976)

## v1.0.5

FEATURE: Initial support for zrok Agent remoting; new `zrok agent enroll` and `zrok agent unenroll` commands that establish opt-in remote Agent management facilities on a per-environment basis. The central API has been augmented to allow for remote control (creating shares and private access instances) of these agents; see the [remoting guide](https://docs.zrok.io/docs/guides/agent/remoting) for details (https://github.com/openziti/zrok/issues/967)

CHANGE: `zrok share public`, `zrok share private`, and `zrok reserve` all default to the "closed" permission mode (they previously defaulted to the "open" permission mode). The `--closed` flag has been replaced with a new `--open` flag. See the [Permission Modes](https://docs.zrok.io/docs/guides/permission-modes/) docs for details (https://github.com/openziti/zrok/issues/971)

FIX: `zrok enable` now handles the case where the user ID does not resolve to a username when generating the default environment description (https://github.com/openziti/zrok/issues/959)

FIX: Linux packages were optimized to avoid manage file revision conflicts (https://github.com/openziti/zrok/issues/817)

## v1.0.4

FIX: `zrok admin bootstrap` and `zrok enable` functionality were broken in `v1.0.3`. A bad combination of dependencies caused issues with marshalling data from the associated controller endpoints

CHANGE: `github.com/openziti/sdk-golang` has been updated to `v1.1.0`, `github.com/openziti/ziti` has been updated to `v1.6.0`. Related dependencies and indirects also updated

CHANGE: Updated to `golang` `v1.24` as the official build toolchain

## v1.0.3

FEATURE: `zrok agent console` now outputs the URL it is attempting to open. New `zrok agent console --headless` option to only emit the agent console URL (https://github.com/openziti/zrok/issues/944)

FEATURE: New `zrok admin unbootstrap` to remove zrok resources from the underlying OpenZiti instance (https://github.com/openziti/zrok/issues/935)

FEATURE: New InfluxDB metrics capture infrastructure for `zrok test canary` framework (https://github.com/openziti/zrok/issues/948)

FEATURE: New `zrok test canary enabler` to validate `enable`/`disable` operations and gather performance metrics around how those paths are operating (https://github.com/openziti/zrok/issues/771)

FEATURE: New `zrok test canary` infrastructure capable of supporting more complex testing scenarios; now capable of streaming canary metrics into an InfluxDB repository; new programming framework for developing additional types of streaming canary metrics (https://github.com/openziti/zrok/issues/948 https://github.com/openziti/zrok/issues/954)

FEATURE: All `zrok test canary` commands that have "min" and "max" values (`--min-pacing` and `--max-pacing` for example) now include a singular version of that flag for setting both "min" and "max" to the same value (`--pacing` for example). The singular version of the flag always overrides any `--min-*` or `--max-*` values that might be set

CHANGE: New _guard_ to prevent users from running potentially dangerous `zrok test canary` commands inadvertently without understanding what they do (https://github.com/openziti/zrok/issues/947)

CHANGE: Updated `npm` dependencies for `ui`, `agent/agentUi` and `website`. Updated `github.com/openziti/sdk-golang` to `v0.24.0`

## v1.0.2

FEATURE: "Auto-rebase" for enabled environments where the `apiEndpoint` is set to `https://api.zrok.io`. This will automatically migrate existing environments to the new `apiEndpoint` for the `v1.0.x` series (https://github.com/openziti/zrok/issues/936)

FEATURE: New `admin/new_account_link` configuration option to allow the insertion of "how do I register for an account?" links into the login form (https://github.com/openziti/zrok/issues/552)

CHANGE: The release environment, share, and access modals in the API console now have a better message letting the user know they will still need to clean up their `zrok` processes (https://github.com/openziti/zrok/issues/910)

CHANGE: The openziti/zrok Docker image has been updated to use the latest version of the ziti CLI, 1.4.3 (https://github.com/openziti/zrok/pull/917)

## v1.0.1

FEATURE: The zrok Agent now persists private accesses and reserved shares between executions. Any `zrok access private` instances or `zrok share reserved` instances created using the agent are now persisted to a registry stored in `${HOME}/.zrok`. When restarting the agent these accesses and reserved shares are re-created from the data in this registry (https://github.com/openziti/zrok/pull/922)

FEATURE: zrok-agent Linux package runs the agent as a user service (https://github.com/openziti/zrok/issues/883)

CHANGE: Updated the "Getting Started" guide to be slightly more streamlined and reflect the `v1.0` changes (https://github.com/openziti/zrok/issues/877)

CHANGE: let the Docker instance set the Caddy HTTPS port (https://github.com/openziti/zrok/pull/920)

CHANGE: Add Traefik option for TLS termination in the Docker instance (https://github.com/openziti/zrok/issues/808)

## v1.0.0

MAJOR RELEASE: zrok reaches version 1.0.0!

FEATURE: Completely redesigned web interface ("API Console"). New implementation provides a dual-mode interface supporting an improved visual network navigator and also a "tabular" view, which provides a more traditional "data" view. New stack built using vite, React, and TypeScript (https://github.com/openziti/zrok/issues/724)

FEATURE: New "zrok Agent", a background manager process for your zrok environments, which allows you to easily manage and work with multiple `zrok share` and `zrok access` processes. New `--subordinate` flag added to `zrok share [public|private|reserved]` and `zrok access private` to operate in a mode that allows an Agent to manage shares and accesses (https://github.com/openziti/zrok/issues/463)

FEATURE: New "zrok Agent UI" a web-based user interface for the zrok Agent, which allows creating and releasing shares and accesses through a web browser. This is just an initial chunk of the new Agent UI, and is considered a "minimum viable" version of this interface (https://github.com/openziti/zrok/issues/221)

FEATURE: `zrok share [public|private|reserved]` and `zrok access private` now auto-detect if the zrok Agent is running in an environment and will automatically service share and access requests through the Agent, rather than in-process if the Agent is running. If the Agent is not running, operation remains as it was in `v0.4.x` and the share or access is handled in-process. New `--force-agent` and `--force-local` flags exist to skip Agent detection and manually select an operating mode (https://github.com/openziti/zrok/issues/751)

FEATURE: `zrok access private` supports a new `--auto` mode, which can automatically find an available open address/port to bind the frontend listener on. Also includes `--auto-address`, `--auto-start-port`, and `--auto-end-port` features with sensible defaults. Supported by both the agent and local operating modes (https://github.com/openziti/zrok/issues/780)

FEATURE: `zrok rebase` commands (`zrok rebase apiEndpoint` and `zrok rebase accountToken`) allows "rebasing" an enabled environment onto a different API endpoint or a different account token. This is useful for migrating already-enabled environments between endpoints supporting different zrok versions, and is also useful when regenerating an account token (https://github.com/openziti/zrok/issues/869, https://github.com/openziti/zrok/issues/897)

FEATURE: `zrok test canary` CLI tree replaces the old `zrok test loop` tree; new `zrok test canary public-proxy` and `zrok test canary private-proxy` provide modernized, updated versions of what the `zrok test loop` commands used to do. This new approach will serve as the foundation for all future zrok testing infrastructure (https://github.com/openziti/zrok/issues/771)

FEATURE: New `/api/v1/versions` endpoint to return comprehensive, full stack version information about the deployed service instance. Currently only returns a single `controllerVersion` property (https://github.com/openziti/zrok/issues/881) 

CHANGE: The default API URL for `v1.0.x` zrok clients is now `https://api-v1.zrok.io` (instead of the older `https://api.zrok.io`). The zrok.io deployment will now be maintaining version-specific DNS for versioned API endpoints.

CHANGE: Refactored API implementation. Cleanup, lint removal, additional data elements added, unused data removed (https://github.com/openziti/zrok/issues/834)

CHANGE: Deprecated the `passwords` configuration stanza. The zrok controller and API console now use a hard-coded set of (what we believe to be) reasonable assumptions about password quality (https://github.com/openziti/zrok/issues/834)

CHANGE: The protocol for determining valid client versions has been changed. Previously a zrok client would do a `GET` against the `/api/v1/version` endpoint and do a local version string comparison (as a normal precondition to any API call) to see if the controller version matched. The protocol has been amended so that any out-of-date client using the old protocol will receive a version string indicating that they need to uprade their client. New clients will do a `POST` against the `/api/v1/clientVersionCheck` endpoint, posting their client version, and the server will check for compatibility. Does not change the security posture in any significant way, but gives more flexibility on the server side for managing client compatibility. Provides a better, cleared out-of-date error message for old clients when accessing `v1.0.0`+ (https://github.com/openziti/zrok/issues/859)

CHANGE: The Node.js SDK is now generated by `openapi-generator` using the `typescript-fetch` template. Examples and SDK components updated to use the `v1.0.0` API and generated client (https://github.com/openziti/zrok/issues/893)

CHANGE: The Python SDK is now generated by `openapi-generator` and requires a newer `urllib3` version 2.1.0. The published Python module, `zrok`, inherits the dependencies of the generated packages (https://github.com/openziti/zrok/issues/894)

## v0.4.49

FIX: Release artifacts now include a reproducible source archive. The archive's download URL is now used by the Homebrew formula when building from source instead of the archive generated on-demand by GitHub (https://github.com/openziti/zrok/issues/858).

FIX: Pre-releases are no longer uploaded to the stable Linux package repo, and workflows that promote stable release artifacts to downstream distribution channels enforce semver stable release tags, i.e., not having a semver hyphenated prerelease suffix.

CHANGE: The release `checksums.txt` has been renamed `checksums.sha256.txt` to reflect the use of a collision-resistant algorithm instead of `shasum`'s default algorithm, SHA-1.

CHANGE: The dependency graph is now published as a release artifact named `sbom-{version}.spdx.json` (https://github.com/openziti/zrok/issues/888).

CHANGE: Pre-releases are uploaded to the pre-release Linux package repo and Docker Hub for testing. [RELEASING.md](./RELEASING.md) describes releaser steps and the events they trigger.

CHANGE: Linux release binaries are now built on the ziti-builder container image based on Ubuntu Focal 20.04 to preserve backward compatibility as the ubuntu-20.04 GitHub runner is end of life.

CHANGE: Container images now include SLSA and SBOM attestations, and these are also published to the Docker Hub registry (https://github.com/openziti/zrok/issues/890).

CHANGE: Release binary and text artifacts are now accompanied by provenance attestations (https://github.com/openziti/zrok/issues/889).

## v0.4.48

FEATURE: The controller configuration now supports a `disable_auto_migration` boolean in the `store` stanza. When set to `true`, the controller will not attempt to auto-migrate (or otherwise validate the migration state) of the underlying database. Leaving `disable_auto_migration` out, or setting it to false will retain the default behavior of auto-migrating when starting the zrok controller. The `zrok admin migrate` command will still perform a migration regardless of how this setting is configured in the controller configuration (https://github.com/openziti/zrok/issues/866)

FIX: the Python SDK erroneously assumed the enabled zrok environment contained a config.json file, and was changed to only load it if the file was present (https://github.com/openziti/zrok/pull/853/).

## v0.4.47

CHANGE: the Docker instance will wait for the ziti container healthy status (contribution from Ben Wong @bwong365 - https://github.com/openziti/zrok/pull/790)

CHANGE: Document solving the DNS propagation timeout for Docker instances that are using Caddy to manage the wildcard certificate.

CHANGE: Add usage hint in `zrok config get --help` to clarify how to list all valid `configName` and their current values by running `zrok status`.

CHANGE: The Python SDK's `Overview()` function was refactored as a class method (https://github.com/openziti/zrok/pull/846).

FEATURE: The Python SDK now includes a `ProxyShare` class providing an HTTP proxy for public and private shares and a
Jupyter notebook example (https://github.com/openziti/zrok/pull/847).

FIX: PyPi publishing was failing due to a CI issue (https://github.com/openziti/zrok/issues/849)

## v0.4.46

FEATURE: Linux service template for systemd user units (https://github.com/openziti/zrok/pull/818)

FIX: Docker share examples had incorrect default path for zrok environment mountpoint

FIX: Clarify how to use DNS providers like Route53 with the zrok Docker instance sample.

CHANGE: Use port 80 for the default Ziti API endpoint in the zrok Docker instance sample (https://github.com/openziti/zrok/issues/793).

CHANGE: Clarify OS requirements for zrok VPN

CHANGE: Set the Windows executable search path in the Windows install guide.

CHANGE: bump macOS runner for Python module from macos-12 to macos-13

## v0.4.45

FEATURE: Minimal support for "organizations". Site admin API endpoints provided to create, list, and delete "organizations". Site admin API endpoints provided to add, list, and remove "organization members" (zrok accounts) with the ability to mark accounts as a "organization admin". API endpoints provided for organization admins to list the members of their organizations, and to also see the overview (environments, shares, and accesses) for any account in their organization. API endpoint for end users to see which organizations their account is a member of (https://github.com/openziti/zrok/issues/537)

CHANGE: briefly mention the backend modes that apply to public and private share concepts

FIX: Update indirect dependency `github.com/golang-jwt/jwt/v4` to version `v4.5.1` (https://github.com/openziti/zrok/issues/794)

FIX: Document unique names

FIX: reduce Docker image sizes (https://github.com/openziti/zrok/pull/783)

FIX: Docker reserved private share startup error (https://github.com/openziti/zrok/pull/801)

FIX: Correct the download URL for the armv7 Linux release (https://github.com/openziti/zrok/issues/782)

## v0.4.44

FIX: Fix for goreleaser build action to align with changed ARM64 build path.

## v0.4.43

CHANGE: Update `github.com/openziti/sdk-golang` to version `v0.23.44`. Remove old `github.com/openziti/fabric` dependency, instead pulling in the modern `github.com/openziti/ziti` dependency.

FIX: Bypass interstitial page for HTTP `OPTIONS` method (https://github.com/openziti/zrok/issues/777)

## v0.4.42

CHANGE: Switch all `Dial` operations made into the OpenZiti overlay to use `DialWithOptions(..., &ziti.DialOptions{ConnectTimeout: 30 * time.Second})`, switching to a 30 second timeout from a 5 second default (https://github.com/openziti/zrok/issues/772)

FIX: Removed the `--basic-auth` flag from `zrok share private` as this was ignored... even if `zrok access private` honored the `ziti.proxy.v1` config to ask for basic auth, it would still be easy to write a custom SDK client that ignored the basic auth and accessed the share directly; better to remove the option than to allow confusing usage (https://github.com/openziti/zrok/issues/770)

FIX: always append common options like `--headless` and conditionally append `--verbose --insecure` if their respective env vars are set to when running in a service manager like systemd or Docker and wrapping the `zrok` command with the `zrok-share.bash` shell script (https://openziti.discourse.group/t/question-about-reserved-public-vs-temp-public-shares/3169)

FIX: Correct registration page CSS to ensure that the entire form is visible

## v0.4.41

FIX: Fixed crash when invoking `zrok share reserved` with no arguments (https://github.com/openziti/zrok/issues/740)

FIX: zrok-share.service on Linux failed to start with a private share in closed permission mode

FIX: Update `gopkg.in/go-jose/go-jose.v2` to `v2.6.3` to fix vulnerability around compressed data (https://github.com/openziti/zrok/issues/761)

## v0.4.40

FEATURE: New endpoint for synchronizing grants for an account (https://github.com/openziti/zrok/pull/744). Useful for updating the `zrok.proxy.v1` config objects containing interstitial setting when the `skip_interstitial_grants` table has been updated.

FIX: prune incorrect troubleshooting advice about listing Caddy's certificates

## v0.4.39

FEATURE: New API endpoint allowing direct creation of accounts in the zrok database. Requires an admin token (specified in the controller configuration yaml) for authentication. See the OpenAPI spec for details of the API endpoint. The `zrok admin create account` CLI was also updated to call the API endpoint, rather than directly operating on the underlying database (https://github.com/openziti/zrok/issues/734). The [Docker](https://github.com/openziti/zrok/pull/736) and [Kubernetes](https://github.com/openziti/helm-charts/pull/249) zrok instance deployments were adapted to the new CLI parameter shape.

FEATURE: Support `html_path` directive in `interstitial` stanza of public frontend configuration to support using an external HTML file for the interstitial page (https://github.com/openziti/zrok/issues/716)

FEATURE: `zrok access private` now includes a `--response-header` flag to add headers to the response for HTTP-based backends. Add flag multiple times to add multiple headers to the response. Expects `key:value` header definitions in this format: `--response-header "Access-Control-Allow-Origin: *"` (https://github.com/openziti/zrok/issues/522)

CHANGE: Update `github.com/openziti/sdk-golang` (and related dependencies) to version `v0.23.40`.

CHANGE: upgrade to ziti v1.1.7 CLI in zrok container image

## v0.4.38

FEATURE: Conditionally enable interstitial page based on `User-Agent` prefix list. See the [frontend configuration template](etc/frontend.yml) for details on the new configuration structure (https://github.com/openziti/zrok/issues/715) 

CHANGE: The interstitial configuration has been modified from a simple `interstitial: <bool>` to a richer structure, but the config version has not been incremented; this feature has not been widely adopted yet. See the [frontend configuration template](etc/frontend.yml) for details on the new structure.

CHANGE: The registration page where a new user's password is set now includes a required checkbox, asking them to acknowledge the terms and conditions presented above the checkbox (https://github.com/openziti/zrok/issues/669)

FIX: The registration page where a new user's password is set now includes better styling of the error message `<div/>` to prevent the entire page from jumping when the message changes.

## v0.4.37

FIX: Fix for setting the `zrok_interstitial` cookie on Chrome-based browsers.

FIX: Fix for `store.IsAccountGrantedSkipInterstitial` to respect the `deleted` flag.

FIX: When an error occurs connecting to the proxied endpoint, the `proxy` backend should return HTTP status `502` (https://github.com/openziti/zrok/issues/703)

## v0.4.36

FEATURE: New interstitial pages that can be enabled per-frontend, and disabled per-account (https://github.com/openziti/zrok/issues/704)

CHANGE: Enable `"declaration": true` in `tsconfig.json` for Node SDK.

FIX: build 32bit build for armhf to fix [the FPE issue](https://github.com/openziti/zrok/issues/654) and [the missing link issue](https://github.com/openziti/zrok/issues/642)

CHANGE: add [cross-build instructions](./BUILD.md) (includes new snapshot build target `armel`)

## v0.4.35

FEATURE: Added import for `github.com/greenpau/caddy-security` to include that Caddy plugin to enable authentication, authorization, and credentials extensions for the `caddy` backend (https://github.com/openziti/zrok/issues/506)

FEATURE: Closed permission mode for Docker and Linux private shares

CHANGE: add example in ./etc/caddy to set X-Real-IP header to public share client IP

CHANGE: auto-update the ziti CLI version that is built in to the openziti/zrok container image

CHANGE: Docker examples set HOME to enable running CLI commands in the container

FIX: Fix for environment count inheritance when using a resource count class to override global environment count (https://github.com/openziti/zrok/issues/695)

## v0.4.34

FEATURE: Linux service support for all private share modes (contribution from Stefan Adelbert @stefanadelbert)

FIX: Fix for mixing limited and unlimited (-1) resource counts in the limits system (https://github.com/openziti/zrok/issues/680)

FIX: Fix for sending multiple warning emails when a warning is applied to an account (https://github.com/openziti/zrok/issues/685)

CHANGE: add Docker compose example for multiple share containers using the same enabled environment in [compose.override.yml](./docker/compose/zrok-public-reserved/compose.override.yml)

CHANGE: bump many GitHub Actions that were using deprecated distributions of Node.js

CHANGE: bump macOS runner for Node SDK from macos-11 to macos-12

## v0.4.33

FIX: Fix for log message in `Agent.CanAccessShare` (`"account '#%d' over frontends per share limit '%d'"`), which was not returning the correct limit value.

FIX: Properly set `permission_mode` in `frontends` when createing a private frontend using `zrok access private` (https://github.com/openziti/zrok/issues/677)

CHANGE: Updated `react-bootstrap` to version `2.10.2` (web console).

CHANGE: Updated `@mui/material` to version `5.15.18` (web console).

CHANGE: Updated `react` and `react-dom` to version `18.3.1` (web console).

CHANGE: Updated `recharts` to version `2.12.7` (web console).

CHANGE: Updated `react-router-dom` to version `6.23.1` (web console).

CHANGE: Updated `axios` to version `1.7.2` for (node SDK).

CHANGE: Updated `@openziti/ziti-sdk-nodejs` to version `0.17.0` (node SDK).

## v0.4.32

FEATURE: New permission mode support for public frontends. Open permission mode frontends are available to all users in the service instance. Closed permission mode frontends reference the new `frontend_grants` table that can be used to control which accounts are allowed to create shares using that frontend. `zrok admin create frontend` now supports `--closed` flag to create closed permission mode frontends (https://github.com/openziti/zrok/issues/539)

FEATURE: New config `defaultFrontend` that specifies the default frontend to be used for an environment. Provides the default `--frontend` for `zrok share public` and `zrok reserve public` (https://github.com/openziti/zrok/issues/663)

FEATURE: Resource count limits now include `share_frontends` to limit the number of frontends that are allowed to make connections to a share (https://github.com/openziti/zrok/issues/650)

CHANGE: The frontend selection flag used by `zrok share public` and `zrok reserve public` has been changed from `--frontends` to `--frontend`

FIX: use controller config spec v4 in the Docker instance

## v0.4.31

FEATURE: New "limits classes" limits implementation (https://github.com/openziti/zrok/issues/606). This new feature allows for extensive limits customization on a per-user basis, with fallback to the global defaults in the controller configuration.

CHANGE: The controller configuration version has been updated to version `4` (`v: 4`) to support the new limits global configuration changes (https://github.com/openziti/zrok/issues/606).

CHANGE: A new `ZROK_CTRL_CONFIG_VERSION` environment variable now exists to temporarily force the controller to assume a specific controller configuration version, regardless of what version exists in the file. This allows two different config versions to potentially be co-mingled in the same controller configuration file. Use with care (https://github.com/openziti/zrok/issues/648)

CHANGE: Log messages that said `backend proxy endpoint` were clarified to say `backend target`.

FIX: Correct the syntax for the Docker and Linux zrok-share "frontdoor" service that broke OAuth email address pattern matching.

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

FEATURE: New CLI commands have been implemented for working with the `drive` share backend mode (part of the "zrok Drives" functionality). These commands include `zrok cp`, `zrok mkdir` `zrok mv`, `zrok ls`, and `zrok rm`. These are initial, minimal versions of these commands and very likely contain bugs and ergonomic annoyances. There is a guide available at (`docs/guides/drives.mdx`) that explains how to work with these tools in detail (https://github.com/openziti/zrok/issues/438)

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
