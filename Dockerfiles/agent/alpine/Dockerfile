# syntax=docker/dockerfile:1
ARG OS_BASE_IMAGE=alpine:3.21
ARG MAJOR_VERSION=7.2
ARG ZBX_VERSION=${MAJOR_VERSION}.5
ARG BUILD_BASE_IMAGE=zabbix-build-mysql:alpine-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} AS builder

FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} \
    ZABBIX_USER_HOME_DIR="/var/lib/zabbix" \
    ZABBIX_CONF_DIR="/etc/zabbix"

ENV ZBX_LOADMODULEPATH="${ZABBIX_USER_HOME_DIR}/modules" \
    ZBX_USERPARAMETERDIR=${ZABBIX_USER_HOME_DIR}/user_scripts

LABEL org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix agent is deployed on a monitoring target to actively monitor local resources and applications" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="AGPL v3.0" \
      org.opencontainers.image.source="${ZBX_SOURCES}" \
      org.opencontainers.image.title="Zabbix agent" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix SIA" \
      org.opencontainers.image.version="${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}-output/agent/sbin/zabbix_agentd", "/usr/sbin/zabbix_agentd"]
COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}-output/general/bin/*", "/usr/bin/"]
COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}-output/agent/conf/", "${ZABBIX_CONF_DIR}/"]

RUN set -eux && \
    INSTALL_PKGS="bash \
            tzdata \
            coreutils \
            iputils \
            pcre2 \
            libgcc \
            libcurl \
            libmodbus \
            libldap \
            zlib" && \
    apk add \
            --no-cache \
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
    mkdir -p ${ZABBIX_CONF_DIR} && \
    mkdir -p ${ZABBIX_CONF_DIR}/zabbix_agentd.d && \
    mkdir -p ${ZABBIX_USER_HOME_DIR} && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/enc && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/enc_internal && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/modules && \
    mkdir -p ${ZBX_USERPARAMETERDIR} && \
    chown --quiet -R zabbix:root ${ZABBIX_CONF_DIR}/zabbix_agentd_modules.conf \
        ${ZABBIX_CONF_DIR}/zabbix_agentd_item_keys.conf \
        ${ZABBIX_CONF_DIR}/zabbix_agentd.d/ \
        ${ZABBIX_USER_HOME_DIR}/ && \
    chown --quiet zabbix:root ${ZABBIX_CONF_DIR}/ && \
    chgrp -R 0 ${ZABBIX_CONF_DIR}/zabbix_agentd_modules.conf \
        ${ZABBIX_CONF_DIR}/zabbix_agentd_item_keys.conf \
        ${ZABBIX_CONF_DIR}/zabbix_agentd.d/ \
        ${ZABBIX_USER_HOME_DIR}/ && \
    chgrp 0 ${ZABBIX_CONF_DIR} && \
    chmod -R g=u ${ZABBIX_CONF_DIR}/zabbix_agentd_modules.conf \
        ${ZABBIX_CONF_DIR}/zabbix_agentd_item_keys.conf \
        ${ZABBIX_CONF_DIR}/zabbix_agentd.d/ \
        ${ZABBIX_USER_HOME_DIR}/ && \
    chmod g=u ${ZABBIX_CONF_DIR} && \
    /usr/sbin/zabbix_agentd -V

EXPOSE 10050/TCP

WORKDIR ${ZABBIX_USER_HOME_DIR}

COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENTRYPOINT ["/usr/bin/docker-entrypoint.sh"]

USER 1997

CMD ["/usr/sbin/zabbix_agentd", "--foreground", "-c", "/etc/zabbix/zabbix_agentd.conf"]
