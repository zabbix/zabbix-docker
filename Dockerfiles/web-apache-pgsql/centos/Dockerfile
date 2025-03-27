# syntax=docker/dockerfile:1
ARG OS_BASE_IMAGE=quay.io/centos/centos:stream9-minimal
ARG MAJOR_VERSION=7.2
ARG ZBX_VERSION=${MAJOR_VERSION}.5
ARG BUILD_BASE_IMAGE=zabbix-build-mysql:centos-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} AS builder

FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} \
    ZABBIX_CONF_DIR="/etc/zabbix" \
    ZABBIX_WWW_ROOT="/usr/share/zabbix"

LABEL org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix web-interface based on Apache2 web server with PostgreSQL database support" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="AGPL v3.0" \
      org.opencontainers.image.source="${ZBX_SOURCES}" \
      org.opencontainers.image.title="Zabbix web-interface (Apache, PostgreSQL)" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix SIA" \
      org.opencontainers.image.version="${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}/ui", "${ZABBIX_WWW_ROOT}"]
COPY ["conf/etc/", "/etc/"]

RUN --mount=type=tmpfs,target=/var/lib/dnf/ \
    --mount=type=tmpfs,target=/var/cache/yum/ \
    set -eux && \
    INSTALL_PKGS="bash \
            curl-minimal \
            httpd \
            mod_ssl \
            php \
            php-curl \
            php-fpm \
            php-bcmath \
            php-gd \
            php-ldap \
            php-mbstring \
            php-pgsql \
            php-json \
            php-xml \
            postgresql \
            findutils \
            glibc-locale-source \
            shadow-utils \
            supervisor" && \
    microdnf -y install \
        --disablerepo "*" \
        --enablerepo "extras-common" \
        --setopt=install_weak_deps=0 \
        --setopt=keepcache=0 \
        --best \
        --nodocs epel-release && \
    microdnf -y module enable \
        --disablerepo "*" \
        --enablerepo "appstream" \
        --setopt=install_weak_deps=0 \
        --setopt=keepcache=0 \
        php:8.2 \
        postgresql:16 && \
    microdnf -y install \
        --disablerepo "*" \
        --enablerepo "baseos" \
        --enablerepo "appstream" \
        --enablerepo="epel" \
        --setopt=install_weak_deps=0 \
        --setopt=keepcache=0 \
        --best \
        --nodocs ${INSTALL_PKGS} && \
    microdnf -y reinstall \
            --enablerepo "baseos" \
            --setopt=install_weak_deps=0 \
            --setopt=keepcache=0 \
            --best \
            --setopt=tsflags=nodocs \
        tzdata && \
    groupadd \
            --system \
            --gid 1995 \
        zabbix && \
    useradd \
            --system \
            --comment "Zabbix monitoring system" \
            -g zabbix \
            --uid 1997 \
            --shell /sbin/nologin \
            --home-dir /var/lib/zabbix/ \
        zabbix && \
    mkdir -p ${ZABBIX_CONF_DIR} && \
    mkdir -p ${ZABBIX_CONF_DIR}/web && \
    mkdir -p ${ZABBIX_CONF_DIR}/web/certs && \
    rm -f "/etc/httpd/conf.d/default.conf" && \
    rm -f "/etc/httpd/conf.d/ssl.conf" && \
    rm -f "/etc/httpd/conf.d/autoindex.conf" && \
    rm -f "/etc/httpd/conf.d/php.conf" && \
    rm -f "/etc/httpd/conf.d/userdir.conf" && \
    rm -f "/etc/httpd/conf.d/welcome.conf" && \
    rm -f /etc/php-fpm.d/www.conf && \
    find /etc/ -name '*.rpmnew' | xargs rm -f && \
    cd ${ZABBIX_WWW_ROOT}/ && \
    rm -f conf/zabbix.conf.php conf/maintenance.inc.php conf/zabbix.conf.php.example && \
    rm -rf tests && \
    rm -f locale/add_new_language.sh locale/update_po.sh locale/make_mo.sh && \
    find ${ZABBIX_WWW_ROOT}/locale -name '*.po' | xargs rm -f && \
    find ${ZABBIX_WWW_ROOT}/locale -name '*.sh' | xargs rm -f && \
    ln -s "${ZABBIX_CONF_DIR}/web/zabbix.conf.php" "${ZABBIX_WWW_ROOT}/conf/zabbix.conf.php" && \
    ln -s "${ZABBIX_CONF_DIR}/web/maintenance.inc.php" "${ZABBIX_WWW_ROOT}/conf/maintenance.inc.php" && \
    cat ${ZABBIX_WWW_ROOT}/include/locales.inc.php | grep display | grep true | awk '{$1=$1};1' | \
        cut -d"'" -f 2 | sort | \
        xargs -I '{}' bash -c 'echo "{}" && localedef -c -i {} -f UTF-8 {}.UTF-8 2>/dev/null' && \
    chown --quiet -R zabbix:root ${ZABBIX_CONF_DIR}/ ${ZABBIX_WWW_ROOT}/include/defines.inc.php ${ZABBIX_WWW_ROOT}/modules/ && \
    chgrp -R 0 ${ZABBIX_CONF_DIR}/ ${ZABBIX_WWW_ROOT}/include/defines.inc.php ${ZABBIX_WWW_ROOT}/modules/ && \
    chmod -R g=u ${ZABBIX_CONF_DIR}/ ${ZABBIX_WWW_ROOT}/include/defines.inc.php ${ZABBIX_WWW_ROOT}/modules/ && \
    chown --quiet -R zabbix:root /etc/httpd/ /etc/php-fpm.d/ /etc/php-fpm.conf && \
    chgrp -R 0 /etc/httpd/ /etc/php-fpm.d/ /etc/php-fpm.conf && \
    chmod -R g=u /etc/httpd/ /etc/php-fpm.d/ /etc/php-fpm.conf && \
    chown --quiet -R zabbix:root /run/httpd/ /var/lib/php/session/ && \
    chgrp -R 0 /run/httpd/ /var/lib/php/session/ && \
    chmod -R g=u /run/httpd/ /var/lib/php/session/ && \
    microdnf -y remove \
            findutils \
            glibc-locale-source

HEALTHCHECK --interval=1m30s --timeout=3s --retries=3 --start-period=40s --start-interval=5s \
    CMD curl -f http://localhost:8080/ping || exit `

EXPOSE 8080/TCP 8443/TCP

WORKDIR ${ZABBIX_WWW_ROOT}

COPY ["docker-entrypoint.sh", "/usr/bin/"]

USER 1997

ENTRYPOINT ["docker-entrypoint.sh"]
