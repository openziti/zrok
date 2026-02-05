#!/usr/bin/env bash
#
# zrok2-controller-run.bash - Launcher script for zrok2 controller service
#
# Verifies config file exists before executing zrok2 controller

set -euo pipefail

CONFIG_FILE="${1:-/etc/zrok2/ctrl.yml}"

if [[ ! -f "$CONFIG_FILE" ]]; then
    echo "ERROR: Configuration file not found: $CONFIG_FILE" >&2
    echo "" >&2
    if [[ -e "$CONFIG_FILE" ]]; then
        echo "File exists but is not readable or not a regular file" >&2
        ls -la "$CONFIG_FILE" 2>&1 >&2 || true
        echo "" >&2
    fi
    if [[ -d "/etc/zrok2" ]]; then
        echo "Directory /etc/zrok2 permissions:" >&2
        ls -lad "/etc/zrok2" 2>&1 >&2 || true
        echo "" >&2
    fi
    echo "To configure the zrok2 controller, create the config file:" >&2
    echo "  sudo cp /etc/zrok2/ctrl.yml.example $CONFIG_FILE" >&2
    echo "  sudo editor $CONFIG_FILE" >&2
    echo "" >&2
    echo "Then start the service:" >&2
    echo "  sudo systemctl start zrok2-controller" >&2
    exit 1
fi

exec /usr/bin/zrok2 controller "$CONFIG_FILE"
