FROM alpine:3.10
LABEL maintainer="Alexey Pustovalov <alexey.pustovalov@zabbix.com>"

ARG BUILD_DATE
ARG VCS_REF

ARG APK_FLAGS_COMMON=""
ARG APK_FLAGS_PERSISTENT="${APK_FLAGS_COMMON} --clean-protected --no-cache"
ARG APK_FLAGS_DEV="${APK_FLAGS_COMMON} --no-cache"
ENV TERM=xterm MIBDIRS=/usr/share/snmp/mibs:/var/lib/zabbix/mibs MIBS=+ALL \
    ZBX_TYPE=server ZBX_DB_TYPE=postgresql ZBX_OPT_TYPE=none

LABEL org.label-schema.name="zabbix-${ZBX_TYPE}-${ZBX_DB_TYPE}-alpine" \
      org.label-schema.vendor="Zabbix LLC" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.description="Zabbix ${ZBX_TYPE} with PostgreSQL database support" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.license="GPL v2.0"

STOPSIGNAL SIGTERM

RUN set -eux && \
    addgroup zabbix && \
    adduser -S \
            -D -G zabbix \
            -h /var/lib/zabbix/ \
        zabbix && \
    adduser zabbix dialout && \
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
    apk update && \
    apk add ${APK_FLAGS_PERSISTENT} \
            tini \
            bash \
            fping \
            iputils \
            libcurl \
            libevent \
            libldap \
            libssh2 \
            libxml2 \
            net-snmp-agent-libs \
            openipmi-libs \
            pcre \
            postgresql-client \
            postgresql-libs \
            unixodbc && \
    rm -rf /var/cache/apk/*

ARG MAJOR_VERSION=4.2
ARG ZBX_VERSION=${MAJOR_VERSION}.6
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ENV ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES}

LABEL org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.version="${ZBX_VERSION}" \
      org.label-schema.vcs-url="${ZBX_SOURCES}" \
      org.label-schema.docker.cmd="docker run --name zabbix-${ZBX_TYPE}-pgsql --link postgres-server:postgres -p 10051:10051 -d zabbix-${ZBX_TYPE}-pgsql:alpine-${ZBX_VERSION}"

RUN set -eux && \
    apk add ${APK_FLAGS_DEV} --virtual build-dependencies \
            alpine-sdk \
            autoconf \
            automake \
            coreutils \
            curl-dev \
            libevent-dev \
            libssh2-dev \
            libxml2-dev \
            net-snmp-dev \
            openipmi-dev \
            openldap-dev \
            pcre-dev \
            postgresql-dev \
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
# Does not support stable iksemel library
#            --with-jabber \
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
    cp database/${ZBX_DB_TYPE}/timescaledb.sql /usr/share/doc/zabbix-${ZBX_TYPE}-${ZBX_DB_TYPE}/ && \
    cd /tmp/ && \
    rm -rf /tmp/zabbix-${ZBX_VERSION}/ && \
    apk del ${APK_FLAGS_COMMON} --purge --no-network \
            build-dependencies && \
    rm -rf /var/cache/apk/*

EXPOSE 10051/TCP

WORKDIR /var/lib/zabbix

VOLUME ["/usr/lib/zabbix/alertscripts", "/usr/lib/zabbix/externalscripts", "/var/lib/zabbix/enc", "/var/lib/zabbix/mibs", "/var/lib/zabbix/modules"]
VOLUME ["/var/lib/zabbix/snmptraps", "/var/lib/zabbix/ssh_keys", "/var/lib/zabbix/ssl/certs", "/var/lib/zabbix/ssl/keys", "/var/lib/zabbix/ssl/ssl_ca"]

COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENTRYPOINT ["/sbin/tini", "--", "/usr/bin/docker-entrypoint.sh"]
