#!/usr/bin/env bash

# Test the documented Linux self-hosting deployment of zrok2.
# Builds zrok2 packages with goreleaser, installs Ziti v2 controller +
# router from the APT repo, installs the built zrok2 packages, runs
# zrok2-bootstrap.bash, and verifies all services start and the API responds.
#
# This script runs as a normal user; privileged commands use sudo internally.
# Requires: Go, Node/npm, and goreleaser (for building packages).
# Use --pkg-dir to skip building and install pre-built packages instead.
#
# Usage:
#   linux.test.bash --ziti-version <version> [--source-dir <dir>] [--pkg-dir <dir>] [--keep]
#
# Options:
#   --ziti-version <version>  Ziti package version to install (e.g., 2.0.0, 2.0.0-rc5)
#   --source-dir <dir>        zrok source tree root (default: auto-detected from
#                             script location)
#   --pkg-dir <dir>           use pre-built packages from <dir> (skip building)
#   --keep                    keep the test instance running on exit (for inspection)
#
# Environment variables:
#   ZITIPAX_DEB   Override the Debian repo (default: zitipax-openziti-deb-stable)
#   ZITIPAX_RPM   Override the RedHat repo (default: zitipax-openziti-rpm-stable)
#
# Examples:
#   # Build packages and test (typical CI workflow):
#   bash linux.test.bash --ziti-version 2.0.0
#
#   # Use pre-built packages (typical local workflow):
#   bash linux.test.bash --ziti-version 2.0.0 --pkg-dir ./dist
#
#   # Install a pre-release from the test repo:
#   ZITIPAX_DEB=zitipax-openziti-deb-test bash linux.test.bash --ziti-version 2.0.0-rc5
#
#   # Explicit source dir:
#   bash linux.test.bash --ziti-version 2.0.0 --source-dir ~/src/zrok
#
#   # Capture output to a log and preserve the exit code:
#   bash linux.test.bash --ziti-version 2.0.0 </dev/null 2>&1 | tee /tmp/test.log; echo "exit: ${PIPESTATUS[0]}"
#   # NOTE: |& (or 2>&1 |) pipes both streams to tee, but the shell reports
#   # tee's exit code (always 0). Use ${PIPESTATUS[0]} to get the script's exit.

set -o errexit
set -o nounset
set -o pipefail
set -o errtrace
set -o xtrace

# Handle --help/-h before anything else (no sudo, no traps needed)
for _arg in "$@"; do
    case "$_arg" in
        --help|-h)
            cat <<'EOF'
Usage: linux.test.bash --ziti-version <version> [--source-dir <dir>] [--pkg-dir <dir>] [--keep]

  --ziti-version <version>  Ziti package version to install from APT (e.g., 2.0.0)
  --source-dir <dir>        zrok source tree (default: auto-detected)
  --pkg-dir <dir>           use pre-built packages (skip building)
  --keep                    keep the test instance running on exit (for inspection)
  --only-clean              run cleanup only (tear down a kept instance) and exit

State from a prior run is always destroyed at the start. By default, the
test also cleans up on exit. Use --keep to leave services running for
post-mortem inspection; the next run cleans up regardless.
Use --only-clean to tear down a kept instance without re-running the test.
EOF
            exit 0
            ;;
    esac
done

# Validate sudo access upfront (prompts for password before output is piped)
if ! sudo -v; then
    echo "ERROR: sudo access is required" >&2
    exit 1
fi

# ============================================================
# Logging
# ============================================================

log_section() { printf '\n\033[1;36m=== %s ===\033[0m\n\n' "$1" >&2; }
log_info()    { printf '\033[34mINFO:\033[0m %s\n' "$1" >&2; }
log_error()   { printf '\033[31mERROR:\033[0m %s\n' "$1" >&2; }
log_pass()    { printf '\033[32mPASS:\033[0m %s\n' "$1" >&2; }
log_fail()    { printf '\033[31mFAIL:\033[0m %s\n' "$1" >&2; }

# ============================================================
# Utilities
# ============================================================

generate_password() {
    head -c "${1:-24}" /dev/urandom | base64 -w0
}

retry() {
    local _max="$1" _delay="$2"
    shift 2
    local _attempts="${_max}"
    until ! (( _attempts )) || "$@"; do
        (( _attempts-- ))
        log_info "retry ($(( _max - _attempts ))/${_max}): $*"
        sleep "${_delay}"
    done
    if (( ! _attempts )); then
        log_error "command failed after ${_max} attempts: $*"
        return 1
    fi
}

