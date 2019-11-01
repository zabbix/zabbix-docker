FROM alpine:3.10
LABEL maintainer="Alexey Pustovalov <alexey.pustovalov@zabbix.com>"

ARG BUILD_DATE
ARG VCS_REF

ARG APK_FLAGS_COMMON=""
ARG APK_FLAGS_PERSISTENT="${APK_FLAGS_COMMON} --clean-protected --no-cache"
ARG APK_FLAGS_DEV="${APK_FLAGS_COMMON} --no-cache"
ENV TERM=xterm \
    ZBX_TYPE=frontend ZBX_DB_TYPE=postgresql ZBX_OPT_TYPE=nginx

LABEL org.label-schema.name="zabbix-web-${ZBX_OPT_TYPE}-${ZBX_DB_TYPE}-alpine" \
      org.label-schema.vendor="Zabbix LLC" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.description="Zabbix web-interface based on Nginx web server with PostgreSQL database support" \
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
            -H \
        zabbix && \
    mkdir -p /etc/zabbix && \
    mkdir -p /etc/zabbix/web && \
    chown --quiet -R zabbix:root /etc/zabbix && \
    apk update && \
    apk add ${APK_FLAGS_PERSISTENT} \
            bash \
            curl \
            nginx \
            php7-bcmath \
            php7-ctype \
            php7-fpm \
            php7-gd \
            php7-gettext \
            php7-json \
            php7-ldap \
            php7-mbstring \
            php7-pgsql \
            php7-session \
            php7-simplexml \
            php7-sockets \
            php7-fileinfo \
            php7-xmlreader \
            php7-xmlwriter \
            postgresql-client \
            supervisor && \
    rm -f /etc/nginx/conf.d/*.conf && \
    rm -rf /var/cache/apk/*

ARG MAJOR_VERSION=4.2
ARG ZBX_VERSION=${MAJOR_VERSION}.8
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ENV ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES}

LABEL org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.version="${ZBX_VERSION}" \
      org.label-schema.vcs-url="${ZBX_SOURCES}" \
      org.label-schema.docker.cmd="docker run --name zabbix-web-${ZBX_OPT_TYPE}-pgsql --link postgres-server:postgres --link zabbix-server:zabbix-server -p 80:80 -d zabbix-web-${ZBX_OPT_TYPE}-pgsql:alpine-${ZBX_VERSION}"

RUN set -eux && \
    apk add ${APK_FLAGS_DEV} --virtual build-dependencies \
            coreutils \
            gettext \
            git && \
    cd /usr/share/ && \
    git clone ${ZBX_SOURCES} --branch ${ZBX_VERSION} --depth 1 --single-branch zabbix-${ZBX_VERSION} && \
    mkdir /usr/share/zabbix/ && \
    cp -R /usr/share/zabbix-${ZBX_VERSION}/frontends/php/* /usr/share/zabbix/ && \
    rm -rf /usr/share/zabbix-${ZBX_VERSION}/ && \
    cd /usr/share/zabbix/ && \
    rm -f conf/zabbix.conf.php && \
    rm -rf tests && \
    ./locale/make_mo.sh && \
    chown --quiet -R nginx:nginx /usr/share/zabbix && \
    apk del ${APK_FLAGS_COMMON} --purge --no-network \
            build-dependencies && \
    rm -rf /var/cache/apk/*

EXPOSE 80/TCP 443/TCP

WORKDIR /usr/share/zabbix

VOLUME ["/etc/ssl/nginx"]

COPY ["conf/etc/supervisor/", "/etc/supervisor/"]
COPY ["conf/etc/zabbix/nginx.conf", "/etc/zabbix/"]
COPY ["conf/etc/zabbix/nginx_ssl.conf", "/etc/zabbix/"]
COPY ["conf/etc/zabbix/web/zabbix.conf.php", "/etc/zabbix/web/"]
COPY ["conf/etc/nginx/nginx.conf", "/etc/nginx/"]
COPY ["conf/etc/php7/php-fpm.conf", "/etc/php7/"]
COPY ["conf/etc/php7/conf.d/99-zabbix.ini", "/etc/php7/conf.d/"]
COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENTRYPOINT ["docker-entrypoint.sh"]
