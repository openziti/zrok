#!/usr/bin/env bash
#
# Wrapper script to build zrok for the host architecture using the cross-build container
# Suppresses all output except errors and the final build report
#

set -o errexit
set -o nounset
set -o pipefail

usage() {
    cat >&2 <<EOF
Usage: $(basename "$0") [OPTIONS]

Build zrok for the host architecture using Docker.

OPTIONS:
    --verbose       Show full build output including Docker build progress
    -h, --help      Show this help message

ENVIRONMENT VARIABLES:
    VERBOSE=1       Same as --verbose flag
    GOCACHE         Go build cache directory (default: \$HOME/.cache/go-build)

EXAMPLES:
    # Build for host architecture (quiet mode)
    ./build.bash

    # Build with verbose output
    ./build.bash --verbose

    # Build with verbose output using environment variable
    VERBOSE=1 ./build.bash

ADVANCED USAGE:
    For more control including cross-compilation for multiple architectures,
    debug mode, and additional options, use the build script directly:

        docker/images/cross-build/linux-build.sh --help

    Or see the cross-build documentation:
    https://github.com/openziti/zrok/blob/main/docker/images/cross-build/README.md

EOF
    exit "${1:-1}"
}

# Check for --verbose flag or VERBOSE environment variable
VERBOSE=false
if [[ "${VERBOSE:-}" == "1" ]]; then
    VERBOSE=true
fi

for arg in "$@"; do
    case "$arg" in
        --verbose)
            VERBOSE=true
            ;;
        -h|--help)
            usage 0
            ;;
        *)
            echo "Error: Unknown argument: $arg" >&2
            echo "" >&2
            usage 1
            ;;
    esac
done

# Detect host architecture
HOST_ARCH=$(uname -m)
case "${HOST_ARCH}" in
    x86_64)
        BUILD_ARCH="amd64"
        ;;
    aarch64|arm64)
        BUILD_ARCH="arm64"
        ;;
    armv7l)
        BUILD_ARCH="armhf"
        ;;
    *)
        echo "Error: Unsupported architecture: ${HOST_ARCH}" >&2
        exit 1
        ;;
esac

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    echo "Error: Docker is not installed or not in PATH" >&2
    exit 1
fi

# Check if the builder image exists and if it's stale
NEEDS_BUILD=false

if ! docker image inspect zrok-builder &> /dev/null; then
    # Image doesn't exist
    NEEDS_BUILD=true
    echo "zrok-builder image not found, building..." >&2
else
    # Image exists, check if it's stale
    IMAGE_CREATED=$(docker image inspect zrok-builder --format='{{.Created}}' 2>/dev/null)
    IMAGE_CREATED_EPOCH=$(date -d "${IMAGE_CREATED}" +%s 2>/dev/null || date -j -f "%Y-%m-%dT%H:%M:%S" "${IMAGE_CREATED%.*}" +%s 2>/dev/null)
    
    # Check if Dockerfile or linux-build.sh are newer than the image
    DOCKERFILE="./docker/images/cross-build/Dockerfile"
    BUILD_SCRIPT="./docker/images/cross-build/linux-build.sh"
    
    for FILE in "${DOCKERFILE}" "${BUILD_SCRIPT}"; do
        if [[ -f "${FILE}" ]]; then
            FILE_MODIFIED_EPOCH=$(stat -c %Y "${FILE}" 2>/dev/null || stat -f %m "${FILE}" 2>/dev/null)
            if [[ "${FILE_MODIFIED_EPOCH}" -gt "${IMAGE_CREATED_EPOCH}" ]]; then
                NEEDS_BUILD=true
                echo "zrok-builder image is stale (${FILE} modified), rebuilding..." >&2
                break
            fi
        fi
    done
fi

if [[ "${NEEDS_BUILD}" == "true" ]]; then
    if [[ "${VERBOSE}" == "true" ]]; then
        docker buildx build -t zrok-builder ./docker/images/cross-build --load
    else
        docker buildx build -t zrok-builder ./docker/images/cross-build --load > /dev/null 2>&1
    fi
fi

# Ensure cache directories exist with correct ownership before docker creates them as root
GOCACHE_DIR="${GOCACHE:-${HOME}/.cache/go-build}"
GOMODCACHE_DIR="${GOMODCACHE:-${HOME}/.cache/go-mod}"

if [[ ! -d "${GOCACHE_DIR}" ]]; then
    mkdir -p "${GOCACHE_DIR}"
fi

if [[ ! -d "${GOMODCACHE_DIR}" ]]; then
    mkdir -p "${GOMODCACHE_DIR}"
fi

# Run the build
docker run --user "${UID}" --rm \
    --volume="${GOCACHE_DIR}:/usr/share/go_cache" \
    --volume="${GOMODCACHE_DIR}:/usr/share/go/pkg/mod" \
    --volume="${PWD}:/mnt" \
    zrok-builder "${BUILD_ARCH}"
