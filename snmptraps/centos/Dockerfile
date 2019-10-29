FROM centos:centos7 as builder

ARG YUM_FLAGS_COMMON="-y"
ARG YUM_FLAGS_DEV="${YUM_FLAGS_COMMON}"

ARG MAJOR_VERSION=4.2
ARG ZBX_VERSION=${MAJOR_VERSION}.8
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ENV ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES}

COPY ["snmptrapfmt_1.14+nmu1ubuntu2.tar.gz", "/tmp/"]

RUN set -eux && \
    yum ${YUM_FLAGS_COMMON} makecache && \
    yum ${YUM_FLAGS_DEV} install \
            autoconf \
            automake \
            gcc \
            patch \
            make \
            libnsl-devel \
            net-snmp-devel && \
    cd /tmp/ && \
    tar -zxf snmptrapfmt_1.14+nmu1ubuntu2.tar.gz && \
    cd /tmp/snmptrapfmt-1.14+nmu1ubuntu1/ && \
    patch -p1 < ./patches/makefile.patch && \
    make -j"$(nproc)" -s

FROM centos:centos7
LABEL maintainer "Alexey Pustovalov <alexey.pustovalov@zabbix.com>"

ARG BUILD_DATE
ARG VCS_REF

ARG YUM_FLAGS_COMMON="-y"
ARG YUM_FLAGS_PERSISTENT="${YUM_FLAGS_COMMON}"

ARG MAJOR_VERSION=4.2
ARG ZBX_VERSION=${MAJOR_VERSION}.8
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ENV ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} \
    TERM=xterm MIBDIRS=/usr/share/snmp/mibs:/var/lib/zabbix/mibs MIBS=+ALL

LABEL org.label-schema.name="zabbix-snmptraps-centos" \
      org.label-schema.vendor="Zabbix LLC" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.description="Zabbix SNMP traps receiver" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.license="GPL v2.0" \
      org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.version="1.14" \
      org.label-schema.vcs-url="https://anonscm.debian.org/gitweb/?p=collab-maint/snmptrapfmt.git" \
      org.label-schema.docker.cmd="docker run --name zabbix-snmptraps --link zabbix-server:zabbix-server -p 162:162/UDP -d zabbix-snmptraps:centos-${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY --from=builder /tmp/snmptrapfmt-1.14+nmu1ubuntu1/snmptrapfmthdlr /usr/sbin/snmptrapfmthdlr
COPY --from=builder /tmp/snmptrapfmt-1.14+nmu1ubuntu1/snmptrapfmt /usr/sbin/snmptrapfmt
COPY --from=builder /tmp/snmptrapfmt-1.14+nmu1ubuntu1/snmptrapfmt.conf /etc/snmp/snmptrapfmt.conf

RUN set -eux && \
    groupadd --system zabbix && \
    adduser -r --shell /sbin/nologin \
            -g zabbix \
            -d /var/lib/zabbix/ \
        zabbix && \
    yum ${YUM_FLAGS_COMMON} makecache && \
    yum ${YUM_FLAGS_PERSISTENT} install epel-release && \
    yum ${YUM_FLAGS_PERSISTENT} install \
            net-snmp \
            supervisor && \
    mkdir -p /var/lib/zabbix && \
    mkdir -p /var/lib/zabbix/snmptraps && \
    mkdir -p /var/lib/zabbix/mibs && \
    chown --quiet -R zabbix:root /var/lib/zabbix && \
    echo "disableAuthorization yes" >> "/etc/snmp/snmptrapd.conf" && \
    echo "traphandle default /usr/sbin/snmptrapfmthdlr" >> "/etc/snmp/snmptrapd.conf" && \
    sed -i \
            -e "/^VARFMT=/s/=.*/=\"%n %v \"/" \
            -e '/^LOGFMT=/s/=.*/=\"$x ZBXTRAP $R $G $S $e $*\"/' \
            -e "/^LOGFILE=/s/=.*/=\"\/var\/lib\/zabbix\/snmptraps\/snmptraps.log\"/" \
        "/etc/snmp/snmptrapfmt.conf" && \
    yum ${YUM_FLAGS_COMMON} clean all && \
    rm -rf /var/cache/yum

EXPOSE 162/UDP

WORKDIR /var/lib/zabbix/snmptraps/

VOLUME ["/var/lib/zabbix/snmptraps", "/var/lib/zabbix/mibs"]

COPY ["conf/etc/supervisor/", "/etc/supervisor/"]
COPY ["conf/etc/logrotate.d/zabbix_snmptraps", "/etc/logrotate.d/"]

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/supervisord.conf"]