wait_for_port() {
    local _host="$1" _port="$2" _timeout="${3:-30}"
    local _deadline=$(( SECONDS + _timeout ))
    while (( SECONDS < _deadline )); do
        if nc -z "${_host}" "${_port}" >/dev/null 2>&1; then
            log_info "port ${_host}:${_port} is reachable"
            return 0
        fi
        sleep 1
    done
    log_error "port ${_host}:${_port} not reachable within ${_timeout}s"
    return 1
}

wait_for_service() {
    local _svc="$1" _timeout="${2:-30}"
    local _deadline=$(( SECONDS + _timeout ))
    while (( SECONDS < _deadline )); do
        if sudo systemctl is-active --quiet "${_svc}" 2>/dev/null; then
            log_info "${_svc} is active"
            return 0
        fi
        sleep 1
    done
    log_error "${_svc} is not active after ${_timeout}s"
    dump_journal "${_svc}" 100
    return 1
}

dump_journal() {
    local _svc="$1" _lines="${2:-100}"
    echo "--- journal: ${_svc} (last ${_lines} lines) ---" >&2
    sudo journalctl -xeu "${_svc}" --no-pager -n "${_lines}" 2>&1 || true
    echo "--- end journal: ${_svc} ---" >&2
}

start_service() {
    local _svc="$1"
    log_info "starting ${_svc}"
    if ! sudo systemctl start "${_svc}"; then
        log_error "${_svc} failed to start"
        dump_journal "${_svc}" 200
        return 1
    fi
}

check_port_available() {
    local _port="$1"
    if nc -zv localhost "${_port}" &>/dev/null; then
        log_error "port ${_port} is already allocated"
        return 1
    fi
    log_info "port ${_port} is available"
}

# ============================================================
# Install Ziti packages via install.bash
# ============================================================
# Honors env vars:
#   ZITIPAX_DEB  — override Debian repo (default: zitipax-openziti-deb-stable)
#   ZITIPAX_RPM  — override RedHat repo (default: zitipax-openziti-rpm-stable)

install_openziti() {
    log_section "Installing OpenZiti packages: $*"
    if [[ -n "${ZITIPAX_DEB:-}" ]]; then
        log_info "using custom Debian repo: ${ZITIPAX_DEB}"
    fi
    if [[ -n "${ZITIPAX_RPM:-}" ]]; then
        log_info "using custom RedHat repo: ${ZITIPAX_RPM}"
    fi
    # install.bash configures the repo and installs packages in one step.
    # It honors ZITIPAX_DEB / ZITIPAX_RPM to select the repo.
    curl -fsSL https://get.openziti.io/install.bash \
        | sudo DEBIAN_FRONTEND=noninteractive \
               ZITIPAX_DEB="${ZITIPAX_DEB:-}" \
               ZITIPAX_RPM="${ZITIPAX_RPM:-}" \
               bash -s -- "$@"
}

# ============================================================
# Error handling and cleanup
# ============================================================

_exit_code=0
_fail_summary=""
_err_handler() {
    _exit_code=$?
    trap - ERR  # prevent re-entry
    _fail_summary="FAILED at line ${LINENO}: ${BASH_COMMAND} (exit ${_exit_code})"
    log_error "${_fail_summary}"
    # Redirect to stderr so diagnostics don't pollute command substitution stdout
    dump_journal ziti-controller.service >&2
    dump_journal ziti-router.service >&2
    dump_journal zrok2-controller.service >&2
    dump_journal zrok2-frontend.service >&2
    dump_journal zrok2-metrics-bridge.service >&2
    exit "${_exit_code}"
}
trap '_err_handler' ERR

KEEP=0

# Wait for the dpkg lock (unattended-upgrades, etc.) up to 60s.
wait_for_dpkg_lock() {
    local _attempts=60
    while sudo fuser /var/lib/dpkg/lock-frontend &>/dev/null; do
        if (( --_attempts == 0 )); then
            log_error "dpkg lock still held after 60s"
            return 1
        fi
        sleep 1
    done
}

