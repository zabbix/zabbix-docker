# syntax=docker/dockerfile:1
ARG OS_BASE_IMAGE=alpine:3.22
FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION=6.0
ARG ZBX_VERSION=${MAJOR_VERSION}.45
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} \
    MIBDIRS=/usr/share/snmp/mibs:/var/lib/zabbix/mibs MIBS=+ALL \
    ZBX_SNMP_TRAP_DATE_FORMAT=+%Y%m%d.%H%M%S ZBX_SNMP_TRAP_FORMAT="\n" \
    ZBX_SNMP_TRAP_USE_DNS=false \
    ZABBIX_USER_HOME_DIR="/var/lib/zabbix"

ENV SNMP_PERSISTENT_DIR="${ZABBIX_USER_HOME_DIR}/snmptrapd_config"

LABEL org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix SNMP traps receiver" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="GPL v2.0" \
      org.opencontainers.image.title="zabbix-snmptraps" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix SIA" \
      org.opencontainers.image.version="${ZBX_VERSION}"

STOPSIGNAL SIGTERM

RUN set -eux && \
    INSTALL_PKGS="bash \
            tzdata \
            net-snmp" && \
    apk add \
            --no-cache \
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
    mkdir -p ${ZABBIX_USER_HOME_DIR} && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/snmptraps && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/mibs && \
    mkdir -p ${SNMP_PERSISTENT_DIR} && \
    chown --quiet -R zabbix:root /etc/snmp/ ${ZABBIX_USER_HOME_DIR}/ /var/tmp/ /var/run/ && \
    chgrp -R 0 /etc/snmp/ ${ZABBIX_USER_HOME_DIR}/ /var/tmp/ /var/run/ && \
    chmod -R g=u /etc/snmp/ ${ZABBIX_USER_HOME_DIR}/ /var/tmp/ /var/run/

EXPOSE 1162/udp

WORKDIR ${ZABBIX_USER_HOME_DIR}/snmptraps/

VOLUME ["${ZABBIX_USER_HOME_DIR}/snmptraps", "${SNMP_PERSISTENT_DIR}"]

COPY ["conf/etc/logrotate.d/zabbix_snmptraps", "/etc/logrotate.d/"]
COPY ["conf/etc/snmp/snmptrapd.conf", "/etc/snmp/"]
COPY ["conf/usr/sbin/zabbix_trap_handler.sh", "/usr/sbin/"]
COPY ["conf/usr/sbin/snmptrapd_runner.sh", "/usr/sbin/"]

USER 1997

CMD ["/usr/sbin/snmptrapd_runner.sh"]
