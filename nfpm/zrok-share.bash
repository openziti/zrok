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
if [[ -n "${STATE_DIRECTORY:-}" ]]; then
  export HOME="${STATE_DIRECTORY%:*}"
else
  echo "WARNING: STATE_DIRECTORY is undefined. Using HOME=${HOME}" >&2
fi
echo "DEBUG: zrok state directory is ${HOME}/.zrok"

: "${ZROK_SHARE_RESERVED:=true}"

echo "DEBUG: ZROK_SHARE_RESERVED=${ZROK_SHARE_RESERVED}"

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

# default mode is 'reserved-public', override modes are reserved-private, temp-public, temp-private.
: "${ZROK_FRONTEND_MODE:=reserved-public}"
if [[ "${ZROK_FRONTEND_MODE:-}" == temp-public ]]; then
  ZROK_CMD="share public --headless ${ZROK_VERBOSE:-}"
elif [[ "${ZROK_FRONTEND_MODE:-}" == temp-private ]]; then
  ZROK_CMD="share private --headless ${ZROK_VERBOSE:-}"
elif [[ -s ~/.zrok/reserved.json ]]; then
  ZROK_RESERVED_TOKEN="$(jq -r '.token' ~/.zrok/reserved.json 2>/dev/null)"
  if [[ -z "${ZROK_RESERVED_TOKEN}" || "${ZROK_RESERVED_TOKEN}" == null ]]; then
    echo "ERROR: invalid reserved.json: '$(jq -c . ~/.zrok/reserved.json)'" >&2
    exit 1
  else
    echo "INFO: zrok backend is already reserved: ${ZROK_RESERVED_TOKEN}"
    ZROK_CMD="${ZROK_RESERVED_TOKEN} ${ZROK_TARGET}"
    ZROK_CMD+=" ${ZROK_VERBOSE:-} ${ZROK_INSECURE:-}"
    if [[ "${ZROK_SHARE_RESERVED}" == true ]]; then
      share_reserved ${ZROK_CMD}
    else
      echo "INFO: finished reserving zrok backend, continuing without sharing"
      exit 0
    fi
  fi
elif [[ "${ZROK_FRONTEND_MODE:-}" == reserved-public ]]; then
  ZROK_CMD="reserve public --json-output ${ZROK_VERBOSE:-}"
elif [[ "${ZROK_FRONTEND_MODE:-}" == reserved-private ]]; then
  ZROK_CMD="reserve private --json-output ${ZROK_VERBOSE:-}"
else
  echo "ERROR: invalid value for ZROK_FRONTEND_MODE '${ZROK_FRONTEND_MODE}'" >&2
  exit 1
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
      echo "INFO: validated backend mode '${ZROK_BACKEND_MODE}' and target '${ZROK_TARGET}'"
    fi
    ;;
  caddy)
    if ! [[ "${ZROK_TARGET}" =~ ^/ ]]; then
      echo "ERROR: ZROK_TARGET='${ZROK_TARGET}' is not an absolute filesystem path" >&2
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
      echo "ERROR: ZROK_TARGET='${ZROK_TARGET}' is not an absolute filesystem path" >&2
      exit 1
    elif ! [[ -d "${ZROK_TARGET}" && -r "${ZROK_TARGET}" ]]; then
      echo "ERROR: ZROK_TARGET='${ZROK_TARGET}' is not a readable directory" >&2
      exit 1
    else
      echo "INFO: validated backend mode ${ZROK_BACKEND_MODE} and target ${ZROK_TARGET}"
    fi
    ;;
  tcpTunnel|udpTunnel|socks|vpn)
    if ! [[ "${ZROK_FRONTEND_MODE}" =~ -private$ ]]; then
      echo "ERROR: ZROK_BACKEND_MODE='${ZROK_BACKEND_MODE}' is a private share backend mode and cannot be used with ZROK_FRONTEND_MODE='${ZROK_FRONTEND_MODE}'" >&2
      exit 1
    else
      case "${ZROK_BACKEND_MODE}" in
        tcpTunnel|udpTunnel)
          echo "INFO: ${ZROK_BACKEND_MODE} backend mode has target '${ZROK_TARGET}'"
          ;;
        vpn)
          if [[ -n "${ZROK_TARGET}" ]]; then
            ZROK_SVC_FILE=/etc/systemd/system/zrok-share.service.d/override.conf
            if ! grep -qE '^AmbientCapabilities=CAP_NET_ADMIN' "${ZROK_SVC_FILE}"; then
              echo "ERROR: you must uncomment 'AmbientCapabilities=CAP_NET_ADMIN' in '${ZROK_SVC_FILE}'"\
                    "and run 'systemctl daemon-reload' to enable VPN mode" >&2
              exit 1
            fi
          fi
          ;;
        socks)
          if [[ -n "${ZROK_TARGET}" ]]; then
            echo "WARNING: ZROK_TARGET='${ZROK_TARGET}' is ignored with ZROK_BACKEND_MODE='${ZROK_BACKEND_MODE}'" >&2
            unset ZROK_TARGET
          fi
          ;;
      esac
    fi
    ;;
  *)
    echo "WARNING: ZROK_BACKEND_MODE='${ZROK_BACKEND_MODE}' is not a recognized mode for a zrok public share."\
          " ZROK_TARGET value will not validated before running." >&2
    ;;
