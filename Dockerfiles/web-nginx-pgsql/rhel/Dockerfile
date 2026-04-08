# syntax=docker/dockerfile:1.6
ARG OS_BASE_IMAGE=registry.access.redhat.com/ubi9/ubi-minimal:9.7
ARG MAJOR_VERSION=7.0
ARG RELEASE=25
ARG ZBX_VERSION=${MAJOR_VERSION}.25
ARG BUILD_BASE_IMAGE=zabbix-build-pgsql:rhel-${ZBX_VERSION}

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

ARG OS_BASE_IMAGE=registry.access.redhat.com/ubi9/ubi-minimal:9.7
FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION
ARG RELEASE
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
    ZBX_SOURCES=${ZBX_SOURCES} \
    ZABBIX_CONF_DIR=/etc/zabbix \
    ZABBIX_USER_HOME_DIR=/var/lib/zabbix \
    ZABBIX_WWW_ROOT=/usr/share/zabbix

LABEL description="Zabbix web-interface based on Nginx web server with PostgreSQL database support" \
      maintainer="alexey.pustovalov@zabbix.com" \
      name="zabbix/zabbix-web-pgsql" \
      release="${RELEASE}" \
      run="docker run --name zabbix-web-nginx --link postgresql:pgsql-server --link zabbix-server:zabbix-server -p 80:80 -d registry.connect.redhat.com/zabbix/zabbix-web-pgsql-70:${ZBX_VERSION}" \
      summary="Zabbix web-interface based on Nginx web server with PostgreSQL database support" \
      url="https://www.zabbix.com/" \
      vendor="Zabbix SIA" \
      version="${MAJOR_VERSION}" \
      io.k8s.description="Zabbix web-interface based on Nginx web server with PostgreSQL database support" \
      io.k8s.display-name="Zabbix Frontend (Nginx)" \
      io.openshift.expose-services="8080:http,8443:https" \
      io.openshift.tags="zabbix,zabbix-web,pgsql,nginx" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.description="Zabbix web-interface based on Nginx web server with PostgreSQL database support" \
      org.label-schema.docker.cmd="docker run --name zabbix-web-nginx --link postgresql:pgsql-server --link zabbix-server:zabbix-server -p 80:80 -d registry.connect.redhat.com/zabbix/zabbix-web-pgsql-70:${ZBX_VERSION}" \
      org.label-schema.license="AGPL v3.0" \
      org.label-schema.name="zabbix-web-pgsql-rhel" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.vcs-url="${ZBX_SOURCES}" \
      org.label-schema.vendor="Zabbix SIA" \
      org.label-schema.version="${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY licenses /licenses/
COPY --from=clean_builder /tmp/zabbix-${ZBX_VERSION}/ui ${ZABBIX_WWW_ROOT}/
COPY conf/etc/ /etc/

RUN --mount=type=tmpfs,target=/var/lib/dnf/ \
    --mount=type=tmpfs,target=/var/cache/yum/ \
    set -eux && \
    INSTALL_PKGS="bash \
            curl-minimal \
            jq \
            supervisor \
            shadow-utils \
            findutils \
            glibc-locale-source \
            nginx \
            postgresql \
            php-bcmath \
            php-curl \
            php-fpm \
            php-gd \
            php-json \
            php-ldap \
            php-mbstring \
            php-pgsql \
            php-xml" && \
    curl --tlsv1.2 -sSf -L https://dl.fedoraproject.org/pub/epel/epel-release-latest-9.noarch.rpm -o /tmp/epel-release-latest-9.noarch.rpm && \
    rpm -ivh /tmp/epel-release-latest-9.noarch.rpm && \
    rm -rf /tmp/epel-release-latest-9.noarch.rpm && \
    ARCH_SUFFIX="$(arch)"; \
    microdnf -y module enable \
        --disablerepo "*" \
        --enablerepo "rhel-9-for-$ARCH_SUFFIX-appstream-rpms" \
        --setopt=install_weak_deps=0 \
        --setopt=keepcache=0 \
        php:8.3 \
        postgresql:16 \
        nginx:1.26 && \
    microdnf -y install \
            --disablerepo="*" \
            --enablerepo="ubi-9-baseos-rpms" \
            --enablerepo="ubi-9-appstream-rpms" \
            --enablerepo="rhel-9-for-$ARCH_SUFFIX-baseos-rpms" \
            --enablerepo="rhel-9-for-$ARCH_SUFFIX-appstream-rpms" \
            --enablerepo="epel" \
            --setopt=install_weak_deps=0 \
            --setopt=keepcache=0 \
            --best \
            --setopt=tsflags=nodocs \
        ${INSTALL_PKGS} && \
    microdnf -y update \
            --disableplugin=subscription-manager \
            --disablerepo="*" \
            --enablerepo="ubi-9-baseos-rpms" \
            --setopt=install_weak_deps=0 \
            --best \
            --setopt=tsflags=nodocs \
        tzdata && \
    microdnf -y reinstall \
            --disableplugin=subscription-manager \
            --disablerepo "*" \
            --enablerepo="ubi-9-baseos-rpms" \
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
        ${ZABBIX_USER_HOME_DIR}/ \
        /var/lib/php/session && \
    rm -rf /etc/nginx/conf.d/*.conf \
        /etc/php-fpm.d/www.conf && \
    ln -sf /dev/stdout /var/log/nginx/access.log && \
    ln -sf /dev/stderr /var/log/nginx/error.log && \
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
        /etc/nginx/ \
        /etc/php-fpm.d/ \
        /etc/php-fpm.conf \
        /var/log/nginx/ \
        /var/lib/php/session/" && \
    chown --quiet -R zabbix:0 ${FIX_PERM_LIST} && \
    chmod -R g=u ${FIX_PERM_LIST} && \
    microdnf -y remove \
            findutils \
            glibc-locale-source

HEALTHCHECK --interval=1m30s --timeout=3s --retries=3 --start-period=40s --start-interval=5s \
    CMD curl -f http://localhost:8080/ping || exit 1

EXPOSE 8080/tcp 8443/tcp

WORKDIR ${ZABBIX_WWW_ROOT}

COPY docker-entrypoint.sh /usr/bin/

USER 1997

ENTRYPOINT ["docker-entrypoint.sh"]
