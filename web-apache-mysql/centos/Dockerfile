FROM centos:centos7
LABEL maintainer="Alexey Pustovalov <alexey.pustovalov@zabbix.com>"

ARG BUILD_DATE
ARG VCS_REF

ARG YUM_FLAGS_COMMON="--quiet -y"
ARG YUM_FLAGS_PERSISTANT="${YUM_FLAGS_COMMON}"
ARG YUM_FLAGS_DEV="${YUM_FLAGS_COMMON}"
ARG DB_TYPE=mysql
ENV TERM=xterm

LABEL org.label-schema.name="zabbix-web-apache-${DB_TYPE}-centos" \
      org.label-schema.vendor="Zabbix LLC" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.description="Zabbix web-interface based on Apache2 web server with MySQL database support" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.license="GPL 2.0"

STOPSIGNAL SIGTERM

RUN groupadd --system zabbix && \
    adduser -r --shell /sbin/nologin \
            -g zabbix \
            -d /var/lib/zabbix/ \
        zabbix && \
    mkdir -p /etc/zabbix && \
    mkdir -p /etc/zabbix/web && \
    chown --quiet -R zabbix:root /etc/zabbix && \
    yum ${YUM_FLAGS_COMMON} makecache && \
    yum ${YUM_FLAGS_DEV} install epel-release && \
    yum ${YUM_FLAGS_PERSISTANT} install \
            dejavu-sans-fonts \
            httpd \
            mariadb \
            php \
            php-bcmath \
            php-gd \
            php-ldap \
            php-mbstring \
            php-mysql \
            php-xml \
            supervisor && \
    yum ${YUM_FLAGS_PERSISTANT} clean all && \
    rm -rf /var/cache/yum/

ARG MAJOR_VERSION=master
ARG ZBX_VERSION=${MAJOR_VERSION}
ARG ZBX_SOURCES=svn://svn.zabbix.com/tags/${ZBX_VERSION}/
ENV ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} DB_TYPE=${DB_TYPE}

LABEL org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.version="${ZBX_VERSION}" \
      org.label-schema.vcs-url="${ZBX_SOURCES}" \
      org.label-schema.docker.cmd="docker run --name zabbix-web-apache-${DB_TYPE} --link mysql-server:mysql --link zabbix-server:zabbix-server -p 80:80 -d zabbix-web-apache-${DB_TYPE}:centos-${ZBX_VERSION}"

COPY ["conf/tmp/font-config", "/tmp/font-config"]

RUN yum --quiet makecache && \
    yum ${YUM_FLAGS_DEV} install \
            gettext \
            patch \
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
    chown -R apache:apache /usr/share/zabbix && \
    yum ${YUM_FLAGS_COMMON} history undo `yum history | sed -n 4p |column -t | cut -d' ' -f1` 1>/dev/null && \
    yum ${YUM_FLAGS_COMMON} clean all && \
    rm -rf /etc/udev/hwdb.bin && \
    rm -rf /var/cache/yum && \
    rm -rf /var/lib/yum/yumdb/* && \
    rm -rf /usr/lib/udev/hwdb.d/* && \
    rm -rf /root/.pki && \
    rm -rf /root/.subversion

EXPOSE 80/TCP 443/TCP

WORKDIR /usr/share/zabbix

VOLUME ["/etc/ssl/apache2"]

COPY ["conf/etc/supervisor/", "/etc/supervisor/"]
COPY ["conf/etc/zabbix/apache.conf", "/etc/zabbix/"]
COPY ["conf/etc/zabbix/apache_ssl.conf", "/etc/zabbix/"]
COPY ["conf/etc/zabbix/web/zabbix.conf.php", "/etc/zabbix/web/"]
COPY ["conf/etc/php.d/99-zabbix.ini", "/etc/php.d/"]
COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENTRYPOINT ["docker-entrypoint.sh"]

CMD ["frontend", "mysql", "apache"]
