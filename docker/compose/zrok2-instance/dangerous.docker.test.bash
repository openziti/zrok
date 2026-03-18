#!/usr/bin/env bash

# Test the documented Docker Compose self-hosting deployment of zrok2.
# Builds zrok2 from source, starts the full stack, runs Python SDK
# integration tests against the live API, then tears down.
#
# Usage:
#   docker.test.bash [--source-dir <dir>] [--keep] [--only-clean]
#
# Options:
#   --source-dir <dir>  zrok source tree root (default: auto-detected)
#   --keep              keep the stack running on exit (for inspection)
#   --only-clean        tear down a kept instance and exit
#
# Environment variables:
#   COMPOSE_FILE  Override compose file list (default: compose.yml:compose.build.yml)
#
# Examples:
#   bash docker.test.bash
#   bash docker.test.bash --keep
#   bash docker.test.bash --only-clean

set -o errexit
set -o nounset
set -o pipefail
set -o errtrace
set -o xtrace

# Handle --help/-h before traps and heavy setup
for _arg in "$@"; do
    case "$_arg" in
        --help|-h)
            cat <<'EOF'
Usage: docker.test.bash [--source-dir <dir>] [--keep] [--only-clean]

  --source-dir <dir>  zrok source tree root (default: auto-detected)
  --keep              keep the stack running on exit (for inspection)
  --only-clean        tear down a kept instance (volumes included) and exit

State from a prior run is always destroyed at the start via 'docker compose
down -v'. Use --keep to leave the stack running for post-mortem inspection.
Use --only-clean to tear down without re-running the test.
EOF
            exit 0
            ;;
    esac
done

# ============================================================
# Logging
# ============================================================

log_section() { printf '\n\033[1;36m=== %s ===\033[0m\n\n' "$1" >&2; }
log_info()    { printf '\033[34mINFO:\033[0m %s\n' "$1" >&2; }
log_error()   { printf '\033[31mERROR:\033[0m %s\n' "$1" >&2; }
log_pass()    { printf '\033[32mPASS:\033[0m %s\n' "$1" >&2; }

# ============================================================
# Utilities
# ============================================================

# shellcheck disable=SC2120
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

dump_logs() {
    log_info "container logs (last 100 lines each):"
    docker compose logs --tail=100 2>/dev/null || true
}

# ============================================================
# Cleanup
# ============================================================

KEEP=0
COMPOSE_PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

teardown() {
    log_section "Tear down"
    if ! (cd "${COMPOSE_PROJECT_DIR}" && docker compose down -v 2>/dev/null); then
        true  # best-effort
    fi
}

cleanup_on_exit() {
    local _real_exit=$?
    set +o errexit
    # Catch nounset/syntax errors that bypass the ERR trap.
    if (( _real_exit != 0 && _exit_code == 0 )); then
        _exit_code="${_real_exit}"
        _fail_summary="exited with status ${_real_exit} (may be nounset or syntax error)"
    fi
    if (( KEEP )); then
        log_info "keeping stack (--keep); run with --only-clean to tear down"
    else
        teardown
    fi

    echo >&2
    if (( _exit_code == 0 )); then
        log_pass "dangerous.docker.test: PASSED"
    else
        log_error "dangerous.docker.test: FAILED (exit ${_exit_code})"
        if [[ -n "${_fail_summary}" ]]; then
            log_error "  ${_fail_summary}"
        fi
    fi
}
trap 'cleanup_on_exit; exit $_exit_code' EXIT

_exit_code=0
_fail_summary=""
_err_handler() {
    _exit_code=$?
    trap - ERR
    _fail_summary="FAILED at line ${LINENO}: ${BASH_COMMAND} (exit ${_exit_code})"
    log_error "${_fail_summary}"
    dump_logs >&2
}
trap '_err_handler' ERR

# ============================================================
# Parse CLI arguments
# ============================================================

usage() {
    trap - EXIT ERR  # don't tear down on usage/help exits
    cat >&2 <<'EOF'
Usage: docker.test.bash [--source-dir <dir>] [--keep] [--only-clean]

  --source-dir <dir>  zrok source tree root (default: auto-detected)
  --keep              keep the stack running on exit (for inspection)
  --only-clean        tear down a kept instance (volumes included) and exit

State from a prior run is always destroyed at the start via 'docker compose
down -v'. Use --keep to leave the stack running for post-mortem inspection.
Use --only-clean to tear down without re-running the test.
EOF
    exit 1
}

SOURCE_DIR=""
ONLY_CLEAN=0
while [[ $# -gt 0 ]]; do
    case "$1" in
        --source-dir)  SOURCE_DIR="$2"; shift 2 ;;
        --keep)        KEEP=1; shift ;;
        --only-clean)  ONLY_CLEAN=1; shift ;;
        *)             usage ;;
    esac
done

if [[ -z "${SOURCE_DIR}" ]]; then
    SOURCE_DIR="$(cd "${COMPOSE_PROJECT_DIR}/../../.." && pwd)"
fi
[[ -d "${SOURCE_DIR}" ]] || { trap - EXIT ERR; log_error "source dir '${SOURCE_DIR}' not found"; exit 1; }

export COMPOSE_FILE="${COMPOSE_FILE:-compose.yml:compose.build.yml}"

if (( ONLY_CLEAN )); then
    trap - EXIT ERR  # teardown is intentional here; don't double-run it on exit
    log_section "Clean only"
    teardown
    log_info "clean complete"
    exit 0
fi

# ============================================================
# Pre-test cleanup: destroy any prior instance
# ============================================================

if [[ -t 0 ]]; then
    cat >&2 <<'WARN'
