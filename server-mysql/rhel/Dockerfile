FROM registry.access.redhat.com/ubi8/ubi
MAINTAINER Alexey Pustovalov <alexey.pustovalov@zabbix.com>

ARG DNF_FLAGS_COMMON="-y"
ARG DNF_FLAGS_PERSISTENT="${DNF_FLAGS_COMMON}"
ARG DNF_FLAGS_DEV="${DNF_FLAGS_COMMON}"

ARG BUILD_DATE
ARG VCS_REF

ARG MAJOR_VERSION=4.4
ARG RELEASE=7
ARG ZBX_VERSION=${MAJOR_VERSION}.7

ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ENV TERM=xterm ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} TINI_VERSION=v0.18.0 \
    MIBDIRS=/usr/share/snmp/mibs:/var/lib/zabbix/mibs MIBS=+ALL

LABEL name="zabbix/zabbix-server-mysql-44" \
      maintainer="alexey.pustovalov@zabbix.com" \
      vendor="Zabbix LLC" \
      version="${MAJOR_VERSION}" \
      release="${RELEASE}" \
      summary="Zabbix server (MySQL)" \
      description="Zabbix server with MySQL database support" \
      url="https://www.zabbix.com/" \
      run="docker run --name zabbix-server --link mysql-server:mysql-server -p 10051:10051 -d registry.connect.redhat.com/zabbix/zabbix-server-mysql-44:${ZBX_VERSION}" \
      io.k8s.description="Zabbix server with MySQL database support" \
      io.k8s.display-name="Zabbix server (MySQL)" \
      io.openshift.expose-services="10051:10051" \
      io.openshift.tags="zabbix,zabbix-server,mysql" \
      org.label-schema.name="zabbix-server-mysql-rhel" \
      org.label-schema.vendor="Zabbix LLC" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.description="Zabbix server with MySQL database support" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.license="GPL v2.0" \
      org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.version="${ZBX_VERSION}" \
      org.label-schema.vcs-url="${ZBX_SOURCES}" \
      org.label-schema.docker.cmd="docker run --name zabbix-server --link mysql-server:mysql-server -p 10051:10051 -d registry.connect.redhat.com/zabbix/zabbix-server-mysql-44:${ZBX_VERSION}"

STOPSIGNAL SIGTERM

### add licenses to this directory
COPY ["licenses", "/licenses"]
COPY ["conf/sbin/tini","/sbin/tini"]

