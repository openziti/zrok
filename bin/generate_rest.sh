#!/bin/bash

set -euo pipefail

# function to build the combined OpenAPI spec from modular source files
build_spec() {
  local src_dir="$1/specs/src"
  local target_file="$2"

  echo "...building combined spec from modular sources"

  # start
  cat "$src_dir/head.yml" > "$target_file"

  # add paths section
  echo "" >> "$target_file"
  echo "paths:" >> "$target_file"

  # combine all path files
  for tag_file in "$src_dir"/{account.yml,admin.yml,agent.yml,environment.yml,metadata.yml,share.yml}; do
    if [[ -f "$tag_file" ]]; then
      # add comment for section and paths with proper indentation
      echo "  #" >> "$target_file"
      echo "  # $(basename "$tag_file")" >> "$target_file"
      echo "  #" >> "$target_file"
      sed 's/^/  /' "$tag_file" >> "$target_file"
      echo "" >> "$target_file"
    fi
  done

  # add definitions section
  echo "" >> "$target_file"
  echo "definitions:" >> "$target_file"
  if [[ -f "$src_dir/definitions.yml" ]]; then
    # add definitions with proper indentation
    sed 's/^/  /' "$src_dir/definitions.yml" >> "$target_file"
  fi

  # add final configuration elements
  cat "$src_dir/tail.yml" >> "$target_file"
}

command -v swagger &>/dev/null || {
  echo >&2 "command 'swagger' not installed. see: https://github.com/go-swagger/go-swagger for installation"
  exit 1
}

command -v openapi-generator-cli &>/dev/null || {
  echo >&2 "command 'openapi-generator-cli' not installed. see: https://www.npmjs.com/package/@openapitools/openapi-generator-cli for installation"
  exit 1
}

# these are the generator versions that reproduce the checked-in tree; any other version rewrites
# generated files the spec change did not touch. openapi-generator-cli needs java; without it, the
# openapitools/openapi-generator-cli:v7.14.0 docker image can stand in for the command.
requiredSwagger="v0.33.2"
requiredOpenapiGenerator="7.14.0"
swaggerVersion=$(swagger version 2>/dev/null | sed -n 's/^version: //p')
openapiGeneratorVersion=$(openapi-generator-cli version 2>/dev/null | grep -Eo '[0-9]+\.[0-9]+\.[0-9]+' | tail -n 1)
if [[ "$swaggerVersion" != "$requiredSwagger" || "$openapiGeneratorVersion" != "$requiredOpenapiGenerator" ]]; then
  echo >&2 "ERROR: requires swagger '$requiredSwagger' (found '$swaggerVersion') and openapi-generator-cli '$requiredOpenapiGenerator' (found '$openapiGeneratorVersion')"
  exit 1
fi

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

# build the combined spec from modular sources
build_spec "$zrokDir" "$zrokSpec"

echo "...clean generate zrok server/client"
rm -rf ./rest_client_zrok ./rest_server_zrok ./rest_model_zrok

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
