# syntax=docker/dockerfile:1.6
ARG OS_BASE_IMAGE=alpine:3.23
ARG MAJOR_VERSION=6.0
ARG ZBX_VERSION=${MAJOR_VERSION}.45
ARG BUILD_BASE_IMAGE=zabbix-build-mysql:alpine-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} AS builder

FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
    ZBX_SOURCES=${ZBX_SOURCES} \
    ZABBIX_USER_HOME_DIR=/var/lib/zabbix \
    ZABBIX_CONF_DIR=/etc/zabbix

LABEL org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix agent 2 is deployed on a monitoring target to actively monitor local resources and applications" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="GPL v2.0, Apache v2.0 for plugins" \
      org.opencontainers.image.source="${ZBX_SOURCES}" \
      org.opencontainers.image.title="Zabbix agent 2" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix SIA" \
      org.opencontainers.image.version="${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/agent2/sbin/zabbix_agent2 /usr/sbin/
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/agent2/sbin/zabbix-agent2-plugin/ /usr/sbin/zabbix-agent2-plugin/
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/general/bin/zabbix_get /usr/bin/
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/general/bin/zabbix_sender /usr/bin/
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/agent2/conf/ ${ZABBIX_CONF_DIR}/

RUN set -eux && \
    INSTALL_PKGS="bash \
            tini \
            tzdata \
            pcre2 \
            coreutils \
            smartmontools \
            sudo \
            iputils" && \
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
    echo "zabbix ALL=(root) NOPASSWD: /usr/sbin/smartctl" >> /etc/sudoers.d/zabbix && \
    mkdir -p ${ZABBIX_CONF_DIR} \
        ${ZABBIX_CONF_DIR}/zabbix_agentd.d \
        ${ZABBIX_CONF_DIR}/zabbix_agent2.d \
        ${ZABBIX_USER_HOME_DIR} \
        ${ZABBIX_USER_HOME_DIR}/enc \
        ${ZABBIX_USER_HOME_DIR}/enc_internal \
        ${ZABBIX_USER_HOME_DIR}/buffer \
        ${ZABBIX_USER_HOME_DIR}/user_scripts && \
    FIX_PERM_LIST="\
        ${ZABBIX_CONF_DIR}/ \
        ${ZABBIX_USER_HOME_DIR}/" && \
    chown --quiet -R zabbix:0 ${FIX_PERM_LIST} && \
    chmod -R g=u ${FIX_PERM_LIST} && \
    /usr/sbin/zabbix_agent2 -V

RUN --mount=from=entrypoints,target=/tmp/entrypoints,ro \
    set -eux && \
    install -d /usr/lib/docker-entrypoint && \
    install -m 0755 /tmp/entrypoints/agent2.sh /usr/bin/docker-entrypoint.sh && \
    install -m 0644 /tmp/entrypoints/lib/*.sh /usr/lib/docker-entrypoint/

EXPOSE 10050/tcp 31999/tcp

WORKDIR ${ZABBIX_USER_HOME_DIR}

ENTRYPOINT ["/sbin/tini", "--", "/usr/bin/docker-entrypoint.sh"]

USER 1997

CMD ["/usr/sbin/zabbix_agent2", "--foreground", "-c", "/etc/zabbix/zabbix_agent2.conf"]
