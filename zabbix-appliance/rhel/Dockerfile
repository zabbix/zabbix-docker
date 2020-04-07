FROM registry.access.redhat.com/ubi8/ubi
MAINTAINER Alexey Pustovalov <alexey.pustovalov@zabbix.com>

ARG DNF_FLAGS_COMMON="-y"
ARG DNF_FLAGS_PERSISTENT="${DNF_FLAGS_COMMON}"
ARG DNF_FLAGS_DEV="${DNF_FLAGS_COMMON}"
ENV TERM=xterm MIBDIRS=/usr/share/snmp/mibs:/var/lib/zabbix/mibs MIBS=+ALL \
    MYSQL_ALLOW_EMPTY_PASSWORD=true DB_SERVER_HOST=localhost MYSQL_USER=zabbix

ARG BUILD_DATE
ARG VCS_REF

ARG MAJOR_VERSION=4.4
ARG RELEASE=7
ARG ZBX_VERSION=${MAJOR_VERSION}.7

ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ENV ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} TINI_VERSION=v0.18.0

LABEL name="zabbix/zabbix-appliance" \
      maintainer="alexey.pustovalov@zabbix.com" \
      vendor="Zabbix LLC" \
      version="${MAJOR_VERSION}" \
      release="${RELEASE}" \
      summary="Zabbix appliance with MySQL database support and Nginx web-server" \
      description="Zabbix appliance contains MySQL database server, Zabbix server, Zabbix Java Gateway and Zabbix frontend based on Nginx web-server." \
      url="https://www.zabbix.com/" \
      run="docker run --name zabbix-appliance -p 80:8080 -p 10051:10051 -d registry.connect.redhat.com/zabbix/zabbix-appliance-44:${ZBX_VERSION}" \
      io.k8s.description="Zabbix appliance with MySQL database support and Nginx web-server" \
      io.k8s.display-name="Zabbix Appliance" \
      io.openshift.expose-services="8080:http,8443:https,10051:10051" \
      io.openshift.tags="zabbix,zabbix-appliance,mysql,nginx" \
      org.label-schema.name="zabbix-appliance-rhel" \
      org.label-schema.vendor="Zabbix LLC" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.description="Zabbix appliance with MySQL database support and Nginx web-server" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.license="GPL v2.0" \
      org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.version="${ZBX_VERSION}" \
      org.label-schema.vcs-url="${ZBX_SOURCES}" \
      org.label-schema.docker.cmd="docker run --name zabbix-appliance -p 80:8080 -p 10051:10051 -d registry.connect.redhat.com/zabbix/zabbix-appliance-44:${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY ["conf/etc/yum.repo.d/nginx.repo", "/etc/yum.repos.d/nginx.repo"]
COPY ["licenses", "/licenses"]
COPY ["conf/sbin/tini","/sbin/tini"]

