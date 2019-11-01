FROM centos:centos7 as builder

ARG YUM_FLAGS_COMMON="-y"
ARG YUM_FLAGS_DEV="${YUM_FLAGS_COMMON}"

ARG MAJOR_VERSION=4.2
ARG ZBX_VERSION=${MAJOR_VERSION}.8
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ENV TERM=xterm ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} \
    ZBX_TYPE=frontend ZBX_DB_TYPE=mysql ZBX_OPT_TYPE=nginx

RUN set -eux && \
    yum --quiet makecache && \
    yum ${YUM_FLAGS_DEV} install \
            gettext \
            patch \
            git && \
    cd /tmp/ && \
    git clone ${ZBX_SOURCES} --branch ${ZBX_VERSION} --depth 1 --single-branch zabbix-${ZBX_VERSION} && \
    mkdir /tmp/zabbix/ && \
    cp -R /tmp/zabbix-${ZBX_VERSION}/frontends/php/* /tmp/zabbix/ && \
    rm -rf /tmp/zabbix-${ZBX_VERSION}/ && \
    cd /tmp/zabbix/ && \
    rm -f conf/zabbix.conf.php && \
    rm -rf tests && \
    ./locale/make_mo.sh

FROM centos:centos7
LABEL maintainer="Alexey Pustovalov <alexey.pustovalov@zabbix.com>"

ARG BUILD_DATE
ARG VCS_REF

ARG YUM_FLAGS_COMMON="-y"
ARG YUM_FLAGS_PERSISTENT="${YUM_FLAGS_COMMON}"

ARG MAJOR_VERSION=4.2
ARG ZBX_VERSION=${MAJOR_VERSION}.8
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ENV TERM=xterm ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} \
    ZBX_TYPE=frontend ZBX_DB_TYPE=mysql ZBX_OPT_TYPE=nginx

LABEL org.label-schema.name="zabbix-web-${ZBX_OPT_TYPE}-${ZBX_DB_TYPE}-centos" \
      org.label-schema.vendor="Zabbix LLC" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.description="Zabbix web-interface based on Nginx web server with MySQL database support" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.license="GPL v2.0" \
      org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.version="${ZBX_VERSION}" \
      org.label-schema.vcs-url="${ZBX_SOURCES}" \
      org.label-schema.docker.cmd="docker run --name zabbix-web-${ZBX_OPT_TYPE}-${ZBX_DB_TYPE} --link mysql-server:mysql --link zabbix-server:zabbix-server -p 80:80 -d zabbix-web-${ZBX_OPT_TYPE}-${ZBX_DB_TYPE}:centos-${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY --from=builder /tmp/zabbix/ /usr/share/zabbix/

RUN set -eux && \
    groupadd --system zabbix && \
    adduser -r --shell /sbin/nologin \
            -g zabbix \
            -d /var/lib/zabbix/ \
        zabbix && \
    mkdir -p /etc/zabbix && \
    mkdir -p /etc/zabbix/web && \
    chown --quiet -R zabbix:root /etc/zabbix && \
    yum ${YUM_FLAGS_COMMON} makecache && \
    yum ${YUM_FLAGS_PERSISTENT} install epel-release && \
    yum ${YUM_FLAGS_PERSISTENT} install \
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
    rm -f /etc/nginx/conf.d/*.conf && \
    cat /usr/share/zabbix/include/locales.inc.php | grep display | grep true | awk '{$1=$1};1' | \
                cut -d"'" -f 2 | sort | \
                xargs -I '{}' bash -c 'echo "{}" && localedef -c -i {} -f UTF-8 {}.UTF-8 2>/dev/null' && \
    chown --quiet -R nginx:nginx /usr/share/zabbix && \
    rm -f /etc/php-fpm.d/www.conf && \
    mkdir -p /var/lib/php/ && \
    chown --quiet -R nginx:nginx /var/lib/php/ && \
    yum ${YUM_FLAGS_PERSISTENT} clean all && \
    rm -rf /var/cache/yum/

EXPOSE 80/TCP 443/TCP

WORKDIR /usr/share/zabbix

VOLUME ["/etc/ssl/nginx"]

COPY ["conf/etc/supervisor/", "/etc/supervisor/"]
COPY ["conf/etc/zabbix/nginx.conf", "/etc/zabbix/"]
COPY ["conf/etc/zabbix/nginx_ssl.conf", "/etc/zabbix/"]
COPY ["conf/etc/zabbix/web/zabbix.conf.php", "/etc/zabbix/web/"]
COPY ["conf/etc/nginx/nginx.conf", "/etc/nginx/"]
COPY ["conf/etc/php-fpm.conf", "/etc/php-fpm.conf"]
COPY ["conf/etc/php.d/99-zabbix.ini", "/etc/php.d/99-zabbix.ini"]
COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENTRYPOINT ["docker-entrypoint.sh"]
