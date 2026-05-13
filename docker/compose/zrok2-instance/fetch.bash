#!/usr/bin/env bash
#
# Download the zrok2 Docker Compose self-hosting project.
#
# Usage:
#   curl -sSfL https://get.openziti.io/zrok2-instance/fetch.bash | bash
#
# Creates a zrok2-instance/ directory with the essential files listed in
# README.md.  See the full guide at
# https://netfoundry.io/docs/zrok/self-hosting/deployment/docker

set -o errexit -o nounset -o pipefail

BASE="https://get.openziti.io/zrok2-instance"

# Essential files from README.md (user-facing compose project).
# CI test scripts (dangerous.*.test.bash, compose.canary.yml, etc.) are
# excluded — they are only needed for development.
# The zrok2-bootstrap.bash library is NOT downloaded — it is included in
# the openziti/zrok2 container image at /usr/local/bin/zrok2-bootstrap
# and sourced automatically by entrypoint-init.bash on first run.
FILES=(
    compose.yml
    compose.caddy.yml
    .env.example
    entrypoint-init.bash
)

DEST="zrok2-instance"

dir_is_empty() { [[ -z "$(ls -A "$1" 2>/dev/null)" ]]; }

if [[ "$(basename "${PWD}")" == "${DEST}" ]]; then
    # Already inside a directory named zrok2-instance
    if dir_is_empty "${PWD}"; then
        DEST="."
    else
        echo "Current directory is '${DEST}' but is not empty." >&2
        exit 1
    fi
elif [[ -d "${DEST}" ]]; then
    if ! dir_is_empty "${DEST}"; then
        echo "Directory '${DEST}' already exists and is not empty." >&2
        exit 1
    fi
    # exists and is empty — reuse it
else
    mkdir "${DEST}"
fi

echo "Downloading zrok2 Docker Compose project to ${DEST}/ ..."

for f in "${FILES[@]}"; do
    curl -sSfL "${BASE}/${f}" -o "${DEST}/${f}"
    echo "  ${f}"
done

if [[ "${DEST}" == "." ]]; then
    echo ""
    echo "Done. Next steps:"
    echo "  cp .env.example .env"
    echo "  # edit .env — set ZROK2_DNS_ZONE, ZROK2_ADMIN_TOKEN, and ZITI_PWD"
    echo "  docker compose up -d"
else
    echo ""
    echo "Done. Next steps:"
    echo "  cd ${DEST}"
    echo "  cp .env.example .env"
    echo "  # edit .env — set ZROK2_DNS_ZONE, ZROK2_ADMIN_TOKEN, and ZITI_PWD"
    echo "  docker compose up -d"
fi
