# syntax=docker/dockerfile:1.6
ARG OS_BASE_IMAGE=ubuntu:resolute
FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION=8.0
ARG ZBX_VERSION=${MAJOR_VERSION}
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
    ZBX_SOURCES=${ZBX_SOURCES} \
    MIBDIRS=/var/lib/mibs/ietf:/var/lib/mibs/iana:/usr/share/snmp/mibs:/var/lib/zabbix/mibs MIBS=+ALL \
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

RUN --mount=type=cache,target=/var/cache/apt/,sharing=locked \
    --mount=type=cache,target=/var/lib/apt/,sharing=locked \
    set -eux && \
    printf '%s\n' '#!/bin/sh' 'exit 101' > /usr/sbin/policy-rc.d && \
    chmod +x /usr/sbin/policy-rc.d && \
    INSTALL_PKGS="bash \
            tzdata \
            snmptrapd \
            snmp-mibs-downloader" && \
    apt-get -y update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y \
            -o Dpkg::Options::="--force-confdef" \
            -o Dpkg::Options::="--force-confold" \
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
    download-mibs && \
    rm -f \
        /var/lib/mibs/ietf/VM-MIB \
        /var/lib/mibs/ietf/TRILL-OAM-MIB \
        /var/lib/mibs/ietf/SMF-MIB \
        /var/lib/mibs/ietf/ENERGY-OBJECT-MIB \
        /var/lib/mibs/ietf/POWER-ATTRIBUTES-MIB \
        /var/lib/mibs/ietf/OLSRv2-MIB \
        /var/lib/mibs/ietf/ENERGY-OBJECT-CONTEXT-MIB \
        /var/lib/mibs/ietf/BFD-STD-MIB \
        /var/lib/mibs/ietf/SNMPv2-PDU && \
    mkdir -p ${ZABBIX_USER_HOME_DIR} \
        ${ZABBIX_USER_HOME_DIR}/snmptraps \
        ${ZABBIX_USER_HOME_DIR}/mibs \
        ${SNMP_PERSISTENT_DIR} \
        /usr/lib/docker-entrypoint && \
    FIX_PERM_LIST="\
        /etc/snmp/ \
        ${ZABBIX_USER_HOME_DIR}/ \
        /var/tmp/ \
        /var/run/ \
        /usr/lib/docker-entrypoint/" && \
    chown --quiet -R zabbix:root ${FIX_PERM_LIST} && \
    chgrp -R 0 ${FIX_PERM_LIST} && \
    chmod -R g=u ${FIX_PERM_LIST}

COPY --chmod=0644 --from=config_templates /logrotate.d/zabbix_snmptraps /etc/logrotate.d/
COPY --chown=1997:0 --chmod=0664 --from=config_templates /snmp/snmptrapd.conf /etc/snmp/
COPY --chmod=0644 --from=scripts /zabbix_trap_handler.sh /usr/sbin/
COPY --chmod=0755 --from=entrypoints /shell/snmptraps.sh /usr/bin/docker-entrypoint.sh
COPY --chmod=0644 --from=entrypoints /shell/hooks.sh /usr/lib/docker-entrypoint/hooks.sh

EXPOSE 1162/udp

WORKDIR ${ZABBIX_USER_HOME_DIR}/snmptraps/

VOLUME ["${ZABBIX_USER_HOME_DIR}/snmptraps", "${SNMP_PERSISTENT_DIR}"]

USER 1997

ENTRYPOINT ["/usr/bin/docker-entrypoint.sh"]

CMD ["/usr/sbin/snmptrapd"]
