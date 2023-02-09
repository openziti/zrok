# v0.3.1

CHANGE: Incorporate initial docker image build (https://github.com/openziti/zrok/issues/217)

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

# v0.3.0

The `v0.2` series was a _proof-of-concept_ implementation for the overall `zrok` architecture and the concept.

`v0.3` is a massive elaboration of the concept, pivoting it from being a simple ephemeral reverse proxy solution, to being the beginnings of a comprehensive sharing platform, complete with public and private sharing (built on top of OpenZiti). 

`v0.3.0` includes the minimal functionality required to produce an early, preview version of the elaborated `zrok` concept, suitable for both production use at `zrok.io`, and also suitable for private self-hosting.

From `v0.3.0` forward, we will begin tracking notable changes in this document.

# v0.2.18

* DEFECT: Token generation has been improved to use an alphabet consisting of `[a-zA-Z0-9]`. Service token generation continues to use a case-insensitive alphabet consisting of `[a-z0-9]` to be DNS-safe.