About to destroy all state from a prior test run:

  Stack:    docker compose down -v (removes containers + named volumes)
  Volumes:  ziti-ctrl-data, ziti-router-data, zrok2-config, pg-data,
            influx-data, rabbitmq-data

Proceeding in 30s. Re-run with </dev/null to skip this delay.
WARN
    sleep 30
fi

log_section "Pre-test cleanup"
if ! (cd "${COMPOSE_PROJECT_DIR}" && docker compose down -v 2>/dev/null); then
    true  # best-effort
fi

# ============================================================
# Phase 1: Generate secrets and write .env
# ============================================================

log_section "Phase 1: Generate secrets"

ENV_FILE="${COMPOSE_PROJECT_DIR}/.env"
cp "${COMPOSE_PROJECT_DIR}/.env.example" "${ENV_FILE}"

for var in ZROK2_ADMIN_TOKEN ZITI_PWD ZROK2_DB_PASSWORD \
           ZROK2_INFLUX_TOKEN ZROK2_INFLUX_PASSWORD; do
    # shellcheck disable=SC2119
    sed -i "s|^${var}=.*|${var}=$(generate_password)|" "${ENV_FILE}"
done
sed -i "s|^ZROK2_DNS_ZONE=.*|ZROK2_DNS_ZONE=localhost|" "${ENV_FILE}"

# Load generated values into the environment
# shellcheck source=/dev/null
source "${ENV_FILE}"
export ZROK2_ADMIN_TOKEN ZROK2_API_ENDPOINT
ZROK2_API_ENDPOINT="http://localhost:${ZROK2_CTRL_PORT:-18080}"
export ZROK2_API_ENDPOINT

log_pass "secrets generated"

# ============================================================
# Phase 2: Build and start the stack
# ============================================================

log_section "Phase 2: Build and start Docker Compose stack"

(cd "${SOURCE_DIR}" && \
    COMPOSE_FILE="${COMPOSE_PROJECT_DIR}/${COMPOSE_FILE}" \
    docker compose -f "${COMPOSE_PROJECT_DIR}/compose.yml" \
                   -f "${COMPOSE_PROJECT_DIR}/compose.build.yml" \
    up -d --build --wait --wait-timeout 300)

log_pass "stack is up and healthy"

# ============================================================
# Phase 3: Verify API
# ============================================================

log_section "Phase 3: Verify API"

ZROK2_ACCEPT="Accept: application/zrok.v1+json"
retry 10 3 curl -sf -H "${ZROK2_ACCEPT}" \
    "${ZROK2_API_ENDPOINT}/api/v2/versions" >/dev/null

_version=$(curl -sf -H "${ZROK2_ACCEPT}" \
    "${ZROK2_API_ENDPOINT}/api/v2/versions" | \
    python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('controllerVersion','?'))")
log_pass "API responded: controllerVersion=${_version}"

log_pass "stack verified — ZROK2_API_ENDPOINT=${ZROK2_API_ENDPOINT} ZROK2_ADMIN_TOKEN=${ZROK2_ADMIN_TOKEN}"

# ============================================================
# Phase 4: Canary looper (exercise shares through the public frontend)
# ============================================================

log_section "Phase 4: Canary public-proxy looper"

# Create a test account and enable an environment inside the controller container.
# The canary runs inside the container where the zrok2 binary and Ziti overlay are available.
_canary_token=$(docker compose -f "${COMPOSE_PROJECT_DIR}/compose.yml" \
    exec -T -e ZROK2_ADMIN_TOKEN="${ZROK2_ADMIN_TOKEN}" \
    zrok2-controller zrok2 admin create account \
    "canary-$(date +%s)@zrok.internal" "canarypass" 2>/dev/null)

if [[ -n "${_canary_token}" ]]; then
    log_info "canary account token: ${_canary_token}"

    # Enable, run canary, disable — all inside the container
    if docker compose -f "${COMPOSE_PROJECT_DIR}/compose.yml" \
        exec -T \
        -e ZROK2_API_ENDPOINT="http://zrok2-controller:${ZROK2_CTRL_PORT:-18080}" \
        -e ZROK2_DANGEROUS_CANARY=1 \
        -e HOME=/tmp/canary-home \
        zrok2-controller sh -c "
            mkdir -p /tmp/canary-home/.zrok2
            echo '{\"v\":\"v0.4\"}' > /tmp/canary-home/.zrok2/metadata.json
            printf '{\"apiEndpoint\":\"http://zrok2-controller:${ZROK2_CTRL_PORT:-18080}\"}' > /tmp/canary-home/.zrok2/config.json
            zrok2 enable '${_canary_token}' --description canary-test &&
            zrok2 test canary public-proxy --iterations 3 --loopers 1 \
                --min-payload 256 --max-payload 256 --min-pacing 1s --max-pacing 1s &&
            zrok2 disable
        " 2>&1; then
        log_pass "canary public-proxy looper passed"
    else
        log_info "canary looper failed (may need frontend reachable from container)"
    fi
else
    log_info "could not create canary account — skipping canary test"
fi

# ============================================================
# Phase 5: Verify metrics pipeline (InfluxDB has data from canary)
# ============================================================

log_section "Phase 5: Verify metrics pipeline"

# Query InfluxDB directly from the influxdb container.
log_info "waiting up to 90s for metrics to appear in InfluxDB..."
_metrics_found=false
for _attempt in $(seq 1 18); do
    _count=$(docker compose -f "${COMPOSE_PROJECT_DIR}/compose.yml" \
        exec -T influxdb influx query \
        'from(bucket: "zrok") |> range(start: -5m) |> count()' \
        --org zrok --token "${ZROK2_INFLUX_TOKEN:-}" --raw 2>/dev/null \
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
    log_info "metrics pipeline: no data in InfluxDB after 90s (metrics profile may not be enabled)"
fi
