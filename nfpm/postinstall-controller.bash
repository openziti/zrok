#!/usr/bin/env bash
#
# postinstall-controller.bash - Post-installation script for zrok-controller package
#

set -euo pipefail

# Package-specific configuration
SERVICE_NAME="zrok-controller"
SERVICE_USER="zrok-controller"
SERVICE_GROUP="zrok-controller"
SERVICE_HOME="/var/lib/zrok-controller"

# Source common functions from installed location
source /usr/share/zrok/postinstall-common.bash

# Run the postinstall logic
run_postinstall "$@"
