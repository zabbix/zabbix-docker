FROM alpine:3.4 as builder

ARG APK_FLAGS_COMMON=""
ARG APK_FLAGS_DEV="${APK_FLAGS_COMMON} --no-cache"

ARG MAJOR_VERSION=4.0
ARG ZBX_VERSION=${MAJOR_VERSION}.2
ARG ZBX_SOURCES=svn://svn.zabbix.com/tags/${ZBX_VERSION}/
ENV ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} \
    PATH=${PATH}:/usr/lib/jvm/default-jvm/bin/ JAVA_HOME=/usr/lib/jvm/default-jvm \
    ZBX_TYPE=server ZBX_DB_TYPE=mysql ZBX_OPT_TYPE=nginx \
    MYSQL_ALLOW_EMPTY_PASSWORD=true ZBX_ADD_SERVER=true ZBX_ADD_WEB=true DB_SERVER_HOST=localhost MYSQL_USER=zabbix ZBX_ADD_JAVA_GATEWAY=true ZBX_JAVAGATEWAY_ENABLE=true ZBX_JAVAGATEWAY=localhost

COPY ["conf/tmp/font-config", "/tmp/font-config"]

RUN apk add ${APK_FLAGS_DEV} --virtual build-dependencies \
            bash \
            alpine-sdk \
            autoconf \
            automake \
            coreutils \
            curl-dev \
            gettext \
            libevent-dev \
            libssh2-dev \
            libxml2-dev \
            mysql-dev \
            net-snmp-dev \
            openipmi-dev \
            openjdk8 \
            openldap-dev \
            pcre-dev \
            subversion \
            unixodbc-dev && \
    cd /tmp/ && \
    svn --quiet export ${ZBX_SOURCES} zabbix-${ZBX_VERSION} && \
    cd /tmp/zabbix-${ZBX_VERSION} && \
    zabbix_revision=`svn info ${ZBX_SOURCES} |grep "Last Changed Rev"|awk '{print $4;}'` && \
    sed -i "s/{ZABBIX_REVISION}/$zabbix_revision/g" include/version.h && \
    sed -i "s/{ZABBIX_REVISION}/$zabbix_revision/g" include/version.h && \
    sed -i "s/{ZABBIX_REVISION}/$zabbix_revision/g" src/zabbix_java/src/com/zabbix/gateway/GeneralInformation.java && \
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
            --enable-java \
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
    cat database/${ZBX_DB_TYPE}/images.sql >> database/${ZBX_DB_TYPE}/create.sql && \
    cat database/${ZBX_DB_TYPE}/data.sql >> database/${ZBX_DB_TYPE}/create.sql && \
    gzip database/${ZBX_DB_TYPE}/create.sql && \
    rm -rf /tmp/zabbix-${ZBX_VERSION}/src/zabbix_java/lib/*.xml && \
    cd frontends/php/ && \
    patch -p3 < /tmp/font-config && \
    rm -f conf/zabbix.conf.php && \
    rm -rf tests && \
    rm /tmp/zabbix-${ZBX_VERSION}/frontends/php/fonts/DejaVuSans.ttf && \
    ./locale/make_mo.sh

FROM alpine:3.4
LABEL maintainer="Alexey Pustovalov <alexey.pustovalov@zabbix.com>"

ARG BUILD_DATE
ARG VCS_REF

ARG APK_FLAGS_COMMON=""
ARG APK_FLAGS_PERSISTENT="${APK_FLAGS_COMMON} --clean-protected --no-cache"

ARG MAJOR_VERSION=4.0
ARG ZBX_VERSION=${MAJOR_VERSION}.2
ARG ZBX_SOURCES=svn://svn.zabbix.com/tags/${ZBX_VERSION}/
ENV ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} \
    MIBDIRS=/usr/share/snmp/mibs:/var/lib/zabbix/mibs MIBS=+ALL PATH=${PATH}:/usr/lib/jvm/default-jvm/bin/ JAVA_HOME=/usr/lib/jvm/default-jvm \
    ZBX_TYPE=server ZBX_DB_TYPE=mysql ZBX_OPT_TYPE=nginx \
    MYSQL_ALLOW_EMPTY_PASSWORD=true ZBX_ADD_SERVER=true ZBX_ADD_WEB=true DB_SERVER_HOST=localhost MYSQL_USER=zabbix ZBX_ADD_JAVA_GATEWAY=true ZBX_JAVAGATEWAY_ENABLE=true ZBX_JAVAGATEWAY=localhost

LABEL org.label-schema.name="zabbix-${ZBX_TYPE}-${ZBX_DB_TYPE}-alpine" \
      org.label-schema.vendor="Zabbix LLC" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.description="Zabbix appliance with MySQL database support and ${ZBX_OPT_TYPE} web-server" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.license="GPL 2.0" \
      org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.version="${ZBX_VERSION}" \
      org.label-schema.vcs-url="${ZBX_SOURCES}" \
      org.label-schema.docker.cmd="docker run --name zabbix-appliance -p 80:80 -p 10051:10051 -d zabbix-appliance:alpine-${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY --from=builder /tmp/zabbix-${ZBX_VERSION}/src/zabbix_${ZBX_TYPE}/zabbix_${ZBX_TYPE} /usr/sbin/zabbix_${ZBX_TYPE}
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}/src/zabbix_get/zabbix_get /usr/bin/zabbix_get
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}/src/zabbix_sender/zabbix_sender /usr/bin/zabbix_sender
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}/conf/zabbix_${ZBX_TYPE}.conf /etc/zabbix/zabbix_${ZBX_TYPE}.conf
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}/database/${ZBX_DB_TYPE}/create.sql.gz /usr/share/doc/zabbix-${ZBX_TYPE}-${ZBX_DB_TYPE}/create.sql.gz

COPY --from=builder /tmp/zabbix-${ZBX_VERSION}/src/zabbix_java/bin/ /usr/sbin/zabbix_java/bin/
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}/src/zabbix_java/lib/ /usr/sbin/zabbix_java/lib/

COPY --from=builder /tmp/zabbix-${ZBX_VERSION}/frontends/php/ /usr/share/zabbix/

RUN addgroup zabbix && \
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
    mkdir -p /usr/share/doc/zabbix-${ZBX_TYPE}-${ZBX_DB_TYPE}/ && \
    ln -s /usr/share/fonts/ttf-dejavu/DejaVuSans.ttf /usr/share/zabbix/fonts/graphfont.ttf && \
    apk update && \
    apk add ${APK_FLAGS_PERSISTENT} \
            bash \
            fping \
            iputils \
            libcurl \
            libevent \
            libldap \
            libssh2 \
            libxml2 \
            mariadb-client \
            mariadb-client-libs \
            mysql \
            net-snmp-agent-libs \
            nginx \
            openipmi-libs \
            openjdk8-jre-base \
            php5-bcmath \
            php5-ctype \
            php5-fpm \
            php5-gd \
            php5-gettext \
            php5-json \
            php5-ldap \
            php5-mysqli \
            php5-sockets \
            php5-xmlreader \
            supervisor \
            ttf-dejavu \
            pcre \
            unixodbc && \
    rm -rf /var/cache/apk/*

EXPOSE 80/TCP 443/TCP 10051/TCP

WORKDIR /var/lib/zabbix

VOLUME ["/etc/ssl/nginx"]
VOLUME ["/usr/lib/zabbix/alertscripts", "/usr/lib/zabbix/externalscripts", "/var/lib/zabbix/enc", "/var/lib/zabbix/mibs", "/var/lib/zabbix/modules"]
VOLUME ["/var/lib/zabbix/snmptraps", "/var/lib/zabbix/ssh_keys", "/var/lib/zabbix/ssl/certs", "/var/lib/zabbix/ssl/keys", "/var/lib/zabbix/ssl/ssl_ca"]

COPY ["conf/etc/supervisor/", "/etc/supervisor/"]
COPY ["conf/etc/zabbix/nginx.conf", "/etc/zabbix/"]
COPY ["conf/etc/zabbix/nginx_ssl.conf", "/etc/zabbix/"]
COPY ["conf/etc/zabbix/web/zabbix.conf.php", "/etc/zabbix/web/"]
COPY ["conf/etc/nginx/nginx.conf", "/etc/nginx/"]
COPY ["conf/etc/php5/php-fpm.conf", "/etc/php5/"]
COPY ["conf/etc/php5/conf.d/99-zabbix.ini", "/etc/php5/conf.d/"]
COPY ["conf/etc/zabbix/zabbix_java_gateway_logback.xml", "/etc/zabbix/"]
COPY ["conf/usr/sbin/zabbix_java_gateway", "/usr/sbin/"]
COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENV ZBX_TYPE=appliance

ENTRYPOINT ["docker-entrypoint.sh"]
