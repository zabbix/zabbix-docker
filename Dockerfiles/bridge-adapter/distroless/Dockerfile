# syntax=docker/dockerfile:1.6

ARG OS_BASE_IMAGE=gcr.io/distroless/static-debian13:latest
ARG GO_BASE_IMAGE=golang:1.27.3-alpine

ARG MAJOR_VERSION=8.0
ARG ZBX_VERSION=${MAJOR_VERSION}.0

FROM ${OS_BASE_IMAGE} AS runtime-base

FROM ${GO_BASE_IMAGE} AS builder

ARG ZBX_VERSION

ENV ZBX_SOURCES_DIR=/tmp/zabbix-${ZBX_VERSION} \
    ZBX_ROOTFS_DIR=/tmp/zabbix-${ZBX_VERSION}-output \
    ZABBIX_USER_HOME_DIR=/var/lib/zabbix \
    ZABBIX_CONF_DIR=/etc/zabbix \
    GOCACHE=/root/.cache/go-build \
    GOMODCACHE=/root/go/pkg/mod \
    GOTELEMETRY=off

RUN set -eux && \
    INSTALL_PKGS="bash \
            binutils \
            make" && \
    apk add \
        ${INSTALL_PKGS}

RUN --mount=from=bridge_adapter_sources,target=/tmp/bridge_adapter_sources,ro \
    set -eux && \
    mkdir -p \
        ${ZBX_SOURCES_DIR} \
        ${ZBX_ROOTFS_DIR}/usr/sbin \
        ${ZBX_ROOTFS_DIR}/usr/bin \
        ${ZBX_ROOTFS_DIR}${ZABBIX_CONF_DIR} \
        ${ZBX_ROOTFS_DIR}${ZABBIX_USER_HOME_DIR} \
        ${ZBX_ROOTFS_DIR}${ZABBIX_USER_HOME_DIR}/bridge-adapter \
        ${ZBX_ROOTFS_DIR}${ZABBIX_USER_HOME_DIR}/enc \
        ${ZBX_ROOTFS_DIR}/run/zabbix/secrets/tls \
        ${ZBX_ROOTFS_DIR}/run/zabbix/secrets_internal \
        ${ZBX_ROOTFS_DIR}${ZABBIX_USER_HOME_DIR}/ssl/certs \
        ${ZBX_ROOTFS_DIR}${ZABBIX_USER_HOME_DIR}/ssl/keys \
        ${ZBX_ROOTFS_DIR}${ZABBIX_USER_HOME_DIR}/ssl/ssl_ca && \
    chown -R 1997:0 \
        ${ZBX_ROOTFS_DIR}${ZABBIX_USER_HOME_DIR} \
        ${ZBX_ROOTFS_DIR}/run/zabbix/secrets_internal && \
    chmod -R g=u \
        ${ZBX_ROOTFS_DIR}${ZABBIX_USER_HOME_DIR} \
        ${ZBX_ROOTFS_DIR}/run/zabbix/secrets_internal && \
    cp -a /tmp/bridge_adapter_sources/. ${ZBX_SOURCES_DIR}/

WORKDIR ${ZBX_SOURCES_DIR}

RUN --mount=from=config_templates,target=/tmp/conf,ro \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    set -eux && \
    make -C bridge_adapter build \
        SYSCONFDIR=${ZABBIX_CONF_DIR} \
        DOCDIR=/usr/share/doc/zabbix-bridge-adapter && \
    strip bridge_adapter/bin/zabbix_bridge_adapter && \
    mv bridge_adapter/bin/zabbix_bridge_adapter ${ZBX_ROOTFS_DIR}/usr/sbin/ && \
    cp -a /tmp/conf/bridge_adapter/. ${ZBX_ROOTFS_DIR}${ZABBIX_CONF_DIR}/

RUN --mount=from=entrypoints,target=/tmp/entrypoints,ro \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    set -eux && \
    mkdir -p ${ZBX_ROOTFS_DIR}/licenses && \
    cd /tmp/entrypoints && \
    CGO_ENABLED=0 go build \
            -buildvcs=false \
            -trimpath \
            -ldflags="-s -w -buildid=" \
            -o ${ZBX_ROOTFS_DIR}/usr/bin/docker-entrypoint \
        ./cmd/bridge-adapter && \
    CGO_ENABLED=0 go build \
            -buildvcs=false \
            -trimpath \
            -ldflags="-s -w -buildid=" \
            -o ${ZBX_ROOTFS_DIR}/usr/bin/http-healthcheck \
        ./cmd/http-healthcheck && \
    cp licenses/go-THIRD-PARTY-NOTICES.txt ${ZBX_ROOTFS_DIR}/licenses/ && \
    cp ${ZBX_SOURCES_DIR}/bridge_adapter/THIRD-PARTY-NOTICES.txt \
        ${ZBX_ROOTFS_DIR}/licenses/bridge-adapter-THIRD-PARTY-NOTICES.txt && \
    cp ${ZBX_SOURCES_DIR}/bridge_adapter/internal/eula/EULA.txt \
        ${ZBX_ROOTFS_DIR}/licenses/Zabbix-Mobile-App-EULA.txt

FROM runtime-base

ARG MAJOR_VERSION
ARG ZBX_VERSION
ARG ZBX_PRODUCT_REF=${ZBX_VERSION}
ARG ZBX_PRODUCT_REVISION=unknown

ENV ZBX_VERSION=${ZBX_VERSION} \
    ZBX_SOURCES=${ZBX_SOURCES} \
    ZABBIX_USER_HOME_DIR=/var/lib/zabbix \
    ZABBIX_CONF_DIR=/etc/zabbix

LABEL org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix Bridge Adapter enables the Zabbix mobile application to communicate with an on-premise Zabbix installation" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="LicenseRef-Zabbix-Mobile-App-EULA" \
      com.zabbix.eula.url="https://www.zabbix.com/mobile_app_eula" \
      com.zabbix.product.ref="${ZBX_PRODUCT_REF}" \
      com.zabbix.product.revision="${ZBX_PRODUCT_REVISION}" \
      org.opencontainers.image.source="https://github.com/zabbix/zabbix-docker" \
      org.opencontainers.image.title="Zabbix Bridge Adapter" \
      org.opencontainers.image.url="https://hub.docker.com/r/zabbix/zabbix-bridge-adapter" \
      org.opencontainers.image.vendor="Zabbix SIA" \
      org.opencontainers.image.version="${ZBX_VERSION}"

COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/ /

RUN ["/usr/sbin/zabbix_bridge_adapter", "-V"]

EXPOSE 10054/tcp 8081/tcp

WORKDIR ${ZABBIX_USER_HOME_DIR}

USER 1997:0

ENTRYPOINT ["/usr/bin/docker-entrypoint"]

HEALTHCHECK --interval=10s --timeout=3s --start-period=90s --start-interval=2s --retries=5 \
    CMD ["/usr/bin/http-healthcheck", "/ready"]

CMD ["/usr/sbin/zabbix_bridge_adapter", "-c", "/etc/zabbix/zabbix_bridge_adapter.conf"]
