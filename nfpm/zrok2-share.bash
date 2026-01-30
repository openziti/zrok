#!/usr/bin/env bash
#
# this script shares the configured backend for a reserved share token
#

set -o errexit
set -o nounset
set -o pipefail

exec_with_common_opts(){
    local zrok_cmd="$* --headless ${ZROK2_VERBOSE:-} ${ZROK2_INSECURE:-}"
    echo "INFO: running: zrok2 ${zrok_cmd}"
    exec zrok2 ${zrok_cmd}
}

exec_share_reserved(){
    local token="$1"
    local target="$2"
    shift 2
    local opts="${*:-}"
    local zrok_cmd="share reserved ${token} ${opts} --override-endpoint ${target}"
    exec_with_common_opts ${zrok_cmd}
}

if ! command -v jq &>/dev/null; then
  echo "ERROR: jq is needed but not installed" >&2
  exit 1
fi

# set HOME to the first colon-sep dir in STATE_DIRECTORY inherited from systemd (/var/lib/zrok2-share) or docker (/mnt)
if [[ -n "${STATE_DIRECTORY:-}" ]]; then
  export HOME="${STATE_DIRECTORY%:*}"
else
  echo "WARNING: STATE_DIRECTORY is undefined. Using HOME=${HOME}" >&2
fi
echo "DEBUG: zrok state directory is ${HOME}/.zrok2"

: "${ZROK2_SHARE_RESERVED:=true}"
echo "DEBUG: ZROK2_SHARE_RESERVED=${ZROK2_SHARE_RESERVED}"

