#!/usr/bin/env bash
#
# zrok2-bootstrap.bash - Idempotent bootstrap script for a zrok2 self-hosted instance
#
# This script automates the full setup of a zrok2 deployment with dynamicProxy:
#   1. Installs and configures RabbitMQ (AMQP broker)
#   2. Installs and configures PostgreSQL (default) or SQLite3 (override)
#   3. Installs and configures InfluxDB for metrics storage
#   4. Generates the controller configuration (ctrl.yml)
#   5. Runs zrok2 admin bootstrap (creates Ziti identities and policies)
#   6. Creates a dynamic frontend
#   7. Creates the dynamicProxyController Ziti identity, service, and policies
#   8. Creates and maps a public namespace to the frontend
#   9. Places identity files for systemd service users
#  10. Generates the frontend configuration (frontend.yml)
#  11. Configures TLS certificate permissions
#  12. Configures OpenZiti controller for metrics events
#  13. Installs and starts the metrics bridge service
#  14. Configures systemd overrides and starts services
#
# Idempotent: safe to re-run. Skips steps that are already complete.
#
# Prerequisites:
#   - OpenZiti controller and router running (see https://netfoundry.io/docs/openziti/category/linux/)
#   - zrok2, zrok2-controller, zrok2-frontend, and zrok2-metrics-bridge packages installed
#   - DNS wildcard record (*.example.com) pointing to this server
#   - TLS wildcard certificate (e.g., from Let's Encrypt certbot)
#   - Ziti CLI authenticated: ziti edge login <controller>:<port> -y -u admin -p <password>
#   - Environment variables set (see below)
#
# Required environment variables:
#   ZROK2_DNS_ZONE          - DNS zone for this instance (e.g., zrok.example.com)
#   ZROK2_ADMIN_TOKEN       - Admin secret for the zrok2 controller
#   ZITI_API_ENDPOINT       - Ziti controller management API (e.g., https://127.0.0.1:1280)
#   ZITI_ADMIN_USER         - Ziti admin username (default: admin)
#   ZITI_ADMIN_PASSWORD     - Ziti admin password
#
# Optional environment variables:
#   ZROK2_CTRL_PORT         - zrok2 controller listen port (default: 18080)
#   ZROK2_TLS_CERT          - Path to TLS fullchain.pem (enables HTTPS)
#   ZROK2_TLS_KEY           - Path to TLS privkey.pem (enables HTTPS)
#   ZROK2_NAMESPACE_TOKEN   - Namespace token for the public namespace (default: public)
#   ZROK2_AMQP_URL          - AMQP broker URL (default: amqp://guest:guest@127.0.0.1:5672)
#   ZROK2_STORE_TYPE        - Database backend: "postgres" (default) or "sqlite3"
#   ZROK2_DB_NAME           - PostgreSQL database name (default: zrok2)
#   ZROK2_DB_USER           - PostgreSQL user name (default: zrok2)
#   ZROK2_DB_PASSWORD       - PostgreSQL password (auto-generated if not set)
#   ZROK2_INFLUX_URL        - InfluxDB URL (default: http://127.0.0.1:8086)
#   ZROK2_INFLUX_ORG        - InfluxDB organization (default: zrok)
#   ZROK2_INFLUX_BUCKET     - InfluxDB bucket (default: zrok)
#   ZROK2_INFLUX_TOKEN      - InfluxDB admin token (auto-generated if not set)
#   ZROK2_INFLUX_PASSWORD   - InfluxDB admin password (auto-generated if not set)
#   ZITI_CTRL_CONFIG        - Path to Ziti controller config (default: auto-detected)
#

# ── Utility functions (always available when sourced) ─────────────────────────

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

info()  { printf "${GREEN}[INFO]${NC} %s\n" "$*"; }
warn()  { printf "${YELLOW}[WARN]${NC} %s\n" "$*"; }
error() { printf "${RED}[ERROR]${NC} %s\n" "$*" >&2; }
die()   { error "$@"; exit 1; }

# Retry a command with backoff.  Usage:
#   retry <max_attempts> <delay_seconds> <description> <command...>
# Retries up to max_attempts times with a fixed delay between attempts.
# Returns the command's exit code on success, or dies after exhausting retries.
retry() {
    local max_attempts="$1" delay="$2" description="$3"
    shift 3
    local attempt=1
    while true; do
        if "$@"; then
            return 0
        fi
        if (( attempt >= max_attempts )); then
            die "$description failed after $max_attempts attempts"
        fi
        warn "$description failed (attempt $attempt/$max_attempts), retrying in ${delay}s..."
        sleep "$delay"
        (( attempt++ ))
    done
}

# Wait for a condition to become true.  Usage:
#   wait_for <timeout_seconds> <poll_interval> <description> <command...>
# Polls the command every poll_interval seconds until it succeeds or the
# timeout is reached.  Returns 0 on success, dies on timeout.
wait_for() {
    local timeout="$1" interval="$2" description="$3"
    shift 3
    local deadline=$(( SECONDS + timeout ))
    while (( SECONDS < deadline )); do
        if "$@"; then
            return 0
        fi
        sleep "$interval"
    done
    die "$description timed out after ${timeout}s"
}

# ── OS family detection and package helpers ──────────────────────────────────

os_family() {
    if [[ -f /etc/os-release ]]; then
        # shellcheck source=/dev/null
        . /etc/os-release
        case "${ID_LIKE:-$ID}" in
            *debian*|debian|ubuntu) echo "debian" ;;
            *rhel*|*fedora*|rhel|centos|almalinux|rocky) echo "redhat" ;;
            *) echo "unknown" ;;
        esac
    elif command -v apt-get &>/dev/null; then
        echo "debian"
    elif command -v dnf &>/dev/null || command -v yum &>/dev/null; then
        echo "redhat"
    else
        echo "unknown"
    fi
}

pkg_install() {
    case "$(os_family)" in
        debian)  DEBIAN_FRONTEND=noninteractive apt-get install -y "$@" ;;
        redhat)  dnf install -y "$@" ;;
        *)       die "Unsupported OS family. Install manually: $*" ;;
    esac
}

pkg_query() {
    case "$(os_family)" in
        debian)  dpkg -l "$1" 2>/dev/null | grep -q "^ii" ;;
        redhat)  rpm -q "$1" &>/dev/null ;;
        *)       return 1 ;;
    esac
}

# ── Initialise variables and validate environment ────────────────────────────
#
# Called by main() for direct execution.  Docker entrypoints may call this
# after overriding path defaults via environment variables.

