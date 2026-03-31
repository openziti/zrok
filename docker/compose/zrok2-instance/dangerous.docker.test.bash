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
# Tag the locally-built image distinctly from published images so the test
# never accidentally uses a stale published openziti/zrok2:latest.
ZROK2_CI_IMAGE="zrok2-ci-test"
ZROK2_CI_TAG="local"
echo "ZROK2_IMAGE=${ZROK2_CI_IMAGE}" >> "${ENV_FILE}"
echo "ZROK2_TAG=${ZROK2_CI_TAG}" >> "${ENV_FILE}"
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

# ============================================================
# Phase 6: Agent graceful shutdown releases named shares
# ============================================================
#
# Verifies the fix for the agent shutdown race: the agent must release its
# named share on the controller before exiting so that a subsequent agent
# start (or different agent) can re-use the same name without hitting a
# 409 shareConflict.

log_section "Phase 6: Agent graceful shutdown releases named shares"

AGENT_NAME="agent-ci-$(date +%s)"
AGENT_HOME="/tmp/agent-${AGENT_NAME}"
FRONTEND_PORT="${ZROK2_FRONTEND_PORT:-8080}"
DNS_ZONE="${ZROK2_DNS_ZONE:-zrok.127.0.0.1.sslip.io}"

# Resolve the locally-built zrok2 image from the running compose stack.
# The compose build overlay tags the image for each service; we grab the
# image used by the controller (any zrok2 service would work).
ZROK2_TEST_IMAGE=$(cd "${COMPOSE_PROJECT_DIR}" && \
    docker compose images zrok2-controller --format json 2>/dev/null \
    | jq -r '.[0] | "\(.Repository):\(.Tag)"' 2>/dev/null)
if [[ -z "${ZROK2_TEST_IMAGE}" ]]; then
    ZROK2_TEST_IMAGE="${ZROK2_IMAGE:-docker.io/openziti/zrok2}:${ZROK2_TAG:-latest}"
    log_info "could not resolve image from compose; falling back to ${ZROK2_TEST_IMAGE}"
fi
log_info "using image: ${ZROK2_TEST_IMAGE}"

# Create a test account and enable an environment inside a helper container.
# The environment state is stored on a named volume so the agent container
# can mount it.
AGENT_VOLUME="zrok2-agent-test-${AGENT_NAME}"
docker volume create "${AGENT_VOLUME}" >/dev/null

# Fix volume ownership (same pattern as compose's zrok-init service).
docker run --rm --user root \
    -v "${AGENT_VOLUME}:${AGENT_HOME}" \
    busybox chown -Rc 2171:2171 "${AGENT_HOME}/"

log_info "creating agent test account and environment..."
docker run --rm --network=host --entrypoint bash \
    -e "ZROK2_API_ENDPOINT=${ZROK2_API_ENDPOINT}" \
    -e "ZROK2_ADMIN_TOKEN=${ZROK2_ADMIN_TOKEN}" \
    -e "HOME=${AGENT_HOME}" \
    -v "${AGENT_VOLUME}:${AGENT_HOME}" \
    "${ZROK2_TEST_IMAGE}" \
    -c "
        set -euo pipefail
        TOKEN=\$(zrok2 admin create account \
            'agent-${AGENT_NAME}@zrok.internal' 'agentpass')
        ZROK2_ENABLE_TOKEN=\"\${TOKEN}\" zrok2-enable
        zrok2 create name '${AGENT_NAME}'
    "
log_pass "agent account enabled, name '${AGENT_NAME}' created"

# Start the agent with a named share in a detached container.
AGENT_CONTAINER="zrok2-agent-test-${AGENT_NAME}"
AGENT_BACKEND_PORT=19191

# Start a test HTTP endpoint in the agent container so the share has a
# reachable backend.  Bind to all interfaces so it's reachable from the
# overlay network, not just loopback.
log_info "starting agent with test endpoint and named share '${AGENT_NAME}'..."
docker run -d --name "${AGENT_CONTAINER}" --network=host --entrypoint bash \
    -e "ZROK2_API_ENDPOINT=${ZROK2_API_ENDPOINT}" \
    -e "HOME=${AGENT_HOME}" \
    -v "${AGENT_VOLUME}:${AGENT_HOME}" \
    "${ZROK2_TEST_IMAGE}" \
    -c "
        set -euo pipefail
        rm -f '${AGENT_HOME}/.zrok2/agent.socket'

        # Start a test HTTP backend on all interfaces.
        zrok2 test endpoint --address 0.0.0.0 --port ${AGENT_BACKEND_PORT} &

        zrok2 agent start &
        AGENT_PID=\$!
        # Forward SIGTERM to the agent so Shutdown() runs on docker stop.
        trap 'kill -TERM \${AGENT_PID} 2>/dev/null; wait \${AGENT_PID} 2>/dev/null' TERM
        sleep 3
        zrok2 share public \
            --name-selection 'public:${AGENT_NAME}' \
            --headless \
            'http://127.0.0.1:${AGENT_BACKEND_PORT}'
        wait \${AGENT_PID}
    "

# Wait for the share to appear on the frontend.
log_info "waiting for agent share '${AGENT_NAME}' to be routable..."
_share_found=false
for _attempt in $(seq 1 30); do
    if curl -sf \
        -H "Host: ${AGENT_NAME}.${DNS_ZONE}" \
        "http://127.0.0.1:${FRONTEND_PORT}/" 2>/dev/null \
        | grep -q "zrok"; then
        _share_found=true
        break
    fi
    sleep 2
