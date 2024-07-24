#!/usr/bin/env bash
#
# build the Linux artifacts for amd64, arm, arm64
#
# runs one background job per desired architecture unless there are too few CPUs
#
# 

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

resolveArch() {
    case ${1} in
        arm|armv7*|arm/v7*|armhf*) echo armhf
        ;;
        armv8*|arm/v8*) echo arm64
        ;;
        *) echo "${1}"
        ;;
    esac
}

# if no architectures supplied then default list of three
if (( ${#} )); then
    typeset -a JOBS=(${@})
else
    typeset -a JOBS=(amd64 arm arm64)
fi

(
    HOME=/tmp/builder
    # Navigate to the "ui" directory and run npm commands
    npm config set cache /mnt/.npm
    cd ./ui/
    mkdir -p $HOME
    npm install
    npm run build
)

for ARCH in "${JOBS[@]}"; do
    goreleaser build \
    --clean \
    --snapshot \
    --output ./dist/ \
    --config "./.goreleaser-linux-$(resolveArch "${ARCH}").yml"
done

