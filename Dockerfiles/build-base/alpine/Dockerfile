# syntax=docker/dockerfile:1
FROM alpine:3.17

ARG MAJOR_VERSION=6.4
ARG ZBX_VERSION=${MAJOR_VERSION}.11

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
    ARCH_SUFFIX="$(cat /etc/apk/arch)"; \
    case "$ARCH_SUFFIX" in \
        x86) \
            openjdk_packages='openjdk8 openjdk8-jre-base'; \
            ;; \
        armv7) \
            openjdk_packages='openjdk8 openjdk8-jre-base'; \
            ;; \
        ppc64le) \
            openjdk_packages='openjdk11 openjdk11-jre-headless'; \
            ;; \
        *) openjdk_packages='openjdk17 openjdk17-jre-headless'; \
    esac; \
    INSTALL_PKGS="bash \
            autoconf \
            automake \
            coreutils \
            curl \
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
            patch \
            pcre2-dev \
            postgresql-dev \
            gettext \
            git \
            g++ \
            go \
            make \
            $openjdk_packages \
            unixodbc-dev" && \
    apk add \
            --no-cache \
            --clean-protected \
        ${INSTALL_PKGS}
