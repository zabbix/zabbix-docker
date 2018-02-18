FROM ubuntu:trusty
LABEL maintainer="Alexey Pustovalov <alexey.pustovalov@zabbix.com>"

ARG BUILD_DATE
ARG VCS_REF

ARG APT_FLAGS_COMMON="-qq -y"
ARG APT_FLAGS_PERSISTANT="${APT_FLAGS_COMMON} --no-install-recommends"
ARG APT_FLAGS_DEV="${APT_FLAGS_COMMON} --no-install-recommends"
ARG DB_TYPE=sqlite3
ENV LANG=en_US.UTF-8 LANGUAGE=en_US:en LC_ALL=en_US.UTF-8 DEBIAN_FRONTEND=noninteractive TERM=xterm \
    MIBDIRS=/var/lib/mibs/iana:/var/lib/mibs/ietf:/usr/share/snmp/mibs:/var/lib/zabbix/mibs MIBS=+ALL

LABEL org.label-schema.name="zabbix-proxy-${DB_TYPE}-ubuntu" \
      org.label-schema.vendor="Zabbix LLC" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.description="Zabbix proxy with SQLite3 database support" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.license="GPL v2.0"

STOPSIGNAL SIGTERM

RUN locale-gen $LC_ALL && \
    echo "#!/bin/sh\nexit 0" > /usr/sbin/policy-rc.d && \
    DISTRIB_CODENAME=$(/bin/bash -c 'source /etc/lsb-release && echo $DISTRIB_CODENAME') && \
    echo "deb http://us.archive.ubuntu.com/ubuntu/ $DISTRIB_CODENAME multiverse" >> /etc/apt/sources.list && \
    addgroup --system --quiet zabbix && \
    adduser --quiet \
            --system --disabled-login \
            --ingroup zabbix \
            --home /var/lib/zabbix/ \
        zabbix && \
    mkdir -p /etc/zabbix && \
    mkdir -p /var/lib/zabbix && \
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
    chown --quiet -R zabbix:root /var/lib/zabbix && \
    apt-get ${APT_FLAGS_COMMON} update && \
    apt-get ${APT_FLAGS_PERSISTANT} install \
            fping \
            libcurl3 \
            libevent-2.0 \
            libopenipmi0 \
            libpcre3 \
            libsnmp30 \
            libsqlite3-0 \
            libssh2-1 \
            libssl1.0.0 \
            libxml2 \
            snmp-mibs-downloader \
            supervisor \
            unixodbc 1>/dev/null && \
    apt-get ${APT_FLAGS_COMMON} autoremove && \
    apt-get ${APT_FLAGS_COMMON} clean && \
    rm -rf /var/lib/apt/lists/*

ARG MAJOR_VERSION=3.4
ARG ZBX_VERSION=${MAJOR_VERSION}.6
ARG ZBX_SOURCES=svn://svn.zabbix.com/tags/${ZBX_VERSION}/
ENV ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} DB_TYPE=${DB_TYPE}

LABEL org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.version="${ZBX_VERSION}" \
      org.label-schema.vcs-url="${ZBX_SOURCES}" \
      org.label-schema.docker.cmd="docker run --name zabbix-proxy-${DB_TYPE} --link zabbix-server:zabbix-server -p 10051:10051 -d zabbix-proxy-${DB_TYPE}:ubuntu-${ZBX_VERSION}"

RUN apt-get ${APT_FLAGS_COMMON} update && \
    apt-get ${APT_FLAGS_DEV} install \
            autoconf \
            automake \
            gcc \
            libc6-dev \
            libcurl4-openssl-dev \
            libevent-dev \
            libldap2-dev \
            libopenipmi-dev \
            libpcre3-dev \
            libsnmp-dev \
            libsqlite3-dev \
            libssh2-1-dev \
            libxml2-dev \
            make \
            pkg-config \
            subversion \
            unixodbc-dev 1>/dev/null && \
    cd /tmp/ && \
    svn --quiet export ${ZBX_SOURCES} zabbix-${ZBX_VERSION} && \
    cd /tmp/zabbix-${ZBX_VERSION} && \
    zabbix_revision=`svn info ${ZBX_SOURCES} |grep "Last Changed Rev"|awk '{print $4;}'` && \
    sed -i "s/{ZABBIX_REVISION}/$zabbix_revision/g" include/version.h && \
    ./bootstrap.sh 1>/dev/null && \
    export CFLAGS="-fPIC -pie -Wl,-z,relro -Wl,-z,now" && \
    ./configure \
            --datadir=/usr/lib \
            --libdir=/usr/lib/zabbix \
            --sysconfdir=/etc/zabbix \
            --prefix=/usr \
            --enable-agent \
            --enable-proxy \
            --with-${DB_TYPE} \
            --with-ldap \
            --with-libcurl \
            --with-libxml2 \
            --with-net-snmp \
            --with-openipmi \
            --with-openssl \
            --with-ssh2 \
            --with-unixodbc \
            --enable-ipv6 \
            --silent && \
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
            autoconf \
            automake \
            gcc \
            libc6-dev \
            libcurl4-openssl-dev \
            libevent-dev \
            libldap2-dev \
            libopenipmi-dev \
            libpcre3-dev \
            libsnmp-dev \
            libsqlite3-dev \
            libssh2-1-dev \
            libxml2-dev \
            make \
            pkg-config \
            subversion \
            unixodbc-dev 1>/dev/null && \
    apt-get ${APT_FLAGS_COMMON} autoremove 1>/dev/null && \
    rm -rf /var/lib/apt/lists/* && \
    rm -rf /root/.subversion

EXPOSE 10051/TCP

WORKDIR /var/lib/zabbix

VOLUME ["/usr/lib/zabbix/externalscripts", "/var/lib/zabbix/enc", "/var/lib/zabbix/modules", "/var/lib/zabbix/snmptraps"]
VOLUME ["/var/lib/zabbix/ssh_keys", "/var/lib/zabbix/ssl/certs", "/var/lib/zabbix/ssl/keys", "/var/lib/zabbix/ssl/ssl_ca"]
      
COPY ["conf/etc/supervisor/", "/etc/supervisor/"]
COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENTRYPOINT ["docker-entrypoint.sh"]

CMD ["proxy", "sqlite3"]
