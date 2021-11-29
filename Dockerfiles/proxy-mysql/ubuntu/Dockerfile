# syntax=docker/dockerfile:1
ARG MAJOR_VERSION=4.0
ARG ZBX_VERSION=${MAJOR_VERSION}.36
ARG BUILD_BASE_IMAGE=zabbix-build-mysql:ubuntu-${ZBX_VERSION}
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

FROM ${BUILD_BASE_IMAGE} as builder

FROM ubuntu:bionic

ARG MAJOR_VERSION
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} \
    MIBDIRS=/var/lib/snmp/mibs/ietf:/var/lib/snmp/mibs/iana:/usr/share/snmp/mibs:/var/lib/zabbix/mibs MIBS=+ALL \
    TINI_VERSION=v0.19.0

LABEL org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix proxy with MySQL database support" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="GPL v2.0" \
      org.opencontainers.image.source="${ZBX_SOURCES}" \
      org.opencontainers.image.title="Zabbix proxy (MySQL)" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix LLC" \
      org.opencontainers.image.version="${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}/src/zabbix_proxy/zabbix_proxy", "/usr/sbin/zabbix_proxy"]
COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}/src/zabbix_get/zabbix_get", "/usr/bin/zabbix_get"]
COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}/src/zabbix_sender/zabbix_sender", "/usr/bin/zabbix_sender"]
COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}/conf/zabbix_proxy.conf", "/etc/zabbix/zabbix_proxy.conf"]
COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}/database/mysql/create_proxy.sql.gz", "/usr/share/doc/zabbix-proxy-mysql/create.sql.gz"]

RUN set -eux && \
    ARCH_SUFFIX="$(arch)"; \
    case "$ARCH_SUFFIX" in \
        i686) export ARCH_SUFFIX='i386' ;; \
        x86_64) [ -f /lib/x86_64-linux-gnu/ld-linux-x86-64.so.2 ] && export ARCH_SUFFIX='amd64' || export ARCH_SUFFIX='i386' ;; \
        aarch64) export ARCH_SUFFIX='arm64' ;; \
        armv7l) export ARCH_SUFFIX='armhf' ;; \
        ppc64el|ppc64le) export ARCH_SUFFIX='ppc64le' ;; \
        s390x) export ARCH_SUFFIX='s390x' ;; \
        *) echo "Unknown ARCH_SUFFIX=${ARCH_SUFFIX-}"; exit 1 ;; \
    esac; \
    echo "#!/bin/sh\nexit 101" > /usr/sbin/policy-rc.d && \
    INSTALL_PKGS="ca-certificates \
            fping \
            libcurl4 \
            libevent-2.1 \
            libmysqlclient20 \
            libopenipmi0 \
            libpcre3 \
            libsnmp30 \
            libssh-4 \
            libssl1.1 \
            libxml2 \
            mysql-client \
            snmp-mibs-downloader \
            unixodbc" && \
    apt-get -y update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y \
            --no-install-recommends install \
        ${INSTALL_PKGS} && \
    INSTALL_TEMP_PKGS="gpg \
            curl \
            ca-certificates \
            dirmngr \
            gpg-agent" && \
    DEBIAN_FRONTEND=noninteractive apt-get -y \
            --no-install-recommends install \
        ${INSTALL_TEMP_PKGS} && \
    curl -L "https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini-${ARCH_SUFFIX}" -o /usr/bin/tini && \
    curl -L "https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini-${ARCH_SUFFIX}.asc" -o /tmp/tini.asc && \
    export GNUPGHOME="$(mktemp -d)" && \
    for server in $(shuf -e ha.pool.sks-keyservers.net \
                            hkp://p80.pool.sks-keyservers.net:80 \
                            ipv4.pool.sks-keyservers.net \
                            keyserver.ubuntu.com \
                            keyserver.pgp.com \
                            pgp.mit.edu) ; do \
        gpg --keyserver "$server" --recv-keys 595E85A6B1B4779EA4DAAEC70B588DFF0527A9B7 && break || : ; \
    done && \
    gpg --batch --verify /tmp/tini.asc /usr/bin/tini && \
    rm -rf "$GNUPGHOME" /tmp/tini.asc && \
    chmod +x /usr/bin/tini && \
    groupadd \
            --system \
            --gid 1995 \
        zabbix && \
    useradd \
            --system \
            --comment "Zabbix monitoring system" \
            -g zabbix \
            -G root \
            --uid 1997 \
            --shell /sbin/nologin \
            --home-dir /var/lib/zabbix/ \
        zabbix && \
    mkdir -p /etc/zabbix && \
    mkdir -p /var/lib/zabbix && \
    mkdir -p /var/lib/zabbix/enc && \
    mkdir -p /usr/lib/zabbix/externalscripts && \
    mkdir -p /var/lib/zabbix/mibs && \
    mkdir -p /var/lib/zabbix/modules && \
    mkdir -p /var/lib/zabbix/snmptraps && \
    mkdir -p /var/lib/zabbix/ssh_keys && \
    mkdir -p /var/lib/zabbix/ssl && \
    mkdir -p /var/lib/zabbix/ssl/certs && \
    mkdir -p /var/lib/zabbix/ssl/keys && \
    mkdir -p /var/lib/zabbix/ssl/ssl_ca && \
    mkdir -p /usr/share/doc/zabbix-proxy-mysql && \
    chown --quiet -R zabbix:root /etc/zabbix/ /var/lib/zabbix/ && \
    chgrp -R 0 /etc/zabbix/ /var/lib/zabbix/ && \
    chmod -R g=u /etc/zabbix/ /var/lib/zabbix/ && \
    DEBIAN_FRONTEND=noninteractive apt-get -y \
        purge \
            curl \
            gpg \
            dirmngr \
            gpg-agent && \
    apt-get -y autoremove && \
    apt-get -y clean && \
    rm -rf /var/lib/apt/lists/*

EXPOSE 10051/TCP

WORKDIR /var/lib/zabbix

VOLUME ["/var/lib/zabbix/snmptraps"]

COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENTRYPOINT ["/usr/bin/tini", "--", "/usr/bin/docker-entrypoint.sh"]

USER 1997

CMD ["/usr/sbin/zabbix_proxy", "--foreground", "-c", "/etc/zabbix/zabbix_proxy.conf"]
