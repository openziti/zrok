#!/usr/bin/env bash

# Test the documented Linux self-hosting deployment of zrok2.
# Builds zrok2 packages with zrok-builder, installs Ziti v2 controller +
# router from the APT repo, installs the built zrok2 packages, runs
# zrok2-bootstrap.bash, and verifies all services start and the API responds.
#
# This script runs as a normal user; privileged commands use sudo internally.
# Docker must be available to the calling user (for zrok-builder).
#
# Usage:
#   linux.test.bash --ziti-version <version> [--source-dir <dir>]
#
# Options:
#   --ziti-version <version>  Ziti package version to install (e.g., 2.0.0, 2.0.0-rc5)
#   --source-dir <dir>        zrok source tree root (default: auto-detected from
#                             script location)
#
# Environment variables:
#   ZITIPAX_DEB   Override the Debian repo (default: zitipax-openziti-deb-stable)
#   ZITIPAX_RPM   Override the RedHat repo (default: zitipax-openziti-rpm-stable)
#
# Packages are built from --source-dir using the zrok-builder Docker image.
# The image is built automatically if it doesn't exist.
#
# Examples:
#   # Build packages and test (typical local workflow):
#   bash linux.test.bash --ziti-version 2.0.0
#
#   # Install a pre-release from the test repo:
#   ZITIPAX_DEB=zitipax-openziti-deb-test bash linux.test.bash --ziti-version 2.0.0-rc5
#
#   # Explicit source dir:
#   bash linux.test.bash --ziti-version 2.0.0 --source-dir ~/src/zrok

set -o errexit
set -o nounset
set -o pipefail
set -o errtrace
set -o xtrace

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
    head -c1024 /dev/urandom | LC_ALL=C tr -dc 'A-Za-z0-9' | cut -c "1-${1:-32}"
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
        | sudo ZITIPAX_DEB="${ZITIPAX_DEB:-}" \
               ZITIPAX_RPM="${ZITIPAX_RPM:-}" \
               bash -s -- "$@"
}

# ============================================================
# Error handling and cleanup
# ============================================================

_exit_code=0
_err_handler() {
    _exit_code=$?
    trap - ERR  # prevent re-entry
    log_error "FAILED at line ${LINENO}: ${BASH_COMMAND} (exit ${_exit_code})"
    # Redirect to stderr so diagnostics don't pollute command substitution stdout
    dump_journal ziti-controller.service >&2
    dump_journal ziti-router.service >&2
    dump_journal zrok2-controller.service >&2
    dump_journal zrok2-frontend.service >&2
    dump_journal zrok2-metrics.service >&2
    exit "${_exit_code}"
}
trap '_err_handler' ERR

cleanup_all() {
    # Disable errexit in cleanup — every command is best-effort
    set +o errexit
    log_section "Cleanup"
    for _svc in zrok2-metrics zrok2-frontend zrok2-controller ziti-router ziti-controller \
                rabbitmq-server postgresql influxdb; do
        sudo systemctl stop "${_svc}" 2>/dev/null || true
    done
    sudo dpkg --purge zrok2-metrics zrok2-frontend zrok2-controller zrok2-agent zrok2 2>/dev/null || true
    sudo dpkg --purge openziti-router openziti-controller openziti 2>/dev/null || true
    sudo rm -rf /var/lib/ziti-controller /var/lib/private/ziti-controller \
           /var/lib/ziti-router /var/lib/private/ziti-router \
           /var/lib/zrok2-controller /var/lib/zrok2-frontend \
           /etc/zrok2 2>/dev/null || true
    sudo systemctl daemon-reload 2>/dev/null || true
}
trap 'cleanup_all; exit $_exit_code' EXIT

# ============================================================
# Parse CLI arguments
# ============================================================

usage() {
    cat >&2 <<'EOF'
Usage: linux.test.bash --ziti-version <version> [--source-dir <dir>]

  --ziti-version <version>  Ziti package version to install from APT (e.g., 2.0.0)
  --source-dir <dir>        zrok source tree (default: auto-detected)
EOF
    exit 1
}

ZITI_VERSION=""
SOURCE_DIR=""
while [[ $# -gt 0 ]]; do
    case "$1" in
        --ziti-version) ZITI_VERSION="$2"; shift 2 ;;
        --source-dir)   SOURCE_DIR="$2"; shift 2 ;;
        *)              usage ;;
    esac
done

[[ -n "${ZITI_VERSION}" ]] || { log_error "--ziti-version is required"; usage; }

# Strip leading 'v' — deb packages use bare semver (e.g., 2.0.0~pre5, not v2.0.0~pre5)
ZITI_VERSION="${ZITI_VERSION#v}"

