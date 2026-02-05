# Cross-build Container for zrok

This container provides advanced build capabilities for zrok, including cross-compilation for multiple Linux architectures. It produces snapshot executables from the current checkout, even if dirty.

**Supported architectures:** `amd64`, `arm64`, `armhf` (arm/v7 hard-float), `armel` (arm/v7 soft-float)

**For simple builds:** Use `./build.bash` in the project root (see [BUILD.md](../../../BUILD.md)). This README covers advanced usage.

## Build the Container Image

The `build.bash` script handles this automatically, but you can build manually if needed:

```bash
# From the project root
docker buildx build -t zrok-builder ./docker/images/cross-build --load
```

**Note:** The image is automatically rebuilt when the Dockerfile or `linux-build.sh` changes.

## Usage

### Basic Build

Run from the project root:

```bash
docker run --user "$(id -u):$(id -g)" --rm \
  --volume="${GOCACHE:-${HOME}/.cache/go-build}:/usr/share/go_cache" \
  --volume="${GOMODCACHE:-${HOME}/.cache/go-mod}:/usr/share/go/pkg/mod" \
  --volume="${PWD}:/mnt" \
  zrok-builder arm64
```

**What happens:**

1. Mounts project root to `/mnt` in container
2. Mounts Go build cache (`GOCACHE`) for faster compilation
3. Mounts Go module cache (`GOMODCACHE`) to avoid re-downloading dependencies
4. Builds UI components with npm/vite
5. Builds Go binary with goreleaser snapshot
6. Outputs binary to `./dist/<binary>_linux_<arch>_<variant>/zrok`

**Output (quiet mode):**

```text
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✓ Build completed successfully (goreleaser snapshot)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

GoReleaser output directory: ./dist/

Built binaries:
  • ./dist/zrok-armv8_linux_arm64_v8.0/zrok
    (config: .goreleaser-linux-arm64.yml)

Embedded UIs:
  • ./ui/dist           → /api/v1/static (main UI)
  • ./agent/agentUi/dist → /agent (agent UI)

Note: GoReleaser also generates archives and metadata in ./dist/
```

### Multiple Architectures

```bash
# Default: amd64 if no architecture specified
docker run --user "$(id -u):$(id -g)" --rm \
  --volume="${GOCACHE:-${HOME}/.cache/go-build}:/usr/share/go_cache" \
  --volume="${GOMODCACHE:-${HOME}/.cache/go-mod}:/usr/share/go/pkg/mod" \
  --volume="${PWD}:/mnt" \
  zrok-builder

# Build different architectures in separate runs
docker run --user "$(id -u):$(id -g)" --rm \
  --volume="${GOCACHE:-${HOME}/.cache/go-build}:/usr/share/go_cache" \
  --volume="${GOMODCACHE:-${HOME}/.cache/go-mod}:/usr/share/go/pkg/mod" \
  --volume="${PWD}:/mnt" \
  zrok-builder arm64

docker run --user "$(id -u):$(id -g)" --rm \
  --volume="${GOCACHE:-${HOME}/.cache/go-build}:/usr/share/go_cache" \
  --volume="${GOMODCACHE:-${HOME}/.cache/go-mod}:/usr/share/go/pkg/mod" \
  --volume="${PWD}:/mnt" \
  zrok-builder armhf
```

**Note:** Each architecture must be built in a separate `docker run` command. Goreleaser cleans the `./dist/` directory at the start of each build, so multiple architectures in a single run are not supported.

### Verbose Output

Show full npm, vite, and goreleaser output:

```bash
# Using flag
docker run --user "$(id -u):$(id -g)" --rm \
  --volume="${GOCACHE:-${HOME}/.cache/go-build}:/usr/share/go_cache" \
  --volume="${GOMODCACHE:-${HOME}/.cache/go-mod}:/usr/share/go/pkg/mod" \
  --volume="${PWD}:/mnt" \
  zrok-builder --verbose arm64

# Using environment variable
docker run --user "$(id -u):$(id -g)" --rm \
  --env VERBOSE=1 \
  --volume="${GOCACHE:-${HOME}/.cache/go-build}:/usr/share/go_cache" \
  --volume="${GOMODCACHE:-${HOME}/.cache/go-mod}:/usr/share/go/pkg/mod" \
  --volume="${PWD}:/mnt" \
  zrok-builder arm64
```

**Shows:** npm install/build output, vite warnings, goreleaser progress, go module downloads

### Debug Mode

Maximum verbosity with bash xtrace (implies `--verbose`):

```bash
# Using flag
docker run --user "$(id -u):$(id -g)" --rm \
  --volume="${GOCACHE:-${HOME}/.cache/go-build}:/usr/share/go_cache" \
  --volume="${GOMODCACHE:-${HOME}/.cache/go-mod}:/usr/share/go/pkg/mod" \
  --volume="${PWD}:/mnt" \
  zrok-builder --debug arm64

# Using environment variable
docker run --user "$(id -u):$(id -g)" --rm \
  --env DEBUG=1 \
  --volume="${GOCACHE:-${HOME}/.cache/go-build}:/usr/share/go_cache" \
  --volume="${GOMODCACHE:-${HOME}/.cache/go-mod}:/usr/share/go/pkg/mod" \
  --volume="${PWD}:/mnt" \
  zrok-builder arm64
```

**Shows:** All verbose output plus bash xtrace (`set -x`) for script debugging

## Output

Built artifacts are placed in `./dist/`:

```text
dist/
├── artifacts.json              # GoReleaser metadata
├── config.yaml                 # GoReleaser config snapshot
├── metadata.json               # Build metadata
└── zrok-<variant>_linux_<arch>_<subarch>/
    └── zrok                    # Executable binary
```

## Notes

* **Go caches:** Two cache mounts significantly speed up builds:
  * `GOCACHE` (build cache): Reuses compiled packages
  * `GOMODCACHE` (module cache): Avoids re-downloading Go modules
* **Dirty builds:** Snapshot builds work with uncommitted changes (dirty working copy)
* **User permissions:** Running with `--user "${UID}"` ensures output files have correct ownership
* **Flag precedence:** Command line flags override environment variables
