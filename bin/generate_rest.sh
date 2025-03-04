#!/bin/bash

set -euo pipefail

command -v swagger >/dev/null 2>&1 || {
  echo >&2 "command 'swagger' not installed. see: https://github.com/go-swagger/go-swagger for installation"
  exit 1
}

command -v swagger-codegen 2>&1 || {
  echo >&2 "command 'swagger-codegen' not installed. see: https://github.com/swagger-api/swagger-codegen for installation"
  exit 1
}

command -v openapi-generator-cli 2>&1 || {
  echo >&2 "command 'openapi-generator-cli' not installed. see: https://www.npmjs.com/package/@openapitools/openapi-generator-cli for installation"
  exit 1
}

command -v realpath 2>&1 || {
  echo >&2 "command 'realpath' not installed. see: https://www.npmjs.com/package/realpath for installation"
  exit 1
}

scriptPath=$(realpath "$0")
scriptDir=$(dirname "$scriptPath")

zrokDir=$(realpath "$scriptDir/..")

zrokSpec=$(realpath "$zrokDir/specs/zrok.yml")

echo "...clean generate zrok server/client"
for GEN in \
  ./rest_client_zrok/ \
  ./rest_model_zrok/ \
  ./rest_server_zrok/
do
  if [[ -d $GEN ]]
  then
    rm -rf "$GEN"
  else
    echo "WARN: not deleting $GEN because it does not exist" >&2
  fi
done

echo "...generating zrok server"
swagger generate server -P rest_model_zrok.Principal -f "$zrokSpec" -s rest_server_zrok -t "$zrokDir" -m "rest_model_zrok" --exclude-main

echo "...generating zrok client"
swagger generate client -P rest_model_zrok.Principal -f "$zrokSpec" -c rest_client_zrok -t "$zrokDir" -m "rest_model_zrok"

echo "...generating api console ts client"
rm -rf ui/src/api
openapi-generator-cli generate -i specs/zrok.yml -o ui/src/api -g typescript-fetch

echo "...generating agent console ts client"
rm -rf agent/agentUi/src/api
openapi-generator-cli generate -i agent/agentGrpc/agent.swagger.json -o agent/agentUi/src/api -g typescript-fetch

echo "...generating nodejs sdk ts client"
rm -rf sdk/nodejs/sdk/src/api
openapi-generator-cli generate -i specs/zrok.yml -o sdk/nodejs/sdk/src/api -g typescript-fetch

echo "...generating python sdk client"
# Delete tracked Python files
PYMOD=sdk/python/src
while IFS= read -r FILE
do
  if [[ -e "${PYMOD}/${FILE}" ]]
  then
    rm -f "${PYMOD}/${FILE}"
  fi
done < ${PYMOD}/.openapi-generator/FILES
# Delete the tracking file
rm -f "${PYMOD}/.openapi-generator/FILES"
# Generate and track new files
openapi-generator-cli generate -i specs/zrok.yml -o "${PYMOD}" -g python \
  --package-name zrok_api --additional-properties projectName=zrok

git checkout rest_server_zrok/configure_zrok.go
