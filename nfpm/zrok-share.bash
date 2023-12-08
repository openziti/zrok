#!/usr/bin/env bash
#
# this script shares the configured backend for a reserved share token
#

set -o errexit
set -o nounset
set -o pipefail

share_reserved(){
    local token="$1"
    local target="$2"
    shift 2
    local opts="${*:-}"
    local zrok_cmd="share reserved ${token} --headless ${opts} --override-endpoint ${target}"
    echo "INFO: running: zrok ${zrok_cmd}"
    exec zrok ${zrok_cmd}
}

if ! command -v jq &>/dev/null; then
  echo "ERROR: jq is needed but not installed" >&2
  exit 1
fi

# set HOME to the first colon-sep dir in STATE_DIRECTORY inherited from systemd (/var/lib/zrok-share) or docker (/mnt)
export HOME="${STATE_DIRECTORY%:*}"

if (( $# )); then
  if [[ -s "$1" ]]; then
    echo "INFO: reading share configuration from $1"
    source "$1"
    shift
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
  echo "INFO: reading share configuration from environment variables"
fi

[[ -n "${ZROK_TARGET:-}" ]] || {
  echo "ERROR: ZROK_TARGET is not defined." >&2
  exit 1
}

# default mode is reserved (public), override mode is temp-public, i.e., "share public" without a reserved subdomain
if [[ "${ZROK_FRONTEND_MODE:-}" == temp-public ]]; then
  ZROK_CMD="share public --headless ${ZROK_VERBOSE:-}"
elif [[ -s ~/.zrok/reserved.json ]]; then
  ZROK_RESERVED_TOKEN="$(jq -r '.token' ~/.zrok/reserved.json 2>/dev/null)"
  if [[ -z "${ZROK_RESERVED_TOKEN}" || "${ZROK_RESERVED_TOKEN}" == null ]]; then
    echo "ERROR: invalid reserved.json: '$(jq -c . ~/.zrok/reserved.json)'" >&2
    exit 1
  else
    echo "INFO: zrok backend is already reserved: ${ZROK_RESERVED_TOKEN}"
    ZITI_CMD="${ZROK_RESERVED_TOKEN} ${ZROK_TARGET}"
    ZITI_CMD+=" ${ZROK_VERBOSE:-} ${ZROK_INSECURE:-}"
    share_reserved ${ZITI_CMD}
  fi
else
  ZROK_CMD="reserve public --json-output ${ZROK_VERBOSE:-}"
fi

[[ -n "${ZROK_BACKEND_MODE:-}" ]] || {
  echo "WARNING: ZROK_BACKEND_MODE was not defined, assuming mode 'proxy'." >&2
  ZROK_BACKEND_MODE="proxy"
}

case "${ZROK_BACKEND_MODE}" in
  proxy)
    if ! [[ "${ZROK_TARGET}" =~ ^https?:// ]]; then
      echo "ERROR: ZROK_TARGET='${ZROK_TARGET}' is not an HTTP URL" >&2
      exit 1
    else
      echo "INFO: validated backend mode ${ZROK_BACKEND_MODE} and target ${ZROK_TARGET}"
    fi
    ;;
  caddy)
    if ! [[ "${ZROK_TARGET}" =~ ^/ ]]; then
      echo "ERROR: ZROK_TARGET='${ZROK_TARGET}' is not an absolute filesystem path." >&2
      exit 1
    elif ! [[ -f "${ZROK_TARGET}" && -r "${ZROK_TARGET}" ]]; then
      echo "ERROR: ZROK_TARGET='${ZROK_TARGET}' is not a readable regular file" >&2
      exit 1
    else
      echo "INFO: validated backend mode ${ZROK_BACKEND_MODE} and target ${ZROK_TARGET}"
    fi
    ;;
  web|drive)
    if ! [[ "${ZROK_TARGET}" =~ ^/ ]]; then
      echo "ERROR: ZROK_TARGET='${ZROK_TARGET}' is not an absolute filesystem path." >&2
      exit 1
    elif ! [[ -d "${ZROK_TARGET}" && -r "${ZROK_TARGET}" ]]; then
      echo "ERROR: ZROK_TARGET='${ZROK_TARGET}' is not a readable directory" >&2
      exit 1
    else
      echo "INFO: validated backend mode ${ZROK_BACKEND_MODE} and target ${ZROK_TARGET}"
    fi
    ;;
  *)
    echo "WARNING: ZROK_BACKEND_MODE='${ZROK_BACKEND_MODE}' is not a recognized mode for a zrok public share."\
          " ZROK_TARGET value will not validated before running." >&2
    ;;
esac

[[ -n "${ZROK_SUBDOMAIN:-}" ]] && {
  ZROK_CMD+=" --unique-name ${ZROK_SUBDOMAIN}"
}

ZROK_CMD+=" --backend-mode ${ZROK_BACKEND_MODE} ${ZROK_TARGET}"

if [[ -n "${ZROK_SHARE_OPTS:-}" ]]; then
  ZROK_CMD+=" ${ZROK_SHARE_OPTS}"
fi

if [[ -n "${ZROK_OAUTH_PROVIDER:-}" ]]; then
  ZROK_CMD+=" --oauth-provider ${ZROK_OAUTH_PROVIDER}"
  if [[ -n "${ZROK_OAUTH_EMAILS:-}" ]]; then
    for EMAIL in ${ZROK_OAUTH_EMAILS}; do
      ZROK_CMD+=" --oauth-email-domains ${EMAIL}"
    done
  fi
elif [[ -n "${ZROK_BASIC_AUTH:-}" ]]; then
  ZROK_CMD+=" --basic-auth ${ZROK_BASIC_AUTH}"
fi

echo "INFO: running: zrok ${ZROK_CMD}"

if [[ "${ZROK_FRONTEND_MODE:-}" == temp-public ]]; then
  # share until exit
  exec zrok ${ZROK_CMD}
else
  # reserve and continue
  zrok ${ZROK_CMD} | jq -rc | tee ~/.zrok/reserved.json
  # share the reserved backend target until exit
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
    ZROK_CMD="${ZROK_RESERVED_TOKEN} ${ZROK_TARGET}"
    ZROK_CMD+=" ${ZROK_VERBOSE:-} ${ZROK_INSECURE:-}"
    share_reserved ${ZROK_CMD}
  fi
fi
