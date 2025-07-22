# syntax=docker/dockerfile:1
ARG OS_BASE_IMAGE=ubuntu:noble
FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION=7.2
ARG ZBX_VERSION=${MAJOR_VERSION}.11

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
    PATH=/usr/lib/go-1.23/bin:$PATH

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
    INSTALL_PKGS="bash \
            autoconf \
            automake \
            binutils \
            curl \
            gcc \
            libc6-dev \
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
            openjdk-17-jdk-headless \
            patch \
            pkg-config \
            gettext \
            git \
            golang-1.23 \
            unzip \
            unixodbc-dev \
            zlib1g-dev" && \
    apt-get -y update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y \
            --no-install-recommends install \
        gpg \
        dirmngr \
        gpg-agent && \
    DISTRIB_CODENAME=$(/bin/bash -c 'source /etc/lsb-release && echo $DISTRIB_CODENAME') && \
    echo "deb http://apt.postgresql.org/pub/repos/apt $DISTRIB_CODENAME-pgdg main" > /etc/apt/sources.list.d/pgdg.list && \
    key='B97B0AFCAA1A47F044F244A07FCC7D46ACCC4CF8' && \
    export GNUPGHOME="$(mktemp -d)" && \
    gpg --batch --keyserver keyserver.ubuntu.com --recv-keys "$key" && \
    gpg --batch --export "$key" > /etc/apt/trusted.gpg.d/postgresql.gpg && \
    gpgconf --kill all && \
    rm -rf "$GNUPGHOME" && \
    apt-get -y update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y \
            --no-install-recommends install \
        ${INSTALL_PKGS} && \
    DEBIAN_FRONTEND=noninteractive apt-get -y \
            purge \
                gpg \
                dirmngr \
                gpg-agent
