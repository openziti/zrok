#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
# set -o xtrace

getZitiPublicFrontend(){
    local RETURNED
    local -A FIELDS
    FIELDS[all]=0
    FIELDS[zid]=1
    FIELDS[name]=2
    FIELDS[type]=3
    FIELDS[attributes]=4
    FIELDS[policy]=5

    if (( $# )); then
        RETURNED="$1"
        shift
    else
        RETURNED="all"
    fi

    if (( $# )); then
        echo "WARN: ignoring unexpected parameters: $*" >&2
    fi

    if [[ -z "${FIELDS[$RETURNED]}" ]]; then
        echo "ERROR: invalid return field $RETURNED" >&2
        return 1
    fi

    ziti edge list identities 'name="public"' --csv \
        | awk -F, '$'${FIELDS[name]}'=="public" {print $'${FIELDS[$RETURNED]}';}'
}

getZrokPublicFrontend(){
    local RETURNED
    local -A FIELDS
    FIELDS[all]=0
    FIELDS[token]=1
    FIELDS[zid]=2
    FIELDS[name]=3
    FIELDS[template]=4
    FIELDS[created]=5
    FIELDS[updated]=6

    if (( $# )); then
        RETURNED="$1"
        shift
    else
        RETURNED="all"
    fi

    if (( $# )); then
        echo "WARN: ignoring unexpected parameters: $*" >&2
    fi

    if [[ -z "${FIELDS[$RETURNED]}" ]]; then
        echo "ERROR: invalid return field $RETURNED" >&2
        return 1
    fi

    # strip ANSI sequences and return the first position from the line with a name exactly matching "public"
    zrok admin list frontends | sed 's/\x1b\[[0-9;]*m//g' \
    | awk '$'${FIELDS[name]}'=="public" {print $'${FIELDS[$RETURNED]}'}'
}

ziti edge login "https://ziti.${ZROK_DNS_ZONE}:${ZITI_CTRL_ADVERTISED_PORT}" \
    --username admin \
    --password "${ZITI_PWD}" \
    --yes

if ! [[ -s ~/.zrok/identities/public.json ]]; then
    mkdir -p ~/.zrok/identities
    ziti edge create identity "public" --jwt-output-file /tmp/public.jwt
    ziti edge enroll --jwt /tmp/public.jwt --out ~/.zrok/identities/public.json
fi

# find Ziti ID of default "public" frontend
ZITI_PUBLIC_ID="$(getZitiPublicFrontend zid)"
until [[ -n "${ZITI_PUBLIC_ID}" ]]; do
    echo "DEBUG: waiting for default frontend "public" Ziti identity to be created"
    sleep 3
    ZITI_PUBLIC_ID="$(getZitiPublicFrontend zid)"
done
echo "DEBUG: 'public' ZITI_PUBLIC_ID=$ZITI_PUBLIC_ID"

until curl -sSf "${ZROK_API_ENDPOINT}" &>/dev/null; do
    echo "DEBUG: waiting for zrok controller API version endpoint to respond"
    sleep 3
done

# if default "public" frontend already exists
ZROK_PUBLIC_TOKEN=$(getZrokPublicFrontend token)
if [[ -n "${ZROK_PUBLIC_TOKEN}" ]]; then
    
    # ensure the Ziti ID of the public frontend's identity is the same in Ziti and zrok
    ZROK_PUBLIC_ZID=$(getZrokPublicFrontend zid)
    if [[ "${ZITI_PUBLIC_ID}" != "${ZROK_PUBLIC_ZID}" ]]; then
        echo "ERROR: existing Ziti Identity named 'public' with id '$ZITI_PUBLIC_ID' is from a previous zrok"\
        "instance life cycle. Delete it then re-run zrok." >&2
        exit 1
    fi

    echo "INFO: updating frontend"
    zrok admin update frontend "${ZROK_PUBLIC_TOKEN}" \
        --url-template "${ZROK_FRONTEND_SCHEME}://{token}.${ZROK_DNS_ZONE}:${ZROK_FRONTEND_PORT}"
else
    echo "INFO: creating frontend"
    zrok admin create frontend "${ZITI_PUBLIC_ID}" public \
        "${ZROK_FRONTEND_SCHEME}://{token}.${ZROK_DNS_ZONE}:${ZROK_FRONTEND_PORT}"
fi

exec "${@}"
