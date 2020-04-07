FROM registry.access.redhat.com/ubi8/ubi
MAINTAINER Alexey Pustovalov <alexey.pustovalov@zabbix.com>

ARG DNF_FLAGS_COMMON="-y"
ARG DNF_FLAGS_PERSISTENT="${DNF_FLAGS_COMMON}"
ARG DNF_FLAGS_DEV="${DNF_FLAGS_COMMON}"
ENV TERM=xterm

ARG BUILD_DATE
ARG VCS_REF

ARG MAJOR_VERSION=4.4
ARG RELEASE=7
ARG ZBX_VERSION=${MAJOR_VERSION}.7

ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ENV ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES}

LABEL name="zabbix/zabbix-web-mysql" \
      maintainer="alexey.pustovalov@zabbix.com" \
      vendor="Zabbix LLC" \
      version="${MAJOR_VERSION}" \
      release="${RELEASE}" \
      summary="Zabbix web-interface based on Nginx web server with MySQL database support" \
      description="Zabbix web-interface based on Nginx web server with MySQL database support" \
      url="https://www.zabbix.com/" \
      run="docker run --name zabbix-web-nginx --link mysql-server:mysql --link zabbix-server:zabbix-server -p 80:80 -d registry.connect.redhat.com/zabbix/zabbix-web-nginx-44:${ZBX_VERSION}" \
      io.k8s.description="Zabbix web-interface based on Nginx web server with MySQL database support" \
      io.k8s.display-name="Zabbix Frontend (Nginx)" \
      io.openshift.expose-services="http:http,https:https" \
      io.openshift.tags="zabbix,zabbix-web,mysql,nginx" \
      org.label-schema.name="zabbix-web-mysql-rhel" \
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
      org.label-schema.docker.cmd="docker run --name zabbix-web-nginx --link mysql-server:mysql --link zabbix-server:zabbix-server -p 80:80 -d registry.connect.redhat.com/zabbix/zabbix-web-nginx-44:${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY ["conf/etc/yum.repo.d/nginx.repo", "/etc/yum.repos.d/nginx.repo"]
            
### add licenses to this directory
COPY ["licenses", "/licenses"]
            
RUN set -eux && INSTALL_PKGS="bash \
            curl \
            mariadb \
            nginx \
            php-bcmath \
            php-fpm \
            php-gd \
            php-ldap \
            php-mbstring \
            python3-setuptools \
            php-mysqlnd \
            php-xml" && \
    REPOLIST="rhel-8-for-x86_64-baseos-rpms,rhel-8-for-x86_64-appstream-rpms,nginx-stable" && \
    dnf ${DNF_FLAGS_COMMON} update-minimal --disablerepo "*" --enablerepo rhel-8-for-x86_64-baseos-rpms --setopt=tsflags=nodocs \
        --security --sec-severity=Important --sec-severity=Critical && \
    dnf ${DNF_FLAGS_COMMON} install --disablerepo "*" --enablerepo "${REPOLIST}" --setopt=tsflags=nodocs ${INSTALL_PKGS} && \
    groupadd -g 1995 --system zabbix && \
    adduser -r --shell /sbin/nologin \
            -g zabbix -G dialout -G root \
            -d /var/lib/zabbix/ -u 1997 \
        zabbix && \
    mkdir -p /etc/zabbix && \
    mkdir -p /etc/zabbix/web && \
    rm -f /etc/php-fpm.d/www.conf && \
    mkdir -p /var/lib/php/ && \
    easy_install-3.6 --install-dir /usr/lib/python3.6/site-packages/ supervisor && \
    ln -s /usr/lib/python3.6/site-packages/supervisord /usr/bin/supervisord && \
    mkdir -p /etc/supervisor/conf.d/ && \
    dnf ${DNF_FLAGS_COMMON} clean all && \
    rm -rf /var/cache/yum /var/lib/yum/yumdb/* /usr/lib/udev/hwdb.d/* && \
    rm -rf /var/cache/dnf /etc/udev/hwdb.bin /root/.pki

COPY ["conf/etc/", "/etc/"]

RUN set -eux && REPOLIST="rhel-8-for-x86_64-baseos-rpms,rhel-8-for-x86_64-appstream-rpms" && \
    INSTALL_PKGS="gettext \
            git" && \
    dnf ${DNF_FLAGS_COMMON} install --disablerepo "*" --enablerepo "${REPOLIST}" --setopt=tsflags=nodocs ${INSTALL_PKGS} && \
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
    dnf ${DNF_FLAGS_COMMON} history undo `dnf -q history | sed -n 3p |column -t | cut -d' ' -f1` && \
    dnf ${DNF_FLAGS_COMMON} clean all && \
    chown --quiet -R zabbix:root /etc/zabbix/ /usr/share/zabbix/conf/ && \
    chgrp -R 0 /etc/zabbix/ /usr/share/zabbix/conf/ && \
    chmod -R g=u /etc/zabbix/ /usr/share/zabbix/conf/ && \
    chown --quiet -R zabbix:root /etc/nginx/ /etc/php.d/ && \
    chgrp -R 0 /etc/nginx/ /etc/php.d/ && \
    chmod -R g=u /etc/nginx/ /etc/php.d/ && \
    chown --quiet -R zabbix:root /var/lib/php/ /var/cache/nginx/ && \
    chgrp -R 0 /var/lib/php/ /var/cache/nginx/ && \
    chmod -R g=u /var/lib/php/ /var/cache/nginx/ && \
    chown --quiet -R zabbix:root /var/cache/nginx/ /var/log/nginx/ && \
    chgrp -R 0 /var/cache/nginx/ /var/log/nginx/ && \
    chmod -R g=u /var/cache/nginx/ /var/log/nginx/ && \
    rm -rf /var/cache/yum /var/lib/yum/yumdb/* /usr/lib/udev/hwdb.d/* && \
    rm -rf /var/cache/dnf /etc/udev/hwdb.bin /root/.pki

EXPOSE 80/TCP 443/TCP

WORKDIR /usr/share/zabbix

VOLUME ["/etc/ssl/nginx"]

COPY ["docker-entrypoint.sh", "/usr/bin/"]

USER 1997

ENTRYPOINT ["docker-entrypoint.sh"]