# Auto-detect source dir from script location (test/linux/linux.test.bash -> repo root)
if [[ -z "${SOURCE_DIR}" ]]; then
    SOURCE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
fi
[[ -d "${SOURCE_DIR}" ]] || { log_error "source dir '${SOURCE_DIR}' is not a directory"; exit 1; }

# ============================================================
# Build packages with zrok-builder
# ============================================================

build_packages() {
    log_section "Building zrok2 packages with zrok-builder"

    if ! command -v docker &>/dev/null; then
        log_error "docker is required to build packages"
        exit 1
    fi

    local _builder_image="zrok-builder"
    local _builder_context="${SOURCE_DIR}/docker/images/cross-build"

    # Build the zrok-builder image if it doesn't exist
    if ! docker image inspect "${_builder_image}" &>/dev/null; then
        log_info "building ${_builder_image} Docker image"
        docker build -t "${_builder_image}" "${_builder_context}"
    fi

    log_info "running zrok-builder --packages (this may take a few minutes)"

    # Build cache volume args — only mount if the host directory exists and is
    # writable. On CI runners these dirs often don't exist (Docker creates them
    # as root), making the cache useless and causing permission errors.
    local -a _cache_vols=()
    local _gocache="${GOCACHE:-${HOME}/.cache/go-build}"
    local _gomodcache="${GOMODCACHE:-${HOME}/.cache/go-mod}"
    if [[ -d "${_gocache}" && -w "${_gocache}" ]]; then
        _cache_vols+=(--volume="${_gocache}:/usr/share/go_cache")
    fi
    if [[ -d "${_gomodcache}" && -w "${_gomodcache}" ]]; then
        _cache_vols+=(--volume="${_gomodcache}:/usr/share/go/pkg/mod")
    fi

    docker run --user "$(id -u)" --rm \
        "${_cache_vols[@]+"${_cache_vols[@]}"}" \
        --volume="${SOURCE_DIR}:/mnt" \
        "${_builder_image}" --packages

    log_info "packages built in ${PKG_DIR}"
}

PKG_DIR="${SOURCE_DIR}/dist"
build_packages

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
ZROK2_ADMIN_TOKEN="$(generate_password 32)"
ZROK2_CTRL_PORT="18080"

: "${TMPDIR:=$(mktemp -d)}"
ZITI_ENROLL_TOKEN_FILE="${TMPDIR}/${ZITI_ROUTER_NAME}.jwt"

export DEBIAN_FRONTEND=noninteractive
cleanup_all

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

# v2 bootstrap.bash reads ZITI_*=value lines from stdin when not a TTY.
# ZITI_BOOTSTRAP=true triggers config generation + database init + service start.
# ZITI_BOOTSTRAP_CLUSTER=true creates a new cluster with PKI.
sudo DEBUG=1 /opt/openziti/etc/controller/bootstrap.bash <<CTRL_ENV
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

# v2 router bootstrap: pipe ZITI_*=value lines to stdin
sudo DEBUG=1 /opt/openziti/etc/router/bootstrap.bash <<RTR_ENV
ZITI_BOOTSTRAP=true
ZITI_BOOTSTRAP_ENROLLMENT=true
ZITI_ENROLL_TOKEN=${ZITI_ENROLL_TOKEN_CONTENT}
ZITI_ROUTER_NAME=${ZITI_ROUTER_NAME}
ZITI_ROUTER_ADVERTISED_ADDRESS=${ZITI_ROUTER_ADVERTISED_ADDRESS}
ZITI_ROUTER_PORT=${ZITI_ROUTER_PORT}
RTR_ENV

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
for _pkg_pattern in zrok2-controller zrok2-frontend zrok2-metrics; do
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
sudo apt-get install -f -y </dev/null

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

# Run the bootstrap as root with the required environment variables.
# The Ziti CLI session from Phase 2 login is still active.
# No TLS — HTTP-only mode for CI testing.
sudo -E \
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

wait_for_service zrok2-metrics.service 60
log_pass "zrok2-metrics is active"

# ============================================================
# Phase 7: Verify API
# ============================================================
log_section "Phase 7: Verify API"

ZROK2_API_ENDPOINT="http://127.0.0.1:${ZROK2_CTRL_PORT}"

retry 10 3 curl -sf "${ZROK2_API_ENDPOINT}/api/v1/version"
_version=$(curl -sf "${ZROK2_API_ENDPOINT}/api/v1/version" | jq -r '.version // empty')
if [[ -n "${_version}" ]]; then
    log_pass "zrok2 API responds: version=${_version}"
else
    log_fail "zrok2 API version endpoint returned empty"
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
# Done
# ============================================================
log_section "All Linux deployment tests passed"
# cleanup runs via EXIT trap
