FROM registry.access.redhat.com/rhel7
MAINTAINER Alexey Pustovalov <alexey.pustovalov@zabbix.com>

ARG YUM_FLAGS_COMMON="-y"
ARG YUM_FLAGS_PERSISTENT="${YUM_FLAGS_COMMON}"
ARG YUM_FLAGS_DEV="${YUM_FLAGS_COMMON}"
ENV TERM=xterm MIBDIRS=/usr/share/snmp/mibs:/var/lib/zabbix/mibs MIBS=+ALL \
    ZBX_TYPE=server ZBX_DB_TYPE=mysql ZBX_OPT_TYPE=nginx \
    MYSQL_ALLOW_EMPTY_PASSWORD=true ZBX_ADD_SERVER=true ZBX_ADD_WEB=true DB_SERVER_HOST=localhost MYSQL_USER=zabbix ZBX_ADD_JAVA_GATEWAY=true ZBX_JAVAGATEWAY_ENABLE=true ZBX_JAVAGATEWAY=localhost

ARG BUILD_DATE
ARG VCS_REF

ARG MAJOR_VERSION=master
ARG RELEASE=
ARG ZBX_VERSION=${MAJOR_VERSION}.${RELEASE}

ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ENV ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES}

LABEL name="zabbix/zabbix-appliance" \
      maintainer="alexey.pustovalov@zabbix.com" \
      vendor="Zabbix LLC" \
      version="${MAJOR_VERSION}" \
      release="${RELEASE}" \
      summary="Zabbix appliance with MySQL database support and ${ZBX_OPT_TYPE} web-server" \
      description="Zabbix appliance contains MySQL database server, Zabbix server, Zabbix Java Gateway and Zabbix frontend based on Nginx web-server." \
      url="https://www.zabbix.com/" \
      run="docker run --name zabbix-appliance -p 80:80 -p 10051:10051 -d registry.connect.redhat.com/zabbix/zabbix-appliance:${ZBX_VERSION}" \
      io.k8s.description="Zabbix appliance with MySQL database support and ${ZBX_OPT_TYPE} web-server" \
      io.k8s.display-name="Zabbix Appliance" \
      io.openshift.expose-services="http:http,https:https,10051:10051" \
      io.openshift.tags="zabbix,zabbix-appliance,mysql,nginx" \
      org.label-schema.name="zabbix-appliance-rhel" \
      org.label-schema.vendor="Zabbix LLC" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.description="Zabbix appliance with MySQL database support and ${ZBX_OPT_TYPE} web-server" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.license="GPL v2.0" \
      org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.version="${ZBX_VERSION}" \
      org.label-schema.vcs-url="${ZBX_SOURCES}" \
      org.label-schema.docker.cmd="docker run --name zabbix-appliance -p 80:80 -p 10051:10051 -d registry.connect.redhat.com/zabbix/zabbix-appliance:${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY ["conf/etc/yum.repo.d/nginx.repo", "/etc/yum.repos.d/nginx.repo"]

### add licenses to this directory
COPY ["licenses", "/licenses"]

