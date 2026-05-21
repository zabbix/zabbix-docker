# syntax=docker/dockerfile:1.6
ARG OS_BASE_IMAGE=registry.access.redhat.com/ubi10/ubi-minimal:10.2
ARG MAJOR_VERSION=8.0
ARG RELEASE=0
ARG ZBX_VERSION=${MAJOR_VERSION}
ARG BUILD_BASE_IMAGE=zabbix-build-mysql:rhel-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} AS builder

FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION
ARG RELEASE
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
    ZBX_SOURCES=${ZBX_SOURCES} \
    ZABBIX_USER_HOME_DIR=/var/lib/zabbix \
    ZABBIX_CONF_DIR=/etc/zabbix

ENV ZBX_LOADMODULEPATH=${ZABBIX_USER_HOME_DIR}/modules \
    ZBX_USERPARAMETERDIR=${ZABBIX_USER_HOME_DIR}/user_scripts

LABEL description="Zabbix agent is deployed on a monitoring target to actively monitor local resources and applications" \
      maintainer="alexey.pustovalov@zabbix.com" \
      name="zabbix/zabbix-agent-trunk" \
      release="${RELEASE}" \
      run="podman run --name zabbix-agent -p 10050:10050 -d registry.connect.redhat.com/zabbix/zabbix-agent-trunk:${ZBX_VERSION}" \
      summary="Zabbix agent" \
      url="https://www.zabbix.com/" \
      vendor="Zabbix SIA" \
      version="${MAJOR_VERSION}" \
      io.k8s.description="Zabbix agent is deployed on a monitoring target to actively monitor local resources and applications" \
      io.k8s.display-name="Zabbix Agent" \
      io.openshift.expose-services="10050:10050" \
      io.openshift.tags="zabbix,zabbix-agent" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.description="Zabbix agent is deployed on a monitoring target to actively monitor local resources and applications" \
      org.label-schema.docker.cmd="podman run --name zabbix-agent -p 10050:10050 -d registry.connect.redhat.com/zabbix/zabbix-agent-trunk:${ZBX_VERSION}" \
      org.label-schema.license="AGPL v3.0" \
      org.label-schema.name="zabbix-agent-rhel" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.vcs-url="${ZBX_SOURCES}" \
      org.label-schema.vendor="Zabbix SIA" \
      org.label-schema.version="${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY licenses /licenses/
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/agent/sbin/zabbix_agentd /usr/sbin/
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/general/bin/zabbix_get /usr/bin/
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/general/bin/zabbix_sender /usr/bin/
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/agent/conf/ ${ZABBIX_CONF_DIR}/

RUN --mount=type=tmpfs,target=/var/lib/dnf/ \
    --mount=type=tmpfs,target=/var/cache/yum/ \
    set -eux && \
    INSTALL_PKGS="bash \
            iputils \
            shadow-utils \
            c-ares \
            pcre2 \
            libmodbus \
            libcurl-minimal \
            openldap \
            zlib" && \
    curl --tlsv1.2 -sSf -L https://dl.fedoraproject.org/pub/epel/epel-release-latest-10.noarch.rpm -o /tmp/epel-release-latest-10.noarch.rpm && \
    rpm -ivh /tmp/epel-release-latest-10.noarch.rpm && \
    rm -rf /tmp/epel-release-latest-10.noarch.rpm && \
    ARCH_SUFFIX="$(arch)"; \
    microdnf -y install \
            --disablerepo="*" \
            --enablerepo="ubi-10-baseos-rpms" \
            --enablerepo="rhel-10-for-$ARCH_SUFFIX-baseos-rpms" \
            --enablerepo="epel" \
            --setopt=install_weak_deps=0 \
            --setopt=keepcache=0 \
            --best \
            --setopt=tsflags=nodocs \
        ${INSTALL_PKGS} && \
    microdnf -y update \
            --disableplugin=subscription-manager \
            --disablerepo="*" \
            --enablerepo="ubi-10-baseos-rpms" \
            --setopt=install_weak_deps=0 \
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
        ${ZABBIX_CONF_DIR}/zabbix_agentd.d \
        ${ZABBIX_USER_HOME_DIR} \
        ${ZABBIX_USER_HOME_DIR}/enc \
        ${ZABBIX_USER_HOME_DIR}/enc_internal \
        ${ZABBIX_USER_HOME_DIR}/modules \
        ${ZBX_USERPARAMETERDIR} && \
    FIX_PERM_LIST="\
        ${ZABBIX_CONF_DIR}/zabbix_agentd_modules.conf \
        ${ZABBIX_CONF_DIR}/zabbix_agentd_item_keys.conf \
        ${ZABBIX_CONF_DIR}/zabbix_agentd.d/ \
        ${ZABBIX_USER_HOME_DIR}/" && \
    chown --quiet -R zabbix:0 ${FIX_PERM_LIST} && \
    chmod -R g=u ${FIX_PERM_LIST} && \
    chown --quiet zabbix:0 ${ZABBIX_CONF_DIR}/ && \
    chmod g=u ${ZABBIX_CONF_DIR} && \
    /usr/sbin/zabbix_agentd -V

RUN --mount=from=entrypoints,target=/tmp/entrypoints,ro \
    set -eux && \
    install -d /usr/lib/docker-entrypoint && \
    install -m 0755 /tmp/entrypoints/agent.sh /usr/bin/docker-entrypoint.sh && \
    install -m 0644 /tmp/entrypoints/lib/*.sh /usr/lib/docker-entrypoint/

EXPOSE 10050/tcp

WORKDIR ${ZABBIX_USER_HOME_DIR}

ENTRYPOINT ["/usr/bin/docker-entrypoint.sh"]

USER 1997

CMD ["/usr/sbin/zabbix_agentd", "--foreground", "-c", "/etc/zabbix/zabbix_agentd.conf"]
