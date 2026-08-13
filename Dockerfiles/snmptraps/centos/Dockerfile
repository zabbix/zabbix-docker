# syntax=docker/dockerfile:1.6
ARG OS_BASE_IMAGE=quay.io/centos/centos:stream10-minimal
ARG ENTRYPOINT_BUILD_IMAGE=golang:1.26.5-alpine

FROM ${ENTRYPOINT_BUILD_IMAGE} AS entrypoint-builder

RUN --mount=from=entrypoints,target=/tmp/entrypoints,ro \
    --mount=type=cache,target=/root/.cache/go-build \
    cd /tmp/entrypoints && \
    CGO_ENABLED=0 go build -buildvcs=false -trimpath -ldflags="-s -w -buildid=" \
            -o /tmp/docker-entrypoint \
        ./cmd/snmptraps

FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION=8.0
ARG ZBX_VERSION=${MAJOR_VERSION}
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
    ZBX_SOURCES=${ZBX_SOURCES} \
    MIBDIRS=/usr/share/snmp/mibs:/var/lib/zabbix/mibs MIBS=+ALL \
    ZBX_SNMP_TRAP_DATE_FORMAT=+%Y-%m-%dT%T%z ZBX_SNMP_TRAP_FORMAT="\n" \
    ZBX_SNMP_TRAP_USE_DNS=false \
    ZABBIX_USER_HOME_DIR=/var/lib/zabbix

ENV SNMP_PERSISTENT_DIR=${ZABBIX_USER_HOME_DIR}/snmptrapd_config

LABEL org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix SNMP traps receiver" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="AGPL v3.0" \
      org.opencontainers.image.title="zabbix-snmptraps" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix SIA" \
      org.opencontainers.image.version="${ZBX_VERSION}"

STOPSIGNAL SIGTERM

RUN --mount=type=tmpfs,target=/var/lib/dnf/ \
    --mount=type=tmpfs,target=/var/cache/yum/ \
    set -eux && \
    INSTALL_PKGS="bash \
            shadow-utils \
            tzdata \
            net-snmp" && \
    microdnf -y install \
            --disablerepo="*" \
            --enablerepo="baseos" \
            --enablerepo="appstream" \
            --setopt=install_weak_deps=0 \
            --setopt=keepcache=0 \
            --best \
            --nodocs \
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
    mkdir -p ${ZABBIX_USER_HOME_DIR} \
        ${ZABBIX_USER_HOME_DIR}/snmptraps \
        ${ZABBIX_USER_HOME_DIR}/mibs \
        ${SNMP_PERSISTENT_DIR} && \
    FIX_PERM_LIST="\
        /etc/snmp/ \
        ${ZABBIX_USER_HOME_DIR}/ \
        /var/tmp/ \
        /var/run/" && \
    chown --quiet -R zabbix:root ${FIX_PERM_LIST} && \
    chgrp -R 0 ${FIX_PERM_LIST} && \
    chmod -R g=u ${FIX_PERM_LIST}

COPY --chmod=0644 --from=config_templates /logrotate.d/zabbix_snmptraps /etc/logrotate.d/
COPY --chown=1997:0 --chmod=0664 --from=config_templates /snmp/snmptrapd.conf /etc/snmp/
COPY --chmod=0644 --from=scripts /zabbix_trap_handler.sh /usr/sbin/
COPY --chmod=0755 --from=entrypoint-builder /tmp/docker-entrypoint /usr/bin/docker-entrypoint
COPY --from=entrypoints /licenses/go-THIRD-PARTY-NOTICES.txt /licenses/

EXPOSE 1162/udp

WORKDIR ${ZABBIX_USER_HOME_DIR}/snmptraps/

VOLUME ["${ZABBIX_USER_HOME_DIR}/snmptraps", "${SNMP_PERSISTENT_DIR}"]

USER 1997

ENTRYPOINT ["/usr/bin/docker-entrypoint"]

CMD ["/usr/sbin/snmptrapd"]
