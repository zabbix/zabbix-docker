# syntax=docker/dockerfile:1
ARG OS_BASE_IMAGE=registry.access.redhat.com/ubi9/ubi-minimal:9.5
ARG MAJOR_VERSION=7.2
ARG RELEASE=5
ARG ZBX_VERSION=${MAJOR_VERSION}.5
ARG BUILD_BASE_IMAGE=zabbix-build-mysql:rhel-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} AS builder

FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION
ARG RELEASE
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} \
    MIBDIRS=/usr/share/snmp/mibs:/var/lib/zabbix/mibs MIBS=+ALL \
    NMAP_PRIVILEGED="" \
    ZABBIX_USER_HOME_DIR="/var/lib/zabbix" \
    ZABBIX_CONF_DIR="/etc/zabbix"

ENV ZBX_DB_NAME="dummy_db_name" \
    ZBX_FPINGLOCATION="/usr/sbin/fping" \
    ZBX_LOADMODULEPATH="${ZABBIX_USER_HOME_DIR}/modules" \
    ZBX_SNMPTRAPPERFILE="${ZABBIX_USER_HOME_DIR}/snmptraps/snmptraps.log" \
    ZBX_SSHKEYLOCATION="${ZABBIX_USER_HOME_DIR}/ssh_keys/" \
    ZBX_SSLCERTLOCATION="${ZABBIX_USER_HOME_DIR}/ssl/certs/" \
    ZBX_SSLKEYLOCATION="${ZABBIX_USER_HOME_DIR}/ssl/keys/" \
    ZBX_SSLCALOCATION="${ZABBIX_USER_HOME_DIR}/ssl/ssl_ca/"

LABEL description="Zabbix proxy with MySQL database support" \
      maintainer="alexey.pustovalov@zabbix.com" \
      name="zabbix/zabbix-proxy-mysql-72" \
      release="${RELEASE}" \
      run="docker run --name zabbix-proxy --link mysql-server:mysql-server -p 10051:10051 -d registry.connect.redhat.com/zabbix/zabbix-proxy-mysql-72:${ZBX_VERSION}" \
      summary="Zabbix proxy (MySQL)" \
      url="https://www.zabbix.com/" \
      vendor="Zabbix SIA" \
      version="${MAJOR_VERSION}" \
      io.k8s.description="Zabbix proxy with MySQL database support" \
      io.k8s.display-name="Zabbix proxy (MySQL)" \
      io.openshift.expose-services="10051:10051" \
      io.openshift.tags="zabbix,zabbix-proxy,mysql" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.description="Zabbix proxy with MySQL database support" \
      org.label-schema.docker.cmd="docker run --name zabbix-proxy --link mysql-server:mysql-server -p 10051:10051 -d registry.connect.redhat.com/zabbix/zabbix-proxy-mysql-72:${ZBX_VERSION}" \
      org.label-schema.license="AGPL v3.0" \
      org.label-schema.name="zabbix-proxy-mysql-rhel" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.vcs-url="${ZBX_SOURCES}" \
      org.label-schema.vendor="Zabbix SIA" \
      org.label-schema.version="${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY ["licenses", "/licenses"]
COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}-output/proxy/sbin/zabbix_proxy", "/usr/sbin/zabbix_proxy"]
COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}-output/general/bin/*", "/usr/bin/"]
COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}-output/proxy/conf/", "${ZABBIX_CONF_DIR}/"]
COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}-output/proxy/database/mysql/", "/usr/share/doc/zabbix-proxy-mysql/"]

RUN --mount=type=tmpfs,target=/var/lib/dnf/ \
    --mount=type=tmpfs,target=/var/cache/yum/ \
#    --mount=type=bind,target=/run/secrets/,src=secrets/ \
    set -eux && \
    INSTALL_PKGS="bash \
            traceroute \
            nmap \
            krb5-workstation \
            shadow-utils \
            fping \
            iputils \
            libcurl-minimal \
            libevent \
            openldap \
            openssl \
            libssh \
            libxml2 \
            mysql \
            mysql-libs \
            net-snmp-agent-libs \
            OpenIPMI-libs \
            pcre2 \
            gzip \
            unixODBC \
            zlib" && \
    curl --tlsv1.2 -sSf -L https://dl.fedoraproject.org/pub/epel/epel-release-latest-9.noarch.rpm -o /tmp/epel-release-latest-9.noarch.rpm && \
    rpm -ivh /tmp/epel-release-latest-9.noarch.rpm && \
    rm -rf /tmp/epel-release-latest-9.noarch.rpm && \
    ARCH_SUFFIX="$(arch)"; \
    microdnf -y install \
            --disablerepo "*" \
            --enablerepo "ubi-9-baseos-rpms" \
            --enablerepo "ubi-9-appstream-rpms" \
            --enablerepo "rhel-9-for-$ARCH_SUFFIX-baseos-rpms" \
            --enablerepo "rhel-9-for-$ARCH_SUFFIX-appstream-rpms" \
            --enablerepo "codeready-builder-for-rhel-9-$ARCH_SUFFIX-rpms" \
            --enablerepo "epel" \
            --setopt=install_weak_deps=0 \
            --setopt=keepcache=0 \
            --best \
            --setopt=tsflags=nodocs \
        ${INSTALL_PKGS} && \
    microdnf -y update \
            --disableplugin=subscription-manager \
            --disablerepo "*" \
            --enablerepo "ubi-9-baseos-rpms" \
            --setopt=install_weak_deps=0 \
            --best \
            --setopt=tsflags=nodocs \
        tzdata && \
    microdnf -y reinstall \
            --disableplugin=subscription-manager \
            --disablerepo "*" \
            --enablerepo "ubi-9-baseos-rpms" \
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
    mkdir -p ${ZABBIX_CONF_DIR}/ && \
    mkdir -p ${ZABBIX_USER_HOME_DIR} && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/enc && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/enc_internal && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/mibs && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/modules && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/snmptraps && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/ssh_keys && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/ssl && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/ssl/certs && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/ssl/keys && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/ssl/ssl_ca && \
    mkdir -p /usr/lib/zabbix/externalscripts && \
    mkdir -p /usr/share/doc/zabbix-proxy-mysql && \
    chown --quiet -R zabbix:root ${ZABBIX_CONF_DIR}/zabbix_proxy_modules.conf ${ZABBIX_USER_HOME_DIR} && \
    chown --quiet zabbix:root ${ZABBIX_CONF_DIR}/ && \
    chgrp -R 0 ${ZABBIX_CONF_DIR}/zabbix_proxy_modules.conf ${ZABBIX_USER_HOME_DIR} && \
    chgrp 0 ${ZABBIX_CONF_DIR} && \
    chmod -R g=u ${ZABBIX_CONF_DIR}/zabbix_proxy_modules.conf ${ZABBIX_USER_HOME_DIR} && \
    chmod g=u ${ZABBIX_CONF_DIR} && \
    /usr/sbin/zabbix_proxy -V


EXPOSE 10051/TCP

WORKDIR ${ZABBIX_USER_HOME_DIR}

VOLUME ["${ZABBIX_USER_HOME_DIR}/snmptraps"]

COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENTRYPOINT ["/usr/bin/docker-entrypoint.sh"]

USER 1997

CMD ["/usr/sbin/zabbix_proxy", "--foreground", "-c", "/etc/zabbix/zabbix_proxy.conf"]
