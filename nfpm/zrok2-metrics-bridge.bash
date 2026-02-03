#!/usr/bin/env bash
#
# zrok-controller-metrics-bridge.bash - Launcher script for zrok metrics bridge service
#
# Verifies config file exists before executing zrok controller metrics bridge

set -euo pipefail

CONFIG_FILE="${1:-/etc/zrok/ctrl.yml}"

if [[ ! -f "$CONFIG_FILE" ]]; then
    echo "ERROR: Configuration file not found: $CONFIG_FILE" >&2
    echo "" >&2
    echo "To configure the zrok metrics bridge, create a minimal config file:" >&2
    echo "" >&2
    cat >&2 <<'EOF'
  sudo tee /etc/zrok/ctrl.yml > /dev/null <<'YAML'
v: 4
bridge:
  source:
    type:           fileSource
    path:           /tmp/fabric-usage.json
  sink:
    type:           amqpSink
    url:            amqp://guest:guest@localhost:5672
    queue_name:     events
YAML
EOF
    echo "" >&2
    echo "Adjust the bridge source and sink settings as needed, then start the service:" >&2
    echo "  sudo systemctl start zrok-metrics" >&2
    exit 1
fi

exec /usr/bin/zrok controller metrics bridge "$CONFIG_FILE"
