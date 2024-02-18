# syntax=docker/dockerfile:1
FROM ubuntu:jammy

ARG MAJOR_VERSION=6.4
ARG ZBX_VERSION=${MAJOR_VERSION}.11

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
    PATH=/usr/lib/go-1.21/bin:$PATH

LABEL org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix build base image contains all required packages to build Zabbix images" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="GPL v2.0" \
      org.opencontainers.image.title="Zabbix build base" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix LLC" \
      org.opencontainers.image.version="${ZBX_VERSION}"

RUN --mount=type=cache,target=/var/lib/apt/,sharing=locked \
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
            golang-1.21 \
            unzip \
            unixodbc-dev" && \
    apt-get -y update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y \
            --no-install-recommends install \
        ${INSTALL_PKGS} && \
    apt-get -y autoremove && \
    apt-get -y clean
