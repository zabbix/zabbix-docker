# syntax=docker/dockerfile:1.6
ARG OS_BASE_IMAGE=alpine:3.24
FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION=6.0
ARG ZBX_VERSION=${MAJOR_VERSION}.48

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
    JAVA_HOME=/usr/lib/jvm/default-jvm \
    GOPATH=/root/go \
    GOCACHE=/root/.cache/go-build \
    GOMODCACHE=/root/go/pkg/mod \
    GOTELEMETRY=off \
    GOFLAGS=-buildvcs=false

LABEL org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix build base image contains all required packages to build Zabbix images" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="GPL-2.0-or-later" \
      org.opencontainers.image.source="https://github.com/zabbix/zabbix-docker" \
      org.opencontainers.image.title="Zabbix build base" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix SIA" \
      org.opencontainers.image.version="${ZBX_VERSION}"

RUN set -eux && \
    ARCH_SUFFIX="$(apk --print-arch)"; \
    case "$ARCH_SUFFIX" in \
        x86) openjdk_packages='openjdk8 openjdk8-jre-base' ;; \
        armv7) openjdk_packages='openjdk8 openjdk8-jre-base' ;; \
        riscv64) openjdk_packages='openjdk21 openjdk21-jre-headless' ;; \
        *)       openjdk_packages='openjdk17 openjdk17-jre-headless' ;; \
    esac; \
    INSTALL_PKGS="bash \
            coreutils \
            curl \
            autoconf \
            automake \
            make \
            g++ \
            go \
            $openjdk_packages \
            gettext \
            git \
            patch \
            curl-dev \
            libevent-dev \
            libmodbus-dev \
            libpq-dev \
            libssh-dev \
            libxml2-dev \
            mysql-dev \
            net-snmp-dev \
            openipmi-dev \
            openldap-dev \
            openssl-dev \
            pcre2-dev \
            sqlite-dev \
            unixodbc-dev \
            zlib-dev" && \
    apk add \
            --no-cache \
            --no-logfile \
            --clean-protected \
        ${INSTALL_PKGS} && \
    go telemetry off
