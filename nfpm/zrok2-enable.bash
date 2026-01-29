#!/usr/bin/env bash
#
# this script uses a zrok enable token to enable a zrok environment in $HOME/.zrok2
#

set -o errexit
set -o nounset
set -o pipefail

BASENAME=$(basename "$0")
DEFAULT_ZROK2_ENVIRONMENT_NAME="zrok2-share service on $(hostname -s 2>/dev/null || echo localhost)"

if (( $# )); then
  case $1 in
    -h|*help)
      echo -e \
        "Usage: ${BASENAME} FILENAME\n"\
        "\tFILENAME\tfile containing environment variables to set"
      exit 0
      ;;
  esac
fi

# set HOME to the first colon-sep dir in STATE_DIRECTORY inherited from systemd, e.g. /var/lib/zrok2-share
if [[ -n "${STATE_DIRECTORY:-}" ]]; then
  export HOME="${STATE_DIRECTORY%:*}"
else
  echo "WARNING: STATE_DIRECTORY is undefined. Using HOME=${HOME}" >&2
fi
echo "DEBUG: zrok state directory is ${HOME}/.zrok2"

if [[ -s ~/.zrok2/environment.json ]]; then
  echo "INFO: zrok environment is already enabled. Delete '$(realpath ~/.zrok2/environment.json)' if you want to create a"\
    "new environment."
  exit 0
fi

if (( $# )); then
  if [[ -s "$1" ]]; then
    echo "INFO: reading enable parameters from $1"
    source "$1"
  else
    echo "ERROR: \$1="$1" is empty or not a readable file" >&2
    exit 1
  fi
else
  echo "INFO: reading enable parameters from environment variables"
fi

if [[ -z "${ZROK2_ENABLE_TOKEN}" ]]; then
  echo "ERROR: ZROK2_ENABLE_TOKEN is not defined" >&2
  exit 1
else
  zrok2 config set apiEndpoint "${ZROK2_API_ENDPOINT:-https://api-v2.zrok.io}"
  echo "INFO: running: zrok2 enable ..."
  exec zrok2 enable --headless --description "${ZROK2_ENVIRONMENT_NAME:-${DEFAULT_ZROK2_ENVIRONMENT_NAME}}" "${ZROK2_ENABLE_TOKEN}"
fi