RUN set -o xtrace && INSTALL_PKGS="OpenIPMI-libs \
            curl \
            fping \
#            iksemel \
            java-1.8.0-openjdk-headless \
            pcre \
            libcurl \
            libssh \
            libevent \
            libxml2 \
            mariadb \
            mariadb-server \
            net-snmp-libs \
            nginx \
            openldap \
            openssl-libs \
            pcre \
            php-bcmath \
            php-fpm \
            php-gd \
            php-ldap \
            php-mbstring \
            python3-setuptools \
            php-mysqlnd \
            php-xml \
            unixODBC" && \
    dnf ${DNF_FLAGS_COMMON} install https://repo.zabbix.com/zabbix/${MAJOR_VERSION}/rhel/8/x86_64/zabbix-release-${MAJOR_VERSION}-1.el8.noarch.rpm && \
    REPOLIST="rhel-8-for-x86_64-baseos-rpms,rhel-8-for-x86_64-appstream-rpms,zabbix-non-supported,nginx-stable" && \
    dnf ${DNF_FLAGS_COMMON} update-minimal --disablerepo "*" --enablerepo rhel-8-for-x86_64-baseos-rpms --setopt=tsflags=nodocs \
        --security --sec-severity=Important --sec-severity=Critical && \
    dnf ${DNF_FLAGS_COMMON} install --disablerepo "*" --enablerepo "${REPOLIST}" --setopt=tsflags=nodocs ${INSTALL_PKGS} && \
    groupadd -g 1995 --system zabbix && \
    adduser -r --shell /sbin/nologin \
            -g zabbix -G dialout -G root \
            -d /var/lib/zabbix/ -u 1997 \
        zabbix && \
    mkdir -p /etc/zabbix && \
    mkdir -p /var/lib/zabbix && \
    mkdir -p /usr/lib/zabbix/alertscripts && \
    mkdir -p /var/lib/zabbix/enc && \
    mkdir -p /usr/lib/zabbix/externalscripts && \
    mkdir -p /var/lib/zabbix/mibs && \
    mkdir -p /var/lib/zabbix/modules && \
    mkdir -p /var/lib/zabbix/snmptraps && \
    mkdir -p /var/lib/zabbix/ssh_keys && \
    mkdir -p /var/lib/zabbix/ssl && \
    mkdir -p /var/lib/zabbix/ssl/certs && \
    mkdir -p /var/lib/zabbix/ssl/keys && \
    mkdir -p /var/lib/zabbix/ssl/ssl_ca && \
    mkdir -p /usr/share/zabbix/ && \
    mkdir -p /usr/sbin/zabbix_java/ && \
    mkdir -p /var/lib/php/ && \
    mkdir -p /usr/share/doc/zabbix-server-mysql/ && \
    easy_install-3.6 --install-dir /usr/lib/python3.6/site-packages/ supervisor && \
    ln -s /usr/lib/python3.6/site-packages/supervisord /usr/bin/supervisord && \
    chmod +x /sbin/tini && \
    dnf ${DNF_FLAGS_COMMON} clean all && \
    rm -f /etc/php-fpm.d/www.conf /etc/my.cnf.d/auth_gssapi.cnf && \
    rm -rf /var/cache/yum /var/lib/yum/yumdb/* /usr/lib/udev/hwdb.d/* && \
    rm -rf /var/cache/dnf /etc/udev/hwdb.bin /root/.pki

COPY ["conf/etc/", "/etc/"]

RUN set -eux && REPOLIST="rhel-8-for-x86_64-baseos-rpms,rhel-8-for-x86_64-appstream-rpms,zabbix-non-supported" && \
    INSTALL_PKGS="autoconf \
            automake \
            gcc \
            gettext \
#            iksemel-devel \
            java-1.8.0-openjdk-devel \
            libcurl-devel \
            libevent-devel \
            libssh-devel \
            libxml2-devel \
            make \
            mariadb-devel \
            pcre-devel \
            net-snmp-devel \
#            OpenIPMI-devel \
            openldap-devel \
            git \
            unixODBC-devel" && \
    dnf ${DNF_FLAGS_COMMON} install --disablerepo "*" --enablerepo "${REPOLIST}" --setopt=tsflags=nodocs ${INSTALL_PKGS} && \
    cd /tmp/ && \
    git clone ${ZBX_SOURCES} --branch ${ZBX_VERSION} --depth 1 --single-branch zabbix-${ZBX_VERSION} && \
    cd /tmp/zabbix-${ZBX_VERSION} && \
    zabbix_revision=`git rev-parse --short HEAD` && \
    sed -i "s/{ZABBIX_REVISION}/$zabbix_revision/g" include/version.h && \
    sed -i "s/{ZABBIX_REVISION}/$zabbix_revision/g" include/version.h && \
    sed -i "s/{ZABBIX_REVISION}/$zabbix_revision/g" src/zabbix_java/src/com/zabbix/gateway/GeneralInformation.java && \
    ./bootstrap.sh && \
    export CFLAGS="-fPIC -pie -Wl,-z,relro -Wl,-z,now" && \
    ./configure \
            --datadir=/usr/lib \
            --libdir=/usr/lib/zabbix \
            --prefix=/usr \
            --sysconfdir=/etc/zabbix \
            --enable-agent \
            --enable-server \
            --with-mysql \
#            --with-jabber \
            --with-ldap \
            --with-libcurl \
            --with-libxml2 \
            --enable-java \
            --with-net-snmp \
#            --with-openipmi \
            --with-openssl \
            --with-ssh \
            --with-unixodbc \
            --enable-ipv6 \
            --silent && \
    make -j"$(nproc)" -s dbschema && \
    make -j"$(nproc)" -s && \
    cp src/zabbix_server/zabbix_server /usr/sbin/zabbix_server && \
    cp src/zabbix_get/zabbix_get /usr/bin/zabbix_get && \
    cp src/zabbix_sender/zabbix_sender /usr/bin/zabbix_sender && \
    cp conf/zabbix_server.conf /etc/zabbix/zabbix_server.conf && \
    cat database/mysql/schema.sql > database/mysql/create.sql && \
    cat database/mysql/images.sql >> database/mysql/create.sql && \
    cat database/mysql/data.sql >> database/mysql/create.sql && \
    gzip database/mysql/create.sql && \
    cp database/mysql/create.sql.gz /usr/share/doc/zabbix-server-mysql/ && \
    cp -r src/zabbix_java/bin /usr/sbin/zabbix_java/ && \
    cp -r src/zabbix_java/lib /usr/sbin/zabbix_java/ && \
    rm -rf /usr/sbin/zabbix_java/lib/*.xml && \
    cp -R /tmp/zabbix-${ZBX_VERSION}/frontends/php/* /usr/share/zabbix/ && \
    cd /tmp/ && \
    rm -rf /tmp/zabbix-${ZBX_VERSION}/ && \
    rm -f /usr/share/zabbix/conf/zabbix.conf.php && \
    rm -rf /usr/share/zabbix/tests/ && \
    cd /usr/share/zabbix/ && \
    ./locale/make_mo.sh && \
#    cat /usr/share/zabbix/include/locales.inc.php | grep display | grep true | awk '{$1=$1};1' | \
#                cut -d"'" -f 2 | sort | \
#                xargs -I '{}' bash -c 'echo "{}" && localedef -c -i {} -f UTF-8 {}.UTF-8' && \
    dnf ${DNF_FLAGS_COMMON} history undo `dnf -q history | sed -n 3p |column -t | cut -d' ' -f1` && \
    dnf ${DNF_FLAGS_COMMON} clean all && \
    chown --quiet -R zabbix:root /etc/zabbix/ /var/lib/zabbix/ /usr/share/zabbix/conf/ && \
    chgrp -R 0 /etc/zabbix/ /var/lib/zabbix/ /usr/share/zabbix/conf/ && \
    chmod -R g=u /etc/zabbix/ /var/lib/zabbix/ /usr/share/zabbix/conf/ && \
    chown --quiet -R zabbix:root /etc/nginx/ /etc/my.cnf.d/ /etc/my.cnf /etc/php.d/ && \
    chgrp -R 0 /etc/nginx/ /etc/my.cnf.d/ /etc/my.cnf /etc/php.d/ && \
    chmod -R g=u /etc/nginx/ /etc/my.cnf.d/ /etc/my.cnf /etc/php.d/ && \
    chown --quiet -R zabbix:root /var/lib/php/ /var/lib/mysql/ /var/cache/nginx/ && \
    chgrp -R 0 /var/lib/php/ /var/lib/mysql/ /var/cache/nginx/ && \
    chmod -R g=u /var/lib/php/ /var/lib/mysql/ /var/cache/nginx/ && \
    chown --quiet -R zabbix:root /var/cache/nginx/ /var/log/nginx/ && \
    chgrp -R 0 /var/cache/nginx/ /var/log/nginx/ && \
    chmod -R g=u /var/cache/nginx/ /var/log/nginx/ && \
    rm -rf /var/cache/yum /var/lib/yum/yumdb/* /usr/lib/udev/hwdb.d/* && \
    rm -rf /var/cache/dnf /etc/udev/hwdb.bin /root/.pki

EXPOSE 80/TCP 443/TCP 10051/TCP

WORKDIR /var/lib/zabbix

VOLUME ["/etc/ssl/nginx", "/var/lib/mysql/"]
VOLUME ["/usr/lib/zabbix/alertscripts", "/usr/lib/zabbix/externalscripts", "/var/lib/zabbix/enc", "/var/lib/zabbix/mibs", "/var/lib/zabbix/modules"]
VOLUME ["/var/lib/zabbix/snmptraps", "/var/lib/zabbix/ssh_keys", "/var/lib/zabbix/ssl/certs", "/var/lib/zabbix/ssl/keys", "/var/lib/zabbix/ssl/ssl_ca"]

COPY ["conf/usr/sbin/zabbix_java_gateway", "/usr/sbin/"]
COPY ["docker-entrypoint.sh", "/usr/bin/"]

USER 1997

ENTRYPOINT ["/sbin/tini", "--", "/usr/bin/docker-entrypoint.sh"]
