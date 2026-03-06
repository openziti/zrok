FROM golang:1.24 AS golang

FROM debian:bookworm-slim
#
# this file mirrors the build params used in the GitHub Actions and enables
# reproducible builds for downstream forks for Ziti contributors 
#

ARG TARGETARCH
ARG go_path=/usr/share/go
ARG go_root=/usr/local/go
ARG go_cache=/usr/share/go_cache

RUN apt-get -y update \
    && apt-get -y install \
        gcc-arm-linux-gnueabi \
        gcc-arm-linux-gnueabihf \
        gcc-aarch64-linux-gnu \
        wget \
        git \
        build-essential

RUN wget -qO- https://deb.nodesource.com/setup_20.x | bash \
    && apt-get -y update \
    && apt-get -y install \
        nodejs

COPY --from=golang /usr/local/go /usr/local/go

# Install goreleaser using the Go toolchain
ENV PATH=/usr/local/go/bin:$PATH
RUN go install github.com/goreleaser/goreleaser/v2@latest \
    && install -m 0755 /root/go/bin/goreleaser /usr/local/bin/goreleaser

COPY ./linux-build.sh /usr/local/bin/linux-build.sh

ENV TARGETARCH=${TARGETARCH}
ENV GOPATH=${go_path}
ENV GOROOT=${go_root}
ENV GOCACHE=${go_cache}
ENV PATH=${go_path}/bin:${go_root}/bin:$PATH

# Create Go cache directories and make them world-writable so any user can build
RUN mkdir -p ${go_cache} ${go_path}/pkg/mod \
    && chmod -R 777 ${go_cache} ${go_path}

WORKDIR /mnt
ENTRYPOINT ["linux-build.sh"]

