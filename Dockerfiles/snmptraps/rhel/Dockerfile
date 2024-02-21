FROM registry.access.redhat.com/ubi9/ubi-minimal:9.3

ARG MAJOR_VERSION=6.4
ARG RELEASE=11
ARG ZBX_VERSION=${MAJOR_VERSION}.11

ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} \
    MIBDIRS=/usr/share/snmp/mibs:/var/lib/zabbix/mibs MIBS=+ALL \
    ZBX_SNMP_TRAP_DATE_FORMAT=+%Y%m%d.%H%M%S ZBX_SNMP_TRAP_FORMAT="\n" \
    ZBX_SNMP_TRAP_USE_DNS=false

LABEL description="Zabbix SNMP traps receiver" \
      maintainer="alexey.pustovalov@zabbix.com" \
      name="zabbix/zabbix-snmptraps-64" \
      release="${RELEASE}" \
      run="docker run --name zabbix-snmptraps --link zabbix-server:zabbix-server -p 162:1162/udp -d registry.connect.redhat.com/zabbix/zabbix-snmptraps-64:${ZBX_VERSION}" \
      summary="Zabbix SNMP traps receiver" \
      url="https://www.zabbix.com/" \
      vendor="Zabbix LLC" \
      version="${MAJOR_VERSION}" \
      io.k8s.description="Zabbix SNMP traps receiver" \
      io.k8s.display-name="Zabbix SNMP traps receiver" \
      io.openshift.expose-services="162:1162" \
      io.openshift.tags="zabbix,zabbix-snmp,snmp-traps" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.description="Zabbix SNMP traps receiver" \
      org.label-schema.docker.cmd="docker run --name zabbix-snmptraps --link zabbix-server:zabbix-server -p 162:1162/udp -d registry.connect.redhat.com/zabbix/zabbix-snmptraps-64:${ZBX_VERSION}" \
      org.label-schema.license="GPL v2.0" \
      org.label-schema.name="zabbix-snmptraps-rhel" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.vcs-url="${ZBX_SOURCES}" \
      org.label-schema.vendor="Zabbix LLC" \
      org.label-schema.version="${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY ["licenses", "/licenses"]

RUN --mount=type=tmpfs,target=/var/lib/dnf/ \
    set -eux && \
    INSTALL_PKGS="bash \
            shadow-utils \
            tzdata \
            net-snmp" && \
    microdnf -y install \
            --disablerepo="*" \
            --enablerepo "ubi-9-baseos-rpms" \
            --enablerepo "ubi-9-appstream-rpms" \
            --setopt=install_weak_deps=0 \
            --setopt=keepcache=0 \
            --best \
            --setopt=tsflags=nodocs \
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
            --home-dir /var/lib/zabbix/ \
        zabbix && \
    mkdir -p /var/lib/zabbix && \
    mkdir -p /var/lib/zabbix/snmptraps && \
    mkdir -p /var/lib/zabbix/mibs && \
    touch /var/lib/net-snmp/snmptrapd.conf && \
    chown --quiet -R zabbix:root /etc/snmp/ /var/lib/zabbix/ /var/tmp/ /var/run/ && \
    chgrp -R 0 /etc/snmp/ /var/lib/zabbix/ /var/tmp/ /var/run/ && \
    chmod -R g=u /etc/snmp/ /var/lib/zabbix/ /var/tmp/ /var/run/ && \
    microdnf -y clean all

EXPOSE 1162/UDP

WORKDIR /var/lib/zabbix/snmptraps/

VOLUME ["/var/lib/zabbix/snmptraps"]

COPY ["conf/etc/logrotate.d/zabbix_snmptraps", "/etc/logrotate.d/"]
COPY ["conf/etc/snmp/snmptrapd.conf", "/etc/snmp/"]
COPY ["conf/usr/sbin/zabbix_trap_handler.sh", "/usr/sbin/"]

USER 1997

CMD ["/usr/sbin/snmptrapd", "-n", "-C", "-c", "/etc/snmp/snmptrapd.conf", "-Lo", "-A"]
