#!/bin/bash

set -euo pipefail

command -v swagger &>/dev/null || {
  echo >&2 "command 'swagger' not installed. see: https://github.com/go-swagger/go-swagger for installation"
  exit 1
}

command -v swagger-codegen &>/dev/null || {
  echo >&2 "command 'swagger-codegen' not installed. see: https://github.com/swagger-api/swagger-codegen for installation"
  exit 1
}

command -v openapi-generator-cli &>/dev/null || {
  echo >&2 "command 'openapi-generator-cli' not installed. see: https://www.npmjs.com/package/@openapitools/openapi-generator-cli for installation"
  exit 1
}

command -v realpath &>/dev/null || {
  echo >&2 "command 'realpath' not installed. see: https://www.npmjs.com/package/realpath for installation"
  exit 1
}

scriptPath=$(realpath "$0")
scriptDir=$(dirname "$scriptPath")

zrokDir=$(realpath "$scriptDir/..")
zrokSpec=$(realpath "$zrokDir/specs/zrok.yml")

# anti-oops in case user runs this script from somewhere else
if [[ "$(realpath "$zrokDir")" != "$(realpath "$(pwd)")" ]]
then
  echo "ERROR: must be run from zrok root" >&2
  exit 1
fi

echo "...clean generate zrok server/client"
rm -rf ./rest_*

echo "...generating zrok server"
swagger generate server -P rest_model_zrok.Principal -f "$zrokSpec" -s rest_server_zrok -t "$zrokDir" -m "rest_model_zrok" --exclude-main

echo "...generating zrok client"
swagger generate client -P rest_model_zrok.Principal -f "$zrokSpec" -c rest_client_zrok -t "$zrokDir" -m "rest_model_zrok"

echo "...generating api console ts client"
rm -rf ui/src/api
openapi-generator-cli generate -i "$zrokSpec" -o ui/src/api -g typescript-fetch

echo "...generating agent console ts client"
rm -rf agent/agentUi/src/api
openapi-generator-cli generate -i agent/agentGrpc/agent.swagger.json -o agent/agentUi/src/api -g typescript-fetch

echo "...generating nodejs sdk ts client"
rm -rf sdk/nodejs/sdk/src/api
openapi-generator-cli generate -i "$zrokSpec" -o sdk/nodejs/sdk/src/api -g typescript-fetch

echo "...generating python sdk client"
pyMod="./sdk/python/src"
rm -rf "$pyMod"/{zrok_api,docs,test,.gitignore,README.md,requirements.txt}
openapi-generator-cli generate -i "$zrokSpec" -o "$pyMod" -g python \
  --package-name zrok_api --additional-properties projectName=zrok

git checkout rest_server_zrok/configure_zrok.go