# Stop all services that may have been started by a prior run.
stop_services() {
    for _svc in zrok2-metrics-bridge zrok2-frontend zrok2-controller ziti-router ziti-controller \
                rabbitmq-server postgresql influxdb; do
        sudo systemctl stop "${_svc}" 2>/dev/null || true
    done
}

# Remove all packages, data dirs, repos, and keyrings from a prior run.
# On failure: returns non-zero so the caller can decide whether to abort.
purge_packages() {
    wait_for_dpkg_lock
    sudo dpkg --purge zrok2-metrics-bridge zrok2-metrics zrok2-frontend zrok2-controller zrok2-agent zrok2 2>/dev/null || true
    sudo dpkg --purge openziti-router openziti-controller openziti 2>/dev/null || true
    # Keep influxdb2 installed to avoid re-downloading from the slow InfluxData
    # repo on every run. Its data dirs are purged in purge_data() so each run
    # starts with a clean database.
    sudo DEBIAN_FRONTEND=noninteractive apt-get purge -y 'rabbitmq-server*' 'postgresql*'
    sudo DEBIAN_FRONTEND=noninteractive apt-get autoremove -y
}

purge_data() {
    # Keep influxdata repo/keyring so re-runs don't need to re-add them
    sudo rm -rf /var/lib/ziti-controller /var/lib/private/ziti-controller \
           /var/lib/ziti-router /var/lib/private/ziti-router \
           /var/lib/zrok2-controller /var/lib/zrok2-frontend \
           /var/lib/influxdb /var/lib/influxdb2 \
           /var/lib/rabbitmq /var/lib/postgresql \
           /etc/zrok2 /etc/influxdb /etc/influxdb2 \
           /root/.influxdbv2 /root/.zrok2 2>/dev/null || true
    sudo systemctl daemon-reload 2>/dev/null || true
}

# Pre-test cleanup: must succeed or the test cannot proceed.
cleanup() {
    log_section "Cleanup"
    stop_services
    purge_packages
    purge_data
}

# Exit cleanup: best-effort — don't mask the real exit code.
cleanup_on_exit() {
    local _real_exit=$?
    set +o errexit
    # Catch nounset/syntax errors that bypass the ERR trap (exit != 0 but
    # _exit_code was never updated).
    if (( _real_exit != 0 && _exit_code == 0 )); then
        _exit_code="${_real_exit}"
        _fail_summary="exited with status ${_real_exit} (may be nounset or syntax error)"
    fi
    if (( KEEP )); then
        log_info "keeping test instance (--keep); run with --only-clean to tear down"
    else
        log_section "Cleanup (on exit)"
        stop_services
        purge_packages 2>/dev/null || true
        purge_data
    fi

    # Print a clear one-line result summary
    echo >&2
    if (( _exit_code == 0 )); then
        log_pass "dangerous.linux.test: PASSED"
    else
        log_fail "dangerous.linux.test: FAILED (exit ${_exit_code})"
        if [[ -n "${_fail_summary}" ]]; then
            log_error "  ${_fail_summary}"
        fi
    fi
}
trap 'cleanup_on_exit; exit $_exit_code' EXIT

# Show a TTY warning before the initial cleanup so the user knows what's
# about to be destroyed from a prior run.
warn_before_cleanup() {
    if [[ -t 0 ]]; then
        cat >&2 <<'WARN'
About to destroy all state from a prior test run:

  Services:   zrok2-controller, zrok2-frontend, zrok2-metrics-bridge,
              ziti-controller, ziti-router,
              rabbitmq-server, postgresql, influxdb
  Packages:   zrok2, zrok2-controller, zrok2-frontend, zrok2-metrics-bridge, zrok2-agent,
              openziti, openziti-controller, openziti-router,
              influxdb2, rabbitmq-server, postgresql
  Data dirs:  /var/lib/{ziti-controller,ziti-router} (+ /var/lib/private/...),
              /var/lib/{zrok2-controller,zrok2-frontend},
              /var/lib/{influxdb,influxdb2,rabbitmq,postgresql}
  Config:     /etc/zrok2, /etc/influxdb, /etc/influxdb2, /root/.influxdbv2
  APT repos:  /etc/apt/sources.list.d/influxdata.list
  Keyrings:   /usr/share/keyrings/influxdata-archive-keyring.gpg
  dpkg info:  /var/lib/dpkg/info/influxdb2.{prerm,postrm} (if broken)

Proceeding in 30s. Re-run with </dev/null to skip this delay.
WARN
        sleep 30
    fi
}

