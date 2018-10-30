FROM alpine:3.7
LABEL maintainer="Alexey Pustovalov <alexey.pustovalov@zabbix.com>"

ARG BUILD_DATE
ARG VCS_REF

ARG APK_FLAGS_COMMON=""
ARG APK_FLAGS_PERSISTENT="${APK_FLAGS_COMMON} --clean-protected --no-cache"
ARG APK_FLAGS_DEV="${APK_FLAGS_COMMON} --no-cache"
ENV TERM=xterm \
    ZBX_TYPE=frontend ZBX_DB_TYPE=postgresql ZBX_OPT_TYPE=apache

LABEL org.label-schema.name="zabbix-web-${ZBX_OPT_TYPE}-${ZBX_DB_TYPE}-alpine" \
      org.label-schema.vendor="Zabbix LLC" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.description="Zabbix web-interface based on Apache2 web server with PostgreSQL database support" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.license="GPL v2.0"

STOPSIGNAL SIGTERM

RUN addgroup zabbix && \
    adduser -S \
            -D -G zabbix \
            -h /var/lib/zabbix/ \
            -H \
        zabbix && \
    mkdir -p /etc/zabbix && \
    mkdir -p /etc/zabbix/web && \
    chown --quiet -R zabbix:root /etc/zabbix && \
    apk update && \
    apk add ${APK_FLAGS_PERSISTENT} \
            apache2 \
            bash \
            curl \
            php7-apache2 \
            php7-bcmath \
            php7-ctype \
            php7-gd \
            php7-gettext \
            php7-json \
            php7-ldap \    
            php7-pgsql \
            php7-mbstring \
            php7-session \
            php7-simplexml \
            php7-sockets \
            php7-xmlreader \
            php7-xmlwriter \
            postgresql-client \
            ttf-dejavu && \
    apk add ${APK_FLAGS_PERSISTENT} --no-scripts apache2-ssl && \
    rm -rf /var/cache/apk/*

ARG MAJOR_VERSION=4.0
ARG ZBX_VERSION=${MAJOR_VERSION}.1
ARG ZBX_SOURCES=svn://svn.zabbix.com/tags/${ZBX_VERSION}/
ENV ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES}

LABEL org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.version="${ZBX_VERSION}" \
      org.label-schema.vcs-url="${ZBX_SOURCES}" \
      org.label-schema.docker.cmd="docker run --name zabbix-web-${ZBX_OPT_TYPE}-pgsql --link postgres-server:postgres --link zabbix-server:zabbix-server -p 80:80 -d zabbix-web-${ZBX_OPT_TYPE}-pgsql:alpine-${ZBX_VERSION}"

COPY ["conf/tmp/font-config", "/tmp/font-config"]

RUN apk add ${APK_FLAGS_DEV} --virtual build-dependencies \
            gettext \
            subversion && \
    cd /usr/share/ && \
    svn --quiet export ${ZBX_SOURCES}/frontends/php/ zabbix && \
    cd /usr/share/zabbix/ && \
    patch -p3 < /tmp/font-config && \
    rm /tmp/font-config && \
    rm -f conf/zabbix.conf.php && \
    rm -rf tests && \
    rm /usr/share/zabbix/fonts/DejaVuSans.ttf && \
    ./locale/make_mo.sh && \
    ln -s /usr/share/fonts/ttf-dejavu/DejaVuSans.ttf /usr/share/zabbix/fonts/graphfont.ttf && \
    apk del ${APK_FLAGS_COMMON} --purge \
            build-dependencies && \
    rm -rf /var/cache/apk/* && \
    rm -rf /root/.subversion

EXPOSE 80/TCP 443/TCP

WORKDIR /usr/share/zabbix

VOLUME ["/etc/ssl/apache2"]

COPY ["conf/etc/zabbix/apache.conf", "/etc/zabbix/"]
COPY ["conf/etc/zabbix/apache_ssl.conf", "/etc/zabbix/"]
COPY ["conf/etc/zabbix/web/zabbix.conf.php", "/etc/zabbix/web/"]
COPY ["conf/etc/php7/conf.d/99-zabbix.ini", "/etc/php7/conf.d/"]
COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENTRYPOINT ["docker-entrypoint.sh"]
