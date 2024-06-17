# syntax=docker/dockerfile:1
ARG MAJOR_VERSION=6.0
ARG ZBX_VERSION=${MAJOR_VERSION}.31
ARG BUILD_BASE_IMAGE=zabbix-build-mysql:centos-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} AS builder

FROM quay.io/centos/centos:stream8

ARG MAJOR_VERSION
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES}

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

COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}-output/agent/sbin/zabbix_agentd", "/usr/sbin/zabbix_agentd"]
COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}-output/general/bin/*", "/usr/bin/"]
COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}-output/agent/conf/", "/etc/zabbix/"]

RUN --mount=type=tmpfs,target=/var/lib/dnf/ \
    set -eux && \
    REPOLIST="baseos,appstream,epel" && \
    INSTALL_PKGS="bash \
            tini \
            iputils \
            pcre2 \
            libcurl-minimal \
            libmodbus \
            openssl-libs \
            zlib" && \
    dnf -y install \
        --disablerepo "*" \
        --enablerepo "extras" \
        --setopt=tsflags=nodocs \
        --setopt=install_weak_deps=False \
        --setopt=keepcache=False \
        --best epel-release && \
    dnf -y install \
        --disablerepo "*" \
        --enablerepo "${REPOLIST}" \
        --setopt=tsflags=nodocs \
        --setopt=install_weak_deps=False \
        --setopt=keepcache=False \
        --best ${INSTALL_PKGS} && \
    dnf -y reinstall \
        --disablerepo "*" \
        --enablerepo "baseos" \
        --setopt=tsflags=nodocs \
        --setopt=install_weak_deps=False \
        --setopt=keepcache=False \
        --best tzdata && \
    groupadd \
            --system \
            --gid 1995 \
        zabbix && \
    useradd \
            --system --comment "Zabbix monitoring system" \
            -g zabbix \
            --uid 1997 \
            --shell /sbin/nologin \
            --home-dir /var/lib/zabbix/ \
        zabbix && \
    mkdir -p /etc/zabbix && \
    mkdir -p /etc/zabbix/zabbix_agentd.d && \
    mkdir -p /var/lib/zabbix && \
    mkdir -p /var/lib/zabbix/enc && \
    mkdir -p /var/lib/zabbix/modules && \
    chown --quiet -R zabbix:root /etc/zabbix/ /var/lib/zabbix/ && \
    chgrp -R 0 /etc/zabbix/ /var/lib/zabbix/ && \
    chmod -R g=u /etc/zabbix/ /var/lib/zabbix/ && \
    dnf -y clean all

EXPOSE 10050/TCP

WORKDIR /var/lib/zabbix

COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENTRYPOINT ["/usr/bin/tini", "--", "/usr/bin/docker-entrypoint.sh"]

USER 1997

CMD ["/usr/sbin/zabbix_agentd", "--foreground", "-c", "/etc/zabbix/zabbix_agentd.conf"]
