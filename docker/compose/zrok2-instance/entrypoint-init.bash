#!/usr/bin/env bash
#
# entrypoint-init.bash — Docker init container for zrok2 self-hosting
#
# Sources zrok2-bootstrap.bash for ALL function definitions (utilities AND
# step_* functions), then calls the relevant steps directly. This avoids
# duplicating the admin operations (frontend, namespace, mapping) that are
# already implemented in the bootstrap script.
#
# Init sequence:
#   1. Generate controller and frontend configuration
#   2. Wait for Ziti controller and log in
#   3. Run zrok2 admin bootstrap (creates Ziti identities/policies)
#   4. Start a temporary controller to run admin commands
#   5. Create the default public frontend, namespace, and mapping (via step_*)
#   6. Save identity files and fix permissions

set -o errexit -o nounset -o pipefail

# Source the bootstrap library — loads ALL function definitions (info, warn,
# die, retry, wait_for, step_create_frontend, step_create_namespace,
# step_map_namespace_frontend, etc.) without executing the main() workflow.
# shellcheck source=../../../nfpm/zrok2-bootstrap.bash
source /bootstrap/zrok2-bootstrap.bash

# ── Map Docker env vars to bootstrap env vars ────────────────────────────────
#
# The bootstrap step_* functions expect ZITI_ADMIN_PASSWORD, ZITI_ADMIN_USER,
# and ZITI_API_ENDPOINT. Docker compose uses ZITI_PWD, ZITI_USER, and
# constructs the endpoint from ZROK2_DNS_ZONE.

export ZITI_ADMIN_PASSWORD="${ZITI_PWD}"
export ZITI_ADMIN_USER="${ZITI_USER:-admin}"
export ZITI_API_ENDPOINT="https://ziti.${ZROK2_DNS_ZONE}:${ZITI_CTRL_PORT:-1280}"
export ZROK2_NAMESPACE_TOKEN="${ZROK2_NAMESPACE_TOKEN:-public}"

# ── Docker-specific configuration ────────────────────────────────────────────

CONFIG_DIR="${ZROK2_CONFIG_DIR:-/var/lib/zrok2/config}"
CTRL_CONFIG="${CONFIG_DIR}/ctrl.yaml"
FRONTEND_CONFIG="${CONFIG_DIR}/frontend.yaml"

ZROK2_DB_PASSWORD="${ZROK2_DB_PASSWORD:-zrok2defaultpw}"
STORE_TYPE="${ZROK2_STORE_TYPE:-postgres}"

if [[ "$STORE_TYPE" == "postgres" ]]; then
    STORE_PATH="host=postgresql port=5432 user=zrok2 password=${ZROK2_DB_PASSWORD} dbname=zrok2 sslmode=disable"
else
    STORE_PATH="/var/lib/zrok2/zrok2.sqlite3"
fi

ZROK2_PORT="${ZROK2_CTRL_PORT:-18080}"

export ZROK2_API_ENDPOINT="http://localhost:${ZROK2_PORT}"
export ZROK2_ADMIN_TOKEN

# ── Helper: commands wrapped for retry/wait_for ──────────────────────────────

_ziti_login() {
    ziti edge login "$ZITI_API_ENDPOINT" \
        --username "${ZITI_ADMIN_USER}" \
        --password "${ZITI_ADMIN_PASSWORD}" \
        --yes 2>/dev/null
}

_zrok2_bootstrap() {
    zrok2 admin bootstrap "$CTRL_CONFIG" 2>&1 \
        | tee /dev/stderr \
        | grep -qE '(bootstrap complete|already bootstrapped)'
}

_zrok2_ctrl_healthy() {
    curl -sf "http://127.0.0.1:${ZROK2_PORT}/api/v1/version" &>/dev/null
}

_zrok2_ctrl_alive() {
    kill -0 "$CTRL_PID" 2>/dev/null
}

