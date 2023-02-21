
# Cross-build Container for Building the Linux Executables for this zrok Project

Running this container produces three executables for zrok, one for each platform architecture: amd64, arm, arm64. You may specify the target CPU architecture as a parameter to the `docker run` command.

## Build the Container Image

You only need to build the container image once unless you change the Dockerfile or `./linux-build.sh`.

```bash
# build a container image named "zrok-builder" with the same version of Go that's declared in go.mod
docker buildx build \
    --tag=zrok-builder \
    --build-arg uid=$UID \
    --build-arg gid=$GID \
    --build-arg golang_version=$(grep -Po '^go\s+\K\d+\.\d+(\.\d+)?$' go.mod) \
    --load \
    ./docker/images/cross-build/
```

## Run the Container to Build Executables for the Desired Architectures

Executing the following `docker run` command will:

1. Mount the top-level of this repo on the container's `/mnt`
2. Run `linux-build.sh ${@}` inside the container
3. Deposit built executables in `./dist/`

```bash
# build for all three architectures: amd64 arm arm64
docker run \
    --rm \
    --name=zrok-builder \
    --volume=$PWD:/mnt \
    zrok-builder

# build only amd64 
docker run \
    --rm \
    --name=zrok-builder \
    --volume=$PWD:/mnt \
    zrok-builder \
        amd64
```

You will find the built artifacts in `./dist/`.
