#!/usr/bin/env bash
#
# build the Linux artifact for amd64, armhf, armel, or arm64
#

set -o errexit
set -o nounset
set -o pipefail

# Set HOME to a writable location when running as non-root user (--user flag)
# This ensures Go cache and other tools work properly
export HOME=/tmp/builder
mkdir -p "${HOME}"

# Ensure Go cache directories are set and writable
# GOCACHE should already be set by Dockerfile to /usr/share/go_cache (mounted from host)
# GOMODCACHE defaults to $GOPATH/pkg/mod which should be /usr/share/go/pkg/mod
export GOCACHE="${GOCACHE:-/usr/share/go_cache}"
export GOMODCACHE="${GOMODCACHE:-/usr/share/go/pkg/mod}"

# Verify cache directories are writable
if [[ ! -w "${GOCACHE}" ]]; then
    echo "WARNING: GOCACHE directory not writable: ${GOCACHE}" >&2
    echo "This may cause slow builds. Check volume mount permissions." >&2
fi
if [[ ! -w "${GOMODCACHE}" ]]; then
    echo "WARNING: GOMODCACHE directory not writable: ${GOMODCACHE}" >&2
    echo "This may cause slow builds. Check volume mount permissions." >&2
fi

usage() {
    cat >&2 <<EOF
Usage: $(basename "$0") [OPTIONS] [ARCHITECTURE]

Build zrok2 binary for a specified Linux architecture using goreleaser snapshots.

ARCHITECTURE:
    amd64           x86_64 architecture (default if none specified)
    arm64           ARM 64-bit (aarch64)
    armhf           ARM 32-bit hard-float (armv7)
    armel           ARM 32-bit soft-float (armv7)

OPTIONS:
    --packages      Build all nfpm packages (deb, rpm) and report file locations
    --verbose       Show full build output (npm, goreleaser, etc.)
    --debug         Enable bash xtrace for debugging (implies --verbose)
    -h, --help      Show this help message

ENVIRONMENT VARIABLES:
    VERBOSE=1       Same as --verbose flag
    DEBUG=1         Same as --debug flag (implies VERBOSE=1)

EXAMPLES:
    # Build for amd64 (default, quiet mode)
    $(basename "$0")

    # Build for arm64
    $(basename "$0") arm64

    # Build with verbose output
    $(basename "$0") --verbose arm64

    # Build with debug tracing
    $(basename "$0") --debug amd64

    # Build packages (deb, rpm)
    $(basename "$0") --packages amd64

OUTPUT:
    Binaries are placed in ./dist/<binary>_linux_<arch>_<variant>/zrok2

NOTE:
    Only one architecture can be built per run. The ./dist/ directory is
    cleaned at the start of each build. To build multiple architectures,
    run this script multiple times with different architecture arguments.

EOF
    exit "${1:-1}"
}

# Check for --verbose, --debug, and --packages flags, or environment variables
VERBOSE=false
DEBUG=false
BUILD_PACKAGES=false

# Check environment variables first
if [[ "${DEBUG:-}" == "1" ]]; then
    DEBUG=true
    VERBOSE=true  # DEBUG=1 implies VERBOSE=1
fi
if [[ "${VERBOSE:-}" == "1" ]]; then
    VERBOSE=true
fi

# Process command line arguments (flags override environment variables)
ARGS=()
for arg in "$@"; do
    case "$arg" in
        --packages)
            BUILD_PACKAGES=true
            ;;
        --debug)
            DEBUG=true
            VERBOSE=true  # --debug implies --verbose
            ;;
        --verbose)
            VERBOSE=true
            ;;
        -h|--help)
            usage 0
            ;;
        amd64|arm64|armhf|armel)
            ARGS+=("$arg")
            ;;
        -*)
            echo "Error: Unknown option: $arg" >&2
            echo "" >&2
            usage 1
            ;;
        *)
            echo "Error: Unknown architecture: $arg" >&2
            echo "Valid architectures: amd64, arm64, armhf, armel" >&2
            echo "" >&2
            usage 1
            ;;
    esac
done

# Enable xtrace if verbose or debug
if [[ "$VERBOSE" == "true" ]]; then
    set -o xtrace
fi

