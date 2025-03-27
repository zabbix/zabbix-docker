# syntax=docker/dockerfile:1
ARG OS_BASE_IMAGE=registry.access.redhat.com/ubi9/ubi-minimal:9.5
ARG MAJOR_VERSION=7.2
ARG RELEASE=5
ARG ZBX_VERSION=${MAJOR_VERSION}.5
ARG BUILD_BASE_IMAGE=zabbix-build-mysql:rhel-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} AS builder

FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION
ARG RELEASE
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} \
    ZABBIX_USER_HOME_DIR="/var/lib/zabbix" \
    ZABBIX_CONF_DIR="/etc/zabbix"

ENV ZBX_PERSISTENTBUFFERFILE="$ZABBIX_USER_HOME_DIR/buffer/agent2.db" \
    ZBX_USERPARAMETERDIR=${ZABBIX_USER_HOME_DIR}/user_scripts

LABEL description="Zabbix agent 2 is deployed on a monitoring target to actively monitor local resources and applications" \
      maintainer="alexey.pustovalov@zabbix.com" \
      name="zabbix/zabbix-agent2-72" \
      release="${RELEASE}" \
      run="docker run --name zabbix-agent2 --link zabbix-server:zabbix-server -p 10050:10050 -d registry.connect.redhat.com/zabbix/zabbix-agent2-72:${ZBX_VERSION}" \
      summary="Zabbix agent" \
      url="https://www.zabbix.com/" \
      vendor="Zabbix SIA" \
      version="${MAJOR_VERSION}" \
      io.k8s.description="Zabbix agent 2 is deployed on a monitoring target to actively monitor local resources and applications" \
      io.k8s.display-name="Zabbix Agent 2" \
      io.openshift.expose-services="10050:10050" \
      io.openshift.tags="zabbix,zabbix-agent" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.description="Zabbix agent 2 is deployed on a monitoring target to actively monitor local resources and applications" \
      org.label-schema.docker.cmd="docker run --name zabbix-agent2 --link zabbix-server:zabbix-server -p 10050:10050 -d registry.connect.redhat.com/zabbix/zabbix-agent2-72:${ZBX_VERSION}" \
      org.label-schema.license="AGPL v3.0" \
      org.label-schema.name="zabbix-agent2-rhel" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.vcs-url="${ZBX_SOURCES}" \
      org.label-schema.vendor="Zabbix SIA" \
      org.label-schema.version="${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY ["licenses", "/licenses"]
COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}-output/agent2/sbin/", "/usr/sbin/"]
COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}-output/general/bin/*", "/usr/bin/"]
COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}-output/agent2/conf/", "${ZABBIX_CONF_DIR}/"]

RUN --mount=type=tmpfs,target=/var/lib/dnf/ \
    --mount=type=tmpfs,target=/var/cache/yum/ \
    set -eux && \
    INSTALL_PKGS="bash \
            iputils \
            shadow-utils \
            pcre2 \
            smartmontools \
            sudo \
            libcurl-minimal" && \
    microdnf -y install \
            --disableplugin=subscription-manager \
            --disablerepo "*" \
            --enablerepo "ubi-9-baseos-rpms" \
            --setopt=install_weak_deps=0 \
            --setopt=keepcache=0 \
            --best \
            --setopt=tsflags=nodocs \
        ${INSTALL_PKGS} && \
    microdnf -y update \
            --disableplugin=subscription-manager \
            --disablerepo "*" \
            --enablerepo "ubi-9-baseos-rpms" \
            --setopt=install_weak_deps=0 \
            --best \
            --setopt=tsflags=nodocs \
        tzdata && \
    microdnf -y reinstall \
            --disableplugin=subscription-manager \
            --disablerepo "*" \
            --enablerepo "ubi-9-baseos-rpms" \
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
    echo "zabbix ALL=(root) NOPASSWD: /usr/sbin/smartctl" >> /etc/sudoers.d/zabbix && \
    mkdir -p ${ZABBIX_CONF_DIR} && \
    mkdir -p ${ZABBIX_CONF_DIR}/zabbix_agentd.d && \
    mkdir -p ${ZABBIX_CONF_DIR}/zabbix_agent2.d && \
    mkdir -p ${ZABBIX_USER_HOME_DIR} && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/enc && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/enc_internal && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/buffer && \
    mkdir -p ${ZBX_USERPARAMETERDIR} && \
    chown --quiet -R zabbix:root ${ZABBIX_CONF_DIR}/zabbix_agent2_item_keys.conf \
        ${ZABBIX_CONF_DIR}/zabbix_agentd.d/ \
        ${ZABBIX_CONF_DIR}/zabbix_agent2.d/ \
        ${ZABBIX_USER_HOME_DIR}/ && \
    chown --quiet zabbix:root ${ZABBIX_CONF_DIR}/ && \
    chgrp -R 0 ${ZABBIX_CONF_DIR}/zabbix_agent2_item_keys.conf \
        ${ZABBIX_CONF_DIR}/zabbix_agentd.d/ \
        ${ZABBIX_CONF_DIR}/zabbix_agent2.d/ \
        ${ZABBIX_USER_HOME_DIR}/ && \
    chgrp 0 ${ZABBIX_CONF_DIR} && \
    chmod -R g=u ${ZABBIX_CONF_DIR}/zabbix_agent2_item_keys.conf \
        ${ZABBIX_CONF_DIR}/zabbix_agentd.d/ \
        ${ZABBIX_CONF_DIR}/zabbix_agent2.d/ \
        ${ZABBIX_USER_HOME_DIR}/ && \
    chmod g=u ${ZABBIX_CONF_DIR} && \
    /usr/sbin/zabbix_agent2 -V

EXPOSE 10050/TCP 31999/TCP

WORKDIR ${ZABBIX_USER_HOME_DIR}

COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENTRYPOINT ["/usr/bin/docker-entrypoint.sh"]

USER 1997

CMD ["/usr/sbin/zabbix_agent2", "--foreground", "-c", "/etc/zabbix/zabbix_agent2.conf"]
