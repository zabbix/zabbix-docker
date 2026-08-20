# syntax=docker/dockerfile:1.6
ARG OS_BASE_IMAGE=alpine:3.24
FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION=6.0
ARG ZBX_VERSION=${MAJOR_VERSION}.48
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ARG ZBX_PRODUCT_REF=${ZBX_VERSION}
ARG ZBX_PRODUCT_REVISION=unknown

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
    ZBX_SOURCES=${ZBX_SOURCES} \
    MIBDIRS=/usr/share/snmp/mibs:/var/lib/zabbix/mibs MIBS=+ALL \
    ZBX_SNMP_TRAP_DATE_FORMAT=+%Y%m%d.%H%M%S ZBX_SNMP_TRAP_FORMAT="\n" \
    ZBX_SNMP_TRAP_USE_DNS=false \
    ZABBIX_USER_HOME_DIR=/var/lib/zabbix

ENV SNMP_PERSISTENT_DIR=${ZABBIX_USER_HOME_DIR}/snmptrapd_config

LABEL org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix SNMP traps receiver" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="GPL-2.0-or-later" \
      com.zabbix.product.source="${ZBX_SOURCES}" \
      com.zabbix.product.ref="${ZBX_PRODUCT_REF}" \
      com.zabbix.product.revision="${ZBX_PRODUCT_REVISION}" \
      org.opencontainers.image.source="https://github.com/zabbix/zabbix-docker" \
      org.opencontainers.image.title="zabbix-snmptraps" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix SIA" \
      org.opencontainers.image.version="${ZBX_VERSION}"

RUN --mount=from=config_templates,target=/tmp/conf,ro \
    --mount=from=scripts,target=/tmp/scripts,ro \
    set -eux && \
    INSTALL_PKGS="bash \
            tzdata \
            net-snmp" && \
    apk add \
            --no-cache \
            --no-logfile \
            --clean-protected \
        ${INSTALL_PKGS} && \
    addgroup \
            --system \
            --gid 1995 \
        zabbix && \
    adduser \
            --system \
            --gecos "Zabbix monitoring system" \
            --disabled-password \
            --uid 1997 \
            --ingroup zabbix \
            --shell /sbin/nologin \
            --home ${ZABBIX_USER_HOME_DIR} \
        zabbix && \
    mkdir -p ${ZABBIX_USER_HOME_DIR} \
        ${ZABBIX_USER_HOME_DIR}/snmptraps \
        ${ZABBIX_USER_HOME_DIR}/mibs \
        ${SNMP_PERSISTENT_DIR} && \
    cp /tmp/conf/logrotate.d/zabbix_snmptraps /etc/logrotate.d/ && \
    cp /tmp/conf/snmp/snmptrapd.conf /etc/snmp/ && \
    cp /tmp/scripts/zabbix_trap_handler.sh /usr/sbin/ && \
    cp /tmp/scripts/snmptrapd_runner.sh /usr/sbin/ && \
    FIX_PERM_LIST="\
        /etc/snmp/ \
        ${ZABBIX_USER_HOME_DIR}/ \
        /var/tmp/ \
        /var/run/" && \
    chown --quiet -R zabbix:root ${FIX_PERM_LIST} && \
    chgrp -R 0 ${FIX_PERM_LIST} && \
    chmod -R g=u ${FIX_PERM_LIST}

EXPOSE 1162/udp

WORKDIR ${ZABBIX_USER_HOME_DIR}/snmptraps/

VOLUME ["${ZABBIX_USER_HOME_DIR}/snmptraps", "${SNMP_PERSISTENT_DIR}"]

USER 1997

CMD ["/usr/sbin/snmptrapd_runner.sh"]
