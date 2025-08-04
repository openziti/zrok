# this builds docker.io/openziti/zrok
FROM docker.io/openziti/ziti-cli:1.5.4

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
      description="Run the zrok CLI" \
      org.opencontainers.image.description="Run the zrok CLI" \
      org.opencontainers.image.source="https://github.com/openziti/zrok"

USER root

### add licenses to this directory
RUN mkdir -p -m0755 /licenses
COPY ${DOCKER_BUILD_DIR}/LICENSE /licenses/apache.txt

RUN mkdir -p /usr/local/bin
COPY --chmod=0755 ${ARTIFACTS_DIR}/${TARGETARCH}/${TARGETOS}/zrok \
      ${DOCKER_BUILD_DIR}/nfpm/zrok-enable.bash \
      ${DOCKER_BUILD_DIR}/nfpm/zrok-share.bash \
      /usr/local/bin/

COPY --chmod=0755 ${DOCKER_BUILD_DIR}/docker/images/zrok/entrypoint.bash /entrypoint.bash

USER ziggy
ENTRYPOINT [ "/entrypoint.bash", "zrok" ]
