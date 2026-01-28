#!/usr/bin/env bash
#
# postinstall-metrics.bash - Post-installation script for zrok-metrics package
#

set -euo pipefail

# Package-specific configuration
SERVICE_NAME="zrok-metrics"
SERVICE_USER="zrok-metrics"
SERVICE_GROUP="zrok-metrics"
SERVICE_HOME="/var/lib/zrok-metrics"
ADDITIONAL_GROUPS="zrok-controller"

# Source common functions from installed location
source /usr/share/zrok/postinstall-common.bash

# Run the postinstall logic
run_postinstall "$@"
