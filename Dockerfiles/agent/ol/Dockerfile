# syntax=docker/dockerfile:1.6
ARG OS_BASE_IMAGE=oraclelinux:9-slim
ARG MAJOR_VERSION=6.0
ARG ZBX_VERSION=${MAJOR_VERSION}.45
ARG BUILD_BASE_IMAGE=zabbix-build-mysql:ol-${ZBX_VERSION}

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
      org.opencontainers.image.description="Zabbix agent is deployed on a monitoring target to actively monitor local resources and applications" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="GPL v2.0" \
      org.opencontainers.image.source="${ZBX_SOURCES}" \
      org.opencontainers.image.title="Zabbix agent" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix SIA" \
      org.opencontainers.image.version="${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/agent/sbin/zabbix_agentd /usr/sbin/
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/general/bin/zabbix_get /usr/bin/
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/general/bin/zabbix_sender /usr/bin/
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/agent/conf/ ${ZABBIX_CONF_DIR}/
COPY conf/etc/yum.repos.d/oracle-epel-ol9.repo /etc/yum.repos.d/oracle-epel-ol9.repo

RUN --mount=type=tmpfs,target=/var/lib/dnf/ \
    --mount=type=tmpfs,target=/var/cache/yum/ \
    set -eux && \
    INSTALL_PKGS="bash \
            tini \
            tzdata \
            iputils \
            pcre2 \
            libcurl \
            libmodbus \
            openssl-libs \
            zlib" && \
    microdnf -y install \
            --disablerepo="*" \
            --enablerepo="ol9_baseos_latest" \
            --enablerepo="ol9_appstream" \
            --enablerepo="ol9_developer_EPEL" \
            --setopt=install_weak_deps=0 \
            --setopt=keepcache=0 \
            --best \
            --nodocs \
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
    mkdir -p ${ZABBIX_CONF_DIR} \
        ${ZABBIX_CONF_DIR}/zabbix_agentd.d \
        ${ZABBIX_USER_HOME_DIR} \
        ${ZABBIX_USER_HOME_DIR}/enc \
        ${ZABBIX_USER_HOME_DIR}/enc_internal \
        ${ZABBIX_USER_HOME_DIR}/modules \
        ${ZABBIX_USER_HOME_DIR}/user_scripts && \
    FIX_PERM_LIST="\
        ${ZABBIX_CONF_DIR}/ \
        ${ZABBIX_USER_HOME_DIR}/" && \
    chown --quiet -R zabbix:0 ${FIX_PERM_LIST} && \
    chmod -R g=u ${FIX_PERM_LIST} && \
    /usr/sbin/zabbix_agentd -V

EXPOSE 10050/tcp

WORKDIR ${ZABBIX_USER_HOME_DIR}

COPY docker-entrypoint.sh /usr/bin/

ENTRYPOINT ["/usr/bin/tini", "--", "/usr/bin/docker-entrypoint.sh"]

USER 1997

CMD ["/usr/sbin/zabbix_agentd", "--foreground", "-c", "/etc/zabbix/zabbix_agentd.conf"]
