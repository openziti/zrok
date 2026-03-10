#!/usr/bin/env bash
#
# entrypoint-init.bash — Docker init container for zrok2 self-hosting
#
# Sources zrok2-bootstrap.bash for utility functions (info, warn, die) and
# handles the complete init sequence:
#   1. Generate controller and frontend configuration
#   2. Run zrok2 admin bootstrap (creates Ziti identities/policies)
#   3. Start a temporary controller to run admin commands
#   4. Create the default public frontend, namespace, and mapping
#   5. Save identity files and fix permissions
#
# This replaces the previous zrok2-init and zrok2-post-init inline scripts
# with a single init container.

set -o errexit -o nounset -o pipefail

# Source the bootstrap library for utility functions (info, warn, die)
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

export ZROK2_API_ENDPOINT="http://localhost:${ZROK2_CTRL_PORT:-18080}"
export ZROK2_ADMIN_TOKEN

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
  port: ${ZROK2_CTRL_PORT:-18080}
store:
  path: "${STORE_PATH}"
  type: "${STORE_TYPE}"
ziti:
  api_endpoint: "https://ziti.${ZROK2_DNS_ZONE}:${ZITI_CTRL_PORT:-1280}"
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
until ziti edge login \
    "https://ziti.${ZROK2_DNS_ZONE}:${ZITI_CTRL_PORT:-1280}" \
    --username "${ZITI_USER:-admin}" \
    --password "${ZITI_PWD}" \
    --yes 2>/dev/null; do
    sleep 3
done
info "Logged into Ziti controller"

# ── Step 4: Run zrok2 admin bootstrap (retries until DB ready) ───────────────

info "Running zrok2 admin bootstrap..."
until zrok2 admin bootstrap "$CTRL_CONFIG" 2>&1 \
    | tee /dev/stderr \
    | grep -qE '(bootstrap complete|already bootstrapped)'; do
    info "Bootstrap not ready, retrying in 3s..."
    sleep 3
done
info "zrok2 bootstrap complete"

# ── Step 5: Start temporary controller for admin commands ────────────────────

info "Starting temporary zrok2 controller for admin commands..."
zrok2 controller "$CTRL_CONFIG" &
CTRL_PID=$!

info "Waiting for temporary controller to become healthy..."
for _i in $(seq 1 30); do
    if curl -sf "http://127.0.0.1:${ZROK2_CTRL_PORT:-18080}/api/v1/version" &>/dev/null; then
        break
    fi
    if ! kill -0 "$CTRL_PID" 2>/dev/null; then
        die "Temporary controller exited unexpectedly"
    fi
    sleep 1
done
info "Temporary controller is healthy"

# ── Step 6: Create frontend, namespace, and mapping ──────────────────────────

# Get the Ziti identity ID for "public" (created during bootstrap)
PUBLIC_ZID=$(ziti edge list identities 'name="public"' -j \
    | jq -r '.data[0].id')
info "Public identity Ziti ID = $PUBLIC_ZID"

# Create the dynamic frontend (idempotent)
EXISTING=$(zrok2 admin list frontends 2>/dev/null || true)
if echo "$EXISTING" | grep -q 'public'; then
    FRONTEND_TOKEN=$(echo "$EXISTING" | awk '/public/ {print $1; exit}')
    info "Frontend already exists: $FRONTEND_TOKEN"
else
    OUTPUT=$(zrok2 admin create frontend \
        --dynamic "$PUBLIC_ZID" public \
        "https://{token}.${ZROK2_DNS_ZONE}" 2>&1)
    FRONTEND_TOKEN=$(echo "$OUTPUT" \
        | grep -oP "(?<=frontend ').*(?=')" || true)
    if [[ -z "$FRONTEND_TOKEN" ]]; then
        FRONTEND_TOKEN=$(zrok2 admin list frontends 2>/dev/null \
            | awk '/public/ {print $1; exit}')
    fi
    info "Created frontend: $FRONTEND_TOKEN"
fi

# Create the public namespace (idempotent)
EXISTING_NS=$(zrok2 admin list namespaces 2>/dev/null || true)
if echo "$EXISTING_NS" | grep -q 'public'; then
    info "Namespace 'public' already exists"
else
    zrok2 admin create namespace \
        --open --token public "${ZROK2_DNS_ZONE}"
    info "Created namespace 'public'"
fi

# Map namespace to frontend (idempotent)
EXISTING_MAP=$(zrok2 admin list namespace-frontend public \
    2>/dev/null || true)
if echo "$EXISTING_MAP" | grep -q "$FRONTEND_TOKEN"; then
    info "Namespace-frontend mapping already exists"
else
    zrok2 admin create namespace-frontend \
        --default public "$FRONTEND_TOKEN"
    info "Mapped namespace 'public' to frontend '$FRONTEND_TOKEN'"
fi

# ── Step 7: Save frontend identity ───────────────────────────────────────────

if [[ -f "${HOME}/.zrok2/identities/public.json" ]]; then
    cp "${HOME}/.zrok2/identities/public.json" "${CONFIG_DIR}/public.json"
    info "Saved public frontend identity"
fi

# ── Step 8: Stop temporary controller ────────────────────────────────────────

info "Stopping temporary controller..."
kill "$CTRL_PID" 2>/dev/null || true
wait "$CTRL_PID" 2>/dev/null || true

# ── Step 9: Fix ownership for non-root zrok2 containers ─────────────────────

chown -R 2171:2171 /var/lib/zrok2

info "zrok2-init complete"
