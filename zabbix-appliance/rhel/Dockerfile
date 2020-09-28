FROM registry.access.redhat.com/ubi8/ubi
MAINTAINER Alexey Pustovalov <alexey.pustovalov@zabbix.com>

ARG MAJOR_VERSION=5.0
ARG RELEASE=4
ARG ZBX_VERSION=${MAJOR_VERSION}.4

ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ENV TERM=xterm ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} \
    MIBDIRS=/usr/share/snmp/mibs:/var/lib/zabbix/mibs MIBS=+ALL
ENV TINI_VERSION=v0.19.0

LABEL name="zabbix/zabbix-appliance" \
      maintainer="alexey.pustovalov@zabbix.com" \
      vendor="Zabbix LLC" \
      version="${MAJOR_VERSION}" \
      release="${RELEASE}" \
      summary="Zabbix appliance with MySQL database support and Nginx web-server" \
      description="Zabbix appliance contains MySQL database server, Zabbix server, Zabbix Java Gateway and Zabbix frontend based on Nginx web-server." \
      url="https://www.zabbix.com/" \
      run="docker run --name zabbix-appliance -p 80:8080 -p 10051:10051 -d registry.connect.redhat.com/zabbix/zabbix-appliance-50:${ZBX_VERSION}" \
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
      org.label-schema.docker.cmd="docker run --name zabbix-appliance -p 80:8080 -p 10051:10051 -d registry.connect.redhat.com/zabbix/zabbix-appliance-50:${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY ["conf/etc/yum.repo.d/nginx.repo", "/etc/yum.repos.d/nginx.repo"]
COPY ["licenses", "/licenses"]

