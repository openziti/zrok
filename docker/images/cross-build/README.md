# Building zrok2 with Docker

This directory covers all Docker-based approaches to building zrok2. For
builds without Docker, see [BUILD.md](../../../BUILD.md).

## Quick start: `bin/build.bash`

`bin/build.bash` wraps the `zrok-builder` container. It auto-detects your host
architecture, builds the container image on first run, and requires only Docker:

```bash
bin/build.bash            # quiet mode
bin/build.bash --verbose  # show full build output
```

**Requirements:** Docker installed and running. No Go, Node.js, or other
toolchain needed on the host.

**Output:** `dist/zrok2_linux_<arch>_<variant>/zrok2`

## Build the `openziti/zrok2` container image

The multi-stage `Dockerfile.build` compiles the UI assets (Vite) and Go binary
(with CGO for SQLite3) and packages the result into the same
`openziti/ziti-cli` base image used by published releases.

```bash
# From the project root
docker build -f docker/compose/zrok2-instance/Dockerfile.build \
  -t openziti/zrok2:local .
```

The resulting image contains:

- `/usr/local/bin/zrok2` — binary with embedded UIs
- `/usr/local/bin/zrok2-enable` — enable helper script
- `/usr/local/bin/zrok2-bootstrap` — bootstrap script

Use the image with Docker Compose via the build overlay:

```bash
cd docker/compose/zrok2-instance
cp .env.example .env
# edit .env with required values
COMPOSE_FILE=compose.yml:compose.build.yml docker compose up -d --build --wait
```

## Advanced: `zrok-builder` container directly

The `zrok-builder` image supports cross-compilation for multiple architectures.

### Build the container image

`bin/build.bash` handles this automatically, but you can build manually:

```bash
# From the project root
docker buildx build -t zrok-builder ./docker/images/cross-build --load
```

The image is automatically rebuilt when `Dockerfile` or `linux-build.sh` changes.

**Supported architectures:** `amd64`, `arm64`, `armhf` (arm/v7 hard-float), `armel` (arm/v7 soft-float)

### Build for a specific architecture

```bash
docker run --user "$(id -u):$(id -g)" --rm \
  --volume="${GOCACHE:-${HOME}/.cache/go-build}:/usr/share/go_cache" \
  --volume="${GOMODCACHE:-${HOME}/.cache/go-mod}:/usr/share/go/pkg/mod" \
  --volume="${PWD}:/mnt" \
  zrok-builder arm64
```

Each architecture must be built in a separate `docker run` — goreleaser cleans
`./dist/` at the start of each build.

### Verbose output

```bash
docker run --user "$(id -u):$(id -g)" --rm \
  --volume="${GOCACHE:-${HOME}/.cache/go-build}:/usr/share/go_cache" \
  --volume="${GOMODCACHE:-${HOME}/.cache/go-mod}:/usr/share/go/pkg/mod" \
  --volume="${PWD}:/mnt" \
  zrok-builder --verbose arm64
```

Or set `VERBOSE=1` in the environment.

### Debug mode

Maximum verbosity with bash xtrace (implies `--verbose`):

```bash
docker run --user "$(id -u):$(id -g)" --rm \
  --volume="${GOCACHE:-${HOME}/.cache/go-build}:/usr/share/go_cache" \
  --volume="${GOMODCACHE:-${HOME}/.cache/go-mod}:/usr/share/go/pkg/mod" \
  --volume="${PWD}:/mnt" \
  zrok-builder --debug arm64
```

Or set `DEBUG=1` in the environment.

## Build output

```text
dist/
├── artifacts.json
├── config.yaml
├── metadata.json
└── zrok2_linux_<arch>_<variant>/
    └── zrok2
```

## Notes

- **Go caches:** Mount `GOCACHE` and `GOMODCACHE` to avoid re-downloading
  modules and recompiling unchanged packages across builds.
- **Dirty builds:** Snapshot builds work with uncommitted changes.
- **User permissions:** `--user "$(id -u):$(id -g)"` ensures output files
  have correct ownership.
