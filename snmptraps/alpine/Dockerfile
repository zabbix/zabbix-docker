FROM alpine:3.10
LABEL maintainer "Alexey Pustovalov <alexey.pustovalov@zabbix.com>"

ARG BUILD_DATE
ARG VCS_REF

ARG APK_FLAGS_COMMON=""
ARG APK_FLAGS_PERSISTENT="${APK_FLAGS_COMMON} --clean-protected --no-cache"
ARG APK_FLAGS_DEV="${APK_FLAGS_COMMON} --no-cache"
ENV TERM=xterm MIBDIRS=/usr/share/snmp/mibs:/var/lib/zabbix/mibs MIBS=+ALL

LABEL org.label-schema.name="zabbix-snmptraps-alpine" \
      org.label-schema.vendor="Zabbix LLC" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.description="Zabbix SNMP traps receiver" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.license="GPL v2.0"

ARG MAJOR_VERSION=4.2
ARG ZBX_VERSION=${MAJOR_VERSION}.6
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ENV ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES}

LABEL org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.version="1.14" \
      org.label-schema.vcs-url="https://anonscm.debian.org/gitweb/?p=collab-maint/snmptrapfmt.git" \
      org.label-schema.docker.cmd="docker run --name zabbix-snmptraps --link zabbix-server:zabbix-server -p 162:162/UDP -d zabbix-snmptraps:alpine-${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY ["snmptrapfmt_1.14+nmu1ubuntu2.tar.gz", "/tmp/"]

RUN set -eux && \
    addgroup zabbix && \
    adduser -S \
            -D -G zabbix \
            -h /var/lib/zabbix/ \
        zabbix && \
    apk update && \
    apk add ${APK_FLAGS_PERSISTENT} \
            net-snmp \
            supervisor && \
    apk add ${APK_FLAGS_DEV} --virtual build-dependencies \
            alpine-sdk \
            autoconf \
            automake \
            libnsl-dev \
            net-snmp-dev && \
    mkdir -p /var/lib/zabbix && \
    mkdir -p /var/lib/zabbix/snmptraps && \
    mkdir -p /var/lib/zabbix/mibs && \
    chown --quiet -R zabbix:root /var/lib/zabbix && \
    cd /tmp/ && \
    tar -zxvf snmptrapfmt_1.14+nmu1ubuntu2.tar.gz && \
    cd /tmp/snmptrapfmt-1.14+nmu1ubuntu1/ && \
    patch -p1 < ./patches/makefile.patch && \
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
    apk del ${APK_FLAGS_COMMON} --purge --no-network \
            build-dependencies && \
    rm -rf /var/cache/apk/*

EXPOSE 162/UDP

WORKDIR /var/lib/zabbix/snmptraps/

VOLUME ["/var/lib/zabbix/snmptraps", "/var/lib/zabbix/mibs"]

COPY ["conf/etc/supervisor/", "/etc/supervisor/"]
COPY ["conf/etc/logrotate.d/zabbix_snmptraps", "/etc/logrotate.d/"]

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/supervisord.conf"]