esac

if [[ "${ZROK_FRONTEND_MODE:-}" =~ ^reserved- && -n "${ZROK_UNIQUE_NAME:-}" ]]; then
  ZROK_CMD+=" --unique-name ${ZROK_UNIQUE_NAME}"
elif [[ -n "${ZROK_UNIQUE_NAME:-}" ]]; then
  echo "WARNING: ZROK_UNIQUE_NAME='${ZROK_UNIQUE_NAME}' is ignored with ZROK_FRONTEND_MODE='${ZROK_FRONTEND_MODE}'" >&2
fi

if [[ "${ZROK_FRONTEND_MODE:-}" =~ -private$ && "${ZROK_PERMISSION_MODE:-}" == closed ]]; then
  ZROK_CMD+=" --closed"
  if [[ -n "${ZROK_ACCESS_GRANTS:-}" ]]; then
    for ACCESS_GRANT in ${ZROK_ACCESS_GRANTS}; do
      ZROK_CMD+=" --access-grant ${ACCESS_GRANT}"
    done
  else
    echo "WARNING: ZROK_PERMISSION_MODE='${ZROK_PERMISSION_MODE}' and no additional ZROK_ACCESS_GRANTS; will be granted access" >&2
  fi
elif [[ "${ZROK_FRONTEND_MODE:-}" =~ -private$ && -n "${ZROK_PERMISSION_MODE:-}" && "${ZROK_PERMISSION_MODE}" != open ]]; then
  echo "WARNING: ZROK_PERMISSION_MODE='${ZROK_PERMISSION_MODE}' is not a recognized value'" >&2
elif [[ "${ZROK_FRONTEND_MODE:-}" =~ -public$ && -n "${ZROK_PERMISSION_MODE:-}" ]]; then
  echo "WARNING: ZROK_PERMISSION_MODE='${ZROK_PERMISSION_MODE}' is ignored with ZROK_FRONTEND_MODE='${ZROK_FRONTEND_MODE}'" >&2
fi

ZROK_CMD+=" --backend-mode ${ZROK_BACKEND_MODE} ${ZROK_TARGET}"

if [[ -n "${ZROK_SHARE_OPTS:-}" ]]; then
  ZROK_CMD+=" ${ZROK_SHARE_OPTS}"
fi

if [[ -n "${ZROK_OAUTH_PROVIDER:-}" ]]; then
  ZROK_CMD+=" --oauth-provider ${ZROK_OAUTH_PROVIDER}"
  if [[ -n "${ZROK_OAUTH_EMAILS:-}" ]]; then
    for EMAIL in ${ZROK_OAUTH_EMAILS}; do
      ZROK_CMD+=" --oauth-email-address-patterns ${EMAIL}"
    done
  fi
elif [[ -n "${ZROK_BASIC_AUTH:-}" ]]; then
  ZROK_CMD+=" --basic-auth ${ZROK_BASIC_AUTH}"
fi

echo "INFO: running: zrok ${ZROK_CMD}"

if [[ "${ZROK_FRONTEND_MODE:-}" =~ ^temp- ]]; then
  # frontend mode starts with 'temp-', so is temporary.
  # share without reserving until exit.
  exec zrok ${ZROK_CMD}
else
  # reserve and continue
  zrok ${ZROK_CMD} > ~/.zrok/reserved.json
  # share the reserved backend target until exit
  if ! [[ -s ~/.zrok/reserved.json ]]; then
    echo "ERROR: empty or missing $(realpath ~/.zrok)/reserved.json" >&2
    exit 1
  elif ! jq . < ~/.zrok/reserved.json &>/dev/null; then
    echo "ERROR: invalid JSON in $(realpath ~/.zrok)/reserved.json" >&2
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
    if [[ "${ZROK_SHARE_RESERVED}" == true ]]; then
      share_reserved ${ZROK_CMD}
    else
      echo "INFO: finished reserving zrok backend, continuing without sharing"
      exit 0
    fi
  fi
fi
