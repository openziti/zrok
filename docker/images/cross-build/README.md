
# Cross-build Container for Building the Linux Executables for this zrok Project

Running this container produces one snapshot executable for zrok from the current checkout, even if dirty. Platform choices are: `amd64`, `arm64`, `arm` (the arm/v7 "armhf" ABI target), and `armel` (the arm/v7 "armel" ABI). You may specify the target architecture as a parameter to the `docker run` command.

## Build the Container Image

You only need to build the container image once unless you change the Dockerfile or `./linux-build.sh`.

```bash
# build a container image named "zrok-builder"
docker buildx build -t zrok-builder ./docker/images/cross-build --load;
```

## Run the Container to Build Executables for the Desired Architectures

Executing the following `docker run` command will:

1. Mount the top-level of this repo on the container's `/mnt`
2. Run `linux-build.sh ${@}` inside the container
3. Deposit built executable in `./dist/`

```bash
# cross-build for arm64/aarch64 architecture on Linux
docker run --user "${UID}" --rm --volume=${HOME}/.cache/go-build:/usr/share/go --volume "${PWD}:/mnt" zrok-builder arm64
```

You will find the built artifacts in `./dist/`.