# Export SHELLOPTS to propagate xtrace to called shell scripts if debug mode
if [[ "$DEBUG" == "true" ]]; then
    export SHELLOPTS
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
    # Only one architecture can be built per run due to goreleaser's dist handling
    if (( ${#ARGS[@]} > 1 )); then
        echo "Error: Only one architecture can be built per run." >&2
        echo "The ./dist/ directory is cleaned at the start of each build." >&2
        echo "To build multiple architectures, run this script multiple times." >&2
        echo "" >&2
        echo "Example:" >&2
        echo "  $(basename "$0") arm64" >&2
        echo "  $(basename "$0") amd64" >&2
        exit 1
    fi
    typeset -a JOBS=("${ARGS[@]}")
else
    typeset -a JOBS=(amd64)
fi

# Redirect output if not verbose (keep stderr visible for errors)
if [[ "$VERBOSE" == "false" ]]; then
    exec 3>&1  # Save original stdout
    exec 1>/dev/null  # Redirect stdout to /dev/null, but keep stderr visible
fi

(
    # pwd is probably /mnt mountpoint in the container
    npm config set cache $(pwd)/.npm
    for UI in ./ui ./agent/agentUi
    do
        pushd ${UI}
        if [[ "$VERBOSE" == "true" ]]; then
            npm ci
            npm run build
        else
            npm ci 2>/dev/null
            npm run build 2>/dev/null
        fi
        popd
    done
)

# Track built binaries and packages for report
typeset -a BUILT_BINARIES=()
typeset -a BUILT_PACKAGES=()
typeset -a GORELEASER_CONFIGS=()

# Track if this is the first build (for --clean flag)
FIRST_BUILD=true

# Build with goreleaser for each architecture
for ARCH in "${JOBS[@]}"; do
    RESOLVED_ARCH=$(resolveArch "${ARCH}")
    CONFIG_FILE=".goreleaser-linux-${RESOLVED_ARCH}.yml"
    
    if [[ ! -f "${CONFIG_FILE}" ]]; then
        echo "Error: GoReleaser config not found: ${CONFIG_FILE}" >&2
        exit 1
    fi
    
    # Run goreleaser in snapshot mode (tolerates dirty working copy)
    # Use --clean only on first build to allow multi-arch builds
    if [[ "$FIRST_BUILD" == "true" ]]; then
        CLEAN_FLAG="--clean"
        FIRST_BUILD=false
    else
        CLEAN_FLAG=""
    fi
    
    # Determine which goreleaser command to use
    if [[ "$BUILD_PACKAGES" == "true" ]]; then
        # Use 'release' to build packages, skip publishing steps
        if [[ "$VERBOSE" == "true" ]]; then
            goreleaser release --snapshot ${CLEAN_FLAG} --skip=publish,validate --config "${CONFIG_FILE}"
        else
            goreleaser release --snapshot ${CLEAN_FLAG} --skip=publish,validate --config "${CONFIG_FILE}" >/dev/null 2>&1
        fi
    else
        # Use 'build' for binaries only
        if [[ "$VERBOSE" == "true" ]]; then
            goreleaser build --snapshot ${CLEAN_FLAG} --config "${CONFIG_FILE}"
        else
            goreleaser build --snapshot ${CLEAN_FLAG} --config "${CONFIG_FILE}" >/dev/null 2>&1
        fi
    fi
    
    # Track the binary location (goreleaser uses dist/*_linux_<goarch>*/zrok2)
    # Map our architecture names to goreleaser's output directory patterns
    case "${RESOLVED_ARCH}" in
        amd64)
            PATTERN="*_linux_amd64*"
            ;;
        arm64)
            PATTERN="*_linux_arm64*"
            ;;
        armhf)
            PATTERN="*_linux_arm_6"
            ;;
        armel)
            PATTERN="*_linux_arm_7"
            ;;
        *)
            PATTERN="*_linux_${RESOLVED_ARCH}*"
            ;;
    esac
    
    # Find all zrok2 binaries in dist/ that match the architecture
    shopt -s nullglob
    BINARY_PATTERN="./dist/${PATTERN}/zrok2"
    FOUND=false
    for BINARY in ${BINARY_PATTERN}; do
        if [[ -f "${BINARY}" ]]; then
            BUILT_BINARIES+=("${BINARY}")
            GORELEASER_CONFIGS+=("${CONFIG_FILE}")
            FOUND=true
            break
        fi
    done
    shopt -u nullglob
    
    if [[ "$FOUND" == "false" ]]; then
        echo "Warning: Could not find binary for architecture ${RESOLVED_ARCH}" >&2
    fi
    
    # Find packages if --packages was specified
    if [[ "$BUILD_PACKAGES" == "true" ]]; then
        shopt -s nullglob
        for PKG in ./dist/*.deb ./dist/*.rpm; do
            if [[ -f "${PKG}" ]]; then
                BUILT_PACKAGES+=("${PKG}")
            fi
        done
        shopt -u nullglob
    fi
done

# Restore stdout and print summary if not verbose
if [[ "$VERBOSE" == "false" ]]; then
    exec 1>&3  # Restore original stdout
    exec 3>&-  # Close saved descriptor
    
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "✓ Build completed successfully (goreleaser snapshot)"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "GoReleaser output directory: ./dist/"
    echo ""
    echo "Built binaries:"
    if [[ ${#BUILT_BINARIES[@]} -eq 0 ]]; then
        echo "  (none found - check for errors above)"
    else
        for i in "${!BUILT_BINARIES[@]}"; do
            echo "  • ${BUILT_BINARIES[$i]}"
            echo "    (config: ${GORELEASER_CONFIGS[$i]})"
        done
    fi
    echo ""
    echo "Embedded UIs:"
    echo "  • ./ui/dist           → /api/v1/static (main UI)"
    echo "  • ./agent/agentUi/dist → /agent (agent UI)"
    echo ""
    
    if [[ "$BUILD_PACKAGES" == "true" ]]; then
        echo "Built packages:"
        if [[ ${#BUILT_PACKAGES[@]} -eq 0 ]]; then
            echo "  (none found - check for errors above)"
        else
            for PKG in "${BUILT_PACKAGES[@]}"; do
                echo "  • ${PKG}"
            done
        fi
        echo ""
    fi
    
    echo "Note: GoReleaser also generates archives and metadata in ./dist/"
    echo ""
fi
