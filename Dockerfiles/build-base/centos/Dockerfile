# syntax=docker/dockerfile:1
FROM centos:centos7

ARG MAJOR_VERSION=4.0
ARG ZBX_VERSION=${MAJOR_VERSION}.42

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
    yum -y install \
           --setopt=tsflags=nodocs \
        https://repo.zabbix.com/non-supported/rhel/7/x86_64/iksemel-1.4-2.el7.centos.x86_64.rpm \
        https://repo.zabbix.com/non-supported/rhel/7/x86_64/iksemel-devel-1.4-2.el7.centos.x86_64.rpm && \
    yum -y install epel-release && \
    INSTALL_PKGS="autoconf \
            automake \
            bash \
            gcc \
            pcre-devel \
            libcurl-devel \
            libevent-devel \
            libssh-devel \
            libxml2-devel \
            openssl-devel \
            openldap-devel \
            make \
            mariadb-devel \
            net-snmp-devel \
            OpenIPMI-devel \
            openldap-devel \
            sqlite-devel \
            postgresql-devel \
            java-1.8.0-openjdk-devel \
            git \
            gettext \
            unixODBC-devel" && \
    yum -y install \
            --setopt=tsflags=nodocs \
        ${INSTALL_PKGS} && \
    yum -y clean all && \
    rm -rf /var/cache/yum /var/lib/yum/yumdb/* /usr/lib/udev/hwdb.d/* && \
    rm -rf /etc/udev/hwdb.bin /root/.pki