_init_vars() {
    OS_FAMILY="$(os_family)"

    # Required environment variables
    : "${ZROK2_DNS_ZONE:?Set ZROK2_DNS_ZONE to your DNS zone (e.g., zrok.example.com)}"
    : "${ZROK2_ADMIN_TOKEN:?Set ZROK2_ADMIN_TOKEN to the admin secret for the zrok2 controller}"
    : "${ZITI_API_ENDPOINT:?Set ZITI_API_ENDPOINT to the Ziti management API URL}"
    : "${ZITI_ADMIN_PASSWORD:?Set ZITI_ADMIN_PASSWORD to the Ziti admin password}"

    ZITI_ADMIN_USER="${ZITI_ADMIN_USER:-admin}"
    ZROK2_CTRL_PORT="${ZROK2_CTRL_PORT:-18080}"
    ZROK2_TLS_CERT="${ZROK2_TLS_CERT:-}"
    ZROK2_TLS_KEY="${ZROK2_TLS_KEY:-}"
    ZROK2_NAMESPACE_TOKEN="${ZROK2_NAMESPACE_TOKEN:-public}"
    ZROK2_AMQP_URL="${ZROK2_AMQP_URL:-amqp://guest:guest@127.0.0.1:5672}"

    # Database defaults
    ZROK2_STORE_TYPE="${ZROK2_STORE_TYPE:-postgres}"
    ZROK2_DB_NAME="${ZROK2_DB_NAME:-zrok2}"
    ZROK2_DB_USER="${ZROK2_DB_USER:-zrok2}"
    ZROK2_DB_PASSWORD="${ZROK2_DB_PASSWORD:-}"

    # InfluxDB defaults
    ZROK2_INFLUX_URL="${ZROK2_INFLUX_URL:-http://127.0.0.1:8086}"
    ZROK2_INFLUX_ORG="${ZROK2_INFLUX_ORG:-zrok}"
    ZROK2_INFLUX_BUCKET="${ZROK2_INFLUX_BUCKET:-zrok}"
    ZROK2_INFLUX_TOKEN="${ZROK2_INFLUX_TOKEN:-}"
    ZROK2_INFLUX_PASSWORD="${ZROK2_INFLUX_PASSWORD:-}"

    # Ziti controller config auto-detection
    ZITI_CTRL_CONFIG="${ZITI_CTRL_CONFIG:-}"

    # Paths — overridable via env vars for Docker or non-standard layouts
    CTRL_CONFIG="${CTRL_CONFIG:-/etc/zrok2/ctrl.yml}"
    FRONTEND_CONFIG="${FRONTEND_CONFIG:-/etc/zrok2/frontend.yml}"
    CONTROLLER_HOME="${CONTROLLER_HOME:-/var/lib/zrok2-controller}"
    FRONTEND_HOME="${FRONTEND_HOME:-/var/lib/zrok2-frontend}"
    FABRIC_USAGE_PATH="${FABRIC_USAGE_PATH:-/var/lib/ziti-controller/fabric-usage.json}"

    # Database store — set by step_database() during Linux bootstrap, or
    # pre-set via env vars when called from Docker entrypoints.
    STORE_TYPE="${STORE_TYPE:-${ZROK2_STORE_TYPE}}"
    STORE_PATH="${STORE_PATH:-}"

    # Determine the zrok2 API endpoint from TLS and port settings
    if [[ -n "$ZROK2_TLS_CERT" ]]; then
        ZROK2_API_ENDPOINT="https://${ZROK2_DNS_ZONE}:${ZROK2_CTRL_PORT}"
        ZROK2_FRONTEND_BIND="0.0.0.0:443"
    else
        ZROK2_API_ENDPOINT="${ZROK2_API_ENDPOINT:-http://127.0.0.1:${ZROK2_CTRL_PORT}}"
        ZROK2_FRONTEND_BIND="${ZROK2_FRONTEND_BIND:-0.0.0.0:8080}"
    fi

    export ZROK2_API_ENDPOINT
    export ZROK2_ADMIN_TOKEN

    # When the operator running this script has a personal enabled zrok2
    # environment (~/.zrok2/environment.json), the zrok2 CLI's ApiEndpoint()
    # function ignores ZROK2_API_ENDPOINT and uses the stored endpoint instead
    # (typically api-v2.zrok.io). This would send every "zrok2 admin" call to
    # the wrong server and cause 401 Unauthorized errors.
    #
    # Fix: run all "zrok2 admin" calls with HOME pointing to a clean temp dir
    # so the CLI finds no enabled environment and respects ZROK2_API_ENDPOINT.
    # We cannot override HOME globally because the ziti CLI needs the real HOME
    # for its login session (~/.config/ziti/).
    #
    # The operator's ~/.zrok2 is never modified by this script.
    # Any identity files written by "zrok2 admin create identity" land in
    # _ZROK2_CLEAN_HOME/.zrok2/identities/ and are then copied to the service
    # user's directory by step_dynamic_proxy_controller().
    if [[ -d "${HOME}/.zrok2" ]] && [[ -f "${HOME}/.zrok2/environment.json" ]]; then
        _ZROK2_CLEAN_HOME="$(mktemp -d)"
        warn "HOME overridden to ${_ZROK2_CLEAN_HOME} for zrok2 admin commands (operator has an enabled zrok2 environment at ${HOME}/.zrok2 — it will not be modified)"
    fi
}

# Wrap "zrok2 admin" so ZROK2_API_ENDPOINT is respected even when the
# operator has an enabled environment pointing to a different server.
# All other zrok2 subcommands (e.g. "zrok2 admin bootstrap") connect directly
# to the database and do not use the API endpoint, so they do not need wrapping.
zrok2_admin() {
    if [[ -n "${_ZROK2_CLEAN_HOME:-}" ]]; then
        HOME="$_ZROK2_CLEAN_HOME" zrok2 admin "$@"
    else
        zrok2 admin "$@"
    fi
}

# ── Step 1: Install and configure RabbitMQ ───────────────────────────────────

step_rabbitmq() {
    info "Step 1: RabbitMQ"

    if systemctl is-active --quiet rabbitmq-server 2>/dev/null; then
        info "RabbitMQ is already running, skipping install"
        return
    fi

    if ! command -v rabbitmqctl &>/dev/null; then
        info "Installing rabbitmq-server..."
        pkg_install rabbitmq-server
    fi

    # Bind RabbitMQ to localhost only
    if [[ ! -f /etc/rabbitmq/rabbitmq-env.conf ]] || ! grep -q 'NODE_IP_ADDRESS=127.0.0.1' /etc/rabbitmq/rabbitmq-env.conf; then
        info "Configuring RabbitMQ to bind to localhost..."
        mkdir -p /etc/rabbitmq
        cat > /etc/rabbitmq/rabbitmq-env.conf <<'RABBITEOF'
NODE_IP_ADDRESS=127.0.0.1
SERVER_ADDITIONAL_ERL_ARGS="-kernel inet_dist_use_interface {127,0,0,1}"
RABBITEOF
    fi

    systemctl enable --now rabbitmq-server
    info "RabbitMQ is running"
}

