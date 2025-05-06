FROM golang:1.21 AS golang

FROM debian:bookworm-slim
#
# this file mirrors the build params used in the GitHub Actions and enables
# reproducible builds for downstream forks for Ziti contributors 
#

ARG TARGETARCH
ARG golang_version=1.21.3
ARG go_path=/usr/share/go
ARG go_root=/usr/local/go
ARG go_cache=/usr/share/go

RUN apt-get -y update \
    && apt-get -y install \
        gcc-arm-linux-gnueabi \
        gcc-arm-linux-gnueabihf \
        gcc-aarch64-linux-gnu \
        wget \
        git \
        build-essential


RUN wget -qO- https://deb.nodesource.com/setup_18.x | bash \
    && apt-get -y update \
    && apt-get -y install \
        nodejs


COPY --from=golang /usr/local/go /usr/local/go

COPY ./linux-build.sh /usr/local/bin/linux-build.sh

ENV TARGETARCH=${TARGETARCH}
ENV GOPATH=${go_path}
ENV GOROOT=${go_root}
ENV GOCACHE=${go_cache}
ENV PATH=${go_path}/bin:${go_root}/bin:$PATH

WORKDIR /mnt
ENTRYPOINT ["linux-build.sh"]

