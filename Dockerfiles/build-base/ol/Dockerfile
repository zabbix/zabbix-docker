# syntax=docker/dockerfile:1.6
ARG OS_BASE_IMAGE=container-registry.oracle.com/os/oraclelinux:10-slim
FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION=7.0
ARG ZBX_VERSION=${MAJOR_VERSION}.28

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
      org.opencontainers.image.licenses="AGPL v3.0" \
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
            --enablerepo="ol10_baseos_latest" \
            --setopt=install_weak_deps=0 \
            --setopt=keepcache=0 \
            --best \
            --nodocs \
        oracle-epel-release-el10 && \
    microdnf -y install \
            --disablerepo="*" \
            --enablerepo="ol10_baseos_latest" \
            --enablerepo="ol10_appstream" \
            --enablerepo="ol10_codeready_builder" \
            --enablerepo="ol10_u1_developer_EPEL" \
            --setopt=install_weak_deps=0 \
            --setopt=keepcache=0 \
            --best \
            --nodocs \
        ${INSTALL_PKGS} && \
    go telemetry off