# ── Step 2: Install and configure the database ───────────────────────────────

step_database() {
    info "Step 2: Database ($ZROK2_STORE_TYPE)"

    if [[ "$ZROK2_STORE_TYPE" == "postgres" ]]; then
        step_database_postgres
    elif [[ "$ZROK2_STORE_TYPE" == "sqlite3" ]]; then
        step_database_sqlite3
    else
        die "Unknown ZROK2_STORE_TYPE: $ZROK2_STORE_TYPE (must be 'postgres' or 'sqlite3')"
    fi
}

step_database_postgres() {
    if ! command -v psql &>/dev/null; then
        info "Installing PostgreSQL..."
        case "$OS_FAMILY" in
            debian)  pkg_install postgresql ;;
            redhat)  pkg_install postgresql-server ;;
            *)       die "Unsupported OS family for PostgreSQL install" ;;
        esac
    fi

    # Ensure the data cluster exists. RedHat always requires explicit initdb;
    # Debian auto-creates on install but the cluster may be missing if the data
    # directory was deleted (e.g., cleanup between test runs).
    if [[ "$OS_FAMILY" == "redhat" ]] && command -v postgresql-setup &>/dev/null; then
        postgresql-setup --initdb 2>/dev/null || true
    elif [[ "$OS_FAMILY" == "debian" ]] && command -v pg_lsclusters &>/dev/null; then
        if [[ -z "$(pg_lsclusters -h 2>/dev/null)" ]]; then
            info "No PostgreSQL cluster found — creating one"
            pg_createcluster "$(pg_config --version | grep -oP '\d+')" main --start 2>/dev/null || true
        fi
    fi

    # Discover the PostgreSQL service name (may be versioned on RedHat)
    local pg_service="postgresql"
    if [[ "$OS_FAMILY" == "redhat" ]] && ! systemctl list-unit-files "${pg_service}.service" &>/dev/null; then
        pg_service=$(systemctl list-unit-files 'postgresql*.service' --no-legend | head -1 | awk '{print $1}' | sed 's/.service$//')
        pg_service="${pg_service:-postgresql}"
    fi

    if ! systemctl is-active --quiet "$pg_service" 2>/dev/null; then
        systemctl enable --now "$pg_service"
    fi

    # Wait for PostgreSQL to accept connections (socket may take a moment)
    local _pg_attempts=30
    until sudo -u postgres psql -c '\q' &>/dev/null; do
        if (( --_pg_attempts == 0 )); then
            die "PostgreSQL did not become ready within 30s"
        fi
        sleep 1
    done

    # Generate a password if not provided
    if [[ -z "$ZROK2_DB_PASSWORD" ]]; then
        ZROK2_DB_PASSWORD=$(head -c24 /dev/urandom | base64 -w0)
        info "Generated PostgreSQL password (saved in $CTRL_CONFIG)"
    fi

    # Create the database user if it doesn't exist
    if ! sudo -u postgres psql -tAc "SELECT 1 FROM pg_roles WHERE rolname='${ZROK2_DB_USER}'" | grep -q 1; then
        info "Creating PostgreSQL user: $ZROK2_DB_USER"
        sudo -u postgres psql -c "CREATE USER ${ZROK2_DB_USER} WITH PASSWORD '${ZROK2_DB_PASSWORD}';"
    else
        info "PostgreSQL user $ZROK2_DB_USER already exists"
        # Update password in case it changed
        sudo -u postgres psql -c "ALTER USER ${ZROK2_DB_USER} WITH PASSWORD '${ZROK2_DB_PASSWORD}';"
    fi

    # Create the database if it doesn't exist
    if ! sudo -u postgres psql -tAc "SELECT 1 FROM pg_database WHERE datname='${ZROK2_DB_NAME}'" | grep -q 1; then
        info "Creating PostgreSQL database: $ZROK2_DB_NAME"
        sudo -u postgres psql -c "CREATE DATABASE ${ZROK2_DB_NAME} OWNER ${ZROK2_DB_USER};"
    else
        info "PostgreSQL database $ZROK2_DB_NAME already exists"
    fi

    STORE_PATH="host=127.0.0.1 user=${ZROK2_DB_USER} password=${ZROK2_DB_PASSWORD} dbname=${ZROK2_DB_NAME}"
    STORE_TYPE="postgres"
    info "PostgreSQL is ready"
}

step_database_sqlite3() {
    local sqlite_pkg
    case "$OS_FAMILY" in
        debian)  sqlite_pkg="libsqlite3-0" ;;
        redhat)  sqlite_pkg="sqlite" ;;
        *)       sqlite_pkg="sqlite3" ;;
    esac
    if ! pkg_query "$sqlite_pkg"; then
        info "Installing SQLite3 libraries..."
        pkg_install "$sqlite_pkg"
    fi

    STORE_PATH="zrok.db"
    STORE_TYPE="sqlite3"
    info "SQLite3 will be used (single-controller only)"
}

# ── Step 3: Install and configure InfluxDB ───────────────────────────────────

step_influxdb() {
    info "Step 3: InfluxDB"

    # The influxdb2 package ships the service as "influxdb.service" (not influxdb2)
    if systemctl is-active --quiet influxdb 2>/dev/null; then
        info "InfluxDB is already running"
    else
        # Check for the unit file, not just the package — dpkg can leave
        # a half-removed state where the package appears installed but the
        # unit file is gone.
        if ! systemctl list-unit-files influxdb.service &>/dev/null \
             || ! systemctl list-unit-files influxdb.service --no-legend | grep -q influxdb; then
            info "Installing InfluxDB..."
            case "$OS_FAMILY" in
                debian)
                    # Add InfluxData APT repository if not present
                    if [[ ! -f /etc/apt/sources.list.d/influxdata.list ]]; then
                        info "Adding InfluxData apt repository..."
                        curl -fsSL https://repos.influxdata.com/influxdata-archive.key \
                            | gpg --dearmor -o /usr/share/keyrings/influxdata-archive-keyring.gpg
                        echo "deb [signed-by=/usr/share/keyrings/influxdata-archive-keyring.gpg] https://repos.influxdata.com/debian stable main" \
                            > /etc/apt/sources.list.d/influxdata.list
                        apt-get update
                    fi
                    DEBIAN_FRONTEND=noninteractive apt-get install -y influxdb2
                    ;;
                redhat)
                    # Add InfluxData YUM/DNF repository if not present
                    if [[ ! -f /etc/yum.repos.d/influxdata.repo ]]; then
                        info "Adding InfluxData dnf repository..."
                        cat > /etc/yum.repos.d/influxdata.repo <<'INFLUXEOF'
