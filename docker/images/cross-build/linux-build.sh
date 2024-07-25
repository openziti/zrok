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
    --output "./dist/" \
    --config "./.goreleaser-linux-$(resolveArch "${ARCH}").yml"
done