### Add necessary Red Hat repos here
RUN INSTALL_PKGS="OpenIPMI-libs \
            curl \
            fping \
            iksemel \
            java-1.8.0-openjdk-headless \
            libcurl \
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
            python-setuptools \
            php-mysql \
            php-xml \
            unixODBC" && \
    rpm -ivh https://repo.zabbix.com/zabbix/${MAJOR_VERSION}/rhel/7/x86_64/zabbix-release-${MAJOR_VERSION}-2.el7.noarch.rpm && \
    REPOLIST="rhel-7-server-rpms,rhel-7-server-optional-rpms,zabbix-non-supported,nginx" && \
    yum -y update-minimal --disablerepo "*" --enablerepo rhel-7-server-rpms --setopt=tsflags=nodocs \
      --security --sec-severity=Important --sec-severity=Critical && \
    echo ${REPOLIST} && \
    yum -y install --disablerepo "*" --enablerepo "${REPOLIST}" --setopt=tsflags=nodocs ${INSTALL_PKGS} && \
    groupadd --system --gid 1995 zabbix && \
    useradd \
            --system --comment "Zabbix monitoring system" \
            -g zabbix -G root,dialout \
            --uid 1997 \
            --shell /sbin/nologin \
            --home-dir /var/lib/zabbix/ \
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
    chown --quiet -R zabbix:root /var/lib/zabbix && \
    mkdir -p /usr/share/doc/zabbix-${ZBX_TYPE}-${ZBX_DB_TYPE}/ && \
    rm -f /etc/php-fpm.d/www.conf && \
    mkdir -p /var/lib/php/ && \
    chown --quiet -R nginx:nginx /var/lib/php/ && \
    easy_install supervisor && \
    mkdir -p /etc/supervisor/conf.d/ && \
    yum ${YUM_FLAGS_COMMON} clean all && \
    rm -rf /var/cache/yum && \
    rm -rf /var/lib/yum/yumdb/* && \
    rm -rf /usr/lib/udev/hwdb.d/*

RUN REPOLIST="rhel-7-server-rpms,rhel-7-server-optional-rpms,zabbix-non-supported" && \
    INSTALL_PKGS="autoconf \
            automake \
            gcc \
            gettext \
            iksemel-devel \
            java-1.8.0-openjdk-devel \
            libcurl-devel \
            libevent-devel \
            libssh2-devel \
            libxml2-devel \
            make \
            mariadb-devel \
            net-snmp-devel \
            OpenIPMI-devel \
            openldap-devel \
            git \
            unixODBC-devel" && \
    yum -y install --disablerepo "*" --enablerepo "${REPOLIST}" --setopt=tsflags=nodocs ${INSTALL_PKGS} && \
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
            --enable-${ZBX_TYPE} \
            --with-${ZBX_DB_TYPE} \
            --with-jabber \
            --with-ldap \
            --with-libcurl \
            --with-libxml2 \
            --enable-java \
            --with-net-snmp \
            --with-openipmi \
            --with-openssl \
            --with-ssh2 \
            --with-unixodbc \
            --enable-ipv6 \
            --silent && \
    make -j"$(nproc)" -s dbschema && \
    make -j"$(nproc)" -s && \
    cp src/zabbix_${ZBX_TYPE}/zabbix_${ZBX_TYPE} /usr/sbin/zabbix_${ZBX_TYPE} && \
    cp src/zabbix_get/zabbix_get /usr/bin/zabbix_get && \
    cp src/zabbix_sender/zabbix_sender /usr/bin/zabbix_sender && \
    cp conf/zabbix_${ZBX_TYPE}.conf /etc/zabbix/zabbix_${ZBX_TYPE}.conf && \
    chown --quiet -R zabbix:root /etc/zabbix && \
    cat database/${ZBX_DB_TYPE}/schema.sql > database/${ZBX_DB_TYPE}/create.sql && \
    cat database/${ZBX_DB_TYPE}/images.sql >> database/${ZBX_DB_TYPE}/create.sql && \
    cat database/${ZBX_DB_TYPE}/data.sql >> database/${ZBX_DB_TYPE}/create.sql && \
    gzip database/${ZBX_DB_TYPE}/create.sql && \
    cp database/${ZBX_DB_TYPE}/create.sql.gz /usr/share/doc/zabbix-${ZBX_TYPE}-${ZBX_DB_TYPE}/ && \
    mkdir -p /usr/sbin/zabbix_java/ && \
    cp -r src/zabbix_java/bin /usr/sbin/zabbix_java/ && \
    cp -r src/zabbix_java/lib /usr/sbin/zabbix_java/ && \
    rm -rf /usr/sbin/zabbix_java/lib/*.xml && \
    cd /tmp/ && \
    rm -rf /tmp/zabbix-${ZBX_VERSION}/ && \
    cd /usr/share/ && \
    git clone ${ZBX_SOURCES} --branch ${ZBX_VERSION} --depth 1 --single-branch zabbix-${ZBX_VERSION} && \
    mkdir /usr/share/zabbix/ && \
    cp -R /usr/share/zabbix-${ZBX_VERSION}/frontends/php/* /usr/share/zabbix/ && \
    rm -rf /usr/share/zabbix-${ZBX_VERSION}/ && \
    cd /usr/share/zabbix/ && \
    rm -f conf/zabbix.conf.php && \
    rm -rf tests && \
    ./locale/make_mo.sh && \
    yum ${YUM_FLAGS_COMMON} history undo `yum history | sed -n 4p |column -t | cut -d' ' -f1` && \
    yum ${YUM_FLAGS_COMMON} clean all && \
    rm -rf /var/cache/yum && \
    rm -rf /var/lib/yum/yumdb/* && \
    rm -rf /usr/lib/udev/hwdb.d/* && \
    rm -rf /etc/udev/hwdb.bin && \
    rm -rf /root/.pki

EXPOSE 80/TCP 443/TCP 10051/TCP

WORKDIR /var/lib/zabbix

VOLUME ["/etc/ssl/nginx"]
VOLUME ["/usr/lib/zabbix/alertscripts", "/usr/lib/zabbix/externalscripts", "/var/lib/zabbix/enc", "/var/lib/zabbix/mibs", "/var/lib/zabbix/modules"]
VOLUME ["/var/lib/zabbix/snmptraps", "/var/lib/zabbix/ssh_keys", "/var/lib/zabbix/ssl/certs", "/var/lib/zabbix/ssl/keys", "/var/lib/zabbix/ssl/ssl_ca"]
VOLUME ["/var/lib/mysql/"]

COPY ["conf/etc/supervisor/", "/etc/supervisor/"]
COPY ["conf/etc/zabbix/nginx.conf", "/etc/zabbix/"]
COPY ["conf/etc/zabbix/nginx_ssl.conf", "/etc/zabbix/"]
COPY ["conf/etc/zabbix/web/zabbix.conf.php", "/etc/zabbix/web/"]
COPY ["conf/etc/nginx/nginx.conf", "/etc/nginx/"]
COPY ["conf/etc/php-fpm.conf", "/etc/php-fpm.conf"]
COPY ["conf/etc/php.d/99-zabbix.ini", "/etc/php.d/99-zabbix.ini"]
COPY ["conf/etc/zabbix/zabbix_java_gateway_logback.xml", "/etc/zabbix/"]
COPY ["conf/usr/sbin/zabbix_java_gateway", "/usr/sbin/"]
COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENV ZBX_TYPE=appliance

ENTRYPOINT ["docker-entrypoint.sh"]