[influxdata]
name=InfluxData Repository - Stable
baseurl=https://repos.influxdata.com/stable/$basearch/main
enabled=1
gpgcheck=1
gpgkey=https://repos.influxdata.com/influxdata-archive.key
INFLUXEOF
                    fi
                    dnf install -y influxdb2
                    ;;
                *)
                    die "Unsupported OS family for InfluxDB install"
                    ;;
            esac
        fi

        systemctl enable --now influxdb
        # Give InfluxDB a moment to start
        sleep 2
    fi

    # Set up InfluxDB if not already configured
    # Check if InfluxDB has been set up by testing the /api/v2/setup endpoint
    local setup_status
    # Note: jq's // operator treats false as null, so we must use 'not not'
    # to distinguish false from null. When .allowed is false, InfluxDB is
    # already set up.
    setup_status=$(curl -sS "${ZROK2_INFLUX_URL}/api/v2/setup" 2>/dev/null \
        | jq -r 'if .allowed == false then "false" else "true" end' 2>/dev/null \
        || echo "true")

    if [[ "$setup_status" == "true" ]]; then
        info "Running InfluxDB initial setup..."

        # Generate token if not provided
        if [[ -z "$ZROK2_INFLUX_TOKEN" ]]; then
            ZROK2_INFLUX_TOKEN=$(head -c48 /dev/urandom | base64 -w0)
            info "Generated InfluxDB admin token (saved in $CTRL_CONFIG)"
        fi

        if [[ -z "$ZROK2_INFLUX_PASSWORD" ]]; then
            ZROK2_INFLUX_PASSWORD=$(head -c24 /dev/urandom | base64 -w0)
            info "Generated InfluxDB admin password (saved in comment in $CTRL_CONFIG)"
        fi

        influx setup \
            --org "$ZROK2_INFLUX_ORG" \
            --bucket "$ZROK2_INFLUX_BUCKET" \
            --username admin \
            --password "$ZROK2_INFLUX_PASSWORD" \
            --token "$ZROK2_INFLUX_TOKEN" \
            --retention 0 \
            --name default \
            --force

        info "InfluxDB initial setup complete"
    else
        info "InfluxDB is already set up"
        # If no token was provided, try to get it from the InfluxDB config
        if [[ -z "$ZROK2_INFLUX_TOKEN" ]]; then
            ZROK2_INFLUX_TOKEN=$(influx auth list --json 2>/dev/null | jq -r '.[0].token // empty' 2>/dev/null || true)
            if [[ -z "$ZROK2_INFLUX_TOKEN" ]]; then
                warn "Could not auto-detect InfluxDB token. Set ZROK2_INFLUX_TOKEN if metrics are not working."
            fi
        fi
    fi

    # Ensure the bucket exists (idempotent)
    if ! influx bucket list --org "$ZROK2_INFLUX_ORG" --name "$ZROK2_INFLUX_BUCKET" --token "$ZROK2_INFLUX_TOKEN" &>/dev/null; then
        info "Creating InfluxDB bucket: $ZROK2_INFLUX_BUCKET"
        influx bucket create \
            --org "$ZROK2_INFLUX_ORG" \
            --name "$ZROK2_INFLUX_BUCKET" \
            --token "$ZROK2_INFLUX_TOKEN" \
            --retention 0
    fi

    info "InfluxDB is ready (org=$ZROK2_INFLUX_ORG, bucket=$ZROK2_INFLUX_BUCKET)"
}

# ── Step 4: Generate controller config ───────────────────────────────────────

step_ctrl_config() {
    info "Step 4: Controller configuration"

    if [[ -f "$CTRL_CONFIG" ]]; then
        info "Controller config already exists at $CTRL_CONFIG, skipping generation"
        return
    fi

    # Build store section based on database type
    local store_section
    if [[ "$STORE_TYPE" == "postgres" ]]; then
        store_section=$(cat <<STOREEOF
store:
  path: "${STORE_PATH}"
  type: "postgres"
  enable_locking: true
STOREEOF
        )
    else
        store_section=$(cat <<STOREEOF
store:
  path: ${STORE_PATH}
  type: sqlite3
STOREEOF
        )
    fi

    info "Generating $CTRL_CONFIG..."
    cat > "$CTRL_CONFIG" <<CTRLEOF
#    _____ __ ___ | | __
#   |_  / '__/ _ \| |/ /
#    / /| | | (_) |   <
#   /___|_|  \___/|_|\_\
# controller configuration

v: 4

admin:
  secrets:
    - ${ZROK2_ADMIN_TOKEN}

bridge:
  source:
    type: fileSource
    path: ${FABRIC_USAGE_PATH}
  sink:
    type: amqpSink
    url: ${ZROK2_AMQP_URL}
    queue_name: events

endpoint:
  host: 0.0.0.0
  port: ${ZROK2_CTRL_PORT}

$(if [[ -n "$ZROK2_TLS_CERT" ]]; then
cat <<TLSEOF
tls:
  cert_path: ${ZROK2_TLS_CERT}
  key_path: ${ZROK2_TLS_KEY}
TLSEOF
fi)

invites:
  invites_open: false

limits:
  environments:    -1
  shares:          -1
  reserved_shares: -1
  unique_names:    -1
  share_frontends: -1
  bandwidth:
    period: 5m
    warning:
      rx:    -1
      tx:    -1
      total: 7242880
    limit:
      rx:    -1
      tx:    -1
      total: 10485760
  enforcing: false
  cycle: 5m

maintenance:
  registration:
    expiration_timeout: 24h
    check_frequency:    1h
    batch_limit:        500
  reset_password:
    expiration_timeout: 15m
    check_frequency:    15m
    batch_limit:        500

metrics:
  agent:
    source:
      type: amqpSource
      url: ${ZROK2_AMQP_URL}
      queue_name: events
  influx:
    url: "${ZROK2_INFLUX_URL}"
    bucket: ${ZROK2_INFLUX_BUCKET}
    org: ${ZROK2_INFLUX_ORG}
    token: "${ZROK2_INFLUX_TOKEN}"
        # Influx admin password used during 'influx setup': "${ZROK2_INFLUX_PASSWORD}"

registration:
  registration_url_template: ${ZROK2_API_ENDPOINT}/register

reset_password:
  reset_url_template: ${ZROK2_API_ENDPOINT}/resetPassword

${store_section}

compatibility:
  version_patterns:
    - "^(refs/(heads|tags)/)?v2\\\\.0"
    - "^v0\\\\.0\\\\.0"

ziti:
  api_endpoint: "${ZITI_API_ENDPOINT}"
  username: ${ZITI_ADMIN_USER}
  password: "${ZITI_ADMIN_PASSWORD}"
CTRLEOF

    if id -u zrok2-controller &>/dev/null; then
        chown zrok2-controller:zrok2-controller "$CTRL_CONFIG"
    fi
    chmod 640 "$CTRL_CONFIG"
    info "Controller config written to $CTRL_CONFIG"
}

