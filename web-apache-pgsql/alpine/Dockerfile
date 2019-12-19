FROM alpine:3.10

LABEL org.opencontainers.image.title="Zabbix web-interface (Apache, PostgreSQL)" \
      org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.vendor="Zabbix LLC" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.description="abbix web-interface based on Apache2 web server with PostgreSQL database support" \
      org.opencontainers.image.licenses="GPL v2.0"

STOPSIGNAL SIGTERM

RUN set -eux && \
    addgroup -S -g 1000 zabbix && \
    adduser -S \
            -D -G zabbix \
            -u 999 \
            -h /var/lib/zabbix/ \
            -H \
        zabbix && \
    mkdir -p /etc/zabbix && \
    mkdir -p /etc/zabbix/web && \
    chown --quiet -R zabbix:root /etc/zabbix && \
    apk add --clean-protected --no-cache \
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
            php7-fileinfo \
            php7-xmlreader \
            php7-xmlwriter \
            postgresql-client && \
    apk add --clean-protected --no-cache --no-scripts apache2-ssl && \
    rm -f "/etc/apache2/conf.d/default.conf" && \
    rm -f "/etc/apache2/conf.d/ssl.conf" && \
    sed -ri \
        -e 's!^(\s*CustomLog)\s+\S+!\1 /proc/self/fd/1!g' \
        -e 's!^(\s*ErrorLog)\s+\S+!\1 /proc/self/fd/2!g' \
        "/etc/apache2/httpd.conf" && \
    sed -ri \
        -e 's!^(\s*PidFile)\s+\S+!\1 "/var/run/httpd.pid"!g' \
        "/etc/apache2/conf.d/mpm.conf" && \
    rm -f "/var/run/apache2/apache2.pid" && \
    rm -rf /var/cache/apk/*

ARG MAJOR_VERSION=4.4
ARG ZBX_VERSION=${MAJOR_VERSION}.4
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES}

LABEL org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.version="${ZBX_VERSION}" \
      org.opencontainers.image.source="${ZBX_SOURCES}"

RUN set -eux && \
    apk add --no-cache --virtual build-dependencies \
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
    ln -s "/etc/zabbix/web/zabbix.conf.php" "/usr/share/zabbix/conf/zabbix.conf.php" && \
    apk del --purge --no-network \
            build-dependencies && \
    rm -rf /var/cache/apk/*

EXPOSE 80/TCP 443/TCP

WORKDIR /usr/share/zabbix

VOLUME ["/etc/ssl/apache2"]

COPY ["conf/etc/zabbix/apache.conf", "/etc/zabbix/"]
COPY ["conf/etc/zabbix/apache_ssl.conf", "/etc/zabbix/"]
COPY ["conf/etc/zabbix/web/zabbix.conf.php", "/etc/zabbix/web/"]
COPY ["conf/etc/php7/conf.d/99-zabbix.ini", "/etc/php7/conf.d/"]
COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENTRYPOINT ["docker-entrypoint.sh"]

CMD ["/usr/sbin/httpd", "-D", "FOREGROUND"]
