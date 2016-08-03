FROM ubuntu:trusty
MAINTAINER Alexey Pustovalov <alexey.pustovalov@zabbix.com>

ARG APT_FLAGS_COMMON="-qq -y"
ARG APT_FLAGS_PERSISTANT="${APT_FLAGS_COMMON} --no-install-recommends"
ARG APT_FLAGS_DEV="${APT_FLAGS_COMMON} --no-install-recommends"
ARG DB_TYPE=postgresql
ENV LANG=en_US.UTF-8 LANGUAGE=en_US:en LC_ALL=en_US.UTF-8 DEBIAN_FRONTEND=noninteractive TERM=xterm
ENV MIBDIRS=/var/lib/zabbix/mibs MIBS=+ALL

RUN locale-gen $LC_ALL && \
    echo "#!/bin/sh\nexit 0" > /usr/sbin/policy-rc.d && \
    addgroup --system --quiet zabbix && \
    adduser --quiet \
            --system --disabled-login \
            --ingroup zabbix \
            --home /var/lib/zabbix/ \
        zabbix && \
    mkdir -p /etc/zabbix/ && \
    mkdir -p /var/lib/zabbix && \
    mkdir -p /var/lib/zabbix/enc && \
    mkdir -p /var/lib/zabbix/modules && \
    mkdir -p /var/lib/zabbix/ssh_keys && \
    mkdir -p /var/lib/zabbix/ssl && \
    mkdir -p /var/lib/zabbix/ssl/certs && \
    mkdir -p /var/lib/zabbix/ssl/keys && \
    mkdir -p /var/lib/zabbix/ssl/ssl_ca && \
    mkdir -p /usr/lib/zabbix/externalscripts && \
    mkdir -p /usr/lib/zabbix/alertscripts && \
    mkdir -p /var/lib/zabbix/mibs && \
    mkdir -p /var/lib/zabbix/snmptraps && \
    chown --quiet -R zabbix:root /var/lib/zabbix && \
    mkdir -p /usr/share/doc/zabbix-server-${DB_TYPE} && \
    apt-get ${APT_FLAGS_COMMON} update && \
    apt-get ${APT_FLAGS_PERSISTANT} install \
            supervisor \
            postgresql-client \
            libpq5 \
            libiksemel3 \
            libsnmp30 \
            libcurl3 \
            unixodbc \
            libssh2-1 \
            libssl1.0.0 \
            libxml2 \
            fping \
            libopenipmi0 1>/dev/null && \
    apt-get ${APT_FLAGS_COMMON} autoremove && \
    apt-get ${APT_FLAGS_COMMON} clean && \
    rm -rf /var/lib/apt/lists/*

ARG MAJOR_VERSION=master
ARG ZBX_VERSION=${MAJOR_VERSION}
ARG ZBX_SOURCES=svn://svn.zabbix.com/tags/${ZBX_VERSION}/
ENV ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} DB_TYPE=${DB_TYPE}

RUN apt-get ${APT_FLAGS_COMMON} update && \
    apt-get ${APT_FLAGS_DEV} install \
            gcc \
            make \
            automake \
            libc6-dev \
            pkg-config \
            libpq-dev \
            libsnmp-dev \
            libcurl4-openssl-dev \
            libldap2-dev \
            libiksemel-dev \
            libopenipmi-dev \
            libssh2-1-dev \
            unixodbc-dev \
            libxml2-dev \
            subversion 1>/dev/null && \
    cd /tmp/ && \
    svn --quiet export ${ZBX_SOURCES} zabbix-${ZBX_VERSION} && \
    cd /tmp/zabbix-${ZBX_VERSION} && \
    zabbix_revision=`svn info ${ZBX_SOURCES} |grep "Last Changed Rev"|awk '{print $4;}'` && \
    sed -i "s/{ZABBIX_REVISION}/$zabbix_revision/g" include/version.h && \
    ./bootstrap.sh 1>/dev/null && \
    ./configure \
            --prefix=/usr \
            --silent \
            --sysconfdir=/etc/zabbix \
            --libdir=/usr/lib/zabbix \
            --enable-server \
            --enable-ipv6 \
            --with-jabber \
            --with-ldap \
            --with-net-snmp \
            --with-openipmi \
            --with-ssh2 \
            --with-libcurl \
            --with-unixodbc \
            --with-libxml2 \
            --with-openssl \
            --with-${DB_TYPE} && \
    make -j"$(nproc)" -s dbschema 1>/dev/null && \
    make -j"$(nproc)" -s 1>/dev/null && \
    cp src/zabbix_server/zabbix_server /usr/sbin/zabbix_server && \
    cp conf/zabbix_server.conf /etc/zabbix/zabbix_server.conf && \
    chown --quiet -R zabbix:root /etc/zabbix && \
    cp database/${DB_TYPE}/schema.sql /usr/share/doc/zabbix-server-${DB_TYPE}/ && \
    cp database/${DB_TYPE}/images.sql /usr/share/doc/zabbix-server-${DB_TYPE}/ && \
    cp database/${DB_TYPE}/data.sql /usr/share/doc/zabbix-server-${DB_TYPE}/ && \
    cd /tmp/ && \
    rm -rf /tmp/zabbix-${ZBX_VERSION}/ && \
    apt-get ${APT_FLAGS_COMMON} purge \
            gcc \
            make \
            automake \
            libc6-dev \
            pkg-config \
            libpq-dev \
            libsnmp-dev \
            libcurl4-openssl-dev \
            libldap2-dev \
            libiksemel-dev \
            libopenipmi-dev \
            libssh2-1-dev \
            unixodbc-dev \
            libxml2-dev \
            subversion 1>/dev/null && \
    apt-get ${APT_FLAGS_COMMON} autoremove 1>/dev/null && \
    rm -rf /var/lib/apt/lists/*

EXPOSE 10051/TCP 162/UDP

WORKDIR /var/lib/zabbix

VOLUME ["/usr/lib/zabbix/alertscripts", "/usr/lib/zabbix/externalscripts", "/var/lib/zabbix/enc", "/var/lib/zabbix/modules", "/var/lib/zabbix/ssh_keys"]
VOLUME ["/var/lib/zabbix/ssl/certs", "/var/lib/zabbix/ssl/keys", "/var/lib/zabbix/ssl/ssl_ca", "/var/lib/zabbix/snmptraps", "/var/lib/zabbix/mibs"]
    
ADD conf/etc/supervisor/ /etc/supervisor/
ADD conf/etc/logrotate.d/logrotate_zabbix.conf /etc/logrotate.d/
ADD run_zabbix_component.sh /data/

ENTRYPOINT ["/bin/bash"]

CMD ["/data/run_zabbix_component.sh", "server", "postgresql"]
