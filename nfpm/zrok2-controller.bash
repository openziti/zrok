#!/usr/bin/env bash
#
# zrok-controller-run.bash - Launcher script for zrok controller service
#
# Verifies config file exists before executing zrok controller

set -euo pipefail

CONFIG_FILE="${1:-/etc/zrok/ctrl.yml}"

if [[ ! -f "$CONFIG_FILE" ]]; then
    echo "ERROR: Configuration file not found: $CONFIG_FILE" >&2
    echo "" >&2
    if [[ -e "$CONFIG_FILE" ]]; then
        echo "File exists but is not readable or not a regular file" >&2
        ls -la "$CONFIG_FILE" 2>&1 >&2 || true
        echo "" >&2
    fi
    if [[ -d "/etc/zrok" ]]; then
        echo "Directory /etc/zrok permissions:" >&2
        ls -lad "/etc/zrok" 2>&1 >&2 || true
        echo "" >&2
    fi
    echo "To configure the zrok controller, create the config file:" >&2
    echo "  sudo cp /etc/zrok/ctrl.yml.example $CONFIG_FILE" >&2
    echo "  sudo editor $CONFIG_FILE" >&2
    echo "" >&2
    echo "Then start the service:" >&2
    echo "  sudo systemctl start zrok-controller" >&2
    exit 1
fi

exec /usr/bin/zrok controller "$CONFIG_FILE"
