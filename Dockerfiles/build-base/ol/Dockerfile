# syntax=docker/dockerfile:1
FROM oraclelinux:8-slim

ARG MAJOR_VERSION=6.2
ARG ZBX_VERSION=${MAJOR_VERSION}.4

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
    PATH=/usr/local/go/bin:$PATH

LABEL org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix build base image contains all required packages to build Zabbix images" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="GPL v2.0" \
      org.opencontainers.image.title="Zabbix build base" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix LLC" \
      org.opencontainers.image.version="${ZBX_VERSION}"

RUN set -eux && \
    INSTALL_PKGS="autoconf \
            automake \
            bash \
            gcc \
            go-toolset \
            pcre2-devel \
            libcurl-devel \
            libevent-devel \
            libssh-devel \
            libxml2-devel \
            openssl-devel \
            openldap-devel \
            make \
            mysql-devel \
            net-snmp-devel \
            OpenIPMI-devel \
            openldap-devel \
            sqlite-devel \
            postgresql-devel \
            java-1.8.0-openjdk-devel \
            git \
            gettext \
            unixODBC-devel" && \
    microdnf -y module enable mysql && \
    microdnf module enable postgresql:13 && \
    microdnf -y install \
        --disablerepo "*" \
        --enablerepo "ol8_baseos_latest" \
        --enablerepo "ol8_appstream" \
        --enablerepo "ol8_codeready_builder" \
        --setopt=install_weak_deps=0 \
        --best \
        --nodocs ${INSTALL_PKGS} && \
    microdnf -y clean all && \
    rm -rf /var/cache/yum /var/lib/yum/yumdb/* /usr/lib/udev/hwdb.d/* && \
    rm -rf /var/cache/dnf /etc/udev/hwdb.bin /root/.pki