# ── Step 1: Generate controller config ────────────────────────────────────────
#
# When ZROK2_METRICS_ENABLED=true, includes bridge (fileSource → amqpSink)
# and metrics (amqpSource → InfluxDB) sections for the metrics pipeline.

mkdir -p "$CONFIG_DIR"

AMQP_URL="${ZROK2_AMQP_URL:-amqp://guest:guest@rabbitmq:5672}"
INFLUX_URL="${ZROK2_INFLUX_URL:-http://influxdb:8086}"
INFLUX_ORG="${ZROK2_INFLUX_ORG:-zrok2}"
INFLUX_BUCKET="${ZROK2_INFLUX_BUCKET:-zrok2}"
INFLUX_TOKEN="${ZROK2_INFLUX_TOKEN:-}"
# Path to fabric-usage.json inside the metrics-bridge container (shared volume).
FABRIC_USAGE_PATH="${ZROK2_FABRIC_USAGE_PATH:-/ziti-data/fabric-usage.json}"

if [[ ! -f "$CTRL_CONFIG" ]]; then
    info "Generating $CTRL_CONFIG..."
    cat > "$CTRL_CONFIG" <<CTRLEOF
v: 4
admin:
  secrets:
    - "${ZROK2_ADMIN_TOKEN}"
endpoint:
  host: 0.0.0.0
  port: ${ZROK2_PORT}
store:
  path: "${STORE_PATH}"
  type: "${STORE_TYPE}"
ziti:
  api_endpoint: "${ZITI_API_ENDPOINT}"
  username: "${ZITI_ADMIN_USER}"
  password: "${ZITI_ADMIN_PASSWORD}"
maintenance:
  registration:
    expiration_timeout: 24h
    check_frequency: 1h
    batch_limit: 500
  reset_password:
    expiration_timeout: 15m
    check_frequency: 15m
    batch_limit: 500
CTRLEOF

    # Append metrics pipeline sections when enabled.
    if [[ "${ZROK2_METRICS_ENABLED:-false}" == "true" ]]; then
        info "Metrics enabled — adding bridge and metrics sections to $CTRL_CONFIG"
        cat >> "$CTRL_CONFIG" <<METRICSEOF
bridge:
  source:
    type: fileSource
    path: ${FABRIC_USAGE_PATH}
  sink:
    type: amqpSink
    url: ${AMQP_URL}
    queue_name: events
metrics:
  agent:
    source:
      type: amqpSource
      url: ${AMQP_URL}
      queue_name: events
  influx:
    url: "${INFLUX_URL}"
    bucket: ${INFLUX_BUCKET}
    org: ${INFLUX_ORG}
    token: "${INFLUX_TOKEN}"
METRICSEOF
    fi

    chmod 640 "$CTRL_CONFIG"
    info "Controller config written to $CTRL_CONFIG"
else
    info "Controller config already exists at $CTRL_CONFIG, skipping"
fi

# ── Step 2: Generate frontend config ─────────────────────────────────────────

if [[ ! -f "$FRONTEND_CONFIG" ]]; then
    cat > "$FRONTEND_CONFIG" <<FEEOF
v: 4
host_match: "${ZROK2_DNS_ZONE}"
address: "0.0.0.0:${ZROK2_FRONTEND_PORT:-8080}"
FEEOF
    chmod 640 "$FRONTEND_CONFIG"
    info "Frontend config written to $FRONTEND_CONFIG"
else
    info "Frontend config already exists at $FRONTEND_CONFIG, skipping"
fi

info "Generated ctrl.yaml and frontend.yaml"

# ── Step 3: Wait for Ziti controller ─────────────────────────────────────────

info "Waiting for Ziti controller..."
wait_for 300 3 "Ziti controller login" _ziti_login
info "Logged into Ziti controller"

# ── Step 4: Run zrok2 admin bootstrap ────────────────────────────────────────

info "Running zrok2 admin bootstrap..."
wait_for 120 3 "zrok2 admin bootstrap" _zrok2_bootstrap
info "zrok2 bootstrap complete"