RUN set -o xtrace && INSTALL_PKGS="OpenIPMI-libs \
            tzdata \
            curl \
            fping \
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
            php-json \
            php-ldap \
            php-mbstring \
            php-mysqlnd \
            php-xml \
            python3-pip \
            unixODBC" && \
    REPOLIST="rhel-8-for-x86_64-baseos-rpms,rhel-8-for-x86_64-appstream-rpms,epel,nginx-stable" && \
    dnf -y update-minimal --disablerepo "*" --enablerepo ubi-8-baseos --setopt=tsflags=nodocs \
        --security --sec-severity=Important --sec-severity=Critical && \
    dnf -y install https://dl.fedoraproject.org/pub/epel/epel-release-latest-8.noarch.rpm && \
    dnf -y install --disablerepo "*" --enablerepo "${REPOLIST}" --setopt=install_weak_deps=False --best \
        --setopt=tsflags=nodocs ${INSTALL_PKGS} && \
    pip3 install supervisor && \
    ln -s /usr/local/bin/supervisord /usr/bin/supervisord && \
    groupadd -g 1995 --system zabbix && \
    adduser -r --shell /sbin/nologin \
            -g zabbix -G dialout -G root \
            -d /var/lib/zabbix/ -u 1997 \
        zabbix && \
    mkdir -p /etc/zabbix && \
    mkdir -p /etc/zabbix/web && \
    mkdir -p /etc/zabbix/web/certs && \
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
    mkdir -p /usr/sbin/zabbix_java/ext_lib/ && \
    mkdir -p /var/lib/php/ && \
    mkdir -p /usr/share/doc/zabbix-server-mysql/ && \
    curl -L https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini -o /sbin/tini && \
    curl -L https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini.asc -o /tmp/tini.asc && \
    export GNUPGHOME="$(mktemp -d)" && \
    for server in $(shuf -e ha.pool.sks-keyservers.net \
                            hkp://p80.pool.sks-keyservers.net:80 \
                            ipv4.pool.sks-keyservers.net \
                            keyserver.ubuntu.com \
                            keyserver.pgp.com \
                            pgp.mit.edu) ; do \
        gpg --keyserver "$server" --recv-keys 595E85A6B1B4779EA4DAAEC70B588DFF0527A9B7 && break || : ; \
    done && \
    gpg --batch --verify /tmp/tini.asc /sbin/tini && \
    rm -r "$GNUPGHOME" /tmp/tini.asc && \
    chmod +x /sbin/tini && \
    dnf -y clean all && \
    rm -f /etc/php-fpm.d/www.conf /etc/nginx/conf.d/*.conf /etc/my.cnf.d/auth_gssapi.cnf && \
    rm -rf /var/cache/yum /var/lib/yum/yumdb/* /usr/lib/udev/hwdb.d/* && \
    rm -rf /var/cache/dnf /etc/udev/hwdb.bin /root/.pki

COPY ["conf/etc/", "/etc/"]

RUN set -eux && REPOLIST="rhel-8-for-x86_64-baseos-rpms,rhel-8-for-x86_64-appstream-rpms" && \
    INSTALL_PKGS="autoconf \
            automake \
            gcc \
            gettext \
            glibc-locale-source \
            java-1.8.0-openjdk-devel \
            libcurl-devel \
            libevent-devel \
            libssh-devel \
            libxml2-devel \
            make \
            mariadb-connector-c-devel \
            pcre-devel \
            net-snmp-devel \
#            OpenIPMI-devel \
            openldap-devel \
            git \
            unixODBC-devel" && \
    dnf -y install --disablerepo "*" --enablerepo "${REPOLIST}" --setopt=install_weak_deps=False --best \
        --setopt=tsflags=nodocs ${INSTALL_PKGS} && \
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
    cp -R /tmp/zabbix-${ZBX_VERSION}/ui/* /usr/share/zabbix/ && \
    cd /tmp/ && \
    rm -rf /tmp/zabbix-${ZBX_VERSION}/ && \
    rm -f /usr/share/zabbix/conf/zabbix.conf.php && \
    rm -f /usr/share/zabbix/conf/maintenance.inc.php && \
    rm -rf /usr/share/zabbix/tests/ && \
    cd /usr/share/zabbix/ && \
    ./locale/make_mo.sh && \
    ln -s "/etc/zabbix/web/zabbix.conf.php" "/usr/share/zabbix/conf/zabbix.conf.php" && \
    ln -s "/etc/zabbix/web/maintenance.inc.php" "/usr/share/zabbix/conf/maintenance.inc.php" && \
    cat /usr/share/zabbix/include/locales.inc.php | grep display | grep true | awk '{$1=$1};1' | \
                cut -d"'" -f 2 | sort | \
                xargs -I '{}' bash -c 'echo "{}" && localedef -c -i {} -f UTF-8 {}.UTF-8 2>/dev/null' && \
    chown --quiet -R zabbix:root /etc/zabbix/ /var/lib/zabbix/ /usr/share/zabbix/conf/ && \
    chgrp -R 0 /etc/zabbix/ /var/lib/zabbix/ /usr/share/zabbix/conf/ && \
    chmod -R g=u /etc/zabbix/ /var/lib/zabbix/ /usr/share/zabbix/conf/ && \
    chown --quiet -R zabbix:root /etc/nginx/ /etc/my.cnf.d/ /etc/my.cnf /etc/php-fpm.d/ /etc/php-fpm.conf && \
    chgrp -R 0 /etc/nginx/ /etc/my.cnf.d/ /etc/my.cnf /etc/php-fpm.d/ /etc/php-fpm.conf && \
    chmod -R g=u /etc/nginx/ /etc/my.cnf.d/ /etc/my.cnf /etc/php-fpm.d/ /etc/php-fpm.conf && \
    chown --quiet -R zabbix:root /var/lib/mysql/ /var/lib/php/session/ && \
    chgrp -R 0 /var/lib/mysql/ /var/lib/php/session/ && \
    chmod -R g=u /var/lib/mysql/ /var/lib/php/session/ && \
    chown --quiet -R zabbix:root /usr/share/zabbix/include/defines.inc.php && \
    chgrp -R 0 /usr/share/zabbix/include/defines.inc.php && \
    chmod -R g=u /usr/share/zabbix/include/defines.inc.php && \
    dnf -y history undo `dnf history list last -q | sed -n 3p |column -t | cut -d' ' -f1` && \
    dnf -y clean all && \
    rm -rf /var/cache/yum /var/lib/yum/yumdb/* /usr/lib/udev/hwdb.d/* && \
    rm -rf /var/cache/dnf /etc/udev/hwdb.bin /root/.pki

EXPOSE 8080/TCP 8443/TCP 10051/TCP

WORKDIR /var/lib/zabbix

VOLUME ["/var/lib/zabbix/snmptraps", "/var/lib/zabbix/export"]

COPY ["conf/usr/sbin/zabbix_java_gateway", "/usr/sbin/"]
COPY ["docker-entrypoint.sh", "/usr/bin/"]

USER 1997

ENTRYPOINT ["/sbin/tini", "--", "/usr/bin/docker-entrypoint.sh"]
