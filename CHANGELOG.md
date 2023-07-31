# v0.4.4

FIX: `zrok status`, `zrok enable`, `zrok config`, etc. were all causing a panic when used on systems that had no previous `~/.zrok` directory (https://github.com/openziti/zrok/issues/383)

# v0.4.3

FEATURE: New `zrok overview` command, which returns all of the account details as a single JSON structure. See the OpenAPI spec at `specs/zrok.yml` for more details of the `/api/v1/overview` endpoint (https://github.com/openziti/zrok/issues/374)

FEATURE: New `zrok` SDK (https://github.com/openziti/zrok/issues/34). `pastebin` example illustrates basic SDK usage (see `sdk/examples/pastebin/README.md` for details) ((https://github.com/openziti/zrok/issues/379)

# v0.4.2

Some days are just like this. `v0.4.2` is a re-do of `v0.4.1`. Trying to get Homebrew working and had a bad release. Hopefully this is the one.

# v0.4.1

FEATURE: New `zrok console` command to open the currently configured web console in the local web browser (https://github.com/openziti/zrok/issues/170)

CHANGE: Further tweaks to the release process to automatically get the latest release into Homebrew (https://github.com/openziti/zrok/issues/264)

# v0.4.0

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

# v0.3.7

FIX: Improved TUI word-wrapping (https://github.com/openziti/zrok/issues/180)

# v0.3.6

CHANGE: Additional change to support branch builds (for CI purposes) and additional containerization efforts around k8s.

# v0.3.5

CHANGE: `zrok config set apiEndpoint` now validates that the new API endpoint correctly starts with `http://` or `https://` (https://github.com/openziti/zrok/issues/258)

CHANGE: Additional linting to support homebrew (https://github.com/openziti/zrok/issues/264)

# v0.3.4

CHANGE: `zrok test endpoint` incorporates `--ziti` mode (and related flags) to allow direct endpoint listening on a Ziti service

CHANGE: `zrok test websocket` command to test websockets, whether over TCP or over Ziti

FIX: Websocket support now functional

# v0.3.3

CHANGE: `zrok test loop` has been moved to `zrok test loop public`, making way for additional types of loopback testing tools. The `zrok test endpoint` server now includes an `/echo` endpoint, which provides a simple echo websocket (https://github.com/openziti/zrok/issues/237)

# v0.3.2

FEATURE: New docker infrastructure, including `docker-compose.yml` examples (and documentation) illustrating how to deploy `zrok` in `docker`-based environments

CHANGE: Include missing `--headless` flag for `zrok enable` and `zrok access private` (https://github.com/openziti/zrok/issues/246)

CHANGE: Fix for `zrok enable` error path handling (https://github.com/openziti/zrok/issues/244)

FEATURE: `zrok controller validate` and `zrok access public validate` will both perform a quick syntax validation on controller and public frontend configuration documents (https://github.com/openziti/zrok/issues/238)

	$ zrok controller validate etc/dev.yml 
	[ERROR]: controller config validation failed (error loading controller config 'etc/dev.yml': field 'maintenance': field 'registration': field 'expiration_timeout': got [bool], expected [time.Duration])

CHANGE: `zrok status` no longer shows secrets (secret token, ziti identity) unless the `--secrets` flag is passed (https://github.com/openziti/zrok/issues/243)

# v0.3.1

CHANGE: Incorporate initial docker image build (https://github.com/openziti/zrok/issues/217)

CHANGE: Improve target URL parsing for `zrok share` when using `--backend-mode` proxy (https://github.com/openziti/zrok/issues/211)

	New and improved URL handling for proxy backends:
	
	9090 -> http://127.0.0.1:9090
	localhost:9090 -> http://127.0.0.1:9090
	https://localhost:9090 -> https://localhost:9090

CHANGE: Improve usability of `zrok invite` TUI in low-color environments (https://github.com/openziti/zrok/issues/206)

CHANGE: Better error responses when `zrok invite` fails due to missing token (https://github.com/openziti/zrok/issues/207)

# v0.3.0

CHANGE: Removed some minor web console lint and warnings (https://github.com/openziti/zrok/issues/205)

# v0.3.0-rc6

CHANGE: Better error message when `zrok admin create frontend` runs into a duplicate name collision (https://github.com/openziti/zrok/issues/168)

CHANGE: Gentler CLI error messages by default (https://github.com/openziti/zrok/issues/203)

CHANGE: Add favicon to web console (https://github.com/openziti/zrok/issues/198)

CHANGE: Add configurable "terms of use" link in the controller configuration, and optionally display the link on the login form and registration forms (https://github.com/openziti/zrok/issues/184)

CHANGE: Prevent multiple `zrok enable` commands from succeeding (https://github.com/openziti/zrok/issues/190)

CHANGE: New `--insecure` flag for `share <public|private|reserved>` commands (https://github.com/openziti/zrok/issues/195)

# v0.3.0-rc5

CHANGE: Improvements to controller log messages to assist in operations (https://github.com/openziti/zrok/issues/186)

CHANGE: `armv7` builds for Linux are now shipped with releases; these builds were tested against a Raspberry Pi 4 (https://github.com/openziti/zrok/issues/93)

CHANGE: `zrok config set` now includes a warning when the `apiEndpoint` config is changed and an environment is already enabled; the user will not see the change until `zrok disable` is run. The CLI now includes a `zrok config unset` command (https://github.com/openziti/zrok/issues/188)

# v0.3.0-rc4

CHANGE: Enable notarization for macos binaries (https://github.com/openziti/zrok/issues/92)

# v0.3.0-rc3

> This release increments the configuration version from `1` to `2`. See the note below.

CHANGE: The email "from" configuration moved from `registration/email_from` to `email/from`. **NOTE: This change increments the configuration `V` from `1` to `2`.**

CHANGE: Replaced un-salted sha512 password hashing with salted hashing based on Argon2 **NOTE: This version will _invalidate_ all account passwords, and will require all users to use the 'Forgot Password?' function to reset their password.** (https://github.com/openziti/zrok/issues/156)

CHANGE: Switched from `ubuntu-latest` (`22.04`) for the Linux builds to `ubuntu-20.04`. Should improve `glibc` compatibility with older Linux distributions (https://github.com/openziti/zrok/issues/179)

CHANGE: `zrok admin generate` now outputs the generated tokens to `stdout` after successfully provisioning the tokens (https://github.com/openziti/zrok/issues/181)

FIX: Fixed log message in `resetPasswordRequest.go` (https://github.com/openziti/zrok/issues/175)

FIX: Fixed `-v` (verbose mode) on in TUI-based `zrok share` and `zrok access` (https://github.com/openziti/zrok/issues/174)

# v0.3.0-rc2

FEATURE: Allow users to reset their password (https://github.com/openziti/zrok/issues/65)

CHANGE: Improved email styling for new user invite emails (https://github.com/openziti/zrok/issues/157)

CHANGE: Migrated from `openziti-test-kitchen` to `openziti` (https://github.com/openziti/zrok/issues/158).

CHANGE: Show a hint when `zrok invite` fails, indicating that the user should check to see if they need to be using the `--token` flag and token-based invites (https://github.com/openziti/zrok/issues/172).

FIX: Fixed PostgreSQL migration issue where sequences got reset and resulted in primary key collisions on a couple of tables (https://github.com/openziti/zrok/issues/160).

FIX: Remove `frontend` instances when `zrok disable`-ing an environment containing them (https://github.com/openziti/zrok/issues/171)

# v0.3.x Series

The `v0.2` series was a _proof-of-concept_ implementation for the overall `zrok` architecture and the concept.

`v0.3` is a massive elaboration of the concept, pivoting it from being a simple ephemeral reverse proxy solution, to being the beginnings of a comprehensive sharing platform, complete with public and private sharing (built on top of OpenZiti). 

`v0.3.0` includes the minimal functionality required to produce an early, preview version of the elaborated `zrok` concept, suitable for both production use at `zrok.io`, and also suitable for private self-hosting.

From `v0.3.0` forward, we will begin tracking notable changes in this document.

# v0.2.18

* DEFECT: Token generation has been improved to use an alphabet consisting of `[a-zA-Z0-9]`. Service token generation continues to use a case-insensitive alphabet consisting of `[a-z0-9]` to be DNS-safe.
