# syntax=docker/dockerfile:1.6
ARG OS_BASE_IMAGE=alpine:3.24
ARG MAJOR_VERSION=8.0
ARG ZBX_VERSION=${MAJOR_VERSION}
ARG BUILD_BASE_IMAGE=zabbix-build-pgsql:alpine-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} AS builder

FROM builder AS entrypoint-builder

RUN --mount=from=entrypoints,target=/tmp/entrypoints,ro \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    set -eux && \
    cd /tmp/entrypoints && \
    CGO_ENABLED=0 go build -buildvcs=false -trimpath -ldflags="-s -w -buildid=" \
            -o /tmp/docker-entrypoint \
        ./cmd/server-pgsql

FROM ${OS_BASE_IMAGE}

COPY --chmod=0755 --from=entrypoint-builder /tmp/docker-entrypoint /usr/bin/docker-entrypoint
COPY --from=entrypoints /licenses/go-THIRD-PARTY-NOTICES.txt /licenses/postgresql-THIRD-PARTY-NOTICES.txt /licenses/

ARG MAJOR_VERSION
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ARG ZBX_PRODUCT_REF=${ZBX_VERSION}
ARG ZBX_PRODUCT_REVISION=unknown

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
    ZBX_SOURCES=${ZBX_SOURCES} \
    MIBDIRS=/usr/share/snmp/mibs:/var/lib/zabbix/mibs MIBS=+ALL \
    NMAP_PRIVILEGED= \
    ZABBIX_USER_HOME_DIR=/var/lib/zabbix \
    ZABBIX_CONF_DIR=/etc/zabbix

# Required for runtime control commands executed via container exec
ENV ZBX_DB_NAME=dummy_db_name \
    ZBX_FPINGLOCATION=/usr/sbin/fping \
    ZBX_LOADMODULEPATH=${ZABBIX_USER_HOME_DIR}/modules \
    ZBX_SNMPTRAPPERFILE=${ZABBIX_USER_HOME_DIR}/snmptraps/snmptraps.log \
    ZBX_SSHKEYLOCATION=${ZABBIX_USER_HOME_DIR}/ssh_keys/ \
    ZBX_SSLCERTLOCATION=${ZABBIX_USER_HOME_DIR}/ssl/certs/ \
    ZBX_SSLKEYLOCATION=${ZABBIX_USER_HOME_DIR}/ssl/keys/ \
    ZBX_SSLCALOCATION=${ZABBIX_USER_HOME_DIR}/ssl/ssl_ca/

LABEL org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix server with PostgreSQL database support" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="AGPL-3.0-only" \
      com.zabbix.product.source="${ZBX_SOURCES}" \
      com.zabbix.product.ref="${ZBX_PRODUCT_REF}" \
      com.zabbix.product.revision="${ZBX_PRODUCT_REVISION}" \
      org.opencontainers.image.source="https://github.com/zabbix/zabbix-docker" \
      org.opencontainers.image.title="Zabbix server (PostgreSQL)" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix SIA" \
      org.opencontainers.image.version="${ZBX_VERSION}"


COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/server/sbin/zabbix_server /usr/sbin/zabbix_server
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/general/bin/zabbix_get /usr/bin/
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/general/bin/zabbix_sender /usr/bin/
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/server/conf/ ${ZABBIX_CONF_DIR}/
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/server/database/postgresql/ /usr/share/doc/zabbix-server-postgresql/
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/server/database/clickhouse/ /usr/share/doc/zabbix-server-postgresql/clickhouse/

RUN set -eux && \
    INSTALL_PKGS="bash \
            tzdata \
            c-ares \
            krb5 \
            libcap \
            libcurl \
            libevent \
            libldap \
            libpq \
            libssh \
            libxml2 \
            net-snmp-agent-libs \
            openipmi-libs \
            openssl \
            pcre2 \
            unixodbc \
            zlib \
            fping \
            iputils \
            nmap \
            traceroute" && \
    apk add \
            --no-cache \
            --no-logfile \
            --clean-protected \
        ${INSTALL_PKGS} && \
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
    adduser zabbix dialout && \
    chgrp zabbix /usr/bin/nmap && \
    setcap cap_net_raw+eip /usr/bin/nmap && \
    chmod u+s /usr/sbin/fping && \
    mkdir -p ${ZABBIX_CONF_DIR} \
        ${ZABBIX_USER_HOME_DIR} \
        ${ZABBIX_USER_HOME_DIR}/dbscripts \
        ${ZABBIX_USER_HOME_DIR}/enc \
        ${ZABBIX_USER_HOME_DIR}/enc_internal \
        ${ZABBIX_USER_HOME_DIR}/export \
        ${ZABBIX_USER_HOME_DIR}/mibs \
        ${ZABBIX_USER_HOME_DIR}/modules \
        ${ZABBIX_USER_HOME_DIR}/snmptraps \
        ${ZABBIX_USER_HOME_DIR}/ssh_keys \
        ${ZABBIX_USER_HOME_DIR}/ssl \
        ${ZABBIX_USER_HOME_DIR}/ssl/certs \
        ${ZABBIX_USER_HOME_DIR}/ssl/keys \
        ${ZABBIX_USER_HOME_DIR}/ssl/ssl_ca \
        /usr/lib/zabbix/alertscripts \
        /usr/lib/zabbix/externalscripts \
        /usr/share/doc/zabbix-server-postgresql && \
    FIX_PERM_LIST="\
        ${ZABBIX_CONF_DIR}/zabbix_server_modules.conf \
        ${ZABBIX_USER_HOME_DIR}/" && \
    chown --quiet -R zabbix:0 ${FIX_PERM_LIST} && \
    chmod -R g=u ${FIX_PERM_LIST} && \
    chown --quiet zabbix:0 ${ZABBIX_CONF_DIR}/ && \
    chmod g=u ${ZABBIX_CONF_DIR} && \
    /usr/sbin/zabbix_server -V

EXPOSE 10051/tcp

WORKDIR ${ZABBIX_USER_HOME_DIR}

VOLUME ["${ZABBIX_USER_HOME_DIR}/snmptraps", "${ZABBIX_USER_HOME_DIR}/export"]

USER 1997

ENTRYPOINT ["/usr/bin/docker-entrypoint"]

CMD ["/usr/sbin/zabbix_server", "--foreground", "-c", "/etc/zabbix/zabbix_server.conf"]
