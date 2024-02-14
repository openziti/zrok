#!/usr/bin/env bash
#
# build the Linux artifacts for amd64, arm, arm64
#
# runs one background job per desired architecture unless there are too few CPUs
#
# 

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

# if no architectures supplied then default list of three
if (( ${#} )); then
    typeset -a JOBS=(${@})
else
    typeset -a JOBS=(amd64 arm arm64)
fi

# specify the Go template used by Gox to save the artifacts
GOX_OUTPUT="dist/{{.Arch}}/{{.OS}}/{{.Dir}}"
# count the number of available CPUs for time-efficient parallelism
PROC_COUNT=$(nproc --all)
# compute the number of processors available for each job, rounded down to integer
PROCS_PER_JOB=$((PROC_COUNT / ${#JOBS[@]}))
# if multiple jobs and at least one processor for each job then background, else foreground with all available CPUs-1 (gox default)
if (( ${#JOBS[@]} > 1 && ${PROCS_PER_JOB} )); then 
    BACKGROUND="&"
    # initialize an associative array in which to map background PIDs to the ARCH being built
    typeset -A BUILDS
else
    BACKGROUND=""   # run normally in foreground
    PROCS_PER_JOB=0 # invokes gox default to use all CPUs-1
fi

(
    HOME=/tmp/builder
    # Navigate to the "ui" directory and run npm commands
    npm config set cache /mnt/.npm
    cd ./ui/
    mkdir -p $HOME
    npm install
    npm run build
)

for ARCH in ${JOBS[@]}; do
    GOX_CMD="
        gox \
            -cgo \
            -os=linux \
            -arch=${ARCH} \
            -output=${GOX_OUTPUT} \
            -parallel=${PROCS_PER_JOB} \
            ./cmd/zrok
    "
case ${ARCH} in
        amd64)  eval ${GOX_CMD} ${BACKGROUND}
                (( ${PROCS_PER_JOB} )) && BUILDS[${!}]=${ARCH}  # if greater than zero procs per job then map background pid->arch
        ;;
        arm)    eval CC=arm-linux-gnueabihf-gcc ${GOX_CMD} ${BACKGROUND}
                (( ${PROCS_PER_JOB} )) && BUILDS[${!}]=${ARCH}
        ;;
        arm64)  eval CC=aarch64-linux-gnu-gcc ${GOX_CMD} ${BACKGROUND}
                (( ${PROCS_PER_JOB} )) && BUILDS[${!}]=${ARCH}
        ;;
        *)      echo "ERROR: invalid architecture '${ARCH}', must be one of amd64, arm, arm64" >&2
                exit 1
        ;;
    esac
done

# if not background in parallel then exit now with well earned success
[[ -z "${BACKGROUND:-}" ]] && exit 0

# Wait for builds in the background and exit with an error if any fail
EXIT=0
while true; do
    # "wait -p" requires BASH >=5.1 which is present in Ubuntu 20.10 and Debian Bullseye
    wait -n -p JOB_PID; JOB_RESULT=$?
    echo "Building for ${BUILDS[$JOB_PID]} finished with result ${JOB_RESULT}"
    (( ${JOB_RESULT} )) && EXIT=1
done

exit ${EXIT}
