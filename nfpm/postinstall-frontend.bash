#!/usr/bin/env bash
#
# postinstall-frontend.bash - Post-installation script for zrok-frontend package
#

set -euo pipefail

# Package-specific configuration
SERVICE_NAME="zrok-frontend"

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/postinstall-common.bash"

# Run the postinstall logic
run_postinstall "$@"
