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
    # pwd is probably /mnt mountpoint in the container
    npm config set cache $(pwd)/.npm
    for UI in ./ui ./agent/agentUi
    do
        pushd ${UI}
        npm install
        npm run build
        popd
    done
)

# Get version information
VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v2.0.x")
STEPS=$(git rev-list --count ${VERSION}..HEAD 2>/dev/null || echo "0")
if [ "$STEPS" -gt "0" ]; then
    VERSION="${VERSION}-${STEPS}"
fi

# Check if working copy is dirty
if [ -z "$(git status --porcelain)" ]; then
    # Clean working directory
    HASH=$(git rev-parse --short HEAD)
else
    # Dirty working directory
    HASH="developer build"
fi

for ARCH in "${JOBS[@]}"; do
    LDFLAGS="-s -w -X 'github.com/openziti/zrok/v2/build.Version=${VERSION}' -X 'github.com/openziti/zrok/v2/build.Hash=${HASH}'"
    GOOS=linux GOARCH=$(resolveArch "${ARCH}") \
    go build -o "./dist/$(resolveArch "${ARCH}")/linux/zrok2" \
    -ldflags "${LDFLAGS}" \
    ./cmd/zrok
done
