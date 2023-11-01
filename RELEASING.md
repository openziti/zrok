
# Releasing zrok

## Manual Steps

1. Create a semver Git tag on main starting with a 'v' character.
1. Push the tag to GitHub.
1. Wait for automated steps to complete.
1. In GitHub Releases, edit the draft release as needed and finalize.

## Automated Steps

1. The Release workflow is triggered by creating the Git tag and
    1. uploads Linux packages to Artifactory and
    1. drafts a release in GitHub Releases.
1. The Publish Container Images workflow is triggered by the Releases API and
    1. pushes Docker images to Docker Hub.
