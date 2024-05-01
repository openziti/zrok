#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
# set -o xtrace

requireBashVersion() {
    if (( "${BASH_VERSION%%.*}" < 4 )); then
        echo "This script requires Bash major version 4 or greater."
        echo "Detected version: $BASH_VERSION"
        exit 1;
    fi
}

fetchFile() {

    local url="${1}"
    local path="${2}"

    if [[ -s "$path" ]]; then
        echo "ERROR: file already exists: $path" >&2
        return 1
    fi

    if { command -v curl > /dev/null; } 2>&1; then
        curl -fLsS --output "${path}" "${url}"
    elif { command -v wget > /dev/null; } 2>&1; then
        wget --output-document "${path}" "${url}"
    else
        echo "ERROR: need one of curl or wget to fetch the artifact." >&2
        return 1
    fi
}


requireCommand() {
    if ! command -v "$1" &>/dev/null; then
        logError "this script requires command '$1'. Please install on the search PATH and try again."
        $1
    fi
}

setWorkingDir() {
    workdir="${1}"

    cd "${workdir}"

    # Count non-hidden files
    non_hidden_files=$(find . -maxdepth 1 -not -name '.*' | wc -l)
    # Count hidden files
    if ls -ld .[^.]* 2> /dev/null; then
        hidden_files=0
        for file in .[^.]*; do
            if [[ -f "$file" ]]; then
                hidden_files=$((hidden_files + 1))
            fi
        done
    else
        hidden_files=0
    fi
    # Calculate total number of files
    total_files=$((non_hidden_files + hidden_files))
    if (( total_files > 0 )); then
        echo "WARN: working directory is not empty: ${workdir}" >&2
        return 1
    fi
}

main() {
    requireBashVersion
    declare -a BINS=(unzip find)
    for BIN in "${BINS[@]}"; do
        requireCommand "$BIN"
    done
    setWorkingDir "${1:-$PWD}" || {
        echo "WARN: installing anyway in a few seconds...press Ctrl-C to abort" >&2
        sleep 9
    }
    fetchFile "${ZROK_REPO_ZIP:-"https://github.com/openziti/zrok/archive/refs/heads/main.zip"}" "zrok.zip"
    unzip -j -d . zrok.zip '*/docker/compose/zrok-instance/*'
    rm zrok.zip .gitignore fetch.bash
}

main "${@}"