while (( $# )); do
  if [[ "${1:0:1}" == @ ]]; then
    ZROK2_INSTANCE="${1:1}"
    shift
  elif [[ -s "$1" ]]; then
    echo "INFO: reading share configuration from $1"
    source "$1"
    shift
  fi
done

ZROK2_RESERVATION_FILE="${HOME}/.zrok2/reserved${ZROK2_INSTANCE:+@${ZROK2_INSTANCE}}.json"

[[ -n "${ZROK2_TARGET:-}" ]] || {
  echo "ERROR: ZROK2_TARGET is not defined." >&2
  exit 1
}

# default mode is 'reserved-public', override modes are reserved-private, temp-public, temp-private.
: "${ZROK2_FRONTEND_MODE:=reserved-public}"
if [[ "${ZROK2_FRONTEND_MODE:-}" == temp-public ]]; then
  ZROK2_CMD="share public"
elif [[ "${ZROK2_FRONTEND_MODE:-}" == temp-private ]]; then
  ZROK2_CMD="share private"
elif [[ -s "${ZROK2_RESERVATION_FILE}" ]]; then
  ZROK2_RESERVATION_TOKEN="$(jq -r '.token' "${ZROK2_RESERVATION_FILE}" 2>/dev/null)"
  if [[ -z "${ZROK2_RESERVATION_TOKEN}" || "${ZROK2_RESERVATION_TOKEN}" == null ]]; then
    echo "ERROR: invalid reservation file: '$(jq -c . "${ZROK2_RESERVATION_FILE}")'" >&2
    exit 1
  else
    echo "INFO: zrok backend is already reserved: ${ZROK2_RESERVATION_TOKEN}"
    ZROK2_CMD="${ZROK2_RESERVATION_TOKEN} ${ZROK2_TARGET}"
    if [[ "${ZROK2_SHARE_RESERVED}" == true ]]; then
      exec_share_reserved ${ZROK2_CMD}
    else
      echo "INFO: finished reserving zrok backend, continuing without sharing"
      exit 0
    fi
  fi
elif [[ "${ZROK2_FRONTEND_MODE:-}" == reserved-public ]]; then
  ZROK2_CMD="reserve public --json-output ${ZROK2_VERBOSE:-}"
elif [[ "${ZROK2_FRONTEND_MODE:-}" == reserved-private ]]; then
  ZROK2_CMD="reserve private --json-output ${ZROK2_VERBOSE:-}"
else
  echo "ERROR: invalid value for ZROK2_FRONTEND_MODE '${ZROK2_FRONTEND_MODE}'" >&2
  exit 1
fi

[[ -n "${ZROK2_BACKEND_MODE:-}" ]] || {
  echo "WARNING: ZROK2_BACKEND_MODE was not defined, assuming mode 'proxy'." >&2
  ZROK2_BACKEND_MODE="proxy"
}

case "${ZROK2_BACKEND_MODE}" in
  proxy)
    if ! [[ "${ZROK2_TARGET}" =~ ^https?:// ]]; then
      echo "ERROR: ZROK2_TARGET='${ZROK2_TARGET}' is not an HTTP URL" >&2
      exit 1
    else
      echo "INFO: validated backend mode '${ZROK2_BACKEND_MODE}' and target '${ZROK2_TARGET}'"
    fi
    ;;
  caddy)
    if ! [[ "${ZROK2_TARGET}" =~ ^/ ]]; then
      echo "ERROR: ZROK2_TARGET='${ZROK2_TARGET}' is not an absolute filesystem path" >&2
      exit 1
    elif ! [[ -f "${ZROK2_TARGET}" && -r "${ZROK2_TARGET}" ]]; then
      echo "ERROR: ZROK2_TARGET='${ZROK2_TARGET}' is not a readable regular file" >&2
      exit 1
    else
      echo "INFO: validated backend mode ${ZROK2_BACKEND_MODE} and target ${ZROK2_TARGET}"
    fi
    ;;
  web|drive)
    if ! [[ "${ZROK2_TARGET}" =~ ^/ ]]; then
      echo "ERROR: ZROK2_TARGET='${ZROK2_TARGET}' is not an absolute filesystem path" >&2
      exit 1
    elif ! [[ -d "${ZROK2_TARGET}" && -r "${ZROK2_TARGET}" ]]; then
      echo "ERROR: ZROK2_TARGET='${ZROK2_TARGET}' is not a readable directory" >&2
      exit 1
    else
      echo "INFO: validated backend mode ${ZROK2_BACKEND_MODE} and target ${ZROK2_TARGET}"
    fi
    ;;
  tcpTunnel|udpTunnel|socks)
    if ! [[ "${ZROK2_FRONTEND_MODE}" =~ -private$ ]]; then
      echo "ERROR: ZROK2_BACKEND_MODE='${ZROK2_BACKEND_MODE}' is a private share backend mode and cannot be used with ZROK2_FRONTEND_MODE='${ZROK2_FRONTEND_MODE}'" >&2
      exit 1
    else
      case "${ZROK2_BACKEND_MODE}" in
        tcpTunnel|udpTunnel)
          echo "INFO: ${ZROK2_BACKEND_MODE} backend mode has target '${ZROK2_TARGET}'"
          ;;
        socks)
          if [[ -n "${ZROK2_TARGET}" ]]; then
            echo "WARNING: ZROK2_TARGET='${ZROK2_TARGET}' is ignored with ZROK2_BACKEND_MODE='${ZROK2_BACKEND_MODE}'" >&2
            unset ZROK2_TARGET
          fi
          ;;
      esac
    fi
    ;;
  *)
    echo "WARNING: ZROK2_BACKEND_MODE='${ZROK2_BACKEND_MODE}' is not a recognized mode for a zrok public share."\
          " ZROK2_TARGET value will not validated before running." >&2
    ;;
esac

if [[ "${ZROK2_FRONTEND_MODE:-}" =~ ^reserved- && -n "${ZROK2_UNIQUE_NAME:-}" ]]; then
  ZROK2_CMD+=" --unique-name ${ZROK2_UNIQUE_NAME}"
elif [[ -n "${ZROK2_UNIQUE_NAME:-}" ]]; then
  echo "WARNING: ZROK2_UNIQUE_NAME='${ZROK2_UNIQUE_NAME}' is ignored with ZROK2_FRONTEND_MODE='${ZROK2_FRONTEND_MODE}'" >&2
fi

if [[ "${ZROK2_FRONTEND_MODE:-}" =~ -private$ && "${ZROK2_PERMISSION_MODE:-}" == closed ]]; then
  ZROK2_CMD+=" --closed"
  if [[ -n "${ZROK2_ACCESS_GRANTS:-}" ]]; then
    for ACCESS_GRANT in ${ZROK2_ACCESS_GRANTS}; do
      ZROK2_CMD+=" --access-grant ${ACCESS_GRANT}"
    done
  else
    echo "WARNING: ZROK2_PERMISSION_MODE='${ZROK2_PERMISSION_MODE}' and no additional ZROK2_ACCESS_GRANTS; will be granted access" >&2
  fi
