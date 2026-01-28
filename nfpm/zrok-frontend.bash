#!/usr/bin/env bash
#
# zrok-frontend-run.bash - Launcher script for zrok frontend service
#
# Verifies config file exists before executing zrok access public or dynamicProxy

set -euo pipefail

FRONTEND_TYPE="${1:-public}"
CONFIG_FILE="${2:-/etc/zrok/frontend.yml}"

if [[ "$FRONTEND_TYPE" != "public" && "$FRONTEND_TYPE" != "dynamicProxy" ]]; then
    echo "ERROR: Invalid frontend type: $FRONTEND_TYPE" >&2
    echo "" >&2
    echo "Usage: $0 [public|dynamicProxy] [config_file]" >&2
    echo "" >&2
    echo "Examples:" >&2
    echo "  $0 public /etc/zrok/frontend.yml" >&2
    echo "  $0 dynamicProxy /etc/zrok/dynamic-frontend.yml" >&2
    exit 1
fi

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

exec /usr/bin/zrok access "$FRONTEND_TYPE" "$CONFIG_FILE"
