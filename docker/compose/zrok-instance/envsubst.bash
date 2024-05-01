#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# Read the shell-format template from stdin
TEMPLATE=$(cat)

# consume args as default values from a file or a list of shell commands
while (( $# )); do
    if [ -s "$1" ]; then
        # Read the default values from the file
        # shellcheck disable=SC1090
        source "$1"
    else
        # Use the argument as a shell command
        eval "$1"
    fi
    shift
done

if [ -z "$TEMPLATE" ]; then
    echo "Error: no template provided on stdin" >&2
    exit 1
fi

# obtain the list of required variables
read -ra VARIABLES <<< "$(envsubst "$TEMPLATE" --variables)"

# Check that all required variables are set
for var in "${VARIABLES[@]}"; do
    if [ -z "${!var:-}" ]; then
        echo "Error: $var is null" >&2
        exit 1
    else
        export "${var?}"
    fi
done

# Render the template to stdout
envsubst <<< "$TEMPLATE" 