# ── Step 5: Start temporary controller for admin commands ────────────────────

info "Starting temporary zrok2 controller for admin commands..."
zrok2 controller "$CTRL_CONFIG" &
CTRL_PID=$!

# Ensure the temporary controller is cleaned up on exit (normal or error)
trap 'kill "$CTRL_PID" 2>/dev/null; wait "$CTRL_PID" 2>/dev/null || true' EXIT

info "Waiting for temporary controller to become healthy..."
wait_for 60 1 "temporary zrok2 controller health" _zrok2_ctrl_healthy

# Sanity check — make sure the process didn't crash during startup
if ! _zrok2_ctrl_alive; then
    die "Temporary controller exited unexpectedly"
fi
info "Temporary controller is healthy"

# ── Step 6: Create frontend, namespace, and mapping ──────────────────────────
#
# These call the bootstrap's step_* functions directly — no duplication.
# The functions are idempotent and use ZROK2_DNS_ZONE, ZROK2_NAMESPACE_TOKEN,
# and FRONTEND_TOKEN (script-level var populated by step_create_frontend).

FRONTEND_TOKEN=""
retry 5 3 "create dynamic frontend" step_create_frontend
info "Frontend token: $FRONTEND_TOKEN"

retry 5 3 "create public namespace" step_create_namespace
info "Namespace '${ZROK2_NAMESPACE_TOKEN}' ready"

retry 5 3 "map namespace to frontend" step_map_namespace_frontend
info "Namespace-frontend mapping ready"

# ── Step 6b: Create dynamicProxyController identity and Ziti resources ───────
#
# The dynamic proxy controller provides real-time AMQP-based mapping updates
# to the frontend, required for named public shares.  The bootstrap library's
# step_dynamic_proxy_controller() creates the identity, Ziti service, and
# policies.  CONTROLLER_HOME must be set so the identity file lands in the
# right place.

export CONTROLLER_HOME="${HOME}"
export ZROK2_AMQP_URL="${AMQP_URL}"
export CTRL_CONFIG
retry 5 3 "create dynamicProxyController" step_dynamic_proxy_controller
info "dynamicProxyController ready"

# ── Step 7: Save frontend identity ───────────────────────────────────────────

if [[ -f "${HOME}/.zrok2/identities/public.json" ]]; then
    cp "${HOME}/.zrok2/identities/public.json" "${CONFIG_DIR}/public.json"
    info "Saved public frontend identity"
fi

# ── Step 8: Stop temporary controller ────────────────────────────────────────
# (handled by EXIT trap, but be explicit for logging)

info "Stopping temporary controller..."
kill "$CTRL_PID" 2>/dev/null || true
wait "$CTRL_PID" 2>/dev/null || true
trap - EXIT

# ── Step 8b: Generate full frontend config (AMQP-backed dynamic proxy) ────────
#
# Overwrite the simple frontend.yaml with the full v1 config that uses
# amqp_subscriber + controller (gRPC via dynamicProxyController Ziti service)
# for real-time mapping updates.  Required for named public shares.

info "Generating full AMQP-backed frontend config..."
cat > "$FRONTEND_CONFIG" <<FEEOF
v: 1

frontend_token: ${FRONTEND_TOKEN}
identity: public
bind_address: 0.0.0.0:${ZROK2_FRONTEND_PORT:-8080}
host_match: ${ZROK2_DNS_ZONE}
mapping_refresh_interval: 1m

amqp_subscriber:
  url: ${ZROK2_AMQP_URL:-amqp://guest:guest@rabbitmq:5672}
  exchange_name: dynamicProxy

controller:
  identity_path: ${CONFIG_DIR}/public.json
  service_name: dynamicProxyController
FEEOF
chmod 640 "$FRONTEND_CONFIG"
info "Full frontend config written to $FRONTEND_CONFIG"

# ── Step 9: Fix ownership for non-root zrok2 containers ─────────────────────

chown -R 2171:2171 /var/lib/zrok2

info "zrok2-init complete"
