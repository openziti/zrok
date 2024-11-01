
ARG ZROK_CLI_TAG=latest
ARG ZROK_CLI_IMAGE=openziti/zrok
FROM ${ZROK_CLI_IMAGE}:${ZROK_CLI_TAG}

# set up image as root
USER root

# install envsubst
RUN   INSTALL_PKGS="gettext" && \
      microdnf -y update --setopt=install_weak_deps=0 --setopt=tsflags=nodocs && \
      microdnf -y install --setopt=install_weak_deps=0 --setopt=tsflags=nodocs ${INSTALL_PKGS}

ARG ZROK_DNS_ZONE
ARG ZROK_ADMIN_TOKEN
ARG ZROK_CTRL_PORT
ARG ZITI_CTRL_ADVERTISED_PORT
ARG ZITI_PWD

# render zrok controller config.yml
COPY --chmod=0755 ./envsubst.bash ./bootstrap-controller.bash /usr/local/bin/
COPY ./zrok-controller-config.yml.envsubst /tmp/
RUN mkdir -p /etc/zrok-controller/
RUN envsubst.bash \
      ZROK_DNS_ZONE=${ZROK_DNS_ZONE} \
      ZROK_ADMIN_TOKEN=${ZROK_ADMIN_TOKEN} \
      ZROK_CTRL_PORT=${ZROK_CTRL_PORT} \
      ZITI_CTRL_ADVERTISED_PORT=${ZITI_CTRL_ADVERTISED_PORT} \
      ZITI_PWD=${ZITI_PWD} \
      < /tmp/zrok-controller-config.yml.envsubst > /etc/zrok-controller/config.yml

# run as ziggy (or ZIGGY_UID if set in compose project)
USER ziggy
ENV HOME=/var/lib/zrok-controller
WORKDIR /var/lib/zrok-controller
ENTRYPOINT ["bootstrap-controller.bash"]
