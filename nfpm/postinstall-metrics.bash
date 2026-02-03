#!/usr/bin/env bash
#
# postinstall-metrics.bash - Post-installation script for zrok2-metrics package
#

set -euo pipefail

# Package-specific configuration
SERVICE_NAME="zrok2-metrics"
SERVICE_USER="zrok2-metrics"
SERVICE_GROUP="zrok2-metrics"
SERVICE_HOME="/var/lib/zrok2-metrics"
ADDITIONAL_GROUPS="zrok2-controller"

# Initialize debug output
: "${DEBUG:=0}"
if (( DEBUG )); then
    exec 3>&1
    set -o xtrace
else
    exec 3>/dev/null
fi

# Function to create service user and group
create_service_user() {
    echo "Creating service user and group..." >&3
    
    # Create group if it doesn't exist
    if ! getent group "${SERVICE_GROUP}" >/dev/null 2>&1; then
        groupadd --system "${SERVICE_GROUP}"
        echo "Created group: ${SERVICE_GROUP}" >&3
    fi
    
    # Create user if it doesn't exist
    if ! getent passwd "${SERVICE_USER}" >/dev/null 2>&1; then
        if [[ -n "${ADDITIONAL_GROUPS}" ]]; then
            useradd --system \
                    --gid "${SERVICE_GROUP}" \
                    --groups "${ADDITIONAL_GROUPS}" \
                    --home-dir "${SERVICE_HOME}" \
                    --create-home \
                    --shell /usr/sbin/nologin \
                    --comment "zrok2 service user" \
                    "${SERVICE_USER}"
        else
            useradd --system \
                    --gid "${SERVICE_GROUP}" \
                    --home-dir "${SERVICE_HOME}" \
                    --create-home \
                    --shell /usr/sbin/nologin \
                    --comment "zrok2 service user" \
                    "${SERVICE_USER}"
        fi
        echo "Created user: ${SERVICE_USER}" >&3
    else
        # User exists, add to additional groups if specified
        if [[ -n "${ADDITIONAL_GROUPS}" ]]; then
            usermod --append --groups "${ADDITIONAL_GROUPS}" "${SERVICE_USER}"
            echo "Added ${SERVICE_USER} to groups: ${ADDITIONAL_GROUPS}" >&3
        fi
    fi
    
    # Ensure home directory exists and has correct permissions
    if [[ ! -d "${SERVICE_HOME}" ]]; then
        mkdir -p "${SERVICE_HOME}"
        echo "Created home directory: ${SERVICE_HOME}" >&3
    fi
    chown "${SERVICE_USER}:${SERVICE_GROUP}" "${SERVICE_HOME}"
    chmod 750 "${SERVICE_HOME}"
}

# Function to handle clean installation
install() {
    echo "Performing clean installation setup for ${SERVICE_NAME}..." >&3
    create_service_user
    
    # Reload systemd
    if command -v systemctl >/dev/null 2>&1; then
        systemctl daemon-reload
        echo "Systemd reloaded. Enable and start service as needed:" >&3
        echo "  systemctl enable --now ${SERVICE_NAME}" >&3
    fi
}

# Function to handle upgrade
upgrade() {
    echo "Performing upgrade setup for ${SERVICE_NAME}..." >&3
    create_service_user
    
    # Reload systemd and restart service if running
    if command -v systemctl >/dev/null 2>&1; then
        systemctl daemon-reload
        
        if systemctl is-active --quiet "${SERVICE_NAME}.service" 2>/dev/null; then
            echo "Restarting ${SERVICE_NAME} service..." >&3
            systemctl restart "${SERVICE_NAME}.service"
        fi
    fi
}

# Main script logic
main() {
    # Determine action based on package manager parameters
    if (( $# )); then
        if [[ $1 == 1 || ($1 == configure && -z ${2:-}) ]]; then
            # RPM: $1=1 for install, DEB: $1=configure with no $2 for install
            action=install
        elif [[ $1 == 2 || ($1 == configure && -n ${2:-}) ]]; then
            # RPM: $1=2 for upgrade, DEB: $1=configure with $2=version for upgrade
            action=upgrade
        else
            echo "ERROR: unexpected action '$1'" >&2
            exit 1
        fi
    else
        echo "ERROR: missing action parameter" >&2
        exit 1
    fi
    
    case "$action" in
        "install")
            install
            printf "\033[32mCompleted clean install of ${SERVICE_NAME}\033[0m\n"
            ;;
        "upgrade")
            upgrade
            printf "\033[32mCompleted upgrade of ${SERVICE_NAME}\033[0m\n"
            ;;
    esac
}

# Run main function with all arguments
main "$@"
