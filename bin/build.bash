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
    --docker        Build the zrok2 container image (implies binary build)
    --push          Push the container image to the registry (requires --docker)
    --tag TAG       Container image tag (default: docker.io/openziti/zrok2-build-debugging:0.0.0-<short-sha>)
    -h, --help      Show this help message

ENVIRONMENT VARIABLES:
    VERBOSE=1       Same as --verbose flag
    GOCACHE         Go build cache directory (default: \$HOME/.cache/go-build)

EXAMPLES:
    # Build for host architecture (quiet mode)
    bin/build.bash

    # Build with verbose output
    bin/build.bash --verbose

    # Build the container image
    bin/build.bash --docker

    # Build and push with a custom tag
    bin/build.bash --docker --push --tag docker.io/openziti/zrok2-build-debugging:0.0.0-$(git rev-parse --short HEAD)

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
BUILD_DOCKER=false
PUSH_IMAGE=false
IMAGE_TAG=""

while (( $# )); do
    case "$1" in
        --verbose)
            VERBOSE=true
            ;;
        --docker)
            BUILD_DOCKER=true
            ;;
        --push)
            PUSH_IMAGE=true
            ;;
        --tag)
            shift
            IMAGE_TAG="${1:?--tag requires a value}"
            ;;
        -h|--help)
            usage 0
            ;;
        *)
            echo "Error: Unknown argument: $1" >&2
            echo "" >&2
            usage 1
            ;;
    esac
    shift
done

if [[ "${PUSH_IMAGE}" == "true" && "${BUILD_DOCKER}" == "false" ]]; then
    echo "Error: --push requires --docker" >&2
    exit 1
fi

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

# Check if the builder image needs to be (re)built.  A sentinel file records
# when we last successfully built the image — compare its mtime against the
# Dockerfile and build script.  This avoids relying on the image's internal
# .Created metadata which reflects cached layer timestamps, not wall-clock.
NEEDS_BUILD=false
SENTINEL=".docker-builder-built"
DOCKERFILE="./docker/images/cross-build/Dockerfile"
BUILD_SCRIPT="./docker/images/cross-build/linux-build.sh"

if ! docker image inspect zrok-builder &> /dev/null; then
    NEEDS_BUILD=true
    echo "zrok-builder image not found, building..." >&2
elif [[ ! -f "${SENTINEL}" ]]; then
    NEEDS_BUILD=true
    echo "zrok-builder sentinel missing, rebuilding..." >&2
else
    for FILE in "${DOCKERFILE}" "${BUILD_SCRIPT}"; do
        if [[ -f "${FILE}" && "${FILE}" -nt "${SENTINEL}" ]]; then
            NEEDS_BUILD=true
            echo "zrok-builder image is stale (${FILE} modified), rebuilding..." >&2
            break
        fi
    done
fi

if [[ "${NEEDS_BUILD}" == "true" ]]; then
    if [[ "${VERBOSE}" == "true" ]]; then
        docker buildx build -t zrok-builder ./docker/images/cross-build --load
    else
        docker buildx build -t zrok-builder ./docker/images/cross-build --load > /dev/null 2>&1
    fi
    touch "${SENTINEL}"
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

# Run the binary build
docker run --user "$(id -u):$(id -g)" --rm \
    --volume="${GOCACHE_DIR}:/usr/share/go_cache" \
    --volume="${GOMODCACHE_DIR}:/usr/share/go/pkg/mod" \
    --volume="${PWD}:/mnt" \
    zrok-builder "${BUILD_ARCH}"

# ── Container image build ─────────────────────────────────────────────────────
if [[ "${BUILD_DOCKER}" == "true" ]]; then
    if [[ -z "${IMAGE_TAG}" ]]; then
        SHORT_SHA=$(git rev-parse --short HEAD)
        IMAGE_TAG="docker.io/openziti/zrok2-build-debugging:0.0.0-${SHORT_SHA}"
    fi

    # Stage the binary where the Dockerfile expects it:
    #   ${ARTIFACTS_DIR}/${TARGETARCH}/${TARGETOS}/zrok2
    # goreleaser puts it in ./dist/*_linux_<goarch>*/zrok2
    DIST_DIR="./dist/${BUILD_ARCH}/linux"
    mkdir -p "${DIST_DIR}"

    GORELEASER_BINARY=$(find ./dist -maxdepth 2 -name zrok2 -path "*linux*" -print -quit 2>/dev/null)
    if [[ -z "${GORELEASER_BINARY}" ]]; then
        echo "Error: could not find zrok2 binary in ./dist/" >&2
        exit 1
    fi
    cp "${GORELEASER_BINARY}" "${DIST_DIR}/zrok2"
    echo "Staged binary: ${GORELEASER_BINARY} → ${DIST_DIR}/zrok2" >&2

    BUILDX_ARGS=(
        --build-arg "ARTIFACTS_DIR=./dist"
        --build-arg "DOCKER_BUILD_DIR=./docker/images/zrok"
        --file ./docker/images/zrok/Dockerfile
        --platform "linux/${BUILD_ARCH}"
        --tag "${IMAGE_TAG}"
    )

    if [[ "${PUSH_IMAGE}" == "true" ]]; then
        BUILDX_ARGS+=(--push)
    else
        BUILDX_ARGS+=(--load)
    fi

    echo "Building container image: ${IMAGE_TAG}" >&2
    if [[ "${VERBOSE}" == "true" ]]; then
        docker buildx build "${BUILDX_ARGS[@]}" .
    else
        docker buildx build "${BUILDX_ARGS[@]}" . 2>&1 | tail -5
    fi

    echo "Container image: ${IMAGE_TAG}" >&2
fi