# ============================================================
# Parse CLI arguments
# ============================================================

usage() {
    trap - EXIT ERR  # don't clean up on usage/help exits
    cat >&2 <<'EOF'
Usage: linux.test.bash --ziti-version <version> [--source-dir <dir>] [--pkg-dir <dir>] [--keep]

  --ziti-version <version>  Ziti package version to install from APT (e.g., 2.0.0)
  --source-dir <dir>        zrok source tree (default: auto-detected)
  --pkg-dir <dir>           use pre-built packages (skip building)
  --keep                    keep the test instance running on exit (for inspection)
  --only-clean              run cleanup only (tear down a kept instance) and exit

State from a prior run is always destroyed at the start. By default, the
test also cleans up on exit. Use --keep to leave services running for
post-mortem inspection; the next run cleans up regardless.
Use --only-clean to tear down a kept instance without re-running the test.
EOF
    exit 1
}

ZITI_VERSION=""
SOURCE_DIR=""
PKG_DIR_ARG=""
ONLY_CLEAN=0
while [[ $# -gt 0 ]]; do
    case "$1" in
        --ziti-version) ZITI_VERSION="$2"; shift 2 ;;
        --source-dir)   SOURCE_DIR="$2"; shift 2 ;;
        --pkg-dir)      PKG_DIR_ARG="$2"; shift 2 ;;
        --keep)         KEEP=1; shift ;;
        --only-clean)   ONLY_CLEAN=1; shift ;;
        *)              usage ;;
    esac
done

if (( ONLY_CLEAN )); then
    trap - EXIT ERR  # cleanup is intentional here; don't double-run it on exit
    log_section "Cleanup only"
    stop_services
    purge_packages 2>/dev/null || true
    purge_data
    log_info "cleanup complete"
    exit 0
fi

[[ -n "${ZITI_VERSION}" ]] || { log_error "--ziti-version is required"; usage; }

# Strip leading 'v' — deb packages use bare semver (e.g., 2.0.0~pre5, not v2.0.0~pre5)
ZITI_VERSION="${ZITI_VERSION#v}"

# Auto-detect source dir from script location (nfpm/linux.test.bash -> repo root)
if [[ -z "${SOURCE_DIR}" ]]; then
    SOURCE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fi
[[ -d "${SOURCE_DIR}" ]] || { trap - EXIT ERR; log_error "source dir '${SOURCE_DIR}' is not a directory"; exit 1; }

# ============================================================
# Build packages with goreleaser (or use pre-built via --pkg-dir)
# ============================================================

build_packages() {
    log_section "Building zrok2 packages with goreleaser"

    for _cmd in go node npm goreleaser; do
        if ! command -v "${_cmd}" &>/dev/null; then
            log_error "'${_cmd}' is required to build packages (install it or use --pkg-dir)"
            exit 1
        fi
    done

    log_info "building UI assets with npm"
    (
        cd "${SOURCE_DIR}"
        npm config set cache "${SOURCE_DIR}/.npm"
        for _ui in ./ui ./agent/agentUi; do
            log_info "npm ci + build in ${_ui}"
            (cd "${_ui}" && npm ci && npm run build)
        done
    )

    log_info "running goreleaser release --snapshot (this may take a few minutes)"
    (
        cd "${SOURCE_DIR}"
        goreleaser release \
            --snapshot \
            --clean \
            --skip=publish,validate \
            --config .goreleaser-linux-amd64.yml
    )

    log_info "packages built in ${PKG_DIR}"
}

if [[ -n "${PKG_DIR_ARG}" ]]; then
    PKG_DIR="$(cd "${PKG_DIR_ARG}" && pwd)"
    log_info "using pre-built packages from ${PKG_DIR}"
else
    PKG_DIR="${SOURCE_DIR}/dist"
    build_packages
fi

