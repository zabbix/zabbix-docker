# syntax=docker/dockerfile:1.6
ARG OS_BASE_IMAGE=alpine:3.24
ARG MAJOR_VERSION=7.4
ARG ZBX_VERSION=${MAJOR_VERSION}.13
ARG BUILD_BASE_IMAGE=zabbix-build-common:alpine-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} AS builder
FROM builder AS clean_builder

ARG ZBX_VERSION

RUN set -eux && \
    rm -rf \
        "/tmp/zabbix-${ZBX_VERSION}/ui/tests" \
        "/tmp/zabbix-${ZBX_VERSION}/ui/conf/zabbix.conf.php" \
        "/tmp/zabbix-${ZBX_VERSION}/ui/conf/zabbix.conf.php.example" \
        "/tmp/zabbix-${ZBX_VERSION}/ui/conf/maintenance.inc.php" && \
    find "/tmp/zabbix-${ZBX_VERSION}/ui/locale" -name '*.po' -delete && \
    find "/tmp/zabbix-${ZBX_VERSION}/ui/locale" -name '*.sh' -delete

FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
    ZBX_SOURCES=${ZBX_SOURCES} \
    ZABBIX_CONF_DIR=/etc/zabbix \
    ZABBIX_USER_HOME_DIR=/var/lib/zabbix \
    ZABBIX_WWW_ROOT=/usr/share/zabbix

ENV APACHE_BIN=/usr/sbin/httpd \
    PHP_FPM_BIN=/usr/sbin/php-fpm85 \
    APACHE_RUN_DIR=/tmp/apache2 \
    HTTPD_CONF_FILE=/etc/apache2/httpd.conf \
    APACHE_SITES_DIR=/etc/apache2/conf.d \
    PHP_FPM_CONFIG=/etc/php85/php-fpm.conf \
    PHP_ZBX_CONFIG_FILE=/etc/php85/php-fpm.d/zabbix.conf

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

COPY --from=clean_builder /tmp/zabbix-${ZBX_VERSION}/ui ${ZABBIX_WWW_ROOT}/
COPY --from=clean_builder /tmp/zabbix-${ZBX_VERSION}-output/web/conf ${ZABBIX_CONF_DIR}/web/
COPY conf/etc/ /etc/

RUN set -eux && \
    INSTALL_PKGS="bash \
            curl \
            jq \
            tzdata \
            apache2 \
            php85-fpm \
            apache2-proxy \
            php85-bcmath \
            php85-ctype \
            php85-curl \
            php85-fileinfo \
            php85-gd \
            php85-gettext \
            php85-json \
            php85-ldap \
            php85-mbstring \
            php85-openssl \
            php85-pgsql \
            php85-session \
            php85-simplexml \
            php85-sockets \
            php85-xmlreader \
            php85-xmlwriter \
            postgresql18-client" && \
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
            --home ${ZABBIX_USER_HOME_DIR} \
        zabbix && \
    mkdir -p ${ZABBIX_CONF_DIR} \
        ${ZABBIX_CONF_DIR}/web \
        ${ZABBIX_CONF_DIR}/web/certs \
        ${ZABBIX_USER_HOME_DIR}/enc \
        ${ZABBIX_USER_HOME_DIR}/enc_internal \
        /var/lib/php/session && \
    rm -rf /etc/apache2/conf.d/default.conf \
        /etc/apache2/conf.d/ssl.conf \
        /etc/apache2/conf.d/info.conf \
        /etc/apache2/conf.d/mpm.conf \
        /etc/apache2/conf.d/proxy.conf \
        /etc/apache2/conf.d/languages.conf \
        /etc/apache2/conf.d/userdir.conf \
        /etc/php85/php-fpm.d/www.conf \
        /var/run/apache2/ && \
    ln -sf "${ZABBIX_CONF_DIR}/web/zabbix.conf.php" "${ZABBIX_WWW_ROOT}/conf/zabbix.conf.php" && \
    ln -sf "${ZABBIX_CONF_DIR}/web/maintenance.inc.php" "${ZABBIX_WWW_ROOT}/conf/maintenance.inc.php" && \
    FIX_PERM_LIST="\
        ${ZABBIX_CONF_DIR}/ \
        ${ZABBIX_WWW_ROOT}/include/defines.inc.php \
        ${ZABBIX_WWW_ROOT}/modules/ \
        ${ZABBIX_USER_HOME_DIR} \
        /etc/apache2/ \
        /etc/php85/php-fpm.d/ \
        /etc/php85/php-fpm.conf \
        /var/lib/php/session/" && \
    chown --quiet -R zabbix:0 ${FIX_PERM_LIST} && \
    chmod -R g=u ${FIX_PERM_LIST}

RUN --mount=from=entrypoints,target=/tmp/entrypoints,ro \
    set -eux && \
    install -d /usr/lib/docker-entrypoint && \
    install -m 0755 /tmp/entrypoints/web-apache-pgsql.sh /usr/bin/docker-entrypoint.sh && \
    install -m 0644 /tmp/entrypoints/lib/*.sh /usr/lib/docker-entrypoint/

HEALTHCHECK --interval=1m30s --timeout=3s --retries=3 --start-period=40s --start-interval=5s \
    CMD curl -f http://localhost:8080/ping || exit 1

EXPOSE 8080/tcp 8443/tcp

WORKDIR ${ZABBIX_WWW_ROOT}

USER 1997

ENTRYPOINT ["docker-entrypoint.sh"]
