# syntax=docker/dockerfile:1.6
ARG OS_BASE_IMAGE=registry.access.redhat.com/ubi10/ubi-minimal:10.2
ARG MAJOR_VERSION=7.4
ARG RELEASE=12
ARG ZBX_VERSION=${MAJOR_VERSION}.12
ARG BUILD_BASE_IMAGE=zabbix-build-sqlite3:rhel-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} AS builder

FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION
ARG RELEASE
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
    ZBX_SOURCES=${ZBX_SOURCES} \
    MIBDIRS=/usr/share/snmp/mibs:/var/lib/zabbix/mibs MIBS=+ALL \
    NMAP_PRIVILEGED= \
    ZABBIX_USER_HOME_DIR=/var/lib/zabbix \
    ZABBIX_CONF_DIR=/etc/zabbix

ENV ZBX_FPINGLOCATION=/usr/sbin/fping \
    ZBX_LOADMODULEPATH=${ZABBIX_USER_HOME_DIR}/modules \
    ZBX_SNMPTRAPPERFILE=${ZABBIX_USER_HOME_DIR}/snmptraps/snmptraps.log \
    ZBX_SSHKEYLOCATION=${ZABBIX_USER_HOME_DIR}/ssh_keys/ \
    ZBX_SSLCERTLOCATION=${ZABBIX_USER_HOME_DIR}/ssl/certs/ \
    ZBX_SSLKEYLOCATION=${ZABBIX_USER_HOME_DIR}/ssl/keys/ \
    ZBX_SSLCALOCATION=${ZABBIX_USER_HOME_DIR}/ssl/ssl_ca/

LABEL description="Zabbix proxy with SQLite3 database support" \
      maintainer="alexey.pustovalov@zabbix.com" \
      name="zabbix/zabbix-proxy-sqlite-74" \
      release="${RELEASE}" \
      run="docker run --name zabbix-proxy -p 10051:10051 -d registry.connect.redhat.com/zabbix/zabbix-proxy-sqlite-74:${ZBX_VERSION}" \
      summary="Zabbix proxy (SQLite3)" \
      url="https://www.zabbix.com/" \
      vendor="Zabbix SIA" \
      version="${MAJOR_VERSION}" \
      io.k8s.description="Zabbix proxy with SQLite3 database support" \
      io.k8s.display-name="Zabbix proxy (SQLite3)" \
      io.openshift.expose-services="10051:10051" \
      io.openshift.tags="zabbix,zabbix-proxy,sqlite3" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.description="Zabbix proxy with SQLite3 database support" \
      org.label-schema.docker.cmd="docker run --name zabbix-proxy -p 10051:10051 -d registry.connect.redhat.com/zabbix/zabbix-proxy-sqlite-74:${ZBX_VERSION}" \
      org.label-schema.license="AGPL v3.0" \
      org.label-schema.name="zabbix-proxy-sqlite3-rhel" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.vcs-url="${ZBX_SOURCES}" \
      org.label-schema.vendor="Zabbix SIA" \
      org.label-schema.version="${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY licenses /licenses/
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/proxy/sbin/zabbix_proxy /usr/sbin/zabbix_proxy
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/general/bin/zabbix_get /usr/bin/
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/general/bin/zabbix_sender /usr/bin/
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/proxy/conf/ ${ZABBIX_CONF_DIR}/
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/proxy/database/sqlite3/ /usr/share/doc/zabbix-proxy-sqlite3/

