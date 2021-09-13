FROM oraclelinux:8-slim

LABEL org.opencontainers.image.title="Zabbix server (MySQL)" \
      org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.vendor="Zabbix LLC" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.description="Zabbix server with MySQL database support" \
      org.opencontainers.image.licenses="GPL v2.0"

STOPSIGNAL SIGTERM

COPY ["conf/etc/yum.repos.d/oracle-epel-ol8.repo", "/etc/yum.repos.d/oracle-epel-ol8.repo"]

RUN set -eux && \
    groupadd --system --gid 1995 zabbix && \
    useradd \
            --system --comment "Zabbix monitoring system" \
            -g zabbix -G root,dialout \
            --uid 1997 \
            --shell /sbin/nologin \
            --home-dir /var/lib/zabbix/ \
        zabbix && \
    mkdir -p /etc/zabbix && \
    mkdir -p /var/lib/zabbix && \
    mkdir -p /var/lib/zabbix/enc && \
    mkdir -p /var/lib/zabbix/export && \
    mkdir -p /var/lib/zabbix/mibs && \
    mkdir -p /var/lib/zabbix/modules && \
    mkdir -p /var/lib/zabbix/snmptraps && \
    mkdir -p /var/lib/zabbix/ssh_keys && \
    mkdir -p /var/lib/zabbix/ssl && \
    mkdir -p /var/lib/zabbix/ssl/certs && \
    mkdir -p /var/lib/zabbix/ssl/keys && \
    mkdir -p /var/lib/zabbix/ssl/ssl_ca && \
    mkdir -p /usr/lib/zabbix/alertscripts && \
    mkdir -p /usr/lib/zabbix/externalscripts && \
    mkdir -p /usr/share/doc/zabbix-server-mysql && \
    microdnf -y module enable mysql && \
    INSTALL_PKGS="fping \
            tini \
            file-libs \
            tzdata \
            iputils \
            traceroute \
            libevent \
            libxml2 \
            libssh \
            systemd \
            mysql \
            mysql-libs \
            net-snmp-libs \
            OpenIPMI-libs \
            openldap \
            openssl-libs \
            pcre \
            zlib \
            unixODBC" && \
    microdnf -y --disablerepo="*" --enablerepo="ol8_baseos_latest" \
        --enablerepo="ol8_appstream" \
        --enablerepo="ol8_developer_EPEL" \
        install --setopt=install_weak_deps=0 --best --nodocs ${INSTALL_PKGS} && \
    microdnf -y clean all && \
    rm -rf /var/cache/yum /var/lib/yum/yumdb/* /usr/lib/udev/hwdb.d/* && \
    rm -rf /var/cache/dnf /etc/udev/hwdb.bin /root/.pki

ARG MAJOR_VERSION=5.4
ARG ZBX_VERSION=${MAJOR_VERSION}.4
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm MIBDIRS=/usr/share/snmp/mibs:/var/lib/zabbix/mibs MIBS=+ALL \
    ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES}

LABEL org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.version="${ZBX_VERSION}" \
      org.opencontainers.image.source="${ZBX_SOURCES}"

RUN set -eux && \
    REPOLIST="ol8_baseos_latest,ol8_appstream,ol8_codeready_builder" && \
    INSTALL_PKGS="autoconf \
            automake \
            gcc \
            pcre-devel \
            libcurl-devel \
            libevent-devel \
            libssh-devel \
            libxml2-devel \
            make \
            mysql-devel \
            net-snmp-devel \
            OpenIPMI-devel \
            openldap-devel \
            git \
            unixODBC-devel" && \
    microdnf -y install --setopt=install_weak_deps=0 --best \
        --nodocs dnf && \
    dnf -y install --disablerepo "*" --enablerepo "${REPOLIST}" --setopt=install_weak_deps=False --best \
        --setopt=tsflags=nodocs ${INSTALL_PKGS} && \
    cd /tmp/ && \
    git -c advice.detachedHead=false clone ${ZBX_SOURCES} --branch ${ZBX_VERSION} --depth 1 --single-branch zabbix-${ZBX_VERSION} && \
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
            --with-ldap \
            --with-libcurl \
            --with-libxml2 \
            --with-net-snmp \
            --with-openipmi \
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
    cat database/mysql/schema.sql > database/mysql/create.sql && \
    cat database/mysql/images.sql >> database/mysql/create.sql && \
    cat database/mysql/data.sql >> database/mysql/create.sql && \
    gzip database/mysql/create.sql && \
    cp database/mysql/create.sql.gz /usr/share/doc/zabbix-server-mysql/ && \
    cd /tmp/ && \
    rm -rf /tmp/zabbix-${ZBX_VERSION}/ && \
    chown --quiet -R zabbix:root /etc/zabbix/ /var/lib/zabbix/ && \
    chgrp -R 0 /etc/zabbix/ /var/lib/zabbix/ && \
    chmod -R g=u /etc/zabbix/ /var/lib/zabbix/ && \
    dnf -y history undo last && \
    dnf -y clean all && \
    microdnf -y remove dnf && \
    rm -rf /var/cache/yum /var/lib/yum/yumdb/* /usr/lib/udev/hwdb.d/* && \
    rm -rf /var/cache/dnf /etc/udev/hwdb.bin /root/.pki

EXPOSE 10051/TCP

WORKDIR /var/lib/zabbix

VOLUME ["/var/lib/zabbix/snmptraps", "/var/lib/zabbix/export"]

COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENTRYPOINT ["/usr/bin/tini", "--", "/usr/bin/docker-entrypoint.sh"]

USER 1997

CMD ["/usr/sbin/zabbix_server", "--foreground", "-c", "/etc/zabbix/zabbix_server.conf"]
