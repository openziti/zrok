# Build

## Quick Start: Containerized Build (Recommended)

The easiest way to build `zrok` is using the provided `build.bash` wrapper script, which uses Docker to create a reproducible build environment:

Run this:

```bash
./build.bash
```

**Expected output (quiet mode):**

```text
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✓ Build completed successfully (goreleaser snapshot)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

GoReleaser output directory: ./dist/

Built binaries:
  • ./dist/zrok-amd64_linux_amd64_v1/zrok
    (config: .goreleaser-linux-amd64.yml)

Embedded UIs:
  • ./ui/dist           → /api/v1/static (main UI)
  • ./agent/agentUi/dist → /agent (agent UI)

Note: GoReleaser also generates archives and metadata in ./dist/
```

**What it does:**

* Automatically detects your host architecture (amd64, arm64, or armhf)
* Builds the `zrok-builder` container image if needed (one-time setup)
* Rebuilds the image automatically if the Dockerfile or build script changes
* Builds zrok with embedded UIs using goreleaser snapshots
* Outputs only errors and a final build report (use `--verbose` for full output)
* Places the binary in `./dist/<binary>_linux_<arch>_<variant>/zrok`

**Requirements:**

* Docker installed and running
* No other dependencies needed (Go, Node.js, etc. are in the container)

**Advantages:**

* Reproducible builds with consistent toolchain versions
* No need to install Go, Node.js, or build tools on your host
* Resembles the CI/CD build environment
* Clean, quiet output by default

For cross-compilation, multiple architectures, or debug mode, see [Cross-build zrok with Docker](#cross-build-zrok-with-docker) below.

## zrok

At this time, building `zrok` is pretty straightforward. You will require `node` v18+ to be installed in order to complete the build as well as `go`. Because `zrok` uses CGO, you will also need to have a working C compiler toolchain. [TDM-GCC](https://jmeubank.github.io/tdm-gcc/download/) works great on Windows (just make sure it's in your PATH).

To build, follow these steps:

* clone the repository
* change to the existing `ui` folder
* run `npm install`
* run `npm run build` (this process takes a while the first time and only needs to be run if the ui changes)
* change back to the checkout root
* change to the existing `agent/agentUi/` folder
* run `npm install`
* run `npm run build` (this process takes a while the first time and only needs to be run if the ui changes)
* change back to the checkout root
* make sure the dist directory exists: `mkdir -p dist`
* build the go project normally: `go build -o dist ./...`

## Cross-build zrok with Docker

For advanced use cases including cross-compilation for multiple architectures, verbose/debug output, or direct control over the build process, see the [cross-build documentation](./docker/images/cross-build/README.md).

**Supported architectures:** amd64, arm64, armhf, armel

**Quick example:**

```bash
# Build for arm64
docker run --user "${UID}" --rm \
  --volume="${GOCACHE:-${HOME}/.cache/go-build}:/usr/share/go_cache" \
  --volume="${GOMODCACHE:-${HOME}/.cache/go-mod}:/usr/share/go/pkg/mod" \
  --volume="${PWD}:/mnt" \
  zrok-builder arm64

# Build for amd64 (in a separate run)
docker run --user "${UID}" --rm \
  --volume="${GOCACHE:-${HOME}/.cache/go-build}:/usr/share/go_cache" \
  --volume="${GOMODCACHE:-${HOME}/.cache/go-mod}:/usr/share/go/pkg/mod" \
  --volume="${PWD}:/mnt" \
  zrok-builder amd64
```

**Note:** Due to goreleaser's dist directory handling, each architecture must be built in a separate `docker run` command. The `./dist/` directory will be cleaned at the start of each build.

## Documentation/Website

The doc website is based on [Docusaurus](https://docusaurus.io/) which in turn will require `npm` to be installed. `yarn`
is another tool which is used to start the Docusaurus dev site.

To build the doc:

* cd to `website`
* run `yarn install` (usually only needed once)
* run `yarn start` to start the development server (make sure port 3000 is open or change the port)

The development server infrequently behaves differently than the 'production' build. If you must use the 'production'
build it is slower, but you can accomplish that with `yarn build`.
