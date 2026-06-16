# syntax=docker/dockerfile:1.6
ARG OS_BASE_IMAGE=ubuntu:resolute
FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION=7.4
ARG ZBX_VERSION=${MAJOR_VERSION}.11

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
      org.opencontainers.image.licenses="AGPL v3.0" \
      org.opencontainers.image.title="Zabbix build base" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix SIA" \
      org.opencontainers.image.version="${ZBX_VERSION}"

RUN --mount=type=cache,target=/var/cache/apt/,sharing=locked \
    --mount=type=cache,target=/var/lib/apt/,sharing=locked \
    set -eux && \
    printf '%s\n' '#!/bin/sh' 'exit 101' > /usr/sbin/policy-rc.d && \
    chmod +x /usr/sbin/policy-rc.d && \
    INSTALL_PKGS="bash \
            autoconf \
            automake \
            binutils \
            curl \
            gcc \
            libc6-dev \
            libc-ares-dev \
            libcurl4-openssl-dev \
            libevent-dev \
            libmodbus-dev \
            libldap2-dev \
            libmysqlclient-dev \
            libopenipmi-dev \
            libpcre2-dev \
            libsnmp-dev \
            libsqlite3-dev \
            libpq-dev \
            libssh-dev \
            libxml2-dev \
            make \
            openjdk-21-jdk-headless \
            patch \
            pkg-config \
            gettext \
            git \
            golang-1.26 \
            unzip \
            unixodbc-dev \
            zlib1g-dev" && \
    apt-get -y update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y \
            -o Dpkg::Options::="--force-confdef" \
            -o Dpkg::Options::="--force-confold" \
            --no-install-recommends install \
        ${INSTALL_PKGS} && \
    go telemetry off
