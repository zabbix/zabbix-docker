FROM centos:centos7

LABEL org.opencontainers.image.title="Zabbix web-interface (Nginx, MySQL)" \
      org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.vendor="Zabbix LLC" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.description="Zabbix web-interface based on Nginx web server with MySQL database support" \
      org.opencontainers.image.licenses="GPL v2.0"

STOPSIGNAL SIGTERM

RUN set -eux && \
    groupadd --system --gid 1995 zabbix && \
    useradd \
            --system --comment "Zabbix monitoring system" \
            -g zabbix -G root \
            --uid 1997 \
            --shell /sbin/nologin \
            --home-dir /var/lib/zabbix/ \
        zabbix && \
    mkdir -p /etc/zabbix && \
    mkdir -p /etc/zabbix/web && \
    mkdir -p /var/lib/php/session && \
    yum --quiet makecache && \
    yum -y install epel-release && \
    yum -y install --setopt=tsflags=nodocs \
            curl \
            mariadb \
            nginx \
            php-bcmath \
            php-fpm \
            php-gd \
            php-ldap \
            php-mbstring \
            php-mysql \
            php-xml \
            supervisor && \
    rm -f /etc/nginx/conf.d/*.conf  && \
    rm -f /etc/php-fpm.d/www.conf && \
    yum -y clean all && \
    rm -rf /var/cache/yum /var/lib/yum/yumdb/* /usr/lib/udev/hwdb.d/* && \
    rm -rf /etc/udev/hwdb.bin /root/.pki

ARG MAJOR_VERSION=master
ARG ZBX_VERSION=${MAJOR_VERSION}
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES}

LABEL org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.version="${ZBX_VERSION}" \
      org.opencontainers.image.source="${ZBX_SOURCES}"

COPY ["conf/etc/", "/etc/"]

RUN set -eux && \
    yum --quiet makecache && \
    yum -y install --setopt=tsflags=nodocs \
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
    cat /usr/share/zabbix/include/locales.inc.php | grep display | grep true | awk '{$1=$1};1' | \
                cut -d"'" -f 2 | sort | \
                xargs -I '{}' bash -c 'echo "{}" && localedef -c -i {} -f UTF-8 {}.UTF-8 2>/dev/null' && \
    chown --quiet -R zabbix:root /etc/zabbix/ && \
    chgrp -R 0 /etc/zabbix/ && \
    chmod -R g=u /etc/zabbix/ && \
    chown --quiet -R zabbix:root /etc/nginx/ /etc/php-fpm.d/ /etc/php-fpm.conf && \
    chgrp -R 0 /etc/nginx/ /etc/php-fpm.d/ /etc/php-fpm.conf && \
    chmod -R g=u /etc/nginx/ /etc/php-fpm.d/ /etc/php-fpm.conf && \
    chown --quiet -R zabbix:root /var/lib/php/session/ && \
    chgrp -R 0 /var/lib/php/session/ && \
    chmod -R g=u /var/lib/php/session/ && \
    chown --quiet -R zabbix:root /usr/share/zabbix/include/defines.inc.php && \
    chgrp -R 0 /usr/share/zabbix/include/defines.inc.php && \
    chmod -R g=u /usr/share/zabbix/include/defines.inc.php && \
    yum -y history undo `yum -q history | sed -n 3p |column -t | cut -d' ' -f1` && \
    yum -y clean all && \
    rm -rf /var/cache/yum /var/lib/yum/yumdb/* /usr/lib/udev/hwdb.d/* && \
    rm -rf /etc/udev/hwdb.bin /root/.pki

EXPOSE 8080/TCP 8443/TCP

WORKDIR /usr/share/zabbix

COPY ["docker-entrypoint.sh", "/usr/bin/"]

USER 1997

ENTRYPOINT ["docker-entrypoint.sh"]
