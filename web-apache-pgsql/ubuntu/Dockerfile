FROM ubuntu:bionic

LABEL org.opencontainers.image.title="Zabbix web-interface (Apache, PostgreSQL)" \
      org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.vendor="Zabbix LLC" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.description="Zabbix web-interface based on Apache2 web server with PostgreSQL database support" \
      org.opencontainers.image.licenses="GPL v2.0"

STOPSIGNAL SIGTERM

RUN set -eux && \
    echo "#!/bin/sh\nexit 0" > /usr/sbin/policy-rc.d && \
    addgroup --system --gid 1995 --quiet zabbix && \
    adduser --quiet \
            --system --disabled-login \
            --ingroup zabbix --ingroup root \
            --uid 1997 \
            --home /var/lib/zabbix/ \
        zabbix && \
    mkdir -p /etc/zabbix && \
    mkdir -p /etc/zabbix/web && \
    apt-get -y update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y --no-install-recommends install \
            apache2 \
            curl \
            libapache2-mod-php \
            ca-certificates \
            locales \
            php7.2-bcmath \
            php7.2-gd \
            php7.2-json \
            php7.2-ldap \
            php7.2-mbstring \
            php7.2-xml \
            php7.2-pgsql \
            postgresql-client && \
    rm -f /etc/apache2/sites-available/* && \
    rm -f /etc/apache2/sites-enabled/* && \
    /usr/sbin/a2enmod ssl && \
    sed -ri \
            -e 's!^(\s*CustomLog)\s+\S+!\1 /proc/self/fd/1!g' \
            -e 's!^(\s*ErrorLog)\s+\S+!\1 /proc/self/fd/2!g' \
            "/etc/apache2/apache2.conf" && \
    sed -ri \
            -e 's!^(\s*CustomLog)\s+\S+!\1 /proc/self/fd/1!g' \
            -e 's!^(\s*ErrorLog)\s+\S+!\1 /proc/self/fd/2!g' \
            "/etc/apache2/conf-available/other-vhosts-access-log.conf" && \
    sed -i 's/Listen 80/Listen 8080/g' /etc/apache2/ports.conf && \
    sed -i 's/Listen 443/Listen 8443/g' /etc/apache2/ports.conf && \
    sed -i 's|/var/run/apache2$SUFFIX|/tmp|g' /etc/apache2/envvars && \
    rm -f /var/run/apache2/apache2.pid && \
    apt-get -y autoremove && \
    apt-get -y clean && \
    rm -rf /var/lib/apt/lists/*

ARG MAJOR_VERSION=master
ARG ZBX_VERSION=${MAJOR_VERSION}
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES}

LABEL org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.version="${ZBX_VERSION}" \
      org.opencontainers.image.source="${ZBX_SOURCES}"

COPY ["conf/etc/", "/etc/"]

RUN set -eux && \
    apt-get -y update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y --no-install-recommends install \
            gettext \
            git && \
    cd /usr/share/ && \
    git clone ${ZBX_SOURCES} --branch ${ZBX_VERSION} --depth 1 --single-branch zabbix-${ZBX_VERSION} && \
    mkdir /usr/share/zabbix/ && \
    cp -R /usr/share/zabbix-${ZBX_VERSION}/frontends/php/* /usr/share/zabbix/ && \
    rm -rf /usr/share/zabbix-${ZBX_VERSION}/ && \
    cd /usr/share/zabbix/ && \
    rm -f conf/zabbix.conf.php && \
    rm -f conf/maintenance.inc.php && \
    rm -rf tests && \
    ./locale/make_mo.sh && \
    ln -s "/etc/zabbix/web/zabbix.conf.php" "/usr/share/zabbix/conf/zabbix.conf.php" && \
    ln -s "/etc/zabbix/web/maintenance.inc.php" "/usr/share/zabbix/conf/maintenance.inc.php" && \
    ./locale/make_mo.sh && \
    mkdir -p /var/lib/locales/supported.d/ && \
    rm -f /var/lib/locales/supported.d/local && \
    cat /usr/share/zabbix/include/locales.inc.php | grep display | grep true | awk '{$1=$1};1' | \
                cut -d"'" -f 2 | sort | \
                xargs -I '{}' bash -c 'echo "{}.UTF-8 UTF-8" >> /var/lib/locales/supported.d/local' && \
    dpkg-reconfigure locales && \
    find /usr/share/zabbix/locale -name '*.po' | xargs rm -f && \
    find /usr/share/zabbix/locale -name '*.sh' | xargs rm -f && \
    chown --quiet -R zabbix:root /etc/zabbix/  && \
    chgrp -R 0 /etc/zabbix/  && \
    chmod -R g=u /etc/zabbix/  && \
    chown --quiet -R zabbix:root /etc/apache2/ /etc/php/7.2/ && \
    chgrp -R 0 /etc/apache2/ /etc/php/7.2/ && \
    chmod -R g=u /etc/apache2/ /etc/php/7.2/ && \
    chown --quiet -R zabbix:root /usr/share/zabbix/include/defines.inc.php && \
    chgrp -R 0 /usr/share/zabbix/include/defines.inc.php && \
    chmod -R g=u /usr/share/zabbix/include/defines.inc.php && \
    DEBIAN_FRONTEND=noninteractive apt-get -y purge \
            gettext \
            git && \
    apt-get -y autoremove && \
    apt-get -y clean && \
    rm -rf /var/lib/apt/lists/*

EXPOSE 8080/TCP 8443/TCP

WORKDIR /usr/share/zabbix

COPY ["docker-entrypoint.sh", "/usr/bin/"]

USER 1997

ENTRYPOINT ["docker-entrypoint.sh"]

CMD ["/usr/sbin/apache2ctl", "-D", "FOREGROUND"]