# ── Step 5: Configure OpenZiti controller for metrics events ─────────────────

step_ziti_events() {
    info "Step 5: OpenZiti controller metrics event configuration"

    # Auto-detect the Ziti controller config
    if [[ -z "$ZITI_CTRL_CONFIG" ]]; then
        for candidate in \
            /var/lib/private/ziti-controller/config.yml \
            /var/lib/ziti-controller/config.yml \
            /etc/openziti/ziti-controller/config.yml; do
            if [[ -f "$candidate" ]]; then
                ZITI_CTRL_CONFIG="$candidate"
                break
            fi
        done
    fi

    if [[ -z "$ZITI_CTRL_CONFIG" || ! -f "$ZITI_CTRL_CONFIG" ]]; then
        warn "Could not locate OpenZiti controller config. Skipping events configuration."
        warn "You must manually add the events and network stanzas to your Ziti controller config."
        warn "See: https://docs.zrok.io/docs/guides/self-hosting/metrics-and-limits/configuring-metrics"
        return
    fi

    info "Ziti controller config: $ZITI_CTRL_CONFIG"

    # Check if events section already exists
    if grep -q 'fabric\.usage' "$ZITI_CTRL_CONFIG" 2>/dev/null; then
        info "Ziti controller already has fabric.usage events configured"
    else
        info "Adding events stanza to Ziti controller config..."

        # Ensure the fabric-usage.json directory exists and is writable
        local usage_dir
        usage_dir=$(dirname "$FABRIC_USAGE_PATH")
        mkdir -p "$usage_dir"

        cat >> "$ZITI_CTRL_CONFIG" <<EVENTSEOF

# zrok metrics: emit fabric.usage events to a file for the metrics bridge
events:
  jsonLogger:
    subscriptions:
      - type: fabric.usage
        version: 3
    handler:
      type: file
      format: json
      path: ${FABRIC_USAGE_PATH}
EVENTSEOF
        info "Events stanza added"
    fi

    # Check if network metrics interval is already tuned
    if grep -q 'metricsReportInterval' "$ZITI_CTRL_CONFIG" 2>/dev/null; then
        info "Ziti controller network metrics interval already configured"
    else
        # Check if there's an existing network: section we can add to
        if grep -q '^network:' "$ZITI_CTRL_CONFIG" 2>/dev/null; then
            warn "Ziti controller has a 'network:' section but no metricsReportInterval."
            warn "Add these to the network section for snappy metrics:"
            warn "  intervalAgeThreshold: 5s"
            warn "  metricsReportInterval: 5s"
        elif grep -q '^#network:' "$ZITI_CTRL_CONFIG" 2>/dev/null; then
            # The network section is commented out, add an uncommented one
            info "Adding network metrics interval to Ziti controller config..."
            cat >> "$ZITI_CTRL_CONFIG" <<NETWORKEOF

# zrok metrics: increase reporting frequency for responsive metrics
network:
  intervalAgeThreshold: 5s
  metricsReportInterval: 5s
NETWORKEOF
            info "Network metrics interval added"
        else
            info "Adding network metrics interval to Ziti controller config..."
            cat >> "$ZITI_CTRL_CONFIG" <<NETWORKEOF

# zrok metrics: increase reporting frequency for responsive metrics
network:
  intervalAgeThreshold: 5s
  metricsReportInterval: 5s
NETWORKEOF
            info "Network metrics interval added"
        fi
    fi

    # Do NOT restart the Ziti controller here. The events config will be
    # picked up when step_start_services restarts the controller at the end
    # of the bootstrap, after all Ziti admin commands (Steps 6-12) are done.
    # Restarting now would break the router's TLS connection and cause
    # Steps 6-8 (which need the Ziti data plane) to fail.
    info "Events config written; Ziti controller restart deferred to Step 17"
}

# ── Step 6: Bootstrap the Ziti network for zrok ──────────────────────────────

step_bootstrap() {
    info "Step 6: zrok2 admin bootstrap"

    # Check if already bootstrapped by looking for the zrok database
    if [[ "$STORE_TYPE" == "sqlite3" && -f "${CONTROLLER_HOME}/zrok.db" ]]; then
        info "Bootstrap already complete (zrok.db exists), skipping"
        return
    fi

    if [[ "$STORE_TYPE" == "postgres" ]]; then
        local table_count
        table_count=$(PGPASSWORD="$ZROK2_DB_PASSWORD" psql -h 127.0.0.1 -U "$ZROK2_DB_USER" -d "$ZROK2_DB_NAME" \
            -tAc "SELECT count(*) FROM information_schema.tables WHERE table_schema='public'" 2>/dev/null || echo "0")
        if (( table_count > 0 )); then
            info "Bootstrap already complete (PostgreSQL tables exist), skipping"
            return
        fi
    fi

    info "Running zrok2 admin bootstrap..."
    if id -u zrok2-controller &>/dev/null; then
        sudo -u zrok2-controller \
            ZROK2_ADMIN_TOKEN="$ZROK2_ADMIN_TOKEN" \
            zrok2 admin bootstrap "$CTRL_CONFIG"
    else
        ZROK2_ADMIN_TOKEN="$ZROK2_ADMIN_TOKEN" \
            zrok2 admin bootstrap "$CTRL_CONFIG"
    fi

    info "Bootstrap complete"
}

# ── Step 7: Start the controller (needed for subsequent admin commands) ──────

step_start_controller() {
    info "Step 7: Start zrok2-controller"

    if systemctl is-active --quiet zrok2-controller 2>/dev/null; then
        info "zrok2-controller is already running"
        return
    fi

    systemctl enable --now zrok2-controller

    # Wait for the controller API to be ready (up to 30s)
    local _attempts=15
    while (( _attempts-- > 0 )); do
        if curl -sf -o /dev/null \
               -H 'Accept: application/zrok.v1+json' \
               "${ZROK2_API_ENDPOINT}/api/v2/versions" 2>/dev/null; then
            info "zrok2-controller started and API is ready"
            return
        fi
        sleep 2
    done
    warn "zrok2-controller started but API readiness check timed out (proceeding anyway)"
}

# ── Step 8: Create a dynamic frontend ────────────────────────────────────────

