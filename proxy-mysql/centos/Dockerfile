FROM centos:centos7 as builder

ARG YUM_FLAGS_COMMON="-y"
ARG YUM_FLAGS_DEV="${YUM_FLAGS_COMMON}"

ARG MAJOR_VERSION=3.4
ARG ZBX_VERSION=${MAJOR_VERSION}.15
ARG ZBX_SOURCES=svn://svn.zabbix.com/tags/${ZBX_VERSION}/
ENV ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} \
    ZBX_TYPE=proxy ZBX_DB_TYPE=mysql

RUN yum --quiet makecache && \
    yum ${YUM_FLAGS_DEV} install \
            autoconf \
            automake \
            gcc \
            libcurl-devel \
            libevent-devel \
            libssh2-devel \
            libxml2-devel \
            make \
            mariadb-devel \
            net-snmp-devel \
            OpenIPMI-devel \
            openldap-devel \
            subversion \
            unixODBC-devel && \
    cd /tmp/ && \
    svn --quiet export ${ZBX_SOURCES} zabbix-${ZBX_VERSION} && \
    cd /tmp/zabbix-${ZBX_VERSION} && \
    zabbix_revision=`svn info ${ZBX_SOURCES} | grep "Last Changed Rev"|awk '{print $4;}'` && \
    sed -i "s/{ZABBIX_REVISION}/$zabbix_revision/g" include/version.h && \
    ./bootstrap.sh && \
    export CFLAGS="-fPIC -pie -Wl,-z,relro -Wl,-z,now" && \
    ./configure \
            --datadir=/usr/lib \
            --libdir=/usr/lib/zabbix \
            --sysconfdir=/etc/zabbix \
            --prefix=/usr \
            --enable-agent \
            --enable-${ZBX_TYPE} \
            --with-${ZBX_DB_TYPE} \
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
    cat database/${ZBX_DB_TYPE}/schema.sql > database/${ZBX_DB_TYPE}/create.sql && \
    gzip database/${ZBX_DB_TYPE}/create.sql

FROM centos:centos7
LABEL maintainer="Alexey Pustovalov <alexey.pustovalov@zabbix.com>"

ARG BUILD_DATE
ARG VCS_REF

ARG YUM_FLAGS_COMMON="-y"
ARG YUM_FLAGS_PERSISTENT="${YUM_FLAGS_COMMON}"

ARG MAJOR_VERSION=3.4
ARG ZBX_VERSION=${MAJOR_VERSION}.15
ARG ZBX_SOURCES=svn://svn.zabbix.com/tags/${ZBX_VERSION}/
ENV TERM=xterm MIBDIRS=/usr/share/snmp/mibs:/var/lib/zabbix/mibs MIBS=+ALL \
    ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} \
    ZBX_TYPE=proxy ZBX_DB_TYPE=mysql ZBX_OPT_TYPE=none
ENV TINI_VERSION v0.18.0

LABEL org.label-schema.name="zabbix-${ZBX_TYPE}-${ZBX_DB_TYPE}-centos" \
      org.label-schema.vendor="Zabbix LLC" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.description="Zabbix ${ZBX_TYPE} with MySQL database support" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.license="GPL v2.0" \
      org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.version="${ZBX_VERSION}" \
      org.label-schema.vcs-url="${ZBX_SOURCES}" \
      org.label-schema.docker.cmd="docker run --name zabbix-${ZBX_TYPE}-${ZBX_DB_TYPE} --link mysql-server:mysql --link zabbix-server:zabbix-server -p 10051:10051 -d zabbix-${ZBX_TYPE}-${ZBX_DB_TYPE}:centos-${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY --from=builder /tmp/zabbix-${ZBX_VERSION}/src/zabbix_${ZBX_TYPE}/zabbix_${ZBX_TYPE} /usr/sbin/zabbix_${ZBX_TYPE}
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}/src/zabbix_get/zabbix_get /usr/bin/zabbix_get
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}/src/zabbix_sender/zabbix_sender /usr/bin/zabbix_sender
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}/conf/zabbix_${ZBX_TYPE}.conf /etc/zabbix/zabbix_${ZBX_TYPE}.conf
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}/database/${ZBX_DB_TYPE}/create.sql.gz /usr/share/doc/zabbix-${ZBX_TYPE}-${ZBX_DB_TYPE}/create.sql.gz
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /sbin/tini

RUN groupadd --system zabbix && \
    adduser -r --shell /sbin/nologin \
            -g zabbix \
            -d /var/lib/zabbix/ \
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
    mkdir -p /usr/share/doc/zabbix-${ZBX_TYPE}-${ZBX_DB_TYPE} && \
    yum ${YUM_FLAGS_COMMON} makecache && \
    yum ${YUM_FLAGS_PERSISTENT} install http://repo.zabbix.com/non-supported/rhel/7/x86_64/fping-3.10-1.el7.x86_64.rpm && \
    yum ${YUM_FLAGS_PERSISTENT} install \
            libcurl \
            libevent \
            libxml2 \
            mariadb \
            net-snmp-libs \
            OpenIPMI-libs \
            openldap \
            openssl-libs \
            pcre \
            unixODBC && \
    yum ${YUM_FLAGS_PERSISTANT} clean all && \
    rm -rf /var/cache/yum/ && \
    chmod +x /sbin/tini

EXPOSE 10051/TCP

WORKDIR /var/lib/zabbix

VOLUME ["/usr/lib/zabbix/externalscripts", "/var/lib/zabbix/enc", "/var/lib/zabbix/modules", "/var/lib/zabbix/snmptraps"]
VOLUME ["/var/lib/zabbix/ssh_keys", "/var/lib/zabbix/ssl/certs", "/var/lib/zabbix/ssl/keys", "/var/lib/zabbix/ssl/ssl_ca"]

COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENTRYPOINT ["/sbin/tini", "--"]

CMD ["docker-entrypoint.sh"]