RUN --mount=type=tmpfs,target=/var/lib/dnf/ \
    --mount=type=tmpfs,target=/var/cache/yum/ \
    set -eux && \
    INSTALL_PKGS="bash \
            traceroute \
            nmap \
            krb5-workstation \
            c-ares \
            shadow-utils \
            fping \
            iputils \
            libcurl-minimal \
            libevent \
            openldap \
            openssl \
            libssh \
            libxml2 \
            net-snmp-agent-libs \
            OpenIPMI-libs \
            pcre2 \
            sqlite \
            sqlite-libs \
            unixODBC \
            gzip \
            zlib" && \
    curl --tlsv1.2 -sSf -L https://dl.fedoraproject.org/pub/epel/epel-release-latest-10.noarch.rpm -o /tmp/epel-release-latest-10.noarch.rpm && \
    rpm -ivh /tmp/epel-release-latest-10.noarch.rpm && \
    rm -rf /tmp/epel-release-latest-10.noarch.rpm && \
    ARCH_SUFFIX="$(arch)"; \
    microdnf -y install \
            --disablerepo="*" \
            --enablerepo="ubi-10-baseos-rpms" \
            --enablerepo="ubi-10-appstream-rpms" \
            --enablerepo="rhel-10-for-$ARCH_SUFFIX-baseos-rpms" \
            --enablerepo="rhel-10-for-$ARCH_SUFFIX-appstream-rpms" \
            --enablerepo="epel" \
            --setopt=install_weak_deps=0 \
            --setopt=keepcache=0 \
            --best \
            --setopt=tsflags=nodocs \
        ${INSTALL_PKGS} && \
    microdnf -y update \
            --disableplugin=subscription-manager \
            --disablerepo="*" \
            --enablerepo="ubi-10-baseos-rpms" \
            --setopt=install_weak_deps=0 \
            --best \
            --setopt=tsflags=nodocs \
        tzdata && \
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
    chgrp zabbix /usr/bin/nmap && \
    setcap cap_net_raw+eip /usr/bin/nmap && \
    mkdir -p ${ZABBIX_CONF_DIR} \
        ${ZABBIX_USER_HOME_DIR} \
        ${ZABBIX_USER_HOME_DIR}/db_data \
        ${ZABBIX_USER_HOME_DIR}/enc \
        ${ZABBIX_USER_HOME_DIR}/enc_internal \
        ${ZABBIX_USER_HOME_DIR}/mibs \
        ${ZABBIX_USER_HOME_DIR}/modules \
        ${ZABBIX_USER_HOME_DIR}/snmptraps \
        ${ZABBIX_USER_HOME_DIR}/ssh_keys \
        ${ZABBIX_USER_HOME_DIR}/ssl \
        ${ZABBIX_USER_HOME_DIR}/ssl/certs \
        ${ZABBIX_USER_HOME_DIR}/ssl/keys \
        ${ZABBIX_USER_HOME_DIR}/ssl/ssl_ca \
        /usr/lib/zabbix/externalscripts && \
    FIX_PERM_LIST="\
        ${ZABBIX_CONF_DIR}/zabbix_proxy_modules.conf \
        ${ZABBIX_USER_HOME_DIR}/" && \
    chown --quiet -R zabbix:0 ${FIX_PERM_LIST} && \
    chmod -R g=u ${FIX_PERM_LIST} && \
    chown --quiet zabbix:0 ${ZABBIX_CONF_DIR}/ && \
    chmod g=u ${ZABBIX_CONF_DIR} && \
    /usr/sbin/zabbix_proxy -V

RUN --mount=from=entrypoints,target=/tmp/entrypoints,ro \
    set -eux && \
    install -d /usr/lib/docker-entrypoint && \
    install -m 0755 /tmp/entrypoints/proxy-sqlite3.sh /usr/bin/docker-entrypoint.sh && \
    install -m 0644 /tmp/entrypoints/lib/*.sh /usr/lib/docker-entrypoint/

EXPOSE 10051/tcp

WORKDIR ${ZABBIX_USER_HOME_DIR}

VOLUME ["${ZABBIX_USER_HOME_DIR}/snmptraps"]

USER 1997

ENTRYPOINT ["/usr/bin/docker-entrypoint.sh"]

CMD ["/usr/sbin/zabbix_proxy", "--foreground", "-c", "/etc/zabbix/zabbix_proxy.conf"]
