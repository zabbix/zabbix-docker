FROM alpine:3.4
LABEL maintainer "Alexey Pustovalov <alexey.pustovalov@zabbix.com>"

ARG APK_FLAGS_COMMON="-q"
ARG APK_FLAGS_PERSISTANT="${APK_FLAGS_COMMON} --clean-protected --no-cache"
ARG APK_FLAGS_DEV="${APK_FLAGS_COMMON} --no-cache"
ARG DB_TYPE=postgresql
ENV TERM=xterm

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
    apk add ${APK_FLAGS_PERSISTANT} \
            supervisor \
            bash \
            postgresql-client \
            nginx \
            php5-fpm \
            php5-pgsql \
            php5-ctype \
            php5-sockets \
            php5-gd \
            php5-gettext \
            php5-bcmath \
            php5-xmlreader \
            php5-ldap \
            php5-json \
            ttf-dejavu && \
    rm -rf /var/cache/apk/*

ARG MAJOR_VERSION=3.4
ARG ZBX_VERSION=${MAJOR_VERSION}.3
ARG ZBX_SOURCES=svn://svn.zabbix.com/tags/${ZBX_VERSION}/
ENV ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} DB_TYPE=${DB_TYPE}

ADD conf/tmp/font-config /tmp/font-config

RUN apk add ${APK_FLAGS_DEV} --virtual build-dependencies \
            coreutils \
            gettext \
            subversion && \
    cd /usr/share/ && \
    svn --quiet export ${ZBX_SOURCES}/frontends/php/ zabbix 1>/dev/null && \
    cd /usr/share/zabbix/ && \
    patch -p3 < /tmp/font-config && \
    rm /tmp/font-config && \
    rm -f conf/zabbix.conf.php && \
    rm -rf tests && \
    rm /usr/share/zabbix/fonts/DejaVuSans.ttf && \
    ./locale/make_mo.sh 2>/dev/null && \
    ln -s /usr/share/fonts/ttf-dejavu/DejaVuSans.ttf /usr/share/zabbix/fonts/graphfont.ttf && \
    apk del ${APK_FLAGS_COMMON} --purge \
            build-dependencies && \
    rm -rf /var/cache/apk/*

EXPOSE 80/TCP 443/TCP

WORKDIR /usr/share/zabbix

VOLUME ["/etc/ssl/nginx"]

ADD conf/etc/supervisor/ /etc/supervisor/
ADD conf/etc/zabbix/nginx.conf /etc/zabbix/
ADD conf/etc/zabbix/nginx_ssl.conf /etc/zabbix/
ADD conf/etc/zabbix/web/zabbix.conf.php /etc/zabbix/web/
ADD conf/etc/nginx/nginx.conf /etc/nginx/
ADD conf/etc/php5/php-fpm.conf /etc/php5/
ADD conf/etc/php5/conf.d/99-zabbix.ini /etc/php5/conf.d/
ADD run_zabbix_component.sh /

ENTRYPOINT ["/bin/bash"]

CMD ["/run_zabbix_component.sh", "frontend", "postgresql", "nginx"]
