FROM ubuntu:bionic
LABEL maintainer "Alexey Pustovalov <alexey.pustovalov@zabbix.com>"

ARG BUILD_DATE
ARG VCS_REF

ARG APT_FLAGS_COMMON="-y"
ARG APT_FLAGS_PERSISTENT="${APT_FLAGS_COMMON} --no-install-recommends"
ARG APT_FLAGS_DEV="${APT_FLAGS_COMMON} --no-install-recommends"
ENV LANG=en_US.UTF-8 LANGUAGE=en_US:en LC_ALL=en_US.UTF-8 TERM=xterm \
    MIBDIRS=/var/lib/mibs/iana:/var/lib/mibs/ietf:/usr/share/snmp/mibs:/var/lib/zabbix/mibs MIBS=+ALL

LABEL org.label-schema.name="zabbix-snmptraps-ubuntu" \
      org.label-schema.vendor="Zabbix LLC" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.description="Zabbix SNMP traps receiver" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.license="GPL v2.0"

ARG MAJOR_VERSION=4.2
ARG ZBX_VERSION=${MAJOR_VERSION}.8
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ENV ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES}

LABEL org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.version="1.14" \
      org.label-schema.vcs-url="https://anonscm.debian.org/gitweb/?p=collab-maint/snmptrapfmt.git" \
      org.label-schema.docker.cmd="docker run --name zabbix-snmptraps --link zabbix-server:zabbix-server -p 162:162/UDP -d zabbix-snmptraps:ubuntu-${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY ["snmptrapfmt_1.14+nmu1ubuntu2_amd64.deb", "/tmp/"]

RUN set -eux && \
    apt-get ${APT_FLAGS_COMMON} update && \
    DEBIAN_FRONTEND=noninteractive apt-get ${APT_FLAGS_PERSISTENT} install locales && \
    locale-gen $LC_ALL && \
    echo "#!/bin/sh\nexit 0" > /usr/sbin/policy-rc.d && \
    addgroup --system --quiet zabbix && \
    adduser --quiet \
            --system --disabled-login \
            --ingroup zabbix \
            --home /var/lib/zabbix/ \
        zabbix && \
    apt-get ${APT_FLAGS_COMMON} update && \
    DEBIAN_FRONTEND=noninteractive apt-get ${APT_FLAGS_PERSISTENT} install \
            snmp-mibs-downloader \
            snmptrapd \
            supervisor && \
    download-mibs && \
    dpkg -i /tmp/snmptrapfmt_1.14+nmu1ubuntu2_amd64.deb && \
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
    rm -rf /tmp/snmptrapfmt_1.14+nmu1ubuntu2_amd64.deb && \
    apt-get ${APT_FLAGS_COMMON} autoremove 1>/dev/null && \
    rm -rf /var/lib/apt/lists/*

EXPOSE 162/UDP

WORKDIR /var/lib/zabbix/snmptraps/

VOLUME ["/var/lib/zabbix/snmptraps", "/var/lib/zabbix/mibs"]

COPY ["conf/etc/supervisor/", "/etc/supervisor/"]
COPY ["conf/etc/logrotate.d/zabbix_snmptraps", "/etc/logrotate.d/"]

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/supervisord.conf"]
