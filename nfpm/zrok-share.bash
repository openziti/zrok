#!/usr/bin/env bash
#
# this script shares the configured backend for a reserved share token
#

set -o errexit
set -o nounset
set -o pipefail

if ! command -v jq &>/dev/null; then
  echo "ERROR: jq is needed but not installed" >&2
  exit 1
fi

# set HOME to the first colon-sep dir in STATE_DIRECTORY inherited from systemd, e.g. /var/lib/zrok-share
export HOME="${STATE_DIRECTORY%:*}" 

if (( $# )); then
  if [[ -s "$1" ]]; then
    source "$1"
  else
    echo "ERROR: '$1' is empty or not readable" >&2
    exit 1
  fi
else
  # TODO: consider defining a default environment file
  # if [[ -s /opt/openziti/etc/zrok.env ]]; then
  #   source /opt/openziti/etc/zrok.env
  # else
  #   echo "ERROR: need /opt/openziti/etc/zrok.env or filename argument to read share configuration" >&2
  #   exit 1
  # fi
  echo "ERROR: need filename argument to read share configuration" >&2
  exit 1
fi

if [[ -s ~/.zrok/reserved.json ]]; then
  ZROK_RESERVED_TOKEN="$(jq '.token' ~/.zrok/reserved.json 2>/dev/null)"
  if [[ -z "${ZROK_RESERVED_TOKEN}" || "${ZROK_RESERVED_TOKEN}" == null ]]; then
    echo "ERROR: invalid reserved.json: '$(jq -c . ~/.zrok/reserved.json)'" >&2
    exit 1
  else
    echo "INFO: zrok backend is already reserved: ${ZROK_RESERVED_TOKEN}"
  fi
else
  ZROK_CMD="reserve public --json-output ${ZROK_VERBOSE:-}"
  case "${ZROK_BACKEND:-}" in
    http://*|https://*)
        ZROK_BACKEND_MODE="proxy"
      ;;
    *Caddyfile)
        ZROK_BACKEND_MODE="caddy"
        if ! [[ -r "${ZROK_BACKEND}" ]]; then
          echo "ERROR: ZROK_BACKEND='${ZROK_BACKEND}' is not a readable Caddyfile" >&2
          exit 1
        fi
      ;;
    /*)
        ZROK_BACKEND_MODE="web"
        if ! [[ -d "${ZROK_BACKEND}" && -r "${ZROK_BACKEND}" ]]; then
          echo "ERROR: ZROK_BACKEND='${ZROK_BACKEND}' is not a readable directory" >&2
          exit 1
        fi
      ;;
    *)
      echo "ERROR: ZROK_BACKEND='${ZROK_BACKEND}' is not a valid backend" >&2
      exit 1
      ;;
  esac
  ZROK_CMD+=" --backend-mode ${ZROK_BACKEND_MODE} ${ZROK_BACKEND}"
  if [[ -n "${ZROK_SHARE_OPTS:-}" ]]; then
    ZROK_CMD+=" ${ZROK_SHARE_OPTS}"
  fi
  if [[ -n "${ZROK_OAUTH_PROVIDER:-}" ]]; then
    ZROK_CMD+=" --oauth-provider ${ZROK_OAUTH_PROVIDER}"
  fi
  if [[ -n "${ZROK_OAUTH_EMAILS:-}" ]]; then
    for EMAIL in ${ZROK_OAUTH_EMAILS}; do
      if ! [[ ${EMAIL} =~ @ ]]; then
        echo "WARNING: '${EMAIL}' does not contain '@' so it may match more than one email domain!" >&2
      fi
      ZROK_CMD+=" --oauth-email-domains ${EMAIL}"
    done
  fi
  echo "INFO: running: zrok ${ZROK_CMD}"
  zrok ${ZROK_CMD} | jq -rc | tee ~/.zrok/reserved.json
fi

if ! [[ -s ~/.zrok/reserved.json ]]; then
  echo "ERROR: empty or missing $(realpath ~/.zrok)/reserved.json" >&2
  exit 1
else
  ZROK_PUBLIC_URLS=$(jq -cr '.frontend_endpoints' ~/.zrok/reserved.json 2>/dev/null)
  if [[ -z "${ZROK_PUBLIC_URLS}" || "${ZROK_PUBLIC_URLS}" == null ]]; then
    echo "ERROR: frontend endpoints not defined in $(realpath ~/.zrok)/reserved.json" >&2
    exit 1
  else 
    echo "INFO: zrok public URLs: ${ZROK_PUBLIC_URLS}"
  fi
  ZROK_RESERVED_TOKEN=$(jq -r '.token' ~/.zrok/reserved.json 2>/dev/null)
  if [[ -z "${ZROK_RESERVED_TOKEN}" || "${ZROK_RESERVED_TOKEN}" == null ]]; then
    echo "ERROR: zrok reservation token not defined in $(realpath ~/.zrok)/reserved.json" >&2
    exit 1
  fi
  ZROK_CMD="share reserved ${ZROK_RESERVED_TOKEN} --headless --override-endpoint ${ZROK_BACKEND}"
  ZROK_CMD+=" ${ZROK_VERBOSE:-} ${ZROK_INSECURE:-}"
  if [[ -n "${ZROK_SHARE_OPTS:-}" ]]; then
    ZROK_CMD+=" ${ZROK_SHARE_OPTS}"
  fi
  echo "INFO: running: zrok ${ZROK_CMD}"
  exec zrok ${ZROK_CMD}
fi
