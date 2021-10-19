# syntax=docker/dockerfile:1
ARG MAJOR_VERSION=5.0
ARG ZBX_VERSION=${MAJOR_VERSION}.17
ARG BUILD_BASE_IMAGE=zabbix-build-mysql:ubuntu-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} as builder

FROM ubuntu:focal

ARG MAJOR_VERSION
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES}

LABEL org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix web-interface based on Nginx web server with MySQL database support" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="GPL v2.0" \
      org.opencontainers.image.source="${ZBX_SOURCES}" \
      org.opencontainers.image.title="Zabbix web-interface (Nginx, MySQL)" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix LLC" \
      org.opencontainers.image.version="${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}/ui", "/usr/share/zabbix"]
COPY ["conf/etc/", "/etc/"]

RUN set -eux && \
    echo "#!/bin/sh\nexit 101" > /usr/sbin/policy-rc.d && \
    INSTALL_PKGS="bash \
            ca-certificates \
            curl \
            mysql-client \
            nginx \
            locales \
            php7.4-bcmath \
            php7.4-fpm \
            php7.4-gd \
            php7.4-json \
            php7.4-ldap \
            php7.4-mbstring \
            php7.4-mysql \
            php7.4-xml \
            supervisor" && \
    INSTALL_TEMP_PKGS="gpg \
            ca-certificates \
            dirmngr \
            gpg-agent" && \
    apt-get -y update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y \
            --no-install-recommends install \
        ${INSTALL_TEMP_PKGS} && \
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
    DEBIAN_FRONTEND=noninteractive apt-get -y \
            -o Dpkg::Options::="--force-confdef" \
            -o Dpkg::Options::="--force-confold" \
            --no-install-recommends install \
        ${INSTALL_PKGS} && \
    groupadd \
            --system \
            --gid 1995 \
        zabbix && \
    useradd \
            --system \
            --comment "Zabbix monitoring system" \
            -g zabbix \
            -G root \
            --uid 1997 \
            --shell /sbin/nologin \
            --home-dir /var/lib/zabbix/ \
        zabbix && \
    mkdir -p /etc/zabbix && \
    mkdir -p /etc/zabbix/web && \
    mkdir -p /etc/zabbix/web/certs && \
    mkdir -p /var/lib/php/session && \
    rm -f /etc/nginx/conf.d/*.conf && \
    rm -rf /var/cache/nginx/ && \
    rm -f /etc/php/7.4/fpm/pool.d/www.conf && \
    ln -sf /dev/fd/2 /var/log/nginx/error.log && \
    rm -f /etc/php/7.4/fpm/php-fpm.conf.dpkg-dist && \
    cd /usr/share/zabbix/ && \
    rm -f conf/zabbix.conf.php conf/maintenance.inc.php conf/zabbix.conf.php.example && \
    rm -rf tests && \
    rm -f locale/add_new_language.sh locale/update_po.sh locale/make_mo.sh && \
    find /usr/share/zabbix/locale -name '*.po' | xargs rm -f && \
    find /usr/share/zabbix/locale -name '*.sh' | xargs rm -f && \
    ln -s "/etc/zabbix/web/zabbix.conf.php" "/usr/share/zabbix/conf/zabbix.conf.php" && \
    ln -s "/etc/zabbix/web/maintenance.inc.php" "/usr/share/zabbix/conf/maintenance.inc.php" && \
    mkdir -p /var/lib/locales/supported.d/ && \
    rm -f /var/lib/locales/supported.d/local && \
    cat /usr/share/zabbix/include/locales.inc.php | grep display | grep true | awk '{$1=$1};1' | \
        cut -d"'" -f 2 | sort | \
        xargs -I '{}' bash -c 'echo "{}.UTF-8 UTF-8" >> /var/lib/locales/supported.d/local' && \
    dpkg-reconfigure locales && \
    chown --quiet -R zabbix:root /etc/zabbix/ /usr/share/zabbix/include/defines.inc.php /usr/share/zabbix/modules/ && \
    chgrp -R 0 /etc/zabbix/ /usr/share/zabbix/include/defines.inc.php /usr/share/zabbix/modules/ && \
    chmod -R g=u /etc/zabbix/ /usr/share/zabbix/include/defines.inc.php /usr/share/zabbix/modules/ && \
    chown --quiet -R zabbix:root /etc/nginx/ /etc/php/7.4/fpm/php-fpm.conf /etc/php/7.4/fpm/pool.d/ && \
    chgrp -R 0 /etc/nginx/ /etc/php/7.4/fpm/php-fpm.conf /etc/php/7.4/fpm/pool.d/ && \
    chmod -R g=u /etc/nginx/ /etc/php/7.4/fpm/php-fpm.conf /etc/php/7.4/fpm/pool.d/ && \
    chown --quiet -R zabbix:root /var/lib/php/session/ && \
    chgrp -R 0 /var/lib/php/session/ && \
    chmod -R g=u /var/lib/php/session/ && \
    DEBIAN_FRONTEND=noninteractive apt-get -y \
        purge \
            gpg \
            dirmngr \
            gpg-agent && \
    apt-get -y autoremove && \
    apt-get -y clean && \
    rm -rf /var/lib/apt/lists/*

EXPOSE 8080/TCP 8443/TCP

WORKDIR /usr/share/zabbix

COPY ["docker-entrypoint.sh", "/usr/bin/"]

USER 1997

ENTRYPOINT ["docker-entrypoint.sh"]
