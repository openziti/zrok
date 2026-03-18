#!/usr/bin/env bash

# Test the documented Docker Compose self-hosting deployment of zrok2.
# Builds zrok2 from source, starts the full stack, runs Python SDK
# integration tests against the live API, then tears down.
#
# Usage:
#   docker.test.bash [OPTIONS]
#
# Options:
#   --source-dir <dir>       zrok source tree root (default: auto-detected)
#   --ziti-repo <repo>       Ziti image repo prefix (default: docker.io/openziti)
#   --ziti-tag <tag>         Ziti controller+router image tag (default: latest)
#   --keep                   keep the stack running on exit (for inspection)
#   --only-clean             tear down a kept instance and exit
#
# Environment variables:
#   COMPOSE_FILE  Override compose file list (default: compose.yml:compose.build.yml)
#
# Examples:
#   bash docker.test.bash
#   bash docker.test.bash --ziti-repo docker.io/openziti --ziti-tag 1.6.14 --keep
#   bash docker.test.bash --ziti-repo docker.io/kbingham --ziti-tag latest --keep
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
Usage: docker.test.bash [OPTIONS]

  --source-dir <dir>       zrok source tree root (default: auto-detected)
  --ziti-repo <repo>       Ziti image repo prefix (default: docker.io/openziti)
  --ziti-tag <tag>         Ziti controller+router image tag (default: latest)
  --keep                   keep the stack running on exit (for inspection)
  --only-clean             tear down a kept instance (volumes included) and exit

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
    (cd "${COMPOSE_PROJECT_DIR}" && docker compose logs --tail=100) 2>/dev/null || true
}

# ============================================================
# Cleanup
# ============================================================

KEEP=0
COMPOSE_PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

teardown() {
    log_section "Tear down"
    if ! (cd "${COMPOSE_PROJECT_DIR}" && docker compose --profile metrics --profile canary down -v --remove-orphans 2>/dev/null); then
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
Usage: docker.test.bash [OPTIONS]

  --source-dir <dir>       zrok source tree root (default: auto-detected)
  --ziti-repo <repo>       Ziti image repo prefix (default: docker.io/openziti)
  --ziti-tag <tag>         Ziti controller+router image tag (default: latest)
  --keep                   keep the stack running on exit (for inspection)
  --only-clean             tear down a kept instance (volumes included) and exit

State from a prior run is always destroyed at the start via 'docker compose
down -v'. Use --keep to leave the stack running for post-mortem inspection.
Use --only-clean to tear down without re-running the test.
EOF
    exit 1
}

SOURCE_DIR=""
ZITI_REPO=""
ZITI_TAG=""
ONLY_CLEAN=0
while [[ $# -gt 0 ]]; do
    case "$1" in
        --source-dir)  SOURCE_DIR="$2"; shift 2 ;;
        --ziti-repo)   ZITI_REPO="$2"; shift 2 ;;
        --ziti-tag)    ZITI_TAG="$2"; shift 2 ;;
        --keep)        KEEP=1; shift ;;
        --only-clean)  ONLY_CLEAN=1; shift ;;
        *)             usage ;;
    esac
done

if [[ -z "${SOURCE_DIR}" ]]; then
    SOURCE_DIR="$(cd "${COMPOSE_PROJECT_DIR}/../../.." && pwd)"
fi
[[ -d "${SOURCE_DIR}" ]] || { trap - EXIT ERR; log_error "source dir '${SOURCE_DIR}' not found"; exit 1; }

