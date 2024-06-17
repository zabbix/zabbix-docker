# syntax=docker/dockerfile:1
FROM quay.io/centos/centos:stream8

ARG MAJOR_VERSION=6.0
ARG ZBX_VERSION=${MAJOR_VERSION}.31

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
    PATH=/usr/local/go/bin:$PATH

LABEL org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix build base image contains all required packages to build Zabbix images" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="GPL v2.0" \
      org.opencontainers.image.title="Zabbix build base" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix SIA" \
      org.opencontainers.image.version="${ZBX_VERSION}"

RUN --mount=type=tmpfs,target=/var/lib/dnf/ \
    set -eux && \
    REPOLIST="baseos,appstream,powertools,epel" && \
    INSTALL_PKGS="autoconf \
            automake \
            bash \
            gcc \
            patch \
            pcre2-devel \
            libcurl-devel \
            libevent-devel \
            libssh-devel \
            libmodbus-devel \
            libxml2-devel \
            openssl-devel \
            openldap-devel \
            make \
            mysql-devel \
            net-snmp-devel \
            OpenIPMI-devel \
            openldap-devel \
            sqlite-devel \
            postgresql-private-devel \
            postgresql-server-devel \
            java-17-openjdk-devel \
            git \
            gettext \
            go-toolset \
            unixODBC-devel \
            zlib-devel" && \
    dnf -y install \
        --disablerepo "*" \
        --enablerepo "extras" \
        --setopt=tsflags=nodocs \
        --setopt=install_weak_deps=False \
        --setopt=keepcache=False \
        --best epel-release && \
    dnf -y module enable \
        --disablerepo "*" \
        --enablerepo "appstream" \
        --setopt=install_weak_deps=False \
        --setopt=keepcache=False \
        postgresql:16 mysql && \
    dnf -y install \
        --disablerepo "*" \
        --enablerepo "${REPOLIST}" \
        --setopt=tsflags=nodocs \
        --setopt=install_weak_deps=False \
        --setopt=keepcache=False \
        --best ${INSTALL_PKGS} && \
    dnf -y clean all