step_create_frontend() {
    info "Step 8: Create dynamic frontend"

    # Check if a frontend already exists
    local existing
    existing=$(zrok2_admin list frontends 2>/dev/null || true)
    if echo "$existing" | grep -q 'public'; then
        FRONTEND_TOKEN=$(echo "$existing" | awk '/public/ {print $1; exit}')
        info "Frontend already exists with token: $FRONTEND_TOKEN"

        # Ensure it's marked dynamic
        local is_dynamic
        is_dynamic=$(echo "$existing" | awk '/public/ {for(i=1;i<=NF;i++) if($i=="true" || $i=="false") print $i; exit}')
        if [[ "$is_dynamic" != *"true"* ]]; then
            warn "Updating frontend to dynamic mode..."
            zrok2_admin update frontend "$FRONTEND_TOKEN" --dynamic
        fi
        return
    fi

    info "Creating dynamic frontend..."

    # Look up the Ziti identity ID for the 'public' frontend identity
    # (created by zrok2 admin bootstrap)
    local public_ziti_id
    public_ziti_id=$(ziti edge list identities 'name="public"' -j \
        | jq -r '.data[0].id' 2>/dev/null || true)
    if [[ -z "$public_ziti_id" || "$public_ziti_id" == "null" ]]; then
        error "Cannot find Ziti identity 'public'. Was 'zrok2 admin bootstrap' run?"
        return 1
    fi
    info "Found public identity Ziti ID: $public_ziti_id"

    local output
    if ! output=$(zrok2_admin create frontend --dynamic "$public_ziti_id" public "https://{token}.${ZROK2_DNS_ZONE}" 2>&1); then
        error "zrok2 admin create frontend failed: $output"
        return 1
    fi
    FRONTEND_TOKEN=$(echo "$output" | grep -oP "(?<=frontend ').*(?=')" || true)

    if [[ -z "$FRONTEND_TOKEN" ]]; then
        # Try to extract from list
        FRONTEND_TOKEN=$(zrok2_admin list frontends 2>/dev/null | awk '/public/ {print $1; exit}')
    fi

    info "Created dynamic frontend with token: $FRONTEND_TOKEN"
}

# ── Step 9: Create dynamicProxyController identity and Ziti resources ────────

step_dynamic_proxy_controller() {
    info "Step 9: dynamicProxyController identity and Ziti resources"

    local SERVICE_NAME="dynamicProxyController"
    local identity_dir="${CONTROLLER_HOME}/.zrok2/identities"

    # Create the zrok identity if not already present
    if [[ -f "${identity_dir}/dynamicProxyController.json" ]]; then
        info "dynamicProxyController identity already exists, skipping creation"
    else
        info "Creating dynamicProxyController identity..."
        zrok2_admin create identity dynamicProxyController

        # Move identity to controller service user's directory.
        # When HOME was overridden to isolate zrok2 from the operator's enabled
        # environment, the identity file lands in _ZROK2_CLEAN_HOME instead.
        local source_dir="${_ZROK2_CLEAN_HOME:+${_ZROK2_CLEAN_HOME}/.zrok2/identities}"
        source_dir="${source_dir:-${HOME}/.zrok2/identities}"
        if [[ -f "${source_dir}/dynamicProxyController.json" ]]; then
            mkdir -p "$identity_dir"
            cp -v "${source_dir}/dynamicProxyController.json" "${identity_dir}/"
            if id -u zrok2-controller &>/dev/null; then
                chown -R zrok2-controller:zrok2-controller "${CONTROLLER_HOME}/.zrok2"
            fi
            info "Identity placed in $identity_dir"
        fi
    fi

    # Get the Ziti ID for the dynamicProxyController identity
    local controller_zid
    controller_zid=$(ziti edge list identities 'name="dynamicProxyController"' -j 2>/dev/null \
        | jq -r '.data[0].id // empty' 2>/dev/null || true)

    if [[ -z "$controller_zid" ]]; then
        warn "Could not find dynamicProxyController Ziti ID; Ziti resources may already exist"
        return
    fi

    info "dynamicProxyController Ziti ID: $controller_zid"

    # Create Ziti service (idempotent - ignore errors if exists)
    if ! ziti edge list services "name=\"${SERVICE_NAME}\"" -j 2>/dev/null | jq -e '.data | length > 0' &>/dev/null; then
        info "Creating Ziti service: $SERVICE_NAME"
        ziti edge create service "$SERVICE_NAME"
    else
        info "Ziti service $SERVICE_NAME already exists"
    fi

    # Create Service Edge Router Policy
    if ! ziti edge list service-edge-router-policies "name=\"${SERVICE_NAME}-serp\"" -j 2>/dev/null | jq -e '.data | length > 0' &>/dev/null; then
        info "Creating SERP: ${SERVICE_NAME}-serp"
        ziti edge create serp "${SERVICE_NAME}-serp" \
            --edge-router-roles '#all' \
            --service-roles "@${SERVICE_NAME}"
    else
        info "SERP ${SERVICE_NAME}-serp already exists"
    fi

    # Create Bind service policy (controller binds the service)
    if ! ziti edge list service-policies "name=\"${SERVICE_NAME}-bind\"" -j 2>/dev/null | jq -e '.data | length > 0' &>/dev/null; then
        info "Creating bind policy: ${SERVICE_NAME}-bind"
        ziti edge create sp "${SERVICE_NAME}-bind" Bind \
            --identity-roles "@${controller_zid}" \
            --service-roles "@${SERVICE_NAME}"
    else
        info "Bind policy ${SERVICE_NAME}-bind already exists"
    fi

    # Create Dial service policy (frontend dials the service)
    if ! ziti edge list service-policies "name=\"${SERVICE_NAME}-dial\"" -j 2>/dev/null | jq -e '.data | length > 0' &>/dev/null; then
        info "Creating dial policy: ${SERVICE_NAME}-dial"
        ziti edge create sp "${SERVICE_NAME}-dial" Dial \
            --identity-roles "@public" \
            --service-roles "@${SERVICE_NAME}"
    else
        info "Dial policy ${SERVICE_NAME}-dial already exists"
    fi

    # Add the dynamic_proxy_controller block to ctrl.yml now that the identity exists.
    # The controller will be restarted in step_start_services to pick this up.
    if ! grep -q 'dynamic_proxy_controller:' "$CTRL_CONFIG" 2>/dev/null; then
        info "Adding dynamic_proxy_controller to $CTRL_CONFIG..."
        cat >> "$CTRL_CONFIG" <<DPCEOF

dynamic_proxy_controller:
  identity_path: ${CONTROLLER_HOME}/.zrok2/identities/dynamicProxyController.json
  service_name: dynamicProxyController
  amqp_publisher:
    url: ${ZROK2_AMQP_URL}
    exchange_name: dynamicProxy
DPCEOF
        info "dynamic_proxy_controller block added to config"
    fi

    info "dynamicProxyController Ziti resources are configured"
}

# ── Step 10: Place frontend identity file ────────────────────────────────────

