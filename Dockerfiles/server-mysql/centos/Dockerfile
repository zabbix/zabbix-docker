# syntax=docker/dockerfile:1.6
ARG OS_BASE_IMAGE=quay.io/centos/centos:stream10-minimal
ARG MAJOR_VERSION=6.0
ARG ZBX_VERSION=${MAJOR_VERSION}.47
ARG BUILD_BASE_IMAGE=zabbix-build-mysql:centos-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} AS builder

FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
    ZBX_SOURCES=${ZBX_SOURCES} \
    MIBDIRS=/usr/share/snmp/mibs:/var/lib/zabbix/mibs MIBS=+ALL \
    NMAP_PRIVILEGED= \
    ZABBIX_USER_HOME_DIR=/var/lib/zabbix \
    ZABBIX_CONF_DIR=/etc/zabbix \
    ZBX_FPINGLOCATION=/usr/sbin/fping \
    DB_ENGINE=mysql

LABEL org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix server with MySQL database support" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="GPL v2.0" \
      org.opencontainers.image.source="${ZBX_SOURCES}" \
      org.opencontainers.image.title="Zabbix server (MySQL)" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix SIA" \
      org.opencontainers.image.version="${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/server/sbin/zabbix_server /usr/sbin/zabbix_server
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/general/bin/zabbix_get /usr/bin/
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/general/bin/zabbix_sender /usr/bin/
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/server/conf/ ${ZABBIX_CONF_DIR}/
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/server/database/mysql/create.sql.gz /usr/share/doc/zabbix-server-mysql/create.sql.gz

RUN --mount=type=tmpfs,target=/var/lib/dnf/ \
    --mount=type=tmpfs,target=/var/cache/yum/ \
    set -eux && \
    INSTALL_PKGS="bash \
            fping \
            traceroute \
            nmap \
            krb5-workstation \
            hostname \
            file-libs \
            iputils \
            libevent \
            libssh \
            libpsl \
            libbrotli \
            libxml2 \
            mysql8.4 \
            mysql8.4-libs \
            net-snmp-libs \
            OpenIPMI-libs \
            openldap \
            openssl \
            openssl-libs \
            pcre2 \
            shadow-utils \
            zlib \
            gzip \
            unixODBC" && \
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
    microdnf -y reinstall \
            --enablerepo="baseos" \
            --setopt=install_weak_deps=0 \
            --setopt=keepcache=0 \
            --best \
            --setopt=tsflags=nodocs \
        tzdata && \
    microdnf swap -y libcurl-minimal libcurl && \
    groupadd \
            --system \
            --gid 1995 \
        zabbix && \
    useradd \
            --system \
            --comment "Zabbix monitoring system" \
            -g zabbix \
            -G dialout \
            --uid 1997 \
            --shell /sbin/nologin \
            --home-dir ${ZABBIX_USER_HOME_DIR} \
        zabbix && \
    chgrp zabbix /usr/bin/nmap && \
    setcap cap_net_raw+eip /usr/bin/nmap && \
    mkdir -p ${ZABBIX_CONF_DIR} \
        ${ZABBIX_USER_HOME_DIR} \
        ${ZABBIX_USER_HOME_DIR}/dbscripts \
        ${ZABBIX_USER_HOME_DIR}/enc \
        ${ZABBIX_USER_HOME_DIR}/enc_internal \
        ${ZABBIX_USER_HOME_DIR}/export \
        ${ZABBIX_USER_HOME_DIR}/mibs \
        ${ZABBIX_USER_HOME_DIR}/modules \
        ${ZABBIX_USER_HOME_DIR}/snmptraps \
        ${ZABBIX_USER_HOME_DIR}/ssh_keys \
        ${ZABBIX_USER_HOME_DIR}/ssl \
        ${ZABBIX_USER_HOME_DIR}/ssl/certs \
        ${ZABBIX_USER_HOME_DIR}/ssl/keys \
        ${ZABBIX_USER_HOME_DIR}/ssl/ssl_ca \
        /usr/lib/zabbix/alertscripts \
        /usr/lib/zabbix/externalscripts \
        /usr/share/doc/zabbix-server-mysql && \
    FIX_PERM_LIST="\
        ${ZABBIX_CONF_DIR}/zabbix_server.conf \
        ${ZABBIX_USER_HOME_DIR}/" && \
    chown --quiet -R zabbix:0 ${FIX_PERM_LIST} && \
    chmod -R g=u ${FIX_PERM_LIST} && \
    chown --quiet zabbix:0 ${ZABBIX_CONF_DIR}/ && \
    chmod g=u ${ZABBIX_CONF_DIR} && \
    /usr/sbin/zabbix_server -V

RUN --mount=from=entrypoints,target=/tmp/entrypoints,ro \
    set -eux && \
    install -d /usr/lib/docker-entrypoint && \
    install -m 0755 /tmp/entrypoints/server-mysql.sh /usr/bin/docker-entrypoint.sh && \
    install -m 0644 /tmp/entrypoints/lib/*.sh /usr/lib/docker-entrypoint/

EXPOSE 10051/tcp

WORKDIR ${ZABBIX_USER_HOME_DIR}

VOLUME ["${ZABBIX_USER_HOME_DIR}/snmptraps", "${ZABBIX_USER_HOME_DIR}/export"]

USER 1997

ENTRYPOINT ["/usr/bin/docker-entrypoint.sh"]

CMD ["/usr/sbin/zabbix_server", "--foreground", "-c", "/etc/zabbix/zabbix_server.conf"]