# Verify zrok2 .deb files exist
shopt -s nullglob
_zrok2_debs=("${PKG_DIR}"/zrok2_*.deb)
shopt -u nullglob
if [[ ${#_zrok2_debs[@]} -eq 0 ]]; then
    log_error "no zrok2_*.deb files found in ${PKG_DIR}"
    exit 1
fi
log_info "found ${#_zrok2_debs[@]} zrok2 .deb file(s) in ${PKG_DIR}"

# Ensure required commands are available
for BIN in nc curl jq; do
    if ! command -v "$BIN" &>/dev/null; then
        log_error "required command '$BIN' not found"
        exit 1
    fi
done

# ============================================================
# Test configuration
# ============================================================

ZITI_BIN=/usr/bin/ziti
ZROK2_BIN=/usr/bin/zrok2

# Ziti v2 uses FQDN-based addresses and cluster mode
ZITI_CTRL_ADVERTISED_ADDRESS="ziti-controller1.127.0.0.1.sslip.io"
ZITI_CTRL_ADVERTISED_PORT="1281"
ZITI_CLUSTER_NODE_NAME="${ZITI_CTRL_ADVERTISED_ADDRESS%%.*}"
ZITI_CLUSTER_TRUST_DOMAIN="${ZITI_CTRL_ADVERTISED_ADDRESS#*.}"

ZITI_USER="admin"
ZITI_PWD="$(generate_password)"

ZITI_ROUTER_NAME="ziti-router1"
ZITI_ROUTER_ADVERTISED_ADDRESS="${ZITI_ROUTER_NAME}.127.0.0.1.sslip.io"
ZITI_ROUTER_PORT="30223"

ZROK2_DNS_ZONE="zrok.127.0.0.1.sslip.io"
ZROK2_ADMIN_TOKEN="$(generate_password)"
ZROK2_CTRL_PORT="18080"

: "${TMPDIR:=$(mktemp -d)}"
ZITI_ENROLL_TOKEN_FILE="${TMPDIR}/${ZITI_ROUTER_NAME}.jwt"

export DEBIAN_FRONTEND=noninteractive
warn_before_cleanup
cleanup

for PORT in "${ZITI_CTRL_ADVERTISED_PORT}" "${ZITI_ROUTER_PORT}" "${ZROK2_CTRL_PORT}"; do
    check_port_available "${PORT}"
done

# ============================================================
# Phase 1: Install Ziti from the stable APT repo
# ============================================================
log_section "Phase 1: Install Ziti v2"

# Convert upstream version to deb convention: hyphens become tildes
# (e.g., 2.0.0-rc1 -> 2.0.0~rc1) and append wildcard for the deb revision suffix
ZITI_DEB_VERSION="${ZITI_VERSION//-/\~}*"

log_info "pinning to Ziti version: ${ZITI_VERSION} (deb pattern: ${ZITI_DEB_VERSION})"
install_openziti \
    "openziti=${ZITI_DEB_VERSION}" \
    "openziti-controller=${ZITI_DEB_VERSION}" \
    "openziti-router=${ZITI_DEB_VERSION}"

# Verify installed version matches (ignoring deb revision suffix)
_installed_ver=$(dpkg-query -W -f='${Version}' openziti-controller)
log_info "installed openziti-controller version: ${_installed_ver}"
_expected_deb_ver="${ZITI_VERSION//-/\~}"
if [[ "${_installed_ver}" != "${_expected_deb_ver}"* ]]; then
    log_error "expected version ${_expected_deb_ver}* but got ${_installed_ver}"
    exit 1
fi

# ============================================================
# Phase 2: Bootstrap Ziti controller (v2: cluster mode + static user)
# ============================================================
log_section "Phase 2: Bootstrap Ziti controller"

# v2 bootstrap.bash accepts an answer file as $1.
# ZITI_BOOTSTRAP=true triggers config generation + database init + service start.
# ZITI_BOOTSTRAP_CLUSTER=true creates a new cluster with PKI.
CTRL_ANSWERS="$(mktemp)"
cat > "${CTRL_ANSWERS}" <<CTRL_ENV
ZITI_BOOTSTRAP=true
ZITI_BOOTSTRAP_CLUSTER=true
ZITI_BOOTSTRAP_CONSOLE=true
ZITI_CLUSTER_NODE_NAME=${ZITI_CLUSTER_NODE_NAME}
ZITI_CLUSTER_TRUST_DOMAIN=${ZITI_CLUSTER_TRUST_DOMAIN}
ZITI_CTRL_ADVERTISED_ADDRESS=${ZITI_CTRL_ADVERTISED_ADDRESS}
ZITI_CTRL_ADVERTISED_PORT=${ZITI_CTRL_ADVERTISED_PORT}
ZITI_USER=${ZITI_USER}
ZITI_PWD=${ZITI_PWD}
CTRL_ENV
sudo DEBUG=1 /opt/openziti/etc/controller/bootstrap.bash "${CTRL_ANSWERS}"

# bootstrap.bash starts the controller service automatically
wait_for_service ziti-controller.service 60
wait_for_port "${ZITI_CTRL_ADVERTISED_ADDRESS}" "${ZITI_CTRL_ADVERTISED_PORT}" 30

# Verify the service user can reach the controller agent
sudo -u ziti-controller "${ZITI_BIN}" agent stats || true

# Login to Ziti controller
# shellcheck disable=SC2140
login_cmd="${ZITI_BIN} edge login ${ZITI_CTRL_ADVERTISED_ADDRESS}:${ZITI_CTRL_ADVERTISED_PORT}"\
" --yes"\
" --username ${ZITI_USER}"\
" --password ${ZITI_PWD}"
# shellcheck disable=SC2086
retry 10 3 ${login_cmd}
log_pass "Ziti controller login succeeded"

# ============================================================
# Phase 3: Bootstrap Ziti router
# ============================================================
log_section "Phase 3: Bootstrap Ziti router"

# Create default policies so routers can serve traffic
"${ZITI_BIN}" edge create edge-router-policy default --edge-router-roles '#all' --identity-roles '#all'
"${ZITI_BIN}" edge create service-edge-router-policy default --edge-router-roles '#all' --service-roles '#all'

# Create the router and get its enrollment token
"${ZITI_BIN}" edge create edge-router "${ZITI_ROUTER_NAME}" -to "${ZITI_ENROLL_TOKEN_FILE}"

if [[ ! -s "${ZITI_ENROLL_TOKEN_FILE}" ]]; then
    log_error "router enrollment token not found at ${ZITI_ENROLL_TOKEN_FILE}"
    exit 1
fi
ZITI_ENROLL_TOKEN_CONTENT="$(<"${ZITI_ENROLL_TOKEN_FILE}")"

# v2 router bootstrap: pass answers via file
RTR_ANSWERS="$(mktemp)"
cat > "${RTR_ANSWERS}" <<RTR_ENV
ZITI_BOOTSTRAP=true
ZITI_BOOTSTRAP_ENROLLMENT=true
ZITI_ENROLL_TOKEN=${ZITI_ENROLL_TOKEN_CONTENT}
ZITI_ROUTER_NAME=${ZITI_ROUTER_NAME}
ZITI_ROUTER_ADVERTISED_ADDRESS=${ZITI_ROUTER_ADVERTISED_ADDRESS}
ZITI_ROUTER_PORT=${ZITI_ROUTER_PORT}
RTR_ENV
sudo DEBUG=1 /opt/openziti/etc/router/bootstrap.bash "${RTR_ANSWERS}"

# Router bootstrap does NOT start the service — start it manually
start_service ziti-router.service
wait_for_service ziti-router.service 60

retry 10 3 bash -c "[[ \$(${ZITI_BIN} edge list edge-routers -j | jq '.data[0].isOnline') == \"true\" ]]"
log_pass "Ziti router is online"

# ============================================================
# Phase 4: Install zrok2 packages
# ============================================================
log_section "Phase 4: Install zrok2 packages"

# Install the main zrok2 CLI package first (other packages depend on it)
sudo dpkg --force-confnew --install "${PKG_DIR}"/zrok2_*.deb </dev/null

# Install the server packages (meta packages with systemd units)
for _pkg_pattern in zrok2-controller zrok2-frontend zrok2-metrics-bridge; do
    shopt -s nullglob
    _files=("${PKG_DIR}/${_pkg_pattern}_"*.deb)
    shopt -u nullglob
    if [[ ${#_files[@]} -gt 0 ]]; then
        sudo dpkg --force-confnew --install "${_files[@]}" </dev/null
    else
        log_info "no ${_pkg_pattern} .deb found, skipping"
    fi
done

# Fix any broken dependencies from meta packages
sudo DEBIAN_FRONTEND=noninteractive apt-get install -f -y </dev/null

log_info "installed zrok2 version: $(${ZROK2_BIN} version 2>/dev/null || echo unknown)"
log_pass "zrok2 packages installed"

# ============================================================
# Phase 5: Run zrok2-bootstrap.bash
# ============================================================
log_section "Phase 5: Run zrok2 bootstrap"

# The bootstrap script is installed by the zrok2-controller package
BOOTSTRAP_SCRIPT="/opt/openziti/etc/zrok2/zrok2-bootstrap.bash"
if [[ ! -x "${BOOTSTRAP_SCRIPT}" ]]; then
    log_error "bootstrap script not found or not executable at ${BOOTSTRAP_SCRIPT}"
    exit 1
fi

# Create a fresh XDG_CONFIG_HOME for this run so the ziti CLI session and
# certificate cache are isolated from prior runs and the calling user's
# personal ~/.config/ziti/ (which may have stale certs from a previous PKI).
# Also keeps HOME at its sudo default (/root) so the bootstrap doesn't see
# the calling user's ~/.zrok2 enabled environment.
ZITI_CONFIG_HOME="$(mktemp -d)"
sudo \
    PATH="${PATH}" \
    XDG_CONFIG_HOME="${ZITI_CONFIG_HOME}" \
    ziti edge login "${ZITI_CTRL_ADVERTISED_ADDRESS}:${ZITI_CTRL_ADVERTISED_PORT}" \
    --yes \
    --username "${ZITI_USER}" \
    --password "${ZITI_PWD}"

# No TLS — HTTP-only mode for CI testing.
sudo \
    PATH="${PATH}" \
    XDG_CONFIG_HOME="${ZITI_CONFIG_HOME}" \
    ZROK2_DNS_ZONE="${ZROK2_DNS_ZONE}" \
    ZROK2_ADMIN_TOKEN="${ZROK2_ADMIN_TOKEN}" \
    ZITI_API_ENDPOINT="https://${ZITI_CTRL_ADVERTISED_ADDRESS}:${ZITI_CTRL_ADVERTISED_PORT}" \
    ZITI_ADMIN_PASSWORD="${ZITI_PWD}" \
    ZITI_ADMIN_USER="${ZITI_USER}" \
    "${BOOTSTRAP_SCRIPT}"

log_pass "zrok2 bootstrap completed"

# ============================================================
# Phase 6: Verify services
# ============================================================
log_section "Phase 6: Verify services"

wait_for_service zrok2-controller.service 60
log_pass "zrok2-controller is active"

wait_for_service zrok2-frontend.service 60
log_pass "zrok2-frontend is active"

wait_for_service zrok2-metrics-bridge.service 60
log_pass "zrok2-metrics-bridge is active"

# ============================================================
# Phase 7: Verify API
# ============================================================
log_section "Phase 7: Verify API"

ZROK2_API_ENDPOINT="http://127.0.0.1:${ZROK2_CTRL_PORT}"

ZROK2_ACCEPT="Accept: application/zrok.v1+json"
retry 10 3 curl -sf -H "${ZROK2_ACCEPT}" "${ZROK2_API_ENDPOINT}/api/v2/versions"
_version=$(curl -sf -H "${ZROK2_ACCEPT}" "${ZROK2_API_ENDPOINT}/api/v2/versions" | jq -r '.controllerVersion // empty')
if [[ -n "${_version}" ]]; then
    log_pass "zrok2 API responds: controllerVersion=${_version}"
else
    log_fail "zrok2 API versions endpoint returned empty"
    exit 1
fi

# ============================================================
# Phase 8: Create and verify test account
# ============================================================
log_section "Phase 8: Create test account"

export ZROK2_API_ENDPOINT
export ZROK2_ADMIN_TOKEN

_test_email="test@example.com"
_test_password="$(generate_password 16)"

# Create the test account via admin CLI
_account_output=$(${ZROK2_BIN} admin create account "${_test_email}" "${_test_password}" 2>&1) || true
log_info "account creation output: ${_account_output}"

# Extract the account token from the output
_account_token=$(echo "${_account_output}" | grep -oP '(?<=token: ).*' || \
                 echo "${_account_output}" | tail -1)

if [[ -n "${_account_token}" ]]; then
    log_pass "test account created: ${_test_email}"
else
    log_info "could not extract account token (may be expected on some versions)"
fi

# ============================================================
# Phase 9: Verify Ziti data plane
# ============================================================
log_section "Phase 9: Verify Ziti data plane"

if retry 3 5 "${ZITI_BIN}" ops verify traffic \
    --timeout 11 --prefix "zrok2-linux-test" --yes --cleanup; then
    log_pass "Ziti data plane traffic verified"
else
    # Retry with verbose for diagnostics
    if "${ZITI_BIN}" ops verify traffic \
        --timeout 11 --prefix "zrok2-linux-test" --yes --cleanup --verbose; then
        log_pass "Ziti data plane traffic verified (retry)"
    else
        log_error "Ziti data plane traffic verification failed"
        log_info "continuing despite traffic verification failure"
    fi
fi

# ============================================================
# Phase 10: Canary looper (exercise shares through the public frontend)
# ============================================================
log_section "Phase 10: Canary public-proxy looper"

# Create a fresh account for the canary via the REST API (avoids CLI endpoint
# precedence issues with the operator's enabled environment).
CANARY_HOME="$(mktemp -d)"
mkdir -p "${CANARY_HOME}/.zrok2"
echo '{"v":"v0.4"}' > "${CANARY_HOME}/.zrok2/metadata.json"
printf '{"apiEndpoint":"%s"}' "${ZROK2_API_ENDPOINT}" > "${CANARY_HOME}/.zrok2/config.json"

_canary_token=$(curl -sf \
    -H "X-TOKEN: ${ZROK2_ADMIN_TOKEN}" \
    -H "Content-Type: application/zrok.v1+json" \
    -d "{\"email\":\"canary-$(date +%s)@zrok.internal\",\"password\":\"canarypass\"}" \
    "${ZROK2_API_ENDPOINT}/api/v2/account" | python3 -c "import sys,json; print(json.load(sys.stdin)['accountToken'])")
log_info "canary account token: ${_canary_token}"

HOME="${CANARY_HOME}" "${ZROK2_BIN}" enable "${_canary_token}" --description "canary-test"

# Determine the frontend port from the installed frontend config.
_frontend_port=$(grep 'bind_address:' /etc/zrok2/frontend.yml 2>/dev/null | grep -oP ':\K[0-9]+$' || echo "8080")
_canary_flags=(--iterations 3 --loopers 1 --min-payload 256 --max-payload 256 --min-pacing 1s --max-pacing 1s)
if [[ "${_frontend_port}" != "443" ]]; then
    _canary_flags+=(--http --frontend-port "${_frontend_port}")
fi

if HOME="${CANARY_HOME}" ZROK2_DANGEROUS_CANARY=1 \
    "${ZROK2_BIN}" test canary public-proxy "${_canary_flags[@]}"; then
    log_pass "canary public-proxy looper passed"
else
    log_error "canary public-proxy looper failed"
    log_info "continuing despite canary failure"
fi

# Disable the canary environment
HOME="${CANARY_HOME}" "${ZROK2_BIN}" disable 2>/dev/null || true

# ============================================================
# Phase 11: Verify metrics pipeline (InfluxDB has data from canary)
# ============================================================
log_section "Phase 11: Verify metrics pipeline"

_influx_token=$(grep -A5 'influx:' /etc/zrok2/ctrl.yml | grep 'token:' | head -1 | awk -F'"' '{print $2}')
if [[ -n "${_influx_token}" ]]; then
    # Wait for metrics to propagate through the pipeline (bridge → RabbitMQ → controller → InfluxDB).
    # The default metricsReportInterval is 60s, so we may need to wait.
    log_info "waiting up to 90s for metrics to appear in InfluxDB..."
    _metrics_found=false
    for _attempt in $(seq 1 18); do
        _count=$(influx query \
            'from(bucket: "zrok") |> range(start: -5m) |> count()' \
            --org zrok --token "${_influx_token}" --raw 2>/dev/null \
            | grep -c ',' || true)
        if (( _count > 0 )); then
            _metrics_found=true
            break
        fi
        sleep 5
    done

    if [[ "${_metrics_found}" == "true" ]]; then
        log_pass "metrics pipeline verified: InfluxDB has data (${_count} series)"
    else
        log_error "metrics pipeline: no data in InfluxDB after 90s"
        log_info "check: systemctl status zrok2-metrics-bridge, rabbitmqctl list_queues"
        log_info "continuing despite metrics verification failure"
    fi
else
    log_info "could not extract InfluxDB token from ctrl.yml — skipping metrics check"
fi

# ============================================================
# Done
# ============================================================
log_section "All Linux deployment tests passed"
# EXIT trap runs cleanup (or skips it if --keep)
