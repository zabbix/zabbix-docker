# syntax=docker/dockerfile:1.6
ARG OS_BASE_IMAGE=ubuntu:resolute
ARG MAJOR_VERSION=8.0
ARG ZBX_VERSION=${MAJOR_VERSION}
ARG BUILD_BASE_IMAGE=zabbix-build-pgsql:ubuntu-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} AS builder

FROM builder AS entrypoint-builder

ARG ZBX_VERSION

RUN --mount=from=entrypoints,target=/tmp/entrypoints,ro \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    set -eux && \
    cd /tmp/entrypoints && \
    CGO_ENABLED=0 go build -buildvcs=false -trimpath -ldflags="-s -w -buildid=" \
            -o /tmp/docker-entrypoint \
        ./cmd/web-nginx-pgsql

RUN set -eux && \
    rm -rf \
        "/tmp/zabbix-${ZBX_VERSION}/ui/tests" \
        "/tmp/zabbix-${ZBX_VERSION}/ui/conf/zabbix.conf.php" \
        "/tmp/zabbix-${ZBX_VERSION}/ui/conf/zabbix.conf.php.example" \
        "/tmp/zabbix-${ZBX_VERSION}/ui/conf/maintenance.inc.php" && \
    find "/tmp/zabbix-${ZBX_VERSION}/ui/locale" -name '*.po' -delete && \
    find "/tmp/zabbix-${ZBX_VERSION}/ui/locale" -name '*.sh' -delete

FROM ${OS_BASE_IMAGE}

COPY --chmod=0755 --from=entrypoint-builder /tmp/docker-entrypoint /usr/bin/docker-entrypoint
COPY --from=entrypoints /licenses/go-THIRD-PARTY-NOTICES.txt /licenses/postgresql-THIRD-PARTY-NOTICES.txt /licenses/

ARG MAJOR_VERSION
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
    ZBX_SOURCES=${ZBX_SOURCES} \
    ZABBIX_CONF_DIR=/etc/zabbix \
    ZABBIX_USER_HOME_DIR=/var/lib/zabbix \
    ZABBIX_WWW_ROOT=/usr/share/zabbix \
    PYTHONDONTWRITEBYTECODE=1

ENV DAEMON_USER=www-data \
    DAEMON_GROUP=www-data \
    NGINX_BIN=/usr/sbin/nginx \
    PHP_FPM_BIN=/usr/sbin/php-fpm8.5 \
    NGINX_CONFD_DIR=/etc/nginx/conf.d \
    PHP_FPM_CONFIG=/etc/php/8.5/fpm/php-fpm.conf \
    PHP_ZBX_CONFIG_FILE=/etc/php/8.5/fpm/pool.d/zabbix.conf

LABEL org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix web-interface based on Nginx web server with PostgreSQL database support" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="AGPL-3.0-only" \
      org.opencontainers.image.source="${ZBX_SOURCES}" \
      org.opencontainers.image.title="Zabbix web-interface (Nginx, PostgreSQL)" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix SIA" \
      org.opencontainers.image.version="${ZBX_VERSION}"


COPY --from=entrypoint-builder /tmp/zabbix-${ZBX_VERSION}/ui ${ZABBIX_WWW_ROOT}/
COPY --from=entrypoint-builder /tmp/zabbix-${ZBX_VERSION}-output/web/conf ${ZABBIX_CONF_DIR}/web/
COPY conf/etc/ /etc/

RUN --mount=type=cache,target=/var/cache/apt/,sharing=locked \
    --mount=type=cache,target=/var/lib/apt/,sharing=locked \
    set -eux && \
    INSTALL_PKGS="ca-certificates \
            curl \
            locales \
            tzdata \
            libldap-common \
            nginx \
            php8.5-fpm \
            php8.5-bcmath \
            php8.5-curl \
            php8.5-gd \
            php8.5-ldap \
            php8.5-mbstring \
            php8.5-pgsql \
            php8.5-xml" && \
    apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
            -o Dpkg::Options::="--force-confdef" \
            -o Dpkg::Options::="--force-confold" \
            --no-install-recommends \
        ${INSTALL_PKGS} && \
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
    mkdir -p \
        ${ZABBIX_CONF_DIR} \
        ${ZABBIX_CONF_DIR}/web \
        ${ZABBIX_CONF_DIR}/web/certs \
        ${ZABBIX_USER_HOME_DIR} \
        ${ZABBIX_USER_HOME_DIR}/enc \
        ${ZABBIX_USER_HOME_DIR}/enc_internal \
        /var/lib/php/session \
        /var/lib/locales/supported.d && \
    rm -rf \
        /etc/nginx/conf.d/*.conf \
        /var/cache/nginx/ \
        /etc/php/8.5/fpm/pool.d/www.conf \
        /etc/php/8.5/fpm/php-fpm.conf.dpkg-dist \
        /var/lib/locales/supported.d/local && \
    ln -sf /dev/stdout /var/log/nginx/access.log && \
    ln -sf /dev/stderr /var/log/nginx/error.log && \
    ln -sf "${ZABBIX_CONF_DIR}/web/zabbix.conf.php" "${ZABBIX_WWW_ROOT}/conf/zabbix.conf.php" && \
    ln -sf "${ZABBIX_CONF_DIR}/web/maintenance.inc.php" "${ZABBIX_WWW_ROOT}/conf/maintenance.inc.php" && \
    awk "/display/ && /true/ {gsub(/^[ \t]+|[ \t]+$/, \"\", \$0); print \$0}" "${ZABBIX_WWW_ROOT}/include/locales.inc.php" | \
        cut -d"'" -f 2 | sort -u | \
        xargs -r -I '{}' bash -c 'echo "{}.UTF-8 UTF-8" >> /var/lib/locales/supported.d/local' && \
    dpkg-reconfigure locales && \
    FIX_PERM_LIST="\
        ${ZABBIX_CONF_DIR}/ \
        ${ZABBIX_WWW_ROOT}/modules/ \
        ${ZABBIX_USER_HOME_DIR}/ \
        /etc/nginx/ \
        /etc/php/8.5/fpm/php-fpm.conf \
        /etc/php/8.5/fpm/pool.d/ \
        /var/lib/php/session/" && \
    chown --quiet -R zabbix:0 ${FIX_PERM_LIST} && \
    chmod -R g=u ${FIX_PERM_LIST}

HEALTHCHECK --interval=1m30s --timeout=3s --retries=3 --start-period=40s --start-interval=5s \
    CMD curl -f http://localhost:8080/ping || exit 1

EXPOSE 8080/tcp 8443/tcp

WORKDIR ${ZABBIX_WWW_ROOT}

USER 1997

ENTRYPOINT ["/usr/bin/docker-entrypoint"]
