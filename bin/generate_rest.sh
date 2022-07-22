#!/bin/bash

set -euo pipefail

command -v swagger >/dev/null 2>&1 || {
  echo >&2 "Command 'swagger' not installed. See: https://github.com/go-swagger/go-swagger for installation"
  exit 1
}

scriptPath=$(realpath $0)
scriptDir=$(dirname "$scriptPath")

zrokDir=$(realpath "$scriptDir/..")

zrokSpec=$(realpath "$zrokDir/specs/zrok.yml")

echo "...generating zrok server"
swagger generate server -f "$zrokSpec" -s rest_zrok_server -t "$zrokDir" -m "rest_model"

echo "...generating zrok client"
swagger generate client -f "$zrokSpec" -c rest_zrok_client -t "$zrokDir" -m "rest_model"