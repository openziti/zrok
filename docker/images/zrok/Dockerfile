# this builds docker.io/openziti/zrok
FROM docker.io/openziti/ziti-cli:1.1.15

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
COPY ./LICENSE /licenses/apache.txt

RUN mkdir -p /usr/local/bin
COPY ${ARTIFACTS_DIR}/${TARGETARCH}/${TARGETOS}/zrok \
      ./nfpm/zrok-enable.bash \
      ./nfpm/zrok-share.bash \
      /usr/local/bin/
RUN chmod 0755 \
      /usr/local/bin/zrok \
      /usr/local/bin/zrok-enable.bash \
      /usr/local/bin/zrok-share.bash

USER ziggy
ENTRYPOINT [ "zrok" ]
