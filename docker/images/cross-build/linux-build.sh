#!/usr/bin/env bash
#
# build the Linux artifact for amd64, armhf, armel, or arm64
#

set -o errexit
set -o nounset
set -o pipefail

# Check for --verbose flag
VERBOSE=false
ARGS=()
for arg in "$@"; do
    if [[ "$arg" == "--verbose" ]]; then
        VERBOSE=true
    else
        ARGS+=("$arg")
    fi
done

# Enable xtrace only if verbose
if [[ "$VERBOSE" == "true" ]]; then
    set -o xtrace
fi

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
if (( ${#ARGS[@]} )); then
    typeset -a JOBS=("${ARGS[@]}")
else
    typeset -a JOBS=(amd64)
fi

# Redirect output if not verbose
if [[ "$VERBOSE" == "false" ]]; then
    exec 3>&1 4>&2  # Save original stdout/stderr
    exec 1>/dev/null 2>&1  # Redirect to /dev/null
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
        npm ci
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

# Track built binaries for report
typeset -a BUILT_BINARIES=()

for ARCH in "${JOBS[@]}"; do
    LDFLAGS="-s -w -X 'github.com/openziti/zrok/v2/build.Version=${VERSION}' -X 'github.com/openziti/zrok/v2/build.Hash=${HASH}'"
    BINARY_PATH="./dist/$(resolveArch "${ARCH}")/linux/zrok2"
    GOOS=linux GOARCH=$(resolveArch "${ARCH}") \
    go build -o "${BINARY_PATH}" \
    -ldflags "${LDFLAGS}" \
    ./cmd/zrok
    BUILT_BINARIES+=("${BINARY_PATH}")
done

# Restore stdout/stderr and print summary if not verbose
if [[ "$VERBOSE" == "false" ]]; then
    exec 1>&3 2>&4  # Restore original stdout/stderr
    exec 3>&- 4>&-  # Close saved descriptors
    
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "✓ Build completed successfully"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "Version: ${VERSION}"
    echo "Hash:    ${HASH}"
    echo ""
    echo "Built binaries:"
    for BINARY in "${BUILT_BINARIES[@]}"; do
        echo "  • ${BINARY}"
    done
    echo ""
    echo "Embedded UIs:"
    echo "  • ./ui/dist           → /api/v1/static (main UI)"
    echo "  • ./agent/agentUi/dist → /agent (agent UI)"
    echo ""
fi
