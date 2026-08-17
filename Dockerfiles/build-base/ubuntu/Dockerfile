# syntax=docker/dockerfile:1.6
ARG OS_BASE_IMAGE=ubuntu:resolute
FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION=7.0
ARG ZBX_VERSION=${MAJOR_VERSION}.29

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
    PATH=/usr/lib/go-1.26/bin:$PATH \
    GOPATH=/root/go \
    GOCACHE=/root/.cache/go-build \
    GOMODCACHE=/root/go/pkg/mod \
    GOTELEMETRY=off \
    GOFLAGS=-buildvcs=false

LABEL org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix build base image contains all required packages to build Zabbix images" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="AGPL-3.0-only" \
      org.opencontainers.image.source="https://github.com/zabbix/zabbix-docker" \
      org.opencontainers.image.title="Zabbix build base" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix SIA" \
      org.opencontainers.image.version="${ZBX_VERSION}"

RUN --mount=type=cache,target=/var/cache/apt/,sharing=locked \
    --mount=type=cache,target=/var/lib/apt/,sharing=locked \
    set -eux && \
    INSTALL_PKGS="bash \
            curl \
            autoconf \
            automake \
            make \
            pkg-config \
            binutils \
            gcc \
            golang-1.26 \
            openjdk-21-jdk-headless \
            gettext \
            git \
            patch \
            unzip \
            libc-ares-dev \
            libc6-dev \
            libcurl4-openssl-dev \
            libevent-dev \
            libldap2-dev \
            libmodbus-dev \
            libmysqlclient-dev \
            libopenipmi-dev \
            libpcre2-dev \
            libpq-dev \
            libsnmp-dev \
            libsqlite3-dev \
            libssh-dev \
            libxml2-dev \
            unixodbc-dev \
            zlib1g-dev" && \
    apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
            --no-install-recommends \
        ${INSTALL_PKGS} && \
    go telemetry off
