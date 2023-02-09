# this builds docker.io/openziti/zrok
FROM registry.access.redhat.com/ubi8/ubi-minimal
# This build stage grabs artifacts that are copied into the final image.
# It uses the same base as the final image to maximize docker cache hits.

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

### add licenses to this directory
RUN mkdir -m0755 /licenses
COPY ./LICENSE /licenses/apache.txt

RUN mkdir -p /usr/local/bin
COPY ${ARTIFACTS_DIR}/${TARGETARCH}/${TARGETOS}/zrok /usr/local/bin/
RUN chmod 0755 /usr/local/bin/zrok

USER nobody
ENTRYPOINT [ "zrok" ]