export COMPOSE_FILE="${COMPOSE_FILE:-compose.yml:compose.build.yml}:compose.canary.yml"

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
# Remove all containers, networks, and named volumes from the compose project.
# Use --profile metrics to include metrics services, and --remove-orphans to
# catch services that may have been renamed or removed between runs.
if ! (cd "${COMPOSE_PROJECT_DIR}" && docker compose --profile metrics --profile canary down -v --remove-orphans 2>/dev/null); then
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
sed -i "s|^ZROK2_DNS_ZONE=.*|ZROK2_DNS_ZONE=zrok.127.0.0.1.sslip.io|" "${ENV_FILE}"
# Enable the metrics pipeline so we can verify InfluxDB receives data.
sed -i "s|^# ZROK2_METRICS_ENABLED=.*|ZROK2_METRICS_ENABLED=true|" "${ENV_FILE}"
# Pin Ziti image repo and/or tag if specified.
# --ziti-repo docker.io/kbingham → controller image is docker.io/kbingham/ziti-controller
if [[ -n "${ZITI_REPO}" ]]; then
    sed -i "s|^# ZITI_CONTROLLER_IMAGE=.*|ZITI_CONTROLLER_IMAGE=${ZITI_REPO}/ziti-controller|" "${ENV_FILE}"
    sed -i "s|^# ZITI_ROUTER_IMAGE=.*|ZITI_ROUTER_IMAGE=${ZITI_REPO}/ziti-router|" "${ENV_FILE}"
    log_info "using Ziti image repo: ${ZITI_REPO}"
fi
if [[ -n "${ZITI_TAG}" ]]; then
    # Docker image tags use hyphens (e.g., 2.0.0-rc5), not tildes.
    # The same ZITI_LINUX_VERSION variable may contain tildes from deb
    # convention (2.0.0~rc5) or hyphens — normalize to hyphens.
    ZITI_TAG="${ZITI_TAG//\~/-}"
    ZITI_TAG="${ZITI_TAG#v}"
    sed -i "s|^# ZITI_CONTROLLER_TAG=.*|ZITI_CONTROLLER_TAG=${ZITI_TAG}|" "${ENV_FILE}"
    sed -i "s|^# ZITI_ROUTER_TAG=.*|ZITI_ROUTER_TAG=${ZITI_TAG}|" "${ENV_FILE}"
    log_info "using Ziti image tag: ${ZITI_TAG}"
fi

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
                   --profile metrics \
    up -d --build --wait --wait-timeout 300)

log_pass "stack is up and healthy"

# Patch the Ziti controller config with fabric.usage event logging, then
# restart the overlay.  We exec into the running controller to append the
# events section, then stop+start (not restart) with the config preserved.
# The router must restart after the controller to re-sync signing keys.
# Patch the Ziti controller config at /ziti-controller/config.yml (the
# controller's working directory, on the ziti-ctrl-data named volume).
# Use sed insertion before known anchors — appending to the end of YAML
# files is unreliably parsed by Go YAML parsers.
log_info "patching Ziti controller with fabric.usage events and metrics reporting..."
docker compose -f "${COMPOSE_PROJECT_DIR}/compose.yml" \
    exec -T ziti-controller sh -c '
        CONFIG=/ziti-controller/config.yml
        # The volume is now mounted at /ziti-controller (the controller workdir),
        # so the events file is on the shared volume and visible to the metrics
        # bridge at /ziti-data/fabric-usage.json.
        EVENTS_PATH=/ziti-controller/fabric-usage.json

        # Insert events section before identity: if not present.
        if ! grep -q "fabric.usage" "$CONFIG"; then
            sed -i "/^identity:/i\\
\\
events:\\
  jsonLogger:\\
    subscriptions:\\
      - type: fabric.usage\\
        version: 3\\
    handler:\\
      type: file\\
      format: json\\
      path: $EVENTS_PATH\\
" "$CONFIG"
            echo "events section inserted"
        else
            # Fix the events file path if it points to the wrong location.
            sed -i "s|path:.*fabric-usage.json|path: $EVENTS_PATH|" "$CONFIG"
            echo "events path corrected"
        fi

        # Insert network section before identity: for responsive metrics.
        if ! grep -q "metricsReportInterval" "$CONFIG"; then
            sed -i "/^identity:/i\\
\\
network:\\
  intervalAgeThreshold: 5s\\
  metricsReportInterval: 5s\\
" "$CONFIG"
            echo "network metrics section inserted"
        fi

        # Pre-create the events file so the metrics bridge does not panic.
        touch "$EVENTS_PATH"
    '