step_frontend_identity() {
    info "Step 10: Frontend identity file"

    local frontend_identity_dir="${FRONTEND_HOME}/.zrok2/identities"

    if [[ -f "${frontend_identity_dir}/public.json" ]]; then
        info "Frontend identity already in place, skipping"
        return
    fi

    # The bootstrap process creates the public identity in the admin user's home
    local source="${HOME}/.zrok2/identities/public.json"
    if [[ ! -f "$source" ]]; then
        # Try the controller's copy
        source="${CONTROLLER_HOME}/.zrok2/identities/public.json"
    fi

    if [[ -f "$source" ]]; then
        mkdir -p "$frontend_identity_dir"
        cp -v "$source" "${frontend_identity_dir}/public.json"
        if id -u zrok2-frontend &>/dev/null; then
            chown -R zrok2-frontend:zrok2-frontend "${FRONTEND_HOME}/.zrok2"
        fi
        info "Frontend identity placed in $frontend_identity_dir"
    else
        warn "Could not find public.json identity file. You may need to copy it manually."
        warn "Look for it in ~/.zrok2/identities/public.json after running bootstrap."
    fi
}

# ── Step 11: Create public namespace ─────────────────────────────────────────

step_create_namespace() {
    info "Step 11: Create public namespace"

    local existing
    existing=$(zrok2_admin list namespaces 2>/dev/null || true)
    if echo "$existing" | grep -q "${ZROK2_DNS_ZONE}"; then
        info "Namespace '${ZROK2_DNS_ZONE}' already exists, skipping"
        return
    fi

    info "Creating namespace '${ZROK2_DNS_ZONE}' with token '${ZROK2_NAMESPACE_TOKEN}'..."
    zrok2_admin create namespace \
        --token "${ZROK2_NAMESPACE_TOKEN}" \
        --open \
        "${ZROK2_DNS_ZONE}"

    info "Namespace created"
}

# ── Step 12: Map namespace to frontend ───────────────────────────────────────

step_map_namespace_frontend() {
    info "Step 12: Map namespace to frontend"

    if [[ -z "${FRONTEND_TOKEN:-}" ]]; then
        FRONTEND_TOKEN=$(zrok2_admin list frontends 2>/dev/null | awk '/public/ {print $1; exit}')
    fi

    if [[ -z "${FRONTEND_TOKEN:-}" ]]; then
        die "Cannot determine frontend token. Create a frontend first."
    fi

    local existing
    existing=$(zrok2_admin list namespace-frontend "${ZROK2_NAMESPACE_TOKEN}" 2>/dev/null || true)
    if echo "$existing" | grep -q "${FRONTEND_TOKEN}"; then
        info "Namespace-frontend mapping already exists, skipping"
        return
    fi

    info "Mapping namespace '${ZROK2_NAMESPACE_TOKEN}' to frontend '${FRONTEND_TOKEN}'..."
    zrok2_admin create namespace-frontend "${ZROK2_NAMESPACE_TOKEN}" "${FRONTEND_TOKEN}" --default

    info "Namespace-frontend mapping created"
}

# ── Step 13: Generate frontend config ────────────────────────────────────────

step_frontend_config() {
    info "Step 13: Frontend configuration"

    if [[ -f "$FRONTEND_CONFIG" ]]; then
        info "Frontend config already exists at $FRONTEND_CONFIG, skipping generation"
        return
    fi

    if [[ -z "${FRONTEND_TOKEN:-}" ]]; then
        FRONTEND_TOKEN=$(zrok2_admin list frontends 2>/dev/null | awk '/public/ {print $1; exit}')
    fi

    info "Generating $FRONTEND_CONFIG..."
    cat > "$FRONTEND_CONFIG" <<FEEOF
v: 1

frontend_token: ${FRONTEND_TOKEN}
identity: public
bind_address: ${ZROK2_FRONTEND_BIND}
mapping_refresh_interval: 1m

amqp_subscriber:
  url: ${ZROK2_AMQP_URL}
  exchange_name: dynamicProxy

controller:
  identity_path: ${FRONTEND_HOME}/.zrok2/identities/public.json
  service_name: dynamicProxyController

host_match: ${ZROK2_DNS_ZONE}

$(if [[ -n "$ZROK2_TLS_CERT" ]]; then
cat <<TLSEOF
tls:
  cert_path: ${ZROK2_TLS_CERT}
  key_path: ${ZROK2_TLS_KEY}
TLSEOF
fi)
FEEOF

    if id -u zrok2-frontend &>/dev/null; then
        chown zrok2-frontend:zrok2-frontend "$FRONTEND_CONFIG"
    fi
    chmod 640 "$FRONTEND_CONFIG"
    info "Frontend config written to $FRONTEND_CONFIG"
}

# ── Step 14: TLS certificate permissions ─────────────────────────────────────

