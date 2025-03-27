# syntax=docker/dockerfile:1
ARG OS_BASE_IMAGE=ubuntu:noble
ARG MAJOR_VERSION=7.2
ARG ZBX_VERSION=${MAJOR_VERSION}.5
ARG BUILD_BASE_IMAGE=zabbix-build-sqlite3:ubuntu-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} AS builder

FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} \
    MIBDIRS=/var/lib/mibs/ietf:/var/lib/mibs/iana:/usr/share/snmp/mibs:/var/lib/zabbix/mibs MIBS=+ALL \
    NMAP_PRIVILEGED="" \
    ZABBIX_USER_HOME_DIR="/var/lib/zabbix" \
    ZABBIX_CONF_DIR="/etc/zabbix"

ENV ZBX_FPINGLOCATION="/usr/bin/fping" \
    ZBX_LOADMODULEPATH="${ZABBIX_USER_HOME_DIR}/modules" \
    ZBX_SNMPTRAPPERFILE="${ZABBIX_USER_HOME_DIR}/snmptraps/snmptraps.log" \
    ZBX_SSHKEYLOCATION="${ZABBIX_USER_HOME_DIR}/ssh_keys/" \
    ZBX_SSLCERTLOCATION="${ZABBIX_USER_HOME_DIR}/ssl/certs/" \
    ZBX_SSLKEYLOCATION="${ZABBIX_USER_HOME_DIR}/ssl/keys/" \
    ZBX_SSLCALOCATION="${ZABBIX_USER_HOME_DIR}/ssl/ssl_ca/"

LABEL org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix proxy with SQLite3 database support" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="AGPL v3.0" \
      org.opencontainers.image.source="${ZBX_SOURCES}" \
      org.opencontainers.image.title="Zabbix proxy (SQLite3)" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix SIA" \
      org.opencontainers.image.version="${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}-output/proxy/sbin/zabbix_proxy", "/usr/sbin/zabbix_proxy"]
COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}-output/proxy/conf/", "${ZABBIX_CONF_DIR}/"]
COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}-output/general/bin/", "/usr/bin/"]

RUN --mount=type=cache,target=/var/cache/apt/,sharing=locked \
    --mount=type=cache,target=/var/lib/apt/,sharing=locked \
    set -eux && \
    echo "#!/bin/sh\nexit 101" > /usr/sbin/policy-rc.d && \
    INSTALL_PKGS="bash \
            tzdata \
            traceroute \
            nmap \
            krb5-user \
            ca-certificates \
            fping \
            openssl \
            libcurl4t64 \
            libevent-2.1 \
            libevent-pthreads-2.1 \
            libevent-extra-2.1 \
            libopenipmi0t64 \
            libpcre2-8-0 \
            libsnmp40t64 \
            libsqlite3-0 \
            libssh-4 \
            libssl3t64 \
            libxml2 \
            snmp-mibs-downloader \
            unixodbc \
            zlib1g" && \
    apt-get -y update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y \
            --no-install-recommends install \
        ${INSTALL_PKGS} && \
    groupadd \
            --system \
            --gid 1995 \
        zabbix && \
    useradd \
            --system \
            --comment "Zabbix monitoring system" \
            -g zabbix \
            --uid 1997 \
            --shell /sbin/nologin \
            --home-dir ${ZABBIX_USER_HOME_DIR} \
        zabbix && \
    chgrp zabbix /usr/bin/nmap && \
    setcap cap_net_raw+eip /usr/bin/nmap && \
    mkdir -p ${ZABBIX_CONF_DIR} && \
    mkdir -p ${ZABBIX_USER_HOME_DIR} && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/db_data && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/enc && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/enc_internal && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/mibs && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/modules && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/snmptraps && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/ssh_keys && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/ssl && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/ssl/certs && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/ssl/keys && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/ssl/ssl_ca && \
    mkdir -p /usr/lib/zabbix/externalscripts && \
    chown --quiet -R zabbix:root ${ZABBIX_CONF_DIR}/zabbix_proxy_modules.conf ${ZABBIX_USER_HOME_DIR} && \
    chown --quiet zabbix:root ${ZABBIX_CONF_DIR}/ && \
    chgrp -R 0 ${ZABBIX_CONF_DIR}/zabbix_proxy_modules.conf ${ZABBIX_USER_HOME_DIR} && \
    chgrp 0 ${ZABBIX_CONF_DIR} && \
    chmod -R g=u ${ZABBIX_CONF_DIR}/zabbix_proxy_modules.conf ${ZABBIX_USER_HOME_DIR} && \
    chmod g=u ${ZABBIX_CONF_DIR} && \
    /usr/sbin/zabbix_proxy -V

EXPOSE 10051/TCP

WORKDIR ${ZABBIX_USER_HOME_DIR}

VOLUME ["${ZABBIX_USER_HOME_DIR}/snmptraps"]

COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENTRYPOINT ["/usr/bin/docker-entrypoint.sh"]

USER 1997

CMD ["/usr/sbin/zabbix_proxy", "--foreground", "-c", "/etc/zabbix/zabbix_proxy.conf"]