log_info "restarting Ziti overlay..."
(cd "${COMPOSE_PROJECT_DIR}" && docker compose restart ziti-controller)
retry 30 3 docker compose -f "${COMPOSE_PROJECT_DIR}/compose.yml" \
    exec -T ziti-controller ziti agent stats

(cd "${COMPOSE_PROJECT_DIR}" && docker compose restart ziti-router)
retry 30 3 docker compose -f "${COMPOSE_PROJECT_DIR}/compose.yml" \
    exec -T ziti-router ziti agent stats

# Restart zrok2 services — their Ziti SDK connections are stale after the
# overlay restart (cached router IPs/sessions are invalid).
(cd "${COMPOSE_PROJECT_DIR}" && docker compose restart zrok2-controller)
retry 30 3 docker compose -f "${COMPOSE_PROJECT_DIR}/compose.yml" \
    exec -T zrok2-controller curl -sf -o /dev/null \
    -H "Accept: application/zrok.v1+json" http://127.0.0.1:18080/api/v1/version
(cd "${COMPOSE_PROJECT_DIR}" && docker compose restart zrok2-frontend)
# Restart the metrics bridge so it picks up the now-existing events file.
(cd "${COMPOSE_PROJECT_DIR}" && docker compose --profile metrics restart zrok2-metrics-bridge) 2>/dev/null || true
log_pass "stack restarted with events config"

# ============================================================
# Phase 3: Verify API
# ============================================================

log_section "Phase 3: Verify API"

ZROK2_ACCEPT="Accept: application/zrok.v1+json"
retry 10 3 curl -sf -H "${ZROK2_ACCEPT}" \
    "${ZROK2_API_ENDPOINT}/api/v2/versions" >/dev/null

_version=$(curl -sf -H "${ZROK2_ACCEPT}" \
    "${ZROK2_API_ENDPOINT}/api/v2/versions" | \
    grep -oP '"controllerVersion"\s*:\s*"\K[^"]+' || echo "?")
log_pass "API responded: controllerVersion=${_version}"

log_pass "stack verified — ZROK2_API_ENDPOINT=${ZROK2_API_ENDPOINT} ZROK2_ADMIN_TOKEN=${ZROK2_ADMIN_TOKEN}"

# ============================================================
# Phase 4: Canary looper (exercise shares through the public frontend)
# ============================================================

log_section "Phase 4: Canary public-proxy looper"

# The canary runs in a host-networked container defined in compose.canary.yml.
# Host networking means *.localhost resolves to 127.0.0.1, reaching the
# frontend at its published port.  The canary service creates an account,
# enables an environment, exercises the public frontend, and disables.
(cd "${COMPOSE_PROJECT_DIR}" && docker compose run --rm canary) 2>&1
log_pass "canary public-proxy looper passed"

# ============================================================
# Phase 5: Verify metrics pipeline (InfluxDB has data from canary)
# ============================================================

log_section "Phase 5: Verify metrics pipeline"

# Query InfluxDB inside the influxdb container.  The metrics profile adds
# RabbitMQ + InfluxDB.  Canary traffic from Phase 4 should have produced
# fabric.usage events that flow through the pipeline.
log_info "waiting up to 90s for metrics to appear in InfluxDB..."
_metrics_found=false
for _attempt in $(seq 1 18); do
    _count=$(docker compose -f "${COMPOSE_PROJECT_DIR}/compose.yml" \
        exec -T influxdb influx query \
        'from(bucket: "zrok2") |> range(start: -5m) |> count()' \
        --org zrok2 --token "${ZROK2_INFLUX_TOKEN:-}" --raw 2>/dev/null \
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
    (cd "${COMPOSE_PROJECT_DIR}" && docker compose logs zrok2-metrics-bridge --tail=10) 2>&1 || true
    (cd "${COMPOSE_PROJECT_DIR}" && docker compose exec -T rabbitmq rabbitmqctl list_queues) 2>&1 || true
    docker compose -f "${COMPOSE_PROJECT_DIR}/compose.yml" \
        exec -T ziti-controller wc -l /ziti-controller/fabric-usage.json 2>&1 || true
    exit 1
fi
