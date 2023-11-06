#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

checkSum() {
    for CMD in sha256sum md5sum; do
        if command -v $CMD &>/dev/null; then
            local SUMCMD=$CMD
            break
        fi
    done
    if [ -z "$SUMCMD" ]; then
        echo "ERROR: No checksum command found. Tried 'sha256sum', 'md5sum'." >&2
        exit 1
    fi
    $SUMCMD | awk '{print $1}'
}

installRedHat(){

    for CMD in dnf yum; do
        if command -v $CMD &>/dev/null; then
            local PACKAGER=$CMD
            break
        fi
    done
    if [ -z "$PACKAGER" ]; then
        echo "ERROR: No package manager found. Tried 'dnf', 'yum'." >&2
        exit 1
    fi

    local REPOSRC="[OpenZitiRelease]
name=OpenZiti Release
baseurl=https://packages.openziti.org/${ZITIPAX_RPM:-zitipax-openziti-rpm-stable}/redhat/\$basearch
enabled=1
gpgcheck=0
gpgkey=https://packages.openziti.org/${ZITIPAX_RPM:-zitipax-openziti-rpm-stable}/redhat/\$basearch/repodata/repomd.xml.key
repo_gpgcheck=1"

    local REPOFILE="/etc/yum.repos.d/openziti-release.repo"
    if [ -s $REPOFILE ]; then
        local EXISTINGSUM
        local REPOSUM
        EXISTINGSUM=$(checkSum < $REPOFILE)
        REPOSUM=$(checkSum <<< "$REPOSRC")
        if [ "$EXISTINGSUM" != "$REPOSUM" ]; then
            mv -v $REPOFILE{,".$(date -Iseconds)"}
            echo "$REPOSRC" > $REPOFILE

        fi
    else
        echo "$REPOSRC" >| $REPOFILE

    fi

    $PACKAGER update --assumeyes
    $PACKAGER install --assumeyes "$@"
    for PKG in "$@"; do
        $PACKAGER info "$PKG"
    done
}

installDebian(){

    for CMD in gpg gpg2; do
        if command -v $CMD &>/dev/null; then
            local GNUPGCMD=$CMD
            break
        fi
    done
    if [ -z "$GNUPGCMD" ]; then
        echo "ERROR: No GnuPG CLI found. Tried 'gpg', gpg2." >&2
        exit 1
    fi
    for CMD in wget curl; do
        if command -v $CMD &>/dev/null; then
            local GETTER=$CMD
            break
        fi
    done
    if [ -z "$GETTER" ]; then
        echo "ERROR: No http client found. Tried 'wget', 'curl'." >&2
        exit 1
    else
        case $GETTER in
            wget)
                GETTERCMD="wget -qO-"
                ;;
            curl)
                GETTERCMD="curl -fsSL"
                ;;
        esac
    fi

    # always update the pubkey
    $GETTERCMD https://get.openziti.io/tun/package-repos.gpg \
    | $GNUPGCMD --batch --yes --dearmor --output /usr/share/keyrings/openziti.gpg
    chmod +r /usr/share/keyrings/openziti.gpg

    local REPOSRC="deb [signed-by=/usr/share/keyrings/openziti.gpg] https://packages.openziti.org/${ZITIPAX_DEB:-zitipax-openziti-deb-stable} debian main"

    local REPOFILE="/etc/apt/sources.list.d/openziti-release.list"
    if [ -s $REPOFILE ]; then
        local EXISTINGSUM
        local REPOSUM
        EXISTINGSUM=$(checkSum < $REPOFILE)
        REPOSUM=$(checkSum <<< "$REPOSRC")
        if [ "$EXISTINGSUM" != "$REPOSUM" ]; then
            mv -v $REPOFILE{,".$(date -Iseconds)"}
            echo "$REPOSRC" > $REPOFILE

        fi
    else
        echo "$REPOSRC" >| $REPOFILE

    fi

    apt-get update
    apt-get install --yes "$@"
    for PKG in "$@"; do
        apt-cache show "$PKG=$(dpkg-query -W -f='${Version}' $PKG)"
    done
}

main(){
    if ! (( $# )); then
        echo "ERROR: No arguments provided. Please provide a space-separated list of packages to install from the OpenZiti repo." >&2
        exit 1
    fi
    # Detect the system's distribution family
    if [ -f /etc/redhat-release ]; then
        installRedHat "$@"
    elif [ -f /etc/debian_version ]; then
        installDebian "$@"
    else
        echo "ERROR: Unsupported Linux distribution family. The zrok-share package is availabe as a Debian or Red Hat package." >&2
        exit 1
    fi
}

# ensure the script is not executed before it is fully downloaded if curl'd to bash
main "$@"
