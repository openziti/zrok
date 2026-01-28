#!/usr/bin/env bash
#
# zrok-frontend-run.bash - Launcher script for zrok frontend service
#
# Verifies config file exists before executing zrok access public

set -euo pipefail

CONFIG_FILE="${1:-/etc/zrok/frontend.yml}"

if [[ ! -f "$CONFIG_FILE" ]]; then
    echo "ERROR: Configuration file not found: $CONFIG_FILE" >&2
    echo "" >&2
    echo "To configure the zrok frontend, create the config file:" >&2
    echo "  sudo cp /etc/zrok/frontend.yml.example $CONFIG_FILE" >&2
    echo "  sudo editor $CONFIG_FILE" >&2
    echo "" >&2
    echo "Then start the service:" >&2
    echo "  sudo systemctl start zrok-frontend" >&2
    exit 1
fi

exec /usr/bin/zrok access public "$CONFIG_FILE"
