# syntax=docker/dockerfile:1
ARG MAJOR_VERSION=6.4
ARG ZBX_VERSION=${MAJOR_VERSION}.0
ARG BUILD_BASE_IMAGE=zabbix-build-oracle:centos-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} as builder

FROM quay.io/centos/centos:stream9

ARG MAJOR_VERSION
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} \
    LD_LIBRARY_PATH=/usr/lib/oracle/19.18/client64/lib:/usr/lib/oracle/19.18/client64/bin/:$LD_LIBRARY_PATH \
    ORACLE_HOME=/usr/lib/oracle/19.18/client64/ \
    JAVA_HOME=/usr/java/jdk-11.0.16 \
    PATH=/usr/local/sqlcl/bin:$JAVA_HOME/bin:$PATH
LABEL org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix web-interface based on Apache2 web server with Oracle database support" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="GPL v2.0" \
      org.opencontainers.image.source="${ZBX_SOURCES}" \
      org.opencontainers.image.title="Zabbix web-interface (Apache, Oracle)" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix LLC" \
      org.opencontainers.image.version="${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}/ui", "/usr/share/zabbix"]
COPY --from=builder ["/usr/lib/oracle/19.18/client64", "/usr/lib/oracle/19.18/client64"]
COPY --from=builder ["/usr/lib/jvm/", "/usr/lib/jvm/"]
COPY --from=builder ["/usr/share/", "/usr/share/"]
COPY ["conf/etc/", "/etc/"]

