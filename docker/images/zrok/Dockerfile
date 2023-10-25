# this builds docker.io/openziti/zrok
ARG ZITI_CLI_TAG="0.30.5"
ARG ZITI_CLI_IMAGE="docker.io/openziti/ziti-cli"
# this builds docker.io/openziti/ziti-controller
FROM ${ZITI_CLI_IMAGE}:${ZITI_CLI_TAG}

ARG ARTIFACTS_DIR=./dist
ARG DOCKER_BUILD_DIR=.
# e.g. linux
ARG TARGETOS
# e.g. arm64
ARG TARGETARCH

### Required OpenShift Labels 
LABEL name="openziti/zrok" \
      maintainer="support@zrok.io" \
      vendor="NetFoundry" \
      summary="Run the zrok CLI" \
      description="Run the zrok CLI"

USER root

### install packages (jq introduced in source image in next release 0.30.6)
RUN   INSTALL_PKGS="jq" && \
      microdnf -y update --setopt=install_weak_deps=0 --setopt=tsflags=nodocs && \
      microdnf -y install --setopt=install_weak_deps=0 --setopt=tsflags=nodocs ${INSTALL_PKGS}

### add licenses to this directory
RUN mkdir -p -m0755 /licenses
COPY ./LICENSE /licenses/apache.txt

RUN mkdir -p /usr/local/bin
COPY ${ARTIFACTS_DIR}/${TARGETARCH}/${TARGETOS}/zrok /usr/local/bin/
RUN chmod 0755 /usr/local/bin/zrok

USER nobody
ENTRYPOINT [ "zrok" ]
