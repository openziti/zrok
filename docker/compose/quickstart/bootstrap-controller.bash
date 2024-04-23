#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

for arg in "${@}"; do
    if [[ ! "${arg}" =~ ^- && -s "${arg}" ]]; then
        CONFIG="${arg}"
        break
    fi
done

if [[ -z "${CONFIG}" ]]; then
    echo "ERROR: args '${*}' do not contain a non-empty config file" >&2
    exit 1
fi

# config.yml is first param
zrok admin bootstrap --skip-frontend "${CONFIG}"

exec "$@"
