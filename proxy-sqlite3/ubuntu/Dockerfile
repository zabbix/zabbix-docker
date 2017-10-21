FROM ubuntu:trusty
LABEL maintainer "Alexey Pustovalov <alexey.pustovalov@zabbix.com>"

ARG APT_FLAGS_COMMON="-qq -y"
ARG APT_FLAGS_PERSISTANT="${APT_FLAGS_COMMON} --no-install-recommends"
ARG APT_FLAGS_DEV="${APT_FLAGS_COMMON} --no-install-recommends"
ARG DB_TYPE=sqlite3
ENV LANG=en_US.UTF-8 LANGUAGE=en_US:en LC_ALL=en_US.UTF-8 DEBIAN_FRONTEND=noninteractive TERM=xterm
ENV MIBDIRS=/usr/share/snmp/mibs:/var/lib/zabbix/mibs MIBS=+ALL

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
    mkdir -p /var/lib/zabbix/mibs && \
    mkdir -p /var/lib/zabbix/snmptraps && \
    chown --quiet -R zabbix:root /var/lib/zabbix && \
    mkdir -p /usr/lib/zabbix/externalscripts && \
    apt-get ${APT_FLAGS_COMMON} update && \
    apt-get ${APT_FLAGS_PERSISTANT} install \
            supervisor \
            libsqlite3-0 \
            libiksemel3 \
            libsnmp30 \
            libcurl3 \
            unixodbc \
            libssh2-1 \
            libssl1.0.0 \
            libxml2 \
            fping \
            libevent-2.0-5 \
            libpcre3 \
            libopenipmi0 1>/dev/null && \
    apt-get ${APT_FLAGS_COMMON} autoremove && \
    apt-get ${APT_FLAGS_COMMON} clean && \
    rm -rf /var/lib/apt/lists/*

ARG MAJOR_VERSION=3.4
ARG ZBX_VERSION=${MAJOR_VERSION}.3
ARG ZBX_SOURCES=svn://svn.zabbix.com/tags/${ZBX_VERSION}/
ENV ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} DB_TYPE=${DB_TYPE}

RUN apt-get ${APT_FLAGS_COMMON} update && \
    apt-get ${APT_FLAGS_DEV} install \
            gcc \
            make \
            automake \
            libc6-dev \
            pkg-config \
            libsqlite3-dev \
            libsnmp-dev \
            libcurl4-openssl-dev \
            libldap2-dev \
            libiksemel-dev \
            libopenipmi-dev \
            libssh2-1-dev \
            unixodbc-dev \
            libxml2-dev \
            libevent-dev \
            libpcre3-dev \
            subversion 1>/dev/null && \
    cd /tmp/ && \
    svn --quiet export ${ZBX_SOURCES} zabbix-${ZBX_VERSION} && \
    cd /tmp/zabbix-${ZBX_VERSION} && \
    zabbix_revision=`svn info ${ZBX_SOURCES} |grep "Last Changed Rev"|awk '{print $4;}'` && \
    sed -i "s/{ZABBIX_REVISION}/$zabbix_revision/g" include/version.h && \
    ./bootstrap.sh 1>/dev/null && \
    export CFLAGS="-fPIC -pie -Wl,-z,relro -Wl,-z,now" && \
    ./configure \
            --prefix=/usr \
            --silent \
            --sysconfdir=/etc/zabbix \
            --libdir=/usr/lib/zabbix \
            --datadir=/usr/lib \
            --enable-proxy \
            --enable-agent \
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
    cp src/zabbix_proxy/zabbix_proxy /usr/sbin/zabbix_proxy && \
    cp src/zabbix_get/zabbix_get /usr/bin/zabbix_get && \
    cp src/zabbix_sender/zabbix_sender /usr/bin/zabbix_sender && \
    cp conf/zabbix_proxy.conf /etc/zabbix/zabbix_proxy.conf && \
    chown --quiet -R zabbix:root /etc/zabbix && \
    cd /tmp/ && \
    rm -rf /tmp/zabbix-${ZBX_VERSION}/ && \
    apt-get ${APT_FLAGS_COMMON} purge \
            gcc \
            make \
            automake \
            libc6-dev \
            pkg-config \
            libsqlite3-dev \
            libsnmp-dev \
            libcurl4-openssl-dev \
            libldap2-dev \
            libiksemel-dev \
            libopenipmi-dev \
            libssh2-1-dev \
            unixodbc-dev \
            libxml2-dev \
            libevent-dev \
            libpcre3-dev \
            subversion 1>/dev/null && \
    apt-get ${APT_FLAGS_COMMON} autoremove 1>/dev/null && \
    rm -rf /var/lib/apt/lists/*

EXPOSE 10051/TCP 162/UDP

WORKDIR /var/lib/zabbix

VOLUME ["/usr/lib/zabbix/externalscripts", "/var/lib/zabbix/enc", "/var/lib/zabbix/modules", "/var/lib/zabbix/ssh_keys"]
VOLUME ["/var/lib/zabbix/ssl/certs", "/var/lib/zabbix/ssl/keys", "/var/lib/zabbix/ssl/ssl_ca", "/var/lib/zabbix/snmptraps", "/var/lib/zabbix/mibs"]

ADD conf/etc/supervisor/ /etc/supervisor/
ADD run_zabbix_component.sh /

ENTRYPOINT ["/bin/bash"]

CMD ["/run_zabbix_component.sh", "proxy", "sqlite3"]
