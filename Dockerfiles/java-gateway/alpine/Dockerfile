# syntax=docker/dockerfile:1.6
ARG OS_BASE_IMAGE=alpine:3.24
ARG MAJOR_VERSION=7.0
ARG ZBX_VERSION=${MAJOR_VERSION}.30
ARG BUILD_BASE_IMAGE=zabbix-build-common:alpine-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} AS builder

FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ARG ZBX_PRODUCT_REF=${ZBX_VERSION}
ARG ZBX_PRODUCT_REVISION=unknown

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
    ZBX_SOURCES=${ZBX_SOURCES} \
    PATH=${PATH}:/usr/lib/jvm/default-jvm/bin/ JAVA_HOME=/usr/lib/jvm/default-jvm \
    ZABBIX_USER_HOME_DIR=/var/lib/zabbix \
    ZABBIX_CONF_DIR=/etc/zabbix

LABEL org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix Java Gateway performs native support for monitoring JMX applications" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="AGPL-3.0-only" \
      com.zabbix.product.source="${ZBX_SOURCES}" \
      com.zabbix.product.ref="${ZBX_PRODUCT_REF}" \
      com.zabbix.product.revision="${ZBX_PRODUCT_REVISION}" \
      org.opencontainers.image.source="https://github.com/zabbix/zabbix-docker" \
      org.opencontainers.image.title="Zabbix Java Gateway" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix SIA" \
      org.opencontainers.image.version="${ZBX_VERSION}"

COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/java_gateway/sbin/ /usr/sbin/zabbix_java/
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/java_gateway/conf/ ${ZABBIX_CONF_DIR}/

RUN set -eux && \
    INSTALL_PKGS="bash \
            openjdk21-jre-headless" && \
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
    mkdir -p ${ZABBIX_CONF_DIR}/ \
        /usr/sbin/zabbix_java/ext_lib/ && \
    FIX_PERM_LIST="\
        ${ZABBIX_CONF_DIR}/ \
        /usr/sbin/zabbix_java/ext_lib" && \
    chown --quiet -R zabbix:0 ${FIX_PERM_LIST} && \
    chmod -R g=u ${FIX_PERM_LIST}

RUN --mount=from=entrypoints,target=/tmp/entrypoints,ro \
    set -eux && \
    install -d /usr/lib/docker-entrypoint && \
    install -m 0755 /tmp/entrypoints/java-gateway.sh /usr/bin/docker-entrypoint.sh && \
    install -m 0644 /tmp/entrypoints/lib/*.sh /usr/lib/docker-entrypoint/

EXPOSE 10052/tcp

USER 1997

WORKDIR ${ZABBIX_USER_HOME_DIR}

ENTRYPOINT ["docker-entrypoint.sh"]
