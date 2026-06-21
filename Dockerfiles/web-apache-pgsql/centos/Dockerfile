# syntax=docker/dockerfile:1.6
ARG OS_BASE_IMAGE=quay.io/centos/centos:stream10-minimal
ARG MAJOR_VERSION=7.4
ARG ZBX_VERSION=${MAJOR_VERSION}.11
ARG BUILD_BASE_IMAGE=zabbix-build-pgsql:centos-${ZBX_VERSION}

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
    PHP_FPM_BIN=/usr/sbin/php-fpm \
    APACHE_RUN_DIR=/tmp/httpd \
    HTTPD_CONF_FILE=/etc/httpd/conf/httpd.conf \
    APACHE_SITES_DIR=/etc/httpd/conf.d \
    PHP_FPM_CONFIG=/etc/php-fpm.conf \
    PHP_ZBX_CONFIG_FILE=/etc/php-fpm.d/zabbix.conf

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
            postgresql18 \
            findutils \
            glibc-locale-source \
            shadow-utils \
            gzip \
            jq" && \
    microdnf -y install \
            --disablerepo="*" \
            --enablerepo="baseos" \
            --enablerepo="appstream" \
            --setopt=install_weak_deps=0 \
            --setopt=keepcache=0 \
            --best \
          --nodocs \
        ${INSTALL_PKGS} && \
    microdnf -y reinstall \
            --enablerepo="baseos" \
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
            --home-dir ${ZABBIX_USER_HOME_DIR} \
        zabbix && \
    mkdir -p ${ZABBIX_CONF_DIR} \
        ${ZABBIX_CONF_DIR}/web \
        ${ZABBIX_CONF_DIR}/web/certs \
        ${ZABBIX_USER_HOME_DIR}/enc \
        ${ZABBIX_USER_HOME_DIR}/enc_internal \
        /var/lib/php/session && \
    rm -rf /etc/httpd/conf.d/default.conf \
        /etc/httpd/conf.d/ssl.conf \
        /etc/httpd/conf.d/autoindex.conf \
        /etc/httpd/conf.d/php.conf \
        /etc/httpd/conf.d/userdir.conf \
        /etc/httpd/conf.d/welcome.conf \
        /etc/php-fpm.d/www.conf && \
    ln -sf "${ZABBIX_CONF_DIR}/web/zabbix.conf.php" "${ZABBIX_WWW_ROOT}/conf/zabbix.conf.php" && \
    ln -sf "${ZABBIX_CONF_DIR}/web/maintenance.inc.php" "${ZABBIX_WWW_ROOT}/conf/maintenance.inc.php" && \
    awk "/display/ && /true/ {gsub(/^[ \t]+|[ \t]+$/, \"\", \$0); print \$0}" "${ZABBIX_WWW_ROOT}/include/locales.inc.php" | \
        cut -d"'" -f 2 | sort -u | \
        xargs -r -I '{}' bash -c 'echo "{}" && localedef -c -i {} -f UTF-8 {}.UTF-8 2>/dev/null' && \
    FIX_PERM_LIST="\
        ${ZABBIX_CONF_DIR}/ \
        ${ZABBIX_WWW_ROOT}/include/defines.inc.php \
        ${ZABBIX_WWW_ROOT}/modules/ \
        ${ZABBIX_USER_HOME_DIR}/ \
        /etc/httpd/ \
        /etc/php-fpm.d/ \
        /etc/php-fpm.conf \
        /run/httpd/ \
        /var/lib/php/session/" && \
    chown --quiet -R zabbix:0 ${FIX_PERM_LIST} && \
    chmod -R g=u ${FIX_PERM_LIST} && \
    microdnf -y remove \
            findutils \
            glibc-locale-source \
            gzip

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
