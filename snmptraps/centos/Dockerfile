FROM centos:centos7
LABEL maintainer "Alexey Pustovalov <alexey.pustovalov@zabbix.com>"

ARG BUILD_DATE
ARG VCS_REF

ARG YUM_FLAGS_COMMON="--quiet -y"
ARG YUM_FLAGS_PERSISTANT="${YUM_FLAGS_COMMON}"
ARG YUM_FLAGS_DEV="${YUM_FLAGS_COMMON}"
ENV TERM=xterm MIBDIRS=/usr/share/snmp/mibs:/var/lib/zabbix/mibs MIBS=+ALL

LABEL org.label-schema.name="zabbix-snmptraps-centos" \
      org.label-schema.vendor="Zabbix LLC" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.description="Zabbix SNMP traps receiver" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.license="GPL 2.0"

ARG MAJOR_VERSION=master
ARG ZBX_VERSION=${MAJOR_VERSION}
ARG ZBX_SOURCES=svn://svn.zabbix.com/tags/${ZBX_VERSION}/
ENV ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES}

LABEL org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.version="1.14" \
      org.label-schema.vcs-url="https://anonscm.debian.org/gitweb/?p=collab-maint/snmptrapfmt.git" \
      org.label-schema.docker.cmd="docker run --name zabbix-snmptraps --link zabbix-server:zabbix-server -p 162:162/UDP -d zabbix-snmptraps:centos-${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY ["snmptrapfmt_1.14+nmu1ubuntu2.tar.gz", "/tmp/"]

RUN groupadd --system zabbix && \
    adduser -r --shell /sbin/nologin \
            -g zabbix \
            -d /var/lib/zabbix/ \
        zabbix && \
    yum ${YUM_FLAGS_COMMON} makecache && \
    yum ${YUM_FLAGS_DEV} install epel-release && \
    yum ${YUM_FLAGS_PERSISTANT} install \
            net-snmp \
            supervisor && \
    yum ${YUM_FLAGS_PERSISTANT} install \
            autoconf \
            automake \
            gcc \
            patch \
            make \
            libnsl-devel \
            net-snmp-devel && \
    mkdir -p /var/lib/zabbix && \
    mkdir -p /var/lib/zabbix/snmptraps && \
    mkdir -p /var/lib/zabbix/mibs && \
    chown --quiet -R zabbix:root /var/lib/zabbix && \
    cd /tmp/ && \
    tar -zxf snmptrapfmt_1.14+nmu1ubuntu2.tar.gz && \
    ls -lah && \
    cd /tmp/snmptrapfmt-1.14+nmu1ubuntu1/ && \
    patch -p1 < ./patches/makefile.patch && \
    make -j"$(nproc)" -s 1>/dev/null && \
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
    rm -rf /tmp/snmptrapfmt_1.14+nmu1ubuntu2.tar.gz 1>/dev/null && \
    rm -rf /tmp/snmptrapfmt-1.14+nmu1ubuntu1/ 1>/dev/null && \
    yum history && \
    yum ${YUM_FLAGS_COMMON} history undo `yum history | sed -n 4p |column -t | cut -d' ' -f1` 1>/dev/null && \
    yum ${YUM_FLAGS_COMMON} clean all && \
    rm -rf /etc/udev/hwdb.bin && \
    rm -rf /var/cache/yum && \
    rm -rf /var/lib/yum/yumdb/* && \
    rm -rf /usr/lib/udev/hwdb.d/* && \
    rm -rf /root/.pki && \
    rm -rf /root/.subversion

EXPOSE 162/UDP

WORKDIR /var/lib/zabbix/snmptraps/

VOLUME ["/var/lib/zabbix/snmptraps", "/var/lib/zabbix/mibs"]

COPY ["conf/etc/supervisor/", "/etc/supervisor/"]
COPY ["conf/etc/logrotate.d/zabbix_snmptraps", "/etc/logrotate.d/"]

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/supervisord.conf"]
