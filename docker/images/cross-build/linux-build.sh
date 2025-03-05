#!/usr/bin/env bash
#
# build the Linux artifact for amd64, armhf, armel, or arm64
#

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

resolveArch() {
    case ${1} in
        arm|armv7*|arm/v7*) echo armhf
        ;;
        armv8*|arm/v8*) echo arm64
        ;;
        *) echo "${1}"
        ;;
    esac
}

# if no architectures supplied then default to amd64
if (( ${#} )); then
    typeset -a JOBS=(${@})
else
    typeset -a JOBS=(amd64)
fi

(
    HOME=/tmp/builder
    # Navigate to the "ui" directory and run npm commands
    mkdir -p $HOME
    npm config set cache /mnt/.npm
    for UI in ./ui ./agent/agentUi
    do
        pushd ${UI}
        npm install
        npm run build
        popd
    done
)

for ARCH in "${JOBS[@]}"; do
    goreleaser build \
    --clean \
    --snapshot \
    --output "./dist/" \
    --config "./.goreleaser-linux-$(resolveArch "${ARCH}").yml"
done