elif [[ "${ZROK2_FRONTEND_MODE:-}" =~ -private$ && -n "${ZROK2_PERMISSION_MODE:-}" && "${ZROK2_PERMISSION_MODE}" != open ]]; then
  echo "WARNING: ZROK2_PERMISSION_MODE='${ZROK2_PERMISSION_MODE}' is not a recognized value'" >&2
elif [[ "${ZROK2_FRONTEND_MODE:-}" =~ -public$ && -n "${ZROK2_PERMISSION_MODE:-}" ]]; then
  echo "WARNING: ZROK2_PERMISSION_MODE='${ZROK2_PERMISSION_MODE}' is ignored with ZROK2_FRONTEND_MODE='${ZROK2_FRONTEND_MODE}'" >&2
fi

ZROK2_CMD+=" --backend-mode ${ZROK2_BACKEND_MODE} ${ZROK2_TARGET}"

if [[ -n "${ZROK2_SHARE_OPTS:-}" ]]; then
  ZROK2_CMD+=" ${ZROK2_SHARE_OPTS}"
fi

if [[ -n "${ZROK2_OAUTH_PROVIDER:-}" ]]; then
  ZROK2_CMD+=" --oauth-provider ${ZROK2_OAUTH_PROVIDER}"
  if [[ -n "${ZROK2_OAUTH_EMAILS:-}" ]]; then
    for EMAIL in ${ZROK2_OAUTH_EMAILS}; do
      ZROK2_CMD+=" --oauth-email-address-pattern ${EMAIL}"
    done
  fi
elif [[ -n "${ZROK2_BASIC_AUTH:-}" ]]; then
  ZROK2_CMD+=" --basic-auth ${ZROK2_BASIC_AUTH}"
fi

if [[ "${ZROK2_FRONTEND_MODE:-}" =~ ^temp- ]]; then
  # frontend mode starts with 'temp-', so is temporary.
  # share without reserving until exit.
  exec_with_common_opts ${ZROK2_CMD}
else
  # reserve and continue
  zrok2 ${ZROK2_CMD} > "${ZROK2_RESERVATION_FILE}"
  # share the reserved backend target until exit
  if ! [[ -s "${ZROK2_RESERVATION_FILE}" ]]; then
    echo "ERROR: empty or missing $(realpath "${ZROK2_RESERVATION_FILE}")" >&2
    exit 1
  elif ! jq . < "${ZROK2_RESERVATION_FILE}" &>/dev/null; then
    echo "ERROR: invalid JSON in $(realpath "${ZROK2_RESERVATION_FILE}")" >&2
    exit 1
  else
    if [[ "${ZROK2_FRONTEND_MODE:-}" == reserved-public ]]; then
      ZROK2_PUBLIC_URLS=$(jq -cr '.frontend_endpoints' "${ZROK2_RESERVATION_FILE}" 2>/dev/null)
      if [[ -z "${ZROK2_PUBLIC_URLS}" || "${ZROK2_PUBLIC_URLS}" == null ]]; then
        echo "ERROR: frontend endpoints not defined in $(realpath "${ZROK2_RESERVATION_FILE}")" >&2
        exit 1
      else
        echo "INFO: zrok public URLs: ${ZROK2_PUBLIC_URLS}"
      fi
    fi
    ZROK2_RESERVATION_TOKEN=$(jq -r '.token' "${ZROK2_RESERVATION_FILE}" 2>/dev/null)
    if [[ -z "${ZROK2_RESERVATION_TOKEN}" || "${ZROK2_RESERVATION_TOKEN}" == null ]]; then
      echo "ERROR: zrok reservation token not defined in $(realpath "${ZROK2_RESERVATION_FILE}")" >&2
      exit 1
    fi
    ZROK2_CMD="${ZROK2_RESERVATION_TOKEN} ${ZROK2_TARGET}"
    if [[ "${ZROK2_SHARE_RESERVED}" == true ]]; then
      exec_share_reserved ${ZROK2_CMD}
    else
      echo "INFO: finished reserving zrok backend, continuing without sharing"
      exit 0
    fi
  fi
fi
