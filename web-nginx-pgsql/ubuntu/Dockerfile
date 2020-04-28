FROM ubuntu:bionic

LABEL org.opencontainers.image.title="Zabbix web-interface (Nginx, PostgreSQL)" \
      org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.vendor="Zabbix LLC" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.description="Zabbix web-interface based on Nginx web server with PostgreSQL database support" \
      org.opencontainers.image.licenses="GPL v2.0"

STOPSIGNAL SIGTERM

RUN set -eux && \
    addgroup --system --gid 1995 --quiet zabbix && \
    adduser --quiet \
            --system --disabled-login \
            --ingroup zabbix --ingroup root \
            --uid 1997 \
            --home /var/lib/zabbix/ \
        zabbix && \
    mkdir -p /etc/zabbix && \
    mkdir -p /etc/zabbix/web && \
    mkdir -p /var/lib/php/session && \
    apt-get -y update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y --no-install-recommends install \
        gpg \
        dirmngr \
        gpg-agent \
        ca-certificates && \
    NGINX_GPGKEY=573BFD6B3D8FBC641079A6ABABF5BD827BD9BF62; \
    found=''; \
    for server in \
        ha.pool.sks-keyservers.net \
        hkp://keyserver.ubuntu.com:80 \
        hkp://p80.pool.sks-keyservers.net:80 \
        pgp.mit.edu \
    ; do \
        echo "Fetching GPG key $NGINX_GPGKEY from $server"; \
        apt-key adv --keyserver "$server" --keyserver-options timeout=10 --recv-keys "$NGINX_GPGKEY" && found=yes && break; \
    done; \
    test -z "$found" && echo >&2 "error: failed to fetch GPG key $NGINX_GPGKEY" && exit 1; \
    DISTRIB_CODENAME=$(/bin/bash -c 'source /etc/lsb-release && echo $DISTRIB_CODENAME') && \
    echo "deb https://nginx.org/packages/ubuntu/ $DISTRIB_CODENAME nginx" >> /etc/apt/sources.list.d/nginx.list && \
    apt-get -y update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y --no-install-recommends install \
            curl \
            nginx \
            php7.2-bcmath \
            php7.2-fpm \
            php7.2-gd \
            php7.2-json \
            php7.2-ldap \
            php7.2-mbstring \
            php7.2-xml \
            php7.2-pgsql \
            postgresql-client \
            supervisor && \
    rm -f /etc/nginx/conf.d/*.conf && \
    DEBIAN_FRONTEND=noninteractive apt-get ${APT_FLAGS_COMMON} purge \
            wget && \
    apt-get ${APT_FLAGS_COMMON} autoremove && \
    apt-get ${APT_FLAGS_COMMON} clean && \
    mkdir -p /var/lib/php7 && \
    chown --quiet -R www-data:www-data /var/lib/php7 && \
    rm -rf /var/cache/nginx/* && \
    rm -rf /var/lib/apt/lists/*

ARG MAJOR_VERSION=master
ARG ZBX_VERSION=${MAJOR_VERSION}
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ENV ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES}

LABEL org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.version="${ZBX_VERSION}" \
      org.label-schema.vcs-url="${ZBX_SOURCES}" \
      org.label-schema.docker.cmd="docker run --name zabbix-web-${ZBX_OPT_TYPE}-pgsql --link postgres-server:postgres --link zabbix-server:zabbix-server -p 80:80 -d zabbix-web-${ZBX_OPT_TYPE}-pgsql:ubuntu-${ZBX_VERSION}"

RUN set -eux && \
    apt-get ${APT_FLAGS_COMMON} update && \
    DEBIAN_FRONTEND=noninteractive apt-get ${APT_FLAGS_DEV} install \
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
    ./locale/make_mo.sh && \
    mkdir -p /var/lib/locales/supported.d/ && \
    rm -f /var/lib/locales/supported.d/local && \
    cat /usr/share/zabbix/include/locales.inc.php | grep display | grep true | awk '{$1=$1};1' | \
                cut -d"'" -f 2 | sort | \
                xargs -I '{}' bash -c 'echo "{}.UTF-8 UTF-8" >> /var/lib/locales/supported.d/local' && \
    dpkg-reconfigure locales && \
    find /usr/share/zabbix/locale -name '*.po' | xargs rm -f && \
    find /usr/share/zabbix/locale -name '*.sh' | xargs rm -f && \
    chown --quiet -R zabbix:root /etc/zabbix/ /usr/share/zabbix/conf/ && \
    chgrp -R 0 /etc/zabbix/ /usr/share/zabbix/conf/ && \
    chmod -R g=u /etc/zabbix/ /usr/share/zabbix/conf/ && \
    chown --quiet -R zabbix:root /etc/nginx/ /etc/php/7.2/fpm/php-fpm.conf /etc/php/7.2/fpm/pool.d/ && \
    chgrp -R 0 /etc/nginx/ /etc/php/7.2/fpm/php-fpm.conf /etc/php/7.2/fpm/pool.d/ && \
    chmod -R g=u /etc/nginx/ /etc/php/7.2/fpm/php-fpm.conf /etc/php/7.2/fpm/pool.d/ && \
    chown --quiet -R zabbix:root /var/lib/php/session/ && \
    chgrp -R 0 /var/lib/php/session/ && \
    chmod -R g=u /var/lib/php/session/ && \
    DEBIAN_FRONTEND=noninteractive apt-get -y purge \
            gettext \
            git \
            locales && \
    apt-get -y autoremove && \
    apt-get -y clean && \
    rm -rf /var/lib/apt/lists/*

EXPOSE 8080/TCP 8443/TCP

WORKDIR /usr/share/zabbix

VOLUME ["/etc/ssl/nginx"]

COPY ["docker-entrypoint.sh", "/usr/bin/"]

USER 1997

ENTRYPOINT ["docker-entrypoint.sh"]