step_tls_permissions() {
    if [[ -z "$ZROK2_TLS_CERT" ]]; then
        info "Step 14: No TLS configured, skipping cert permissions"
        return
    fi

    info "Step 14: TLS certificate permissions"

    # Ensure service users can read the certificate files
    # Let's Encrypt stores certs as symlinks into /etc/letsencrypt/archive/
    local cert_dir
    cert_dir=$(dirname "$(realpath "$ZROK2_TLS_CERT")")

    # Create a shared group for cert access
    if ! getent group zrok2-tls &>/dev/null; then
        groupadd --system zrok2-tls
    fi

    usermod -aG zrok2-tls zrok2-controller 2>/dev/null || true
    usermod -aG zrok2-tls zrok2-frontend 2>/dev/null || true
    usermod -aG zrok2-tls zrok2-metrics-bridge 2>/dev/null || true

    # Make the archive directory group-readable
    chgrp -R zrok2-tls "$cert_dir"
    chmod g+rx "$cert_dir"
    chmod g+r "$cert_dir"/*

    # Also ensure the live/ and archive/ parent dirs are traversable
    local le_base="/etc/letsencrypt"
    if [[ -d "$le_base" ]]; then
        chmod o+x "$le_base" "$le_base/live" "$le_base/archive" 2>/dev/null || true
        local domain_dir
        domain_dir=$(basename "$(dirname "$ZROK2_TLS_CERT")")
        chmod o+x "$le_base/live/$domain_dir" "$le_base/archive/$domain_dir" 2>/dev/null || true
    fi

    info "TLS certificate permissions configured for service users"
}

# ── Step 15: systemd overrides ───────────────────────────────────────────────

step_systemd_overrides() {
    info "Step 15: systemd service overrides"

    # Frontend: CAP_NET_BIND_SERVICE allows binding privileged ports (443) as non-root
    local frontend_override="/etc/systemd/system/zrok2-frontend.service.d/override.conf"
    if [[ ! -f "$frontend_override" ]]; then
        info "Creating frontend systemd override for CAP_NET_BIND_SERVICE..."
        mkdir -p "$(dirname "$frontend_override")"
        cat > "$frontend_override" <<'FEOVERRIDE'
[Service]
AmbientCapabilities=CAP_NET_BIND_SERVICE
FEOVERRIDE
    else
        info "Frontend systemd override already exists"
    fi

    # Controller: add supplementary group for TLS cert access if TLS is configured
    if [[ -n "$ZROK2_TLS_CERT" ]]; then
        local ctrl_override="/etc/systemd/system/zrok2-controller.service.d/override.conf"
        if [[ ! -f "$ctrl_override" ]]; then
            info "Creating controller systemd override for TLS group..."
            mkdir -p "$(dirname "$ctrl_override")"
            cat > "$ctrl_override" <<'CTRLOVERRIDE'
[Service]
SupplementaryGroups=zrok2-tls
CTRLOVERRIDE
        fi

        # Also add the TLS group to the frontend override
        if ! grep -q 'SupplementaryGroups' "$frontend_override" 2>/dev/null; then
            cat >> "$frontend_override" <<'FEGROUPS'
SupplementaryGroups=zrok2-tls
FEGROUPS
        fi
    fi

    systemctl daemon-reload
    info "systemd overrides applied"
}

# ── Step 16: Start the metrics bridge ────────────────────────────────────────

step_start_metrics_bridge() {
    info "Step 16: Start zrok2-metrics-bridge bridge"

    # The metrics bridge needs read access to the controller config (for the bridge section)
    # and read access to the fabric-usage.json file written by the Ziti controller.
    # The zrok2-metrics-bridge user is added to the zrok2-controller group by postinstall-metrics.bash.

    # Ensure the fabric-usage.json file exists and is readable by the metrics bridge
    if [[ ! -f "$FABRIC_USAGE_PATH" ]]; then
        touch "$FABRIC_USAGE_PATH"
    fi
    # Make the file group-readable by ziti-controller group (not world-readable)
    chown ziti-controller:ziti-controller "$FABRIC_USAGE_PATH"
    chmod 0640 "$FABRIC_USAGE_PATH"

    # The metrics bridge also writes a .ptr file alongside fabric-usage.json to
    # track its read position. The directory must be group-writable by the
    # ziti-controller group so that the zrok2-metrics-bridge user can create it.
    local usage_dir
    usage_dir="$(dirname "$FABRIC_USAGE_PATH")"
    chown ziti-controller:ziti-controller "$usage_dir"
    chmod g+w "$usage_dir"

    # Grant read/write access to the fabric-usage.json directory and file
    # for the metrics bridge user via supplementary group membership
    if getent passwd zrok2-metrics-bridge &>/dev/null; then
        usermod -aG ziti-controller zrok2-metrics-bridge 2>/dev/null || true
    fi

    if systemctl is-active --quiet zrok2-metrics-bridge 2>/dev/null; then
        info "zrok2-metrics-bridge is already running"
    else
        systemctl enable --now zrok2-metrics-bridge
        info "zrok2-metrics-bridge bridge started"
    fi
}

# ── Step 17: Start frontend and restart controller ───────────────────────────

step_start_services() {
    info "Step 17: Start services"

    # Restart the Ziti controller to pick up the events and network stanzas
    # added in Step 5. All Ziti admin commands (Steps 6-12) are complete, so
    # a brief router disconnect is harmless.
    if systemctl is-active --quiet ziti-controller 2>/dev/null; then
        info "Restarting ziti-controller to apply events configuration..."
        local _svc_env="/opt/openziti/etc/controller/service.env"
        local _had_renew_setting=false
        if grep -q '^ZITI_AUTO_RENEW_CERTS=' "$_svc_env" 2>/dev/null; then
            _had_renew_setting=true
        else
            echo "ZITI_AUTO_RENEW_CERTS=false" >> "$_svc_env"
        fi
        systemctl restart ziti-controller
        if ! $_had_renew_setting; then
            sed -i '/^ZITI_AUTO_RENEW_CERTS=false$/d' "$_svc_env"
        fi
        sleep 2
    fi

    # Restart controller to pick up dynamic_proxy_controller config, metrics, and new overrides
    info "Restarting zrok2-controller to apply full configuration..."
    systemctl restart zrok2-controller
    sleep 2

    systemctl enable --now zrok2-frontend
    info "zrok2-frontend started"
}

# ── Main ─────────────────────────────────────────────────────────────────────

main() {
    _init_vars
    info "=== zrok2 Bootstrap ==="
    info "DNS zone:      $ZROK2_DNS_ZONE"
    info "API endpoint:  $ZROK2_API_ENDPOINT"
    info "Database:      $ZROK2_STORE_TYPE"
    info "InfluxDB:      $ZROK2_INFLUX_URL"
    info ""

    step_rabbitmq
    step_database
    step_influxdb
    step_ctrl_config
    step_ziti_events
    step_bootstrap
    step_start_controller
    step_create_frontend
    step_dynamic_proxy_controller
    step_frontend_identity
    step_create_namespace
    step_map_namespace_frontend
    step_frontend_config
    step_tls_permissions
    step_systemd_overrides
    step_start_metrics_bridge
    step_start_services

    info ""
    info "=== Bootstrap Complete ==="
    info ""
    info "Your zrok2 instance is running at: $ZROK2_API_ENDPOINT"
    info "Public namespace: $ZROK2_NAMESPACE_TOKEN (${ZROK2_DNS_ZONE})"
    info "Database: $ZROK2_STORE_TYPE"
    info "Metrics: InfluxDB at $ZROK2_INFLUX_URL (org=$ZROK2_INFLUX_ORG, bucket=$ZROK2_INFLUX_BUCKET)"
    info ""
    info "Next steps:"
    info "  1. Create a user account:"
    info "     zrok2 admin create account <email> <password>"
    info ""
    info "  2. On a client device, enable the environment:"
    info "     zrok2 config set apiEndpoint $ZROK2_API_ENDPOINT"
    info "     zrok2 enable <account-token>"
    info ""
    info "  3. Create a share:"
    info "     zrok2 config set defaultNamespace $ZROK2_NAMESPACE_TOKEN"
    info "     zrok2 share public localhost:8080"
}

# ── Source / Execute guard ────────────────────────────────────────────────────
#
# When executed directly (Linux self-hosting), the full 17-step bootstrap runs.
# When sourced (e.g., Docker entrypoints), only function definitions are loaded
# and the caller drives execution by calling _init_vars() and step_*() as needed.

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    # Direct execution — full Linux bootstrap
    set -o errexit -o nounset -o pipefail
    main "$@"
fi
