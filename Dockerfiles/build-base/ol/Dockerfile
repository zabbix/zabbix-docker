# syntax=docker/dockerfile:1
FROM oraclelinux:8-slim

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

COPY ["conf/etc/yum.repos.d/oracle-epel-ol8.repo", "/etc/yum.repos.d/oracle-epel-ol8.repo"]

RUN --mount=type=tmpfs,target=/var/lib/dnf/ \
    set -eux && \
    INSTALL_PKGS="autoconf \
            automake \
            bash \
            gcc \
            go-toolset \
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
            unixODBC-devel \
            zlib-devel" && \
    microdnf -y module enable \
        --disablerepo "*" \
        --enablerepo "ol8_appstream" \
        --setopt=install_weak_deps=0 \
        --setopt=keepcache=0 \
        postgresql:16 mysql && \
    microdnf -y install \
        --disablerepo "*" \
        --enablerepo "ol8_baseos_latest" \
        --enablerepo "ol8_appstream" \
        --enablerepo "ol8_codeready_builder" \
        --enablerepo="ol8_developer_EPEL" \
        --setopt=install_weak_deps=0 \
        --setopt=keepcache=0 \
        --best \
        --nodocs ${INSTALL_PKGS} && \
    microdnf -y clean all