RUN set -eux && INSTALL_PKGS="bash \
            fping \
            iputils \
            libcurl \
            libevent \
#            libldap \
            libssh \
            libxml2 \
            mariadb \
            net-snmp-agent-libs \
            OpenIPMI-libs \
            pcre \
            unixODBC" && \
    dnf ${DNF_FLAGS_COMMON} install https://repo.zabbix.com/zabbix/${MAJOR_VERSION}/rhel/8/x86_64/zabbix-release-${MAJOR_VERSION}-1.el8.noarch.rpm && \
    REPOLIST="rhel-8-for-x86_64-baseos-rpms,rhel-8-for-x86_64-appstream-rpms,zabbix-non-supported" && \
    dnf ${DNF_FLAGS_COMMON} update-minimal --disablerepo "*" --enablerepo rhel-8-for-x86_64-baseos-rpms --setopt=tsflags=nodocs \
        --security --sec-severity=Important --sec-severity=Critical && \
    dnf ${DNF_FLAGS_COMMON} install --disablerepo "*" --enablerepo "${REPOLIST}" --setopt=tsflags=nodocs ${INSTALL_PKGS} && \
    groupadd -g 1995 --system zabbix && \
    adduser -r --shell /sbin/nologin \
            -g zabbix -G dialout -G root \
            -d /var/lib/zabbix/ -u 1997 \
        zabbix && \
    mkdir -p /etc/zabbix && \
    mkdir -p /var/lib/zabbix && \
    mkdir -p /usr/lib/zabbix/alertscripts && \
    mkdir -p /var/lib/zabbix/enc && \
    mkdir -p /usr/lib/zabbix/externalscripts && \
    mkdir -p /var/lib/zabbix/mibs && \
    mkdir -p /var/lib/zabbix/modules && \
    mkdir -p /var/lib/zabbix/snmptraps && \
    mkdir -p /var/lib/zabbix/ssh_keys && \
    mkdir -p /var/lib/zabbix/ssl && \
    mkdir -p /var/lib/zabbix/ssl/certs && \
    mkdir -p /var/lib/zabbix/ssl/keys && \
    mkdir -p /var/lib/zabbix/ssl/ssl_ca && \
    mkdir -p /usr/share/doc/zabbix-server-mysql && \
    chmod +x /sbin/tini && \
    dnf ${DNF_FLAGS_COMMON} clean all && \
    rm -rf /var/cache/yum /var/lib/yum/yumdb/* /usr/lib/udev/hwdb.d/* && \
    rm -rf /var/cache/dnf /etc/udev/hwdb.bin /root/.pki

RUN set -eux && INSTALL_PKGS="autoconf \
            automake \
            libcurl-devel \
            libevent-devel \
            libssh-devel \
            libxml2-devel \
            mariadb-connector-c-devel \
            net-snmp-devel \
            openldap-devel \
            pcre-devel \
            git \
            gcc \
            make \
            unixODBC-devel" && \
    REPOLIST="rhel-8-for-x86_64-baseos-rpms,rhel-8-for-x86_64-appstream-rpms,zabbix-non-supported" && \
    dnf ${DNF_FLAGS_COMMON} install --disablerepo "*" --enablerepo "${REPOLIST}" --setopt=tsflags=nodocs ${INSTALL_PKGS} && \
    cd /tmp/ && \
    git clone ${ZBX_SOURCES} --branch ${ZBX_VERSION} --depth 1 --single-branch zabbix-${ZBX_VERSION} && \
    cd /tmp/zabbix-${ZBX_VERSION} && \
    zabbix_revision=`git rev-parse --short HEAD` && \
    sed -i "s/{ZABBIX_REVISION}/$zabbix_revision/g" include/version.h && \
    ./bootstrap.sh && \
    export CFLAGS="-fPIC -pie -Wl,-z,relro -Wl,-z,now" && \
    ./configure \
            --datadir=/usr/lib \
            --libdir=/usr/lib/zabbix \
            --prefix=/usr \
            --sysconfdir=/etc/zabbix \
            --enable-agent \
            --enable-server \
            --with-mysql \
# Does not support stable iksemel library
#            --with-jabber \
            --with-ldap \
            --with-libcurl \
            --with-libxml2 \
            --with-net-snmp \
#            --with-openipmi \
            --with-openssl \
            --with-ssh \
            --with-unixodbc \
            --enable-ipv6 \
            --silent && \
    make -j"$(nproc)" -s dbschema && \
    make -j"$(nproc)" -s && \
    cp src/zabbix_server/zabbix_server /usr/sbin/zabbix_server && \
    cp src/zabbix_get/zabbix_get /usr/bin/zabbix_get && \
    cp src/zabbix_sender/zabbix_sender /usr/bin/zabbix_sender && \
    cp conf/zabbix_server.conf /etc/zabbix/zabbix_server.conf && \
    chown --quiet -R zabbix:root /etc/zabbix && \
    cat database/mysql/schema.sql > database/mysql/create.sql && \
    cat database/mysql/images.sql >> database/mysql/create.sql && \
    cat database/mysql/data.sql >> database/mysql/create.sql && \
    gzip database/mysql/create.sql && \
    cp database/mysql/create.sql.gz /usr/share/doc/zabbix-server-mysql/ && \
    cd /tmp/ && \
    rm -rf /tmp/zabbix-${ZBX_VERSION}/ && \
    dnf ${DNF_FLAGS_COMMON} history undo `dnf -q history | sed -n 3p |column -t | cut -d' ' -f1` && \
    dnf ${DNF_FLAGS_COMMON} clean all && \
    chown --quiet -R zabbix:root /etc/zabbix/ /var/lib/zabbix/ && \
    chgrp -R 0 /etc/zabbix/ /var/lib/zabbix/ && \
    chmod -R g=u /etc/zabbix/ /var/lib/zabbix/ && \
    rm -rf /var/cache/yum /var/lib/yum/yumdb/* /usr/lib/udev/hwdb.d/* && \
    rm -rf /var/cache/dnf /etc/udev/hwdb.bin /root/.pki

EXPOSE 10051/TCP

WORKDIR /var/lib/zabbix

VOLUME ["/usr/lib/zabbix/alertscripts", "/usr/lib/zabbix/externalscripts", "/var/lib/zabbix/enc", "/var/lib/zabbix/mibs", "/var/lib/zabbix/modules"]
VOLUME ["/var/lib/zabbix/snmptraps", "/var/lib/zabbix/ssh_keys", "/var/lib/zabbix/ssl/certs", "/var/lib/zabbix/ssl/keys", "/var/lib/zabbix/ssl/ssl_ca"]
VOLUME ["/var/lib/zabbix/export"]

COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENTRYPOINT ["/sbin/tini", "--", "/usr/bin/docker-entrypoint.sh"]

USER 1997

CMD ["/usr/sbin/zabbix_server", "--foreground", "-c", "/etc/zabbix/zabbix_server.conf"]
