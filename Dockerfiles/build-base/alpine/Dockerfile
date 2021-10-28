# syntax=docker/dockerfile:1
FROM alpine:3.14

ARG MAJOR_VERSION=5.4
ARG ZBX_VERSION=${MAJOR_VERSION}.7

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
    PATH=/usr/lib/jvm/default-jvm/bin:$PATH \
    JAVA_HOME=/usr/lib/jvm/default-jvm

LABEL org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix build base image contains all required packages to build Zabbix images" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="GPL v2.0" \
      org.opencontainers.image.title="Zabbix build base" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix LLC" \
      org.opencontainers.image.version="${ZBX_VERSION}"

RUN set -eux && \
    INSTALL_PKGS="bash \
            autoconf \
            automake \
            coreutils \
            curl-dev \
            libevent-dev \
            libssh-dev \
            sqlite-dev \
            libmodbus-dev \
            libxml2-dev \
            mysql-dev \
            net-snmp-dev \
            openipmi-dev \
            openldap-dev \
            openssl-dev \
            pcre-dev \
            postgresql-dev \
            openjdk8 \
            gettext \
            git \
            g++ \
            go \
            make \
            openjdk8-jre-base \
            unixodbc-dev" && \
    apk add \
            --no-cache \
            --clean-protected \
        ${INSTALL_PKGS} && \
    rm -rf /var/cache/apk/*