done

if [[ "${_share_found}" != "true" ]]; then
    log_error "agent share '${AGENT_NAME}' not routable after 60s"
    docker logs "${AGENT_CONTAINER}" --tail=30 2>&1 || true
    docker rm -f "${AGENT_CONTAINER}" 2>/dev/null || true
    docker volume rm "${AGENT_VOLUME}" 2>/dev/null || true
    exit 1
fi
log_pass "agent share '${AGENT_NAME}' is routable"

# Gracefully stop the agent container. The Shutdown() method should block
# until deleteShare completes on the controller.
log_info "stopping agent container gracefully (10s grace)..."
docker stop --timeout 10 "${AGENT_CONTAINER}"
docker rm "${AGENT_CONTAINER}" 2>/dev/null || true
log_info "agent container stopped"

# Verify the name is released: create a throwaway share with the same name.
# If the agent's Shutdown didn't release, this will 409.
# Verify the name is released: try to create a new share with the same name.
# If the agent's Shutdown didn't release, the POST /share returns 409.
# We run the share in the background, wait briefly for the API call to
# succeed (or fail), then check if the process is still alive.  If it exited
# immediately with an error, the name wasn't released.
log_info "verifying name '${AGENT_NAME}' is released on controller..."
VERIFY_CONTAINER="zrok2-verify-release-${AGENT_NAME}"
docker run -d --name "${VERIFY_CONTAINER}" --network=host --entrypoint bash \
    -e "ZROK2_API_ENDPOINT=${ZROK2_API_ENDPOINT}" \
    -e "HOME=${AGENT_HOME}" \
    -v "${AGENT_VOLUME}:${AGENT_HOME}" \
    "${ZROK2_TEST_IMAGE}" \
    -c "
        set -euo pipefail
        zrok2 share public 'http://127.0.0.1:${AGENT_BACKEND_PORT}' \
            --name-selection 'public:${AGENT_NAME}' \
            --backend-mode proxy --headless
    "
# Give the share API call a few seconds to succeed or 409.
sleep 5
if docker inspect "${VERIFY_CONTAINER}" --format '{{.State.Running}}' 2>/dev/null | grep -q true; then
    log_pass "name '${AGENT_NAME}' is available after agent shutdown (no 409)"
    docker stop --time 2 "${VERIFY_CONTAINER}" 2>/dev/null || true
else
    _verify_exit=$(docker inspect "${VERIFY_CONTAINER}" --format '{{.State.ExitCode}}' 2>/dev/null || echo "unknown")
    log_error "name '${AGENT_NAME}' NOT released — share exited with code ${_verify_exit}"
    docker logs "${VERIFY_CONTAINER}" --tail=10 2>&1 || true
    docker rm "${VERIFY_CONTAINER}" 2>/dev/null || true
    docker volume rm "${AGENT_VOLUME}" 2>/dev/null || true
    exit 1
fi
docker rm "${VERIFY_CONTAINER}" 2>/dev/null || true

# Restart the agent and verify it reloads the share from the registry.
log_info "restarting agent to verify registry reload..."
AGENT_CONTAINER2="zrok2-agent-test2-${AGENT_NAME}"
docker run -d --name "${AGENT_CONTAINER2}" --network=host --entrypoint bash \
    -e "ZROK2_API_ENDPOINT=${ZROK2_API_ENDPOINT}" \
    -e "HOME=${AGENT_HOME}" \
    -v "${AGENT_VOLUME}:${AGENT_HOME}" \
    "${ZROK2_TEST_IMAGE}" \
    -c "
        set -euo pipefail
        rm -f '${AGENT_HOME}/.zrok2/agent.socket'
        zrok2 test endpoint --address 0.0.0.0 --port ${AGENT_BACKEND_PORT} &
        zrok2 agent start &
        AGENT_PID=\$!
        trap 'kill -TERM \${AGENT_PID} 2>/dev/null; wait \${AGENT_PID} 2>/dev/null' TERM
        wait \${AGENT_PID}
    "

# Wait for the agent to reload the share from its registry.
log_info "waiting for agent to reload share '${AGENT_NAME}' from registry..."
_reload_ok=false
for _attempt in $(seq 1 30); do
    if curl -sf \
        -H "Host: ${AGENT_NAME}.${DNS_ZONE}" \
        "http://127.0.0.1:${FRONTEND_PORT}/" 2>/dev/null \
        | grep -q "zrok"; then
        _reload_ok=true
        break
    fi
    sleep 2
done

docker stop --timeout 10 "${AGENT_CONTAINER2}" 2>/dev/null || true
docker rm "${AGENT_CONTAINER2}" 2>/dev/null || true

if [[ "${_reload_ok}" == "true" ]]; then
    log_pass "agent reloaded share '${AGENT_NAME}' from registry after restart"
else
    log_error "agent failed to reload share '${AGENT_NAME}' after restart"
    docker logs "${AGENT_CONTAINER2}" --tail=30 2>&1 || true
fi

# Clean up the test volume.
docker volume rm "${AGENT_VOLUME}" 2>/dev/null || true
log_pass "agent graceful shutdown test passed"
