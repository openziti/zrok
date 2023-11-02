#!/usr/bin/env bash
#
# this script uses a zrok enable token to enable a zrok environment in $HOME/.zrok
#

set -o errexit
set -o nounset
set -o pipefail

BASENAME=$(basename "$0")

if (( $# )); then
  case $1 in
    -h|*help)
      echo -e "Usage: ${BASENAME} [filename|-]\n"\
            "\tFILENAME\tfile containing environment variables to set\n"\
            "\t-\t\tread environment variable assignments from stdin\n"
      exit 0
      ;;
  esac
fi

# set HOME to the first colon-sep dir in STATE_DIRECTORY inherited from systemd, e.g. /var/lib/zrok-share
if [[ -n "${STATE_DIRECTORY:-}" ]]; then
  export HOME="${STATE_DIRECTORY%:*}"
else
  # assume we're enabling for a system service, not a user service, if run outside systemd
  export HOME="/var/lib/zrok-share"
fi

if ! [[ -d "${HOME}" && -w "${HOME}" ]]; then
  echo "ERROR: HOME='${HOME}' is not a writeable directory" >&2
  exit 1
fi

if [[ -s ~/.zrok/environment.json ]]; then
  echo "INFO: zrok environment is already enabled. Delete '$(realpath ~/.zrok/environment.json)' if you want to create a"\
        " new environment."
  exit 0
fi

if (( $# )); then
  if [[ "$1" == "-" ]]; then
    source <(cat)
  elif [[ -s "$1" ]]; then
    source "$1"
  else
    echo "ERROR: \$1="$1" is empty or not a readable file" >&2
    exit 1
  fi
else
  if [[ -z "${ZROK_ENABLE_TOKEN:-}" ]]; then
    read -sr -p "Enter zrok enable token: " ZROK_ENABLE_TOKEN
    echo
  fi
  if [[ -z "${ZROK_ENABLE_TOKEN}" ]]; then
    echo "ERROR: need ZROK_ENABLE_TOKEN or filename argument to read env configuration" >&2
    exit 1
  fi
  if [[ -z "${ZROK_ENVIRONMENT_NAME:-}" ]]; then
    DEFAULT_ZROK_ENVIRONMENT_NAME="zrok-share.service on $(hostname -s)"
    read -r -p "Enter zrok environment name [${DEFAULT_ZROK_ENVIRONMENT_NAME}]: " ZROK_ENVIRONMENT_NAME
    : "${ZROK_ENVIRONMENT_NAME:=${DEFAULT_ZROK_ENVIRONMENT_NAME}}"
  fi
  if [[ -z "${ZROK_API_ENDPOINT:-}" ]]; then
    read -r -p "Enter zrok API endpoint [https://api.zrok.io]: " ZROK_API_ENDPOINT
    if [[ -n "${ZROK_API_ENDPOINT}" ]]; then
      export ZROK_API_ENDPOINT
    fi
  fi
fi

if [[ -z "${ZROK_ENABLE_TOKEN}" ]]; then
  echo "ERROR: ZROK_ENABLE_TOKEN is not defined" >&2
  exit 1
else
  zrok config set apiEndpoint "${ZROK_API_ENDPOINT:-https://api.zrok.io}"
  echo "INFO: running: zrok enable ..."
  exec zrok enable --headless --description "${ZROK_ENVIRONMENT_NAME}" "${ZROK_ENABLE_TOKEN}"
fi
