#!/usr/bin/env bash
#
# postinstall-frontend.bash - Post-installation script for zrok-frontend package
#

set -euo pipefail

# Package-specific configuration
SERVICE_NAME="zrok-frontend"
SERVICE_USER="zrok-frontend"
SERVICE_GROUP="zrok-frontend"
SERVICE_HOME="/var/lib/zrok-frontend"

# Source common functions from installed location
source /usr/share/zrok/postinstall-common.bash

# Run the postinstall logic
run_postinstall "$@"
