FROM alpine:3.12

ARG MAJOR_VERSION=5.2
ARG ZBX_VERSION=${MAJOR_VERSION}.3
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} \
    MIBDIRS=/usr/share/snmp/mibs:/var/lib/zabbix/mibs MIBS=+ALL

LABEL org.opencontainers.image.title="zabbix-snmptraps-alpine" \
      org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.vendor="Zabbix LLC" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.description="Zabbix SNMP traps receiver" \
      org.opencontainers.image.licenses="GPL v2.0" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.version="${ZBX_VERSION}"

STOPSIGNAL SIGTERM

RUN set -eux && \
    addgroup --system --gid 1995 zabbix && \
    adduser --system \
            --gecos "Zabbix monitoring system" \
            --disabled-password \
            --uid 1997 \
            --ingroup zabbix \
            --shell /sbin/nologin \
            --home /var/lib/zabbix/ \
        zabbix && \
    adduser zabbix root && \
    apk update && \
    apk add --clean-protected --no-cache \
            tzdata \
            net-snmp && \
    touch /var/lib/net-snmp/snmptrapd.conf && \
    mkdir -p /var/lib/zabbix && \
    mkdir -p /var/lib/zabbix/snmptraps && \
    mkdir -p /var/lib/zabbix/mibs && \
    chown --quiet -R zabbix:root /etc/snmp/ /var/lib/zabbix/ /var/tmp/ /var/run/ && \
    chgrp -R 0 /etc/snmp/ /var/lib/zabbix/ /var/tmp/ /var/run/ && \
    chmod -R g=u /etc/snmp/ /var/lib/zabbix/ /var/tmp/ /var/run/ && \
    rm -rf /var/cache/apk/*

EXPOSE 1162/UDP

WORKDIR /var/lib/zabbix/snmptraps/

VOLUME ["/var/lib/zabbix/snmptraps"]

COPY ["conf/etc/logrotate.d/zabbix_snmptraps", "/etc/logrotate.d/"]
COPY ["conf/etc/snmp/snmptrapd.conf", "/etc/snmp/"]

USER 1997

CMD ["/usr/sbin/snmptrapd", "-n", "-C", "-c", "/etc/snmp/snmptrapd.conf", "-Lo", "-A"]
