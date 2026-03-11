#!/usr/bin/env bash
#
# entrypoint-init.bash — Docker init container for zrok2 self-hosting
#
# Sources zrok2-bootstrap.bash for utility functions (info, warn, die,
# retry, wait_for) and handles the complete init sequence:
#   1. Generate controller and frontend configuration
#   2. Wait for Ziti controller and log in
#   3. Run zrok2 admin bootstrap (creates Ziti identities/policies)
#   4. Start a temporary controller to run admin commands
#   5. Create the default public frontend, namespace, and mapping
#   6. Save identity files and fix permissions
#
# This replaces the previous zrok2-init and zrok2-post-init inline scripts
# with a single init container.

set -o errexit -o nounset -o pipefail

# Source the bootstrap library for utility functions
# shellcheck source=../../../nfpm/zrok2-bootstrap.bash
source /bootstrap/zrok2-bootstrap.bash

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

ZITI_ENDPOINT="https://ziti.${ZROK2_DNS_ZONE}:${ZITI_CTRL_PORT:-1280}"
ZROK2_PORT="${ZROK2_CTRL_PORT:-18080}"

export ZROK2_API_ENDPOINT="http://localhost:${ZROK2_PORT}"
export ZROK2_ADMIN_TOKEN

# ── Helper: commands wrapped for retry/wait_for ──────────────────────────────

_ziti_login() {
    ziti edge login "$ZITI_ENDPOINT" \
        --username "${ZITI_USER:-admin}" \
        --password "${ZITI_PWD}" \
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

# ── Step 1: Generate controller config (Docker-minimal) ─────────────────────
#
# Omits bridge, dynamic_proxy_controller, and metrics sections — those
# require local AMQP/InfluxDB that are external containers in Docker.

mkdir -p "$CONFIG_DIR"

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
  api_endpoint: "${ZITI_ENDPOINT}"
  username: "${ZITI_USER:-admin}"
  password: "${ZITI_PWD}"
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

# Get the Ziti identity ID for "public" (created during bootstrap)
_get_public_zid() {
    local zid
    zid=$(ziti edge list identities 'name="public"' -j 2>/dev/null \
        | jq -r '.data[0].id')
    [[ -n "$zid" && "$zid" != "null" ]]
}

retry 5 3 "Ziti identity lookup for 'public'" _get_public_zid
PUBLIC_ZID=$(ziti edge list identities 'name="public"' -j | jq -r '.data[0].id')
info "Public identity Ziti ID = $PUBLIC_ZID"

# Create the dynamic frontend (idempotent)
_create_frontend() {
    local existing
    existing=$(zrok2 admin list frontends 2>/dev/null || true)
    if echo "$existing" | grep -q 'public'; then
        FRONTEND_TOKEN=$(echo "$existing" | awk '/public/ {print $1; exit}')
        info "Frontend already exists: $FRONTEND_TOKEN"
        return 0
    fi

    local output
    output=$(zrok2 admin create frontend \
        --dynamic "$PUBLIC_ZID" public \
        "https://{token}.${ZROK2_DNS_ZONE}" 2>&1)
    FRONTEND_TOKEN=$(echo "$output" \
        | grep -oP "(?<=frontend ').*(?=')" || true)
    if [[ -z "$FRONTEND_TOKEN" ]]; then
        FRONTEND_TOKEN=$(zrok2 admin list frontends 2>/dev/null \
            | awk '/public/ {print $1; exit}')
    fi
    [[ -n "$FRONTEND_TOKEN" ]]
}

FRONTEND_TOKEN=""
retry 5 3 "create dynamic frontend" _create_frontend
info "Frontend token: $FRONTEND_TOKEN"

# Create the public namespace (idempotent)
_create_namespace() {
    local existing
    existing=$(zrok2 admin list namespaces 2>/dev/null || true)
    if echo "$existing" | grep -q 'public'; then
        info "Namespace 'public' already exists"
        return 0
    fi
    zrok2 admin create namespace \
        --open --token public "${ZROK2_DNS_ZONE}"
}

retry 5 3 "create public namespace" _create_namespace
info "Namespace 'public' ready"

# Map namespace to frontend (idempotent)
_map_namespace_frontend() {
    local existing
    existing=$(zrok2 admin list namespace-frontend public 2>/dev/null || true)
    if echo "$existing" | grep -q "$FRONTEND_TOKEN"; then
        info "Namespace-frontend mapping already exists"
        return 0
    fi
    zrok2 admin create namespace-frontend \
        --default public "$FRONTEND_TOKEN"
}

retry 5 3 "map namespace to frontend" _map_namespace_frontend
info "Namespace-frontend mapping ready"

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

# ── Step 9: Fix ownership for non-root zrok2 containers ─────────────────────

chown -R 2171:2171 /var/lib/zrok2

info "zrok2-init complete"