RUN set -eux && \
    REPOLIST="baseos,appstream,epel" && \
    INSTALL_PKGS="bash \
            curl-minimal \
            httpd \
            sudo \
            mod_ssl \
            php \
            php-fpm \
            php-bcmath \
            php-gd \
            php-ldap \
            php-mbstring \
            php-json \
            php-xml \
            php-devel \
            php-pear \
            libaio \
            libnsl \
            wget \
            unzip \
            cpio \
            glibc-locale-source \
            supervisor" && \
    dnf -y install epel-release && \
    dnf -y install \
            --disablerepo "*" \
            --enablerepo "${REPOLIST}" \
            --setopt=tsflags=nodocs \
            --setopt=install_weak_deps=False \
            --best \
        ${INSTALL_PKGS} && \
    dnf -y install https://rpms.remirepo.net/enterprise/remi-release-9.rpm && \
    dnf -y module install php:remi-8.1 && \
    dnf -y install php-oci8 && \
    wget -O /basic.rpm https://download.oracle.com/otn_software/linux/instantclient/1918000/oracle-instantclient19.18-basic-19.18.0.0.0-2.x86_64.rpm &&\
    wget -O /sqlplus.rpm https://download.oracle.com/otn_software/linux/instantclient/1918000/oracle-instantclient19.18-sqlplus-19.18.0.0.0-2.x86_64.rpm &&\
    wget -O /devel.rpm https://download.oracle.com/otn_software/linux/instantclient/1918000/oracle-instantclient19.18-devel-19.18.0.0.0-2.x86_64.rpm &&\
    wget -O /odbc.rpm https://download.oracle.com/otn_software/linux/instantclient/1918000/oracle-instantclient19.18-odbc-19.18.0.0.0-2.x86_64.rpm  &&\
    wget -O /jdbc.rpm https://download.oracle.com/otn_software/linux/instantclient/1918000/oracle-instantclient19.18-jdbc-19.18.0.0.0-2.x86_64.rpm &&\
    wget -O /basiclite.rpm https://download.oracle.com/otn_software/linux/instantclient/1918000/oracle-instantclient19.18-basiclite-19.18.0.0.0-2.x86_64.rpm &&\
    wget -O /sqlcl.zip https://download.oracle.com/otn_software/java/sqldeveloper/sqlcl-latest.zip && \
    wget -O /basic.zip https://download.oracle.com/otn_software/linux/instantclient/219000/instantclient-basic-linux.x64-21.9.0.0.0dbru.zip && \
    rpm2cpio basic.rpm | cpio -idmv && \
    rpm2cpio sqlplus.rpm | cpio -idmv && \
    rpm2cpio devel.rpm | cpio -idmv && \
    rpm2cpio odbc.rpm | cpio -idmv && \
    rpm2cpio jdbc.rpm | cpio -idmv && \
    rpm2cpio basiclite.rpm | cpio -idmv && \
    unzip -d /usr/local sqlcl.zip && \
    unzip -d /usr/local basic.zip && \
    cp /usr/local/instantclient_21_9/*.so* /usr/lib/oracle/19.18/client64/lib/ && \
    rm -rf /sqlcl.zip *.rpm /usr/local/instantclient_21_9 && \
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
rm -f "/etc/httpd/conf.d/default.conf" && \
rm -f "/etc/httpd/conf.d/ssl.conf" && \
rm -f /etc/php-fpm.d/www.conf && \
sed -ri \
        -e 's!^(\s*CustomLog)\s+\S+!\1 /proc/self/fd/1!g' \
        -e 's!^(\s*ErrorLog)\s+\S+!\1 /proc/self/fd/2!g' \
    "/etc/httpd/conf/httpd.conf" && \
sed -i 's/Listen 80/Listen 8080/g' /etc/httpd/conf/httpd.conf && \
cd /usr/share/zabbix/ && \
rm -f conf/zabbix.conf.php conf/maintenance.inc.php conf/zabbix.conf.php.example && \
rm -rf tests && \
rm -f locale/add_new_language.sh locale/update_po.sh locale/make_mo.sh && \
find /usr/share/zabbix/locale -name '*.po' | xargs rm -f && \
find /usr/share/zabbix/locale -name '*.sh' | xargs rm -f && \
ln -s "/etc/zabbix/web/zabbix.conf.php" "/usr/share/zabbix/conf/zabbix.conf.php" && \
ln -s "/etc/zabbix/web/maintenance.inc.php" "/usr/share/zabbix/conf/maintenance.inc.php" && \
cat /usr/share/zabbix/include/locales.inc.php | grep display | grep true | awk '{$1=$1};1' | \
    cut -d"'" -f 2 | sort | \
    xargs -I '{}' bash -c 'echo "{}" && localedef -c -i {} -f UTF-8 {}.UTF-8 2>/dev/null' && \
chown --quiet -R zabbix:root /etc/zabbix/ /usr/share/zabbix/include/defines.inc.php /usr/share/zabbix/modules/ && \
chgrp -R 0 /etc/zabbix/ /usr/share/zabbix/include/defines.inc.php /usr/share/zabbix/modules/ && \
chmod -R g=u /etc/zabbix/ /usr/share/zabbix/include/defines.inc.php /usr/share/zabbix/modules/ && \
chown --quiet -R zabbix:root /etc/httpd/ /etc/php-fpm.d/ /etc/php-fpm.conf /usr/share/php/ /usr/share/pear/ /usr/share/ && \
chgrp -R 0 /etc/httpd/ /etc/php-fpm.d/ /etc/php-fpm.conf && \
chmod -R g=u /etc/httpd/ /etc/php-fpm.d/ /etc/php-fpm.conf && \
chown --quiet -R zabbix:root /run/httpd/ /var/lib/php/session/ && \
chgrp -R 0 /run/httpd/ /var/lib/php/session/ && \
chmod -R g=u /run/httpd/ /var/lib/php/session/ && \
dnf -y remove \
    findutils \
    glibc-locale-source && \
dnf -y clean all && \
rm -rf /var/cache/yum /var/lib/yum/yumdb/* /usr/lib/udev/hwdb.d/* && \
rm -rf /var/cache/dnf /etc/udev/hwdb.bin /root/.pki

EXPOSE 8080/TCP 8443/TCP

WORKDIR /usr/share/zabbix

COPY ["docker-entrypoint.sh", "/usr/bin/"]

USER 1997

ENTRYPOINT ["docker-entrypoint.sh"]
