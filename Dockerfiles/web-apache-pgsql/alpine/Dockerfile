# syntax=docker/dockerfile:1
ARG OS_BASE_IMAGE=alpine:3.21
ARG MAJOR_VERSION=7.2
ARG ZBX_VERSION=${MAJOR_VERSION}.5
ARG BUILD_BASE_IMAGE=zabbix-build-pgsql:alpine-${ZBX_VERSION}

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

RUN set -eux && \
    INSTALL_PKGS="bash \
            tzdata \
            curl \
            apache2-proxy \
            php83-bcmath \
            php83-ctype \
            php83-curl \
            php83-fpm \
            php83-gd \
            php83-gettext \
            php83-json \
            php83-ldap \
            php83-pgsql \
            php83-mbstring \
            php83-session \
            php83-simplexml \
            php83-sockets \
            php83-fileinfo \
            php83-xmlreader \
            php83-xmlwriter \
            php83-openssl \
            postgresql17-client \
            supervisor" && \
    apk add \
            --no-cache \
            --clean-protected \
        ${INSTALL_PKGS} && \
    apk add \
            --clean-protected \
            --no-cache \
            --no-scripts \
        apache2-ssl && \
    addgroup \
            --system \
            --gid 1995 \
        zabbix && \
    adduser \
            --system \
            --gecos "Zabbix monitoring system" \
            --disabled-password \
            --uid 1997 \
            --ingroup zabbix \
            --shell /sbin/nologin \
            --home /var/lib/zabbix/ \
        zabbix && \
    mkdir -p ${ZABBIX_CONF_DIR} && \
    mkdir -p ${ZABBIX_CONF_DIR}/web && \
    mkdir -p ${ZABBIX_CONF_DIR}/web/certs && \
    rm -f "/etc/apache2/conf.d/default.conf" && \
    rm -f "/etc/apache2/conf.d/ssl.conf" && \
    rm -f "/etc/apache2/conf.d/info.conf" && \
    rm -f "/etc/apache2/conf.d/mpm.conf" && \
    rm -f "/etc/apache2/conf.d/proxy.conf" && \
    rm -f "/etc/apache2/conf.d/languages.conf" && \
    rm -f "/etc/apache2/conf.d/userdir.conf" && \
    mkdir -p /var/lib/php/session && \
    rm -rf /etc/php83/php-fpm.d/www.conf && \
    rm -rf "/var/run/apache2/" && \
    cd ${ZABBIX_WWW_ROOT}/ && \
    rm -f conf/zabbix.conf.php conf/maintenance.inc.php conf/zabbix.conf.php.example && \
    rm -rf tests && \
    rm -f locale/add_new_language.sh locale/update_po.sh locale/make_mo.sh && \
    find ${ZABBIX_WWW_ROOT}/locale -name '*.po' | xargs rm -f && \
    find ${ZABBIX_WWW_ROOT}/locale -name '*.sh' | xargs rm -f && \
    ln -s "${ZABBIX_CONF_DIR}/web/zabbix.conf.php" "${ZABBIX_WWW_ROOT}/conf/zabbix.conf.php" && \
    ln -s "${ZABBIX_CONF_DIR}/web/maintenance.inc.php" "${ZABBIX_WWW_ROOT}/conf/maintenance.inc.php" && \
    chown --quiet -R zabbix:root ${ZABBIX_CONF_DIR}/ ${ZABBIX_WWW_ROOT}/include/defines.inc.php ${ZABBIX_WWW_ROOT}/modules/ && \
    chgrp -R 0 ${ZABBIX_CONF_DIR}/ ${ZABBIX_WWW_ROOT}/include/defines.inc.php ${ZABBIX_WWW_ROOT}/modules/ && \
    chmod -R g=u ${ZABBIX_CONF_DIR}/ ${ZABBIX_WWW_ROOT}/include/defines.inc.php ${ZABBIX_WWW_ROOT}/modules/ && \
    chown --quiet -R zabbix:root /etc/apache2/ /etc/php83/php-fpm.d/ /etc/php83/php-fpm.conf && \
    chgrp -R 0 /etc/apache2/ /etc/php83/php-fpm.d/ /etc/php83/php-fpm.conf && \
    chmod -R g=u /etc/apache2/ /etc/php83/php-fpm.d/ /etc/php83/php-fpm.conf && \
    chown --quiet -R zabbix:root /var/lib/php/session/ && \
    chgrp -R 0 /var/lib/php/session/ && \
    chmod -R g=u /var/lib/php/session/

HEALTHCHECK --interval=1m30s --timeout=3s --retries=3 --start-period=40s --start-interval=5s \
    CMD curl -f http://localhost:8080/ping || exit 1

EXPOSE 8080/TCP 8443/TCP

WORKDIR ${ZABBIX_WWW_ROOT}

COPY ["docker-entrypoint.sh", "/usr/bin/"]

USER 1997

ENTRYPOINT ["docker-entrypoint.sh"]
