
# Releasing zrok

## Manual Steps

> [!NOTE]
> Each trigger is outlined separately, but some may occur simultaneously, e.g., when a draft release is published as stable rather than first publishing it as a pre-release, or a pre-release is promoted to stable and marked as latest at the same time.

1. Push a tag to GitHub like `v*.*.*` to trigger **the pre-release workflow**. Wait for this workflow to complete before marking the release stable (`isPrerelease: false`).
    1. Linux packages are uploaded to Artifactory as pre-releases.
    1. Docker images are uploaded to Docker Hub as pre-releases.
    1. A release is drafted in GitHub.
1. Edit the draft and publish the release as a pre-release (`isPrerelease: true`).
    1. The one-time GitHub "published" event fires, and binaries are available in GitHub.
1. Edit the pre-release to mark it as a stable release (`isPrerelease: false`).
    1. The one-time GitHub "released" event fires.
    1. Linux packages are promoted to "stable" in Artifactory.
    1. Docker images are promoted to `:latest` in Docker Hub.
    1. Homebrew formulae are built and released.
    1. Python wheels are built and released to PyPi.
    1. Node.js packages are built and released to NPM.
1. Edit the stable release to mark it as latest.
    1. https://docs.zrok.io/docs/guides/install/ always serves the "latest" stable version via GitHub binary download URLs.

## Rolling Back Downstreams

The concepts, tools, and procedures for managing existing downstream artifacts in Artifactory and Docker Hub are identical for zrok and ziti. Here's the [RELEASING.md document for ziti](https://github.com/openziti/ziti/blob/main/RELEASING.md#rolling-back-downstreams).
