# syntax=docker/dockerfile:1.6
ARG OS_BASE_IMAGE=quay.io/centos/centos:stream10-minimal
FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION=7.0
ARG ZBX_VERSION=${MAJOR_VERSION}.30

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
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

RUN --mount=type=tmpfs,target=/var/lib/dnf/ \
    --mount=type=tmpfs,target=/var/cache/yum/ \
    set -eux && \
    INSTALL_PKGS="bash \
            autoconf \
            automake \
            make \
            gcc \
            go-toolset \
            java-21-openjdk-devel \
            gettext \
            git \
            patch \
            c-ares-devel \
            libcurl-devel \
            libevent-devel \
            libmodbus-devel \
            libssh-devel \
            libxml2-devel \
            mysql8.4-devel \
            net-snmp-devel \
            OpenIPMI-devel \
            openldap-devel \
            openssl-devel \
            pcre2-devel \
            postgresql18-private-devel \
            postgresql18-server-devel \
            sqlite-devel \
            unixODBC-devel \
            zlib-devel" && \
    microdnf -y install \
            --disablerepo="*" \
            --enablerepo="extras-common" \
            --setopt=install_weak_deps=0 \
            --setopt=keepcache=0 \
            --best \
            --nodocs \
        epel-release && \
    microdnf -y install \
            --disablerepo="*" \
            --enablerepo="baseos" \
            --enablerepo="appstream" \
            --enablerepo="crb" \
            --enablerepo="epel" \
            --setopt=install_weak_deps=0 \
            --setopt=keepcache=0 \
            --best \
            --nodocs \
        ${INSTALL_PKGS} && \
    go telemetry off
