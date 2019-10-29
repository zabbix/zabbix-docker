FROM ubuntu:bionic
LABEL maintainer "Alexey Pustovalov <alexey.pustovalov@zabbix.com>"

ARG BUILD_DATE
ARG VCS_REF

ARG APT_FLAGS_COMMON="-y"
ARG APT_FLAGS_PERSISTENT="${APT_FLAGS_COMMON} --no-install-recommends"
ARG APT_FLAGS_DEV="${APT_FLAGS_COMMON} --no-install-recommends"
ENV LANG=en_US.UTF-8 LANGUAGE=en_US:en LC_ALL=en_US.UTF-8 TERM=xterm \
    MIBDIRS=/var/lib/snmp/mibs/ietf:/var/lib/snmp/mibs/iana:/usr/share/snmp/mibs:/var/lib/zabbix/mibs MIBS=+ALL \
    ZBX_TYPE=server ZBX_DB_TYPE=mysql ZBX_OPT_TYPE=none
ENV TINI_VERSION v0.18.0

LABEL org.label-schema.name="zabbix-${ZBX_TYPE}-${ZBX_DB_TYPE}-ubuntu" \
      org.label-schema.vendor="Zabbix LLC" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.description="Zabbix ${ZBX_TYPE} with MySQL database support" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.license="GPL v2.0"

STOPSIGNAL SIGTERM

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
    usermod -G zabbix,dialout zabbix && \
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
    chown --quiet -R zabbix:root /var/lib/zabbix && \
    mkdir -p /usr/share/doc/zabbix-${ZBX_TYPE}-${ZBX_DB_TYPE} && \
    apt-get ${APT_FLAGS_COMMON} update && \
    DEBIAN_FRONTEND=noninteractive apt-get ${APT_FLAGS_PERSISTENT} install \
            iputils-ping \
            traceroute \
            fping \
            libcurl4 \
            libevent-2.1 \
            libiksemel3 \
            libmysqlclient20 \
            libopenipmi0 \
            libpcre3 \
            libsnmp30 \
            libssh2-1 \
            libssl1.1 \
            libxml2 \
            mysql-client \
            snmp-mibs-downloader \
            ca-certificates \
            unixodbc && \
    apt-get ${APT_FLAGS_COMMON} autoremove && \
    apt-get ${APT_FLAGS_COMMON} clean && \
    rm -rf /var/lib/apt/lists/*

ARG MAJOR_VERSION=4.2
ARG ZBX_VERSION=${MAJOR_VERSION}.8
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ENV ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES}

LABEL org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.version="${ZBX_VERSION}" \
      org.label-schema.vcs-url="${ZBX_SOURCES}" \
      org.label-schema.docker.cmd="docker run --name zabbix-${ZBX_TYPE}-${ZBX_DB_TYPE} --link mysql-server:mysql -p 10051:10051 -d zabbix-${ZBX_TYPE}-${ZBX_DB_TYPE}:ubuntu-${ZBX_VERSION}"

ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /sbin/tini

RUN set -eux && \
    apt-get ${APT_FLAGS_COMMON} update && \
    DEBIAN_FRONTEND=noninteractive apt-get ${APT_FLAGS_DEV} install \
            autoconf \
            automake \
            gcc \
            libc6-dev \
            libcurl4-openssl-dev \
            libevent-dev \
            libiksemel-dev \
            libldap2-dev \
            libmysqlclient-dev \
            libopenipmi-dev \
            libpcre3-dev \
            libsnmp-dev \
            libssh2-1-dev \
            libxml2-dev \
            make \
            pkg-config \
            git \
            unixodbc-dev && \
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
            --enable-${ZBX_TYPE} \
            --with-${ZBX_DB_TYPE} \
            --with-jabber \
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
    make -j"$(nproc)" -s dbschema && \
    make -j"$(nproc)" -s && \
    cp src/zabbix_${ZBX_TYPE}/zabbix_${ZBX_TYPE} /usr/sbin/zabbix_${ZBX_TYPE} && \
    cp src/zabbix_get/zabbix_get /usr/bin/zabbix_get && \
    cp src/zabbix_sender/zabbix_sender /usr/bin/zabbix_sender && \
    cp conf/zabbix_${ZBX_TYPE}.conf /etc/zabbix/zabbix_${ZBX_TYPE}.conf && \
    chown --quiet -R zabbix:root /etc/zabbix && \
    cat database/${ZBX_DB_TYPE}/schema.sql > database/${ZBX_DB_TYPE}/create.sql && \
    cat database/${ZBX_DB_TYPE}/images.sql >> database/${ZBX_DB_TYPE}/create.sql && \
    cat database/${ZBX_DB_TYPE}/data.sql >> database/${ZBX_DB_TYPE}/create.sql && \
    gzip database/${ZBX_DB_TYPE}/create.sql && \
    cp database/${ZBX_DB_TYPE}/create.sql.gz /usr/share/doc/zabbix-${ZBX_TYPE}-${ZBX_DB_TYPE}/ && \
    cd /tmp/ && \
    rm -rf /tmp/zabbix-${ZBX_VERSION}/ && \
    DEBIAN_FRONTEND=noninteractive apt-get ${APT_FLAGS_COMMON} purge \
            autoconf \
            automake \
            gcc \
            libc6-dev \
            libcurl4-openssl-dev \
            libevent-dev \
            libiksemel-dev \
            libldap2-dev \
            libmysqlclient-dev \
            libopenipmi-dev \
            libpcre3-dev \
            libsnmp-dev \
            libssh2-1-dev \
            libxml2-dev \
            make \
            pkg-config \
            git \
            unixodbc-dev && \
    apt-get ${APT_FLAGS_COMMON} autoremove && \
    rm -rf /var/lib/apt/lists/* && \
    chmod +x /sbin/tini

EXPOSE 10051/TCP

WORKDIR /var/lib/zabbix

VOLUME ["/usr/lib/zabbix/alertscripts", "/usr/lib/zabbix/externalscripts", "/var/lib/zabbix/enc", "/var/lib/zabbix/mibs", "/var/lib/zabbix/modules"]
VOLUME ["/var/lib/zabbix/snmptraps", "/var/lib/zabbix/ssh_keys", "/var/lib/zabbix/ssl/certs", "/var/lib/zabbix/ssl/keys", "/var/lib/zabbix/ssl/ssl_ca"]
      
COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENTRYPOINT ["/sbin/tini", "--", "/usr/bin/docker-entrypoint.sh"]
