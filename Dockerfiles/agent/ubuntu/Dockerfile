# syntax=docker/dockerfile:1
ARG MAJOR_VERSION=4.0
ARG ZBX_VERSION=${MAJOR_VERSION}.38
ARG BUILD_BASE_IMAGE=zabbix-build-mysql:ubuntu-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} as builder

FROM ubuntu:bionic

ARG MAJOR_VERSION
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} \
    TINI_VERSION=v0.19.0

LABEL org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix agent is deployed on a monitoring target to actively monitor local resources and applications" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="GPL v2.0" \
      org.opencontainers.image.source="${ZBX_SOURCES}" \
      org.opencontainers.image.title="Zabbix agent" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix LLC" \
      org.opencontainers.image.version="${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}/src/zabbix_agent/zabbix_agentd", "/usr/sbin/zabbix_agentd"]
COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}/src/zabbix_get/zabbix_get", "/usr/bin/zabbix_get"]
COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}/src/zabbix_sender/zabbix_sender", "/usr/bin/zabbix_sender"]
COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}/conf/zabbix_agentd.conf", "/etc/zabbix/zabbix_agentd.conf"]

RUN set -eux && \
    ARCH_SUFFIX="$(arch)"; \
    case "$ARCH_SUFFIX" in \
        i686) export ARCH_SUFFIX='i386' ;; \
        x86_64) [ -f /lib/x86_64-linux-gnu/ld-linux-x86-64.so.2 ] && export ARCH_SUFFIX='amd64' || export ARCH_SUFFIX='i386' ;; \
        aarch64) export ARCH_SUFFIX='arm64' ;; \
        armv7l) export ARCH_SUFFIX='armhf' ;; \
        ppc64el|ppc64le) export ARCH_SUFFIX='ppc64le' ;; \
        s390x) export ARCH_SUFFIX='s390x' ;; \
        *) echo "Unknown ARCH_SUFFIX=${ARCH_SUFFIX-}"; exit 1 ;; \
    esac; \
    echo "#!/bin/sh\nexit 101" > /usr/sbin/policy-rc.d && \
    INSTALL_PKGS="bash \
            tzdata \
            ca-certificates \
            libssl1.1 \
            libcurl4 \
            libldap-2.4" && \
    apt-get -y update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y \
            --no-install-recommends install \
        ${INSTALL_PKGS} && \
    INSTALL_TEMP_PKGS="gpg \
            curl \
            ca-certificates \
            dirmngr \
            gpg-agent" && \
    DEBIAN_FRONTEND=noninteractive apt-get -y \
            --no-install-recommends install \
        ${INSTALL_TEMP_PKGS} && \
    curl -L "https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini-${ARCH_SUFFIX}" -o /usr/bin/tini && \
    curl -L "https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini-${ARCH_SUFFIX}.asc" -o /tmp/tini.asc && \
    export GNUPGHOME="$(mktemp -d)" && \
    for server in $(shuf -e ha.pool.sks-keyservers.net \
                            hkp://p80.pool.sks-keyservers.net:80 \
                            ipv4.pool.sks-keyservers.net \
                            keyserver.ubuntu.com \
                            keyserver.pgp.com \
                            pgp.mit.edu) ; do \
        gpg --keyserver "$server" --recv-keys 595E85A6B1B4779EA4DAAEC70B588DFF0527A9B7 && break || : ; \
    done && \
    gpg --batch --verify /tmp/tini.asc /usr/bin/tini && \
    rm -rf "$GNUPGHOME" /tmp/tini.asc && \
    chmod +x /usr/bin/tini && \
    groupadd \
            --system \
            --gid 1995 \
        zabbix && \
    useradd \
            --system \
            --comment "Zabbix monitoring system" \
            -g zabbix \
            -G root \
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
    DEBIAN_FRONTEND=noninteractive apt-get -y \
        purge \
            curl \
            gpg \
            dirmngr \
            gpg-agent && \
    apt-get -y autoremove && \
    apt-get -y clean && \
    rm -rf /var/lib/apt/lists/*

EXPOSE 10050/TCP

WORKDIR /var/lib/zabbix

COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENTRYPOINT ["/usr/bin/tini", "--", "/usr/bin/docker-entrypoint.sh"]

USER 1997

CMD ["/usr/sbin/zabbix_agentd", "--foreground", "-c", "/etc/zabbix/zabbix_agentd.conf"]
