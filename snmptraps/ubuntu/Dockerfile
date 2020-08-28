FROM ubuntu:bionic

ARG MAJOR_VERSION=master
ARG ZBX_VERSION=${MAJOR_VERSION}
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} \
    MIBDIRS=/usr/share/snmp/mibs:/var/lib/zabbix/mibs MIBS=+ALL

LABEL org.opencontainers.image.title="zabbix-snmptraps-ubuntu" \
      org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.vendor="Zabbix LLC" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.description="Zabbix SNMP traps receiver" \
      org.opencontainers.image.licenses="GPL v2.0" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.version="${ZBX_VERSION}" \
      org.opencontainers.image.source="https://anonscm.debian.org/gitweb/?p=collab-maint/snmptrapfmt.git"

STOPSIGNAL SIGTERM

COPY ["snmptrapfmt_1.14+nmu1ubuntu2.tar.gz", "/tmp/"]

RUN set -eux && \
    echo "#!/bin/sh\nexit 0" > /usr/sbin/policy-rc.d && \
    addgroup --system --gid 1995 --quiet zabbix && \
    adduser --quiet \
            --system --disabled-login \
            --ingroup zabbix --ingroup root \
            --uid 1997 \
            --home /var/lib/zabbix/ \
        zabbix && \
    apt-get -y update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y --no-install-recommends install \
            tzdata \
            snmp-mibs-downloader \
            snmptrapd \
            supervisor && \
    download-mibs && \
    DEBIAN_FRONTEND=noninteractive apt-get -y --no-install-recommends install \
            autoconf \
            automake \
            gcc \
            patch \
            make \
            libc6-dev \
            libsnmp-dev && \
    mkdir -p /var/lib/zabbix && \
    mkdir -p /var/lib/zabbix/snmptraps && \
    mkdir -p /var/lib/zabbix/mibs && \
    cd /tmp/ && \
    tar -zxvf snmptrapfmt_1.14+nmu1ubuntu2.tar.gz && \
    ls -lah && \
    cd /tmp/snmptrapfmt-1.14+nmu1ubuntu1/ && \
    patch -p1 < ./patches/makefile.patch && \
    patch -p1 < ./patches/pid_location.patch && \
    make -j"$(nproc)" -s && \
    cp snmptrapfmthdlr /usr/sbin/snmptrapfmthdlr && \
    cp snmptrapfmt /usr/sbin/snmptrapfmt && \
    cp snmptrapfmt.conf /etc/snmp/snmptrapfmt.conf && \
    echo "disableAuthorization yes" >> "/etc/snmp/snmptrapd.conf" && \
    echo "traphandle default /usr/sbin/snmptrapfmthdlr" >> "/etc/snmp/snmptrapd.conf" && \
    sed -i \
            -e "/^VARFMT=/s/=.*/=\"%n %v \"/" \
            -e '/^LOGFMT=/s/=.*/=\"$x ZBXTRAP $R $G $S $e $*\"/' \
            -e "/^LOGFILE=/s/=.*/=\"\/var\/lib\/zabbix\/snmptraps\/snmptraps.log\"/" \
        "/etc/snmp/snmptrapfmt.conf" && \
    rm -rf /tmp/snmptrapfmt_1.14+nmu1ubuntu2.tar.gz && \
    rm -rf /tmp/snmptrapfmt-1.14+nmu1ubuntu1/ && \
    chown --quiet -R zabbix:root /etc/snmp/ /var/lib/zabbix/ /var/tmp/ /var/run/ && \
    chgrp -R 0 /etc/snmp/ /var/lib/zabbix/ /var/tmp/ /var/run/ && \
    chmod -R g=u /etc/snmp/ /var/lib/zabbix/ /var/tmp/ /var/run/ && \
    DEBIAN_FRONTEND=noninteractive apt-get -y purge \
            autoconf \
            automake \
            gcc \
            patch \
            make \
            libc6-dev \
            libsnmp-dev && \
    apt-get -y autoremove && \
    rm -rf /var/lib/apt/lists/*

EXPOSE 1162/UDP

WORKDIR /var/lib/zabbix/snmptraps/

VOLUME ["/var/lib/zabbix/snmptraps"]

COPY ["conf/etc/supervisor/", "/etc/supervisor/"]
COPY ["conf/etc/logrotate.d/zabbix_snmptraps", "/etc/logrotate.d/"]

USER 1997

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/supervisord.conf"]
