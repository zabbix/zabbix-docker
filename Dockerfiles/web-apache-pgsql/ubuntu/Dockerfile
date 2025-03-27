# syntax=docker/dockerfile:1
ARG OS_BASE_IMAGE=ubuntu:noble
ARG MAJOR_VERSION=7.2
ARG ZBX_VERSION=${MAJOR_VERSION}.5
ARG BUILD_BASE_IMAGE=zabbix-build-mysql:ubuntu-${ZBX_VERSION}

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

RUN --mount=type=cache,target=/var/cache/apt/,sharing=locked \
    --mount=type=cache,target=/var/lib/apt/,sharing=locked \
    set -eux && \
    echo "#!/bin/sh\nexit 101" > /usr/sbin/policy-rc.d && \
    INSTALL_PKGS="bash \
            tzdata \
            curl \
            ca-certificates \
            apache2 \ 
            locales \
            libldap-common \
            php8.3-bcmath \
            php8.3-curl \
            php8.3-fpm \
            php8.3-gd \
            php8.3-ldap \
            php8.3-mbstring \
            php8.3-xml \
            php8.3-pgsql \
            postgresql-client \
            supervisor" && \
    apt-get -y update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y \
            -o Dpkg::Options::="--force-confdef" \
            -o Dpkg::Options::="--force-confold" \
            --no-install-recommends install \
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
            --home-dir /var/lib/zabbix/ \
        zabbix && \
    mkdir -p ${ZABBIX_CONF_DIR} && \
    mkdir -p ${ZABBIX_CONF_DIR}/web && \
    mkdir -p ${ZABBIX_CONF_DIR}/web/certs && \
    mkdir -p /var/lib/php/session && \
    find /etc/ -name '*.dpkg-dist' | xargs rm -f && \
    rm -f /etc/apache2/sites-available/* && \
    rm -f /etc/apache2/sites-enabled/* && \
    rm -f /etc/php/8.3/fpm/pool.d/www.conf && \
    rm -f /var/run/apache2/ && \
    cd ${ZABBIX_WWW_ROOT}/ && \
    rm -f conf/zabbix.conf.php conf/maintenance.inc.php conf/zabbix.conf.php.example && \
    rm -rf tests && \
    rm -f locale/add_new_language.sh locale/update_po.sh locale/make_mo.sh && \
    find ${ZABBIX_WWW_ROOT}/locale -name '*.po' | xargs rm -f && \
    find ${ZABBIX_WWW_ROOT}/locale -name '*.sh' | xargs rm -f && \
    ln -s "${ZABBIX_CONF_DIR}/web/zabbix.conf.php" "${ZABBIX_WWW_ROOT}/conf/zabbix.conf.php" && \
    ln -s "${ZABBIX_CONF_DIR}/web/maintenance.inc.php" "${ZABBIX_WWW_ROOT}/conf/maintenance.inc.php" && \
    mkdir -p /var/lib/locales/supported.d/ && \
    rm -f /var/lib/locales/supported.d/local && \
    cat ${ZABBIX_WWW_ROOT}/include/locales.inc.php | grep display | grep true | awk '{$1=$1};1' | \
        cut -d"'" -f 2 | sort | \
        xargs -I '{}' bash -c 'echo "{}.UTF-8 UTF-8" >> /var/lib/locales/supported.d/local' && \
    dpkg-reconfigure locales && \
    chown --quiet -R zabbix:root ${ZABBIX_CONF_DIR}/ ${ZABBIX_WWW_ROOT}/include/defines.inc.php ${ZABBIX_WWW_ROOT}/modules/ && \
    chgrp -R 0 ${ZABBIX_CONF_DIR}/ ${ZABBIX_WWW_ROOT}/include/defines.inc.php ${ZABBIX_WWW_ROOT}/modules/ && \
    chmod -R g=u ${ZABBIX_CONF_DIR}/ ${ZABBIX_WWW_ROOT}/include/defines.inc.php ${ZABBIX_WWW_ROOT}/modules/ && \
    chown --quiet -R zabbix:root /etc/apache2/ /etc/php/8.3/fpm/ && \
    chgrp -R 0 /etc/apache2/ /etc/php/8.3/fpm/ && \
    chmod -R g=u /etc/apache2/ /etc/php/8.3/fpm/ && \
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
