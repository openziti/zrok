#!/bin/bash

set -euo pipefail

command -v swagger >/dev/null 2>&1 || {
  echo >&2 "command 'swagger' not installed. see: https://github.com/go-swagger/go-swagger for installation"
  exit 1
}

command -v openapi-generator-cli >/dev/null 2>&1 || {
  echo >&2 "command 'openapi-generator-cli' not installed. see: https://github.com/OpenAPITools/openapi-generator-cli#installation for installation"
}

command -v swagger-codegen 2>&1 || {
  echo >&2 "command 'swagger-codegen. see: https://github.com/swagger-api/swagger-codegen for installation"
}

scriptPath=$(realpath $0)
scriptDir=$(dirname "$scriptPath")

zrokDir=$(realpath "$scriptDir/..")

zrokSpec=$(realpath "$zrokDir/specs/zrok.yml")

pythonConfig=$(realpath "$zrokDir/bin/python_config.json")

echo "...generating zrok server"
swagger generate server -P rest_model_zrok.Principal -f "$zrokSpec" -s rest_server_zrok -t "$zrokDir" -m "rest_model_zrok" --exclude-main

echo "...generating zrok client"
swagger generate client -P rest_model_zrok.Principal -f "$zrokSpec" -c rest_client_zrok -t "$zrokDir" -m "rest_model_zrok"

echo "...generating js client"
openapi-generator-cli generate -i specs/zrok.yml -o ui/src/api -g javascript --additional-properties=usePromises=true,useES6=true

echo "...generating python client"
swagger-codegen generate -i specs/zrok.yml -o sdk/python/sdk/zrok -c $pythonConfig -l python

git checkout rest_server_zrok/configure_zrok.go
