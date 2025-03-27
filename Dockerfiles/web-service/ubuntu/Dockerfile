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
    ZABBIX_USER_HOME_DIR="/var/lib/zabbix" \
    ZABBIX_CONF_DIR="/etc/zabbix"

LABEL org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix web service for performing various tasks using headless web browser" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="AGPL v3.0" \
      org.opencontainers.image.source="${ZBX_SOURCES}" \
      org.opencontainers.image.title="Zabbix web service" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix SIA" \
      org.opencontainers.image.version="${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}-output/web_service/sbin/zabbix_web_service", "/usr/sbin/zabbix_web_service"]
COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}-output/web_service/conf/", "${ZABBIX_CONF_DIR}/"]
COPY ["conf/etc/apt/preferences.d/chromium.pref", "/etc/apt/preferences.d/chromium.pref"]

RUN --mount=type=cache,target=/var/cache/apt/,sharing=locked \
    --mount=type=cache,target=/var/lib/apt/,sharing=locked \
    set -eux && \
    echo "#!/bin/sh\nexit 101" > /usr/sbin/policy-rc.d && \
    INSTALL_PKGS="bash \
            ca-certificates \
            chromium \
            chromium-sandbox" && \
    INSTALL_TEMP_PKGS="gnupg dirmngr gpg gpgconf" && \
    apt-get -y update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y \
            --no-install-recommends install \
        ${INSTALL_TEMP_PKGS} && \
    gpg --keyserver keyserver.ubuntu.com --recv-keys 0E98404D386FA1D9 && \
    gpg -o /etc/apt/keyrings/debian.gpg --export 0E98404D386FA1D9 && \
    echo "deb [signed-by=/etc/apt/keyrings/debian.gpg] http://deb.debian.org/debian bullseye main" > /etc/apt/sources.list.d/debian.list && \
    echo "deb [signed-by=/etc/apt/keyrings/debian.gpg] http://deb.debian.org/debian bullseye-updates main" >> /etc/apt/sources.list.d/debian.list && \
    apt-get -y update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y \
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
            --home-dir ${ZABBIX_USER_HOME_DIR} \
        zabbix && \
    mkdir -p ${ZABBIX_CONF_DIR} && \
    mkdir -p ${ZABBIX_USER_HOME_DIR} && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/enc && \
    mkdir -p ${ZABBIX_USER_HOME_DIR}/enc_internal && \
    chown --quiet -R zabbix:root ${ZABBIX_USER_HOME_DIR}/ && \
    chgrp -R 0 ${ZABBIX_USER_HOME_DIR}/ && \
    chmod -R g=u ${ZABBIX_USER_HOME_DIR}/ && \
    DEBIAN_FRONTEND=noninteractive apt-get -y \
            purge \
        ${INSTALL_TEMP_PKGS} && \
    rm -rf ~/.gnupg && \
    apt-get -y autoremove && \
    /usr/sbin/zabbix_web_service -V

EXPOSE 10053/TCP

WORKDIR ${ZABBIX_USER_HOME_DIR}

COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENTRYPOINT ["docker-entrypoint.sh"]

USER 1997

CMD ["/usr/sbin/zabbix_web_service", "-c", "/etc/zabbix/zabbix_web_service.conf"]
