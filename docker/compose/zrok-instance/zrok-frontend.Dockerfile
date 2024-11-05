
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
ARG ZROK_FRONTEND_PORT
ARG ZROK_OAUTH_PORT
ARG ZROK_OAUTH_HASH_KEY
ARG ZROK_OAUTH_GOOGLE_CLIENT_ID
ARG ZROK_OAUTH_GOOGLE_CLIENT_SECRET
ARG ZROK_OAUTH_GITHUB_CLIENT_ID
ARG ZROK_OAUTH_GITHUB_CLIENT_SECRET

# render zrok frontend config.yml
COPY --chmod=0755 ./envsubst.bash ./bootstrap-frontend.bash /usr/local/bin/
COPY ./zrok-frontend-config.yml.envsubst /tmp/
RUN mkdir -p /etc/zrok-frontend/
RUN envsubst.bash \
      ZROK_DNS_ZONE=${ZROK_DNS_ZONE} \
      ZROK_FRONTEND_PORT=${ZROK_FRONTEND_PORT} \
      ZROK_OAUTH_PORT=${ZROK_OAUTH_PORT} \
      ZROK_OAUTH_HASH_KEY=${ZROK_OAUTH_HASH_KEY} \
      ZROK_OAUTH_GOOGLE_CLIENT_ID=${ZROK_OAUTH_GOOGLE_CLIENT_ID} \
      ZROK_OAUTH_GOOGLE_CLIENT_SECRET=${ZROK_OAUTH_GOOGLE_CLIENT_SECRET} \
      ZROK_OAUTH_GITHUB_CLIENT_ID=${ZROK_OAUTH_GITHUB_CLIENT_ID} \
      ZROK_OAUTH_GITHUB_CLIENT_SECRET=${ZROK_OAUTH_GITHUB_CLIENT_SECRET} \
      < /tmp/zrok-frontend-config.yml.envsubst > /etc/zrok-frontend/config.yml

# run as ziggy (or ZIGGY_UID if set in compose project)
USER ziggy
ENV HOME=/var/lib/zrok-frontend
WORKDIR /var/lib/zrok-frontend
ENTRYPOINT ["bootstrap-frontend.bash"]
