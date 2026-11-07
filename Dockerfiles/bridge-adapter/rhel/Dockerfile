# syntax=docker/dockerfile:1.6

ARG OS_BASE_IMAGE=registry.access.redhat.com/ubi10/ubi-micro:10.2
ARG RUNTIME_PREP_BASE_IMAGE=registry.access.redhat.com/ubi10/ubi:10.2
ARG GO_BASE_IMAGE=registry.access.redhat.com/ubi10/go-toolset:1.26

ARG MAJOR_VERSION=8.0
ARG RELEASE=0
ARG ZBX_VERSION=${MAJOR_VERSION}.0

FROM ${GO_BASE_IMAGE} AS builder

ARG ZBX_VERSION

ENV ZBX_SOURCES_DIR=/tmp/zabbix-${ZBX_VERSION} \
    ZBX_ROOTFS_DIR=/tmp/zabbix-${ZBX_VERSION}-output \
    ZABBIX_CONF_DIR=/etc/zabbix \
    GOCACHE=/tmp/.cache/go-build \
    GOMODCACHE=/tmp/go/pkg/mod \
    GOTELEMETRY=off

RUN --mount=from=bridge_adapter_sources,target=/tmp/bridge_adapter_sources,ro \
    set -eux && \
    mkdir -p \
        ${ZBX_SOURCES_DIR} \
        ${ZBX_ROOTFS_DIR}/usr/sbin \
        ${ZBX_ROOTFS_DIR}/usr/bin \
        ${ZBX_ROOTFS_DIR}${ZABBIX_CONF_DIR} && \
    cp -a --no-preserve=ownership /tmp/bridge_adapter_sources/. ${ZBX_SOURCES_DIR}/

WORKDIR ${ZBX_SOURCES_DIR}

RUN --mount=from=config_templates,target=/tmp/conf,ro \
    --mount=type=cache,target=/tmp/.cache/go-build,uid=1001,gid=0 \
    --mount=type=cache,target=/tmp/go/pkg/mod,uid=1001,gid=0 \
    set -eux && \
    make -C bridge_adapter build \
        SYSCONFDIR=${ZABBIX_CONF_DIR} \
        DOCDIR=/usr/share/doc/zabbix-bridge-adapter && \
    strip bridge_adapter/bin/zabbix_bridge_adapter && \
    mv bridge_adapter/bin/zabbix_bridge_adapter ${ZBX_ROOTFS_DIR}/usr/sbin/ && \
    cp -a /tmp/conf/bridge_adapter/. ${ZBX_ROOTFS_DIR}${ZABBIX_CONF_DIR}/

RUN --mount=from=entrypoints,target=/tmp/entrypoints,ro \
    --mount=type=cache,target=/tmp/.cache/go-build,uid=1001,gid=0 \
    --mount=type=cache,target=/tmp/go/pkg/mod,uid=1001,gid=0 \
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

FROM ${OS_BASE_IMAGE} AS runtime-base

FROM ${RUNTIME_PREP_BASE_IMAGE} AS runtime-prep

ARG ZBX_VERSION

ENV ZBX_ROOTFS_DIR=/tmp/zabbix-${ZBX_VERSION}-output \
    ZABBIX_USER_HOME_DIR=/var/lib/zabbix

COPY --from=runtime-base / ${ZBX_ROOTFS_DIR}/

RUN --mount=type=tmpfs,target=${ZBX_ROOTFS_DIR}/var/lib/dnf/ \
    --mount=type=tmpfs,target=${ZBX_ROOTFS_DIR}/var/cache/dnf/ \
    --mount=type=tmpfs,target=${ZBX_ROOTFS_DIR}/var/log/ \
    set -eux && \
    ARCH_SUFFIX="$(arch)"; \
    dnf -y install \
            --installroot=${ZBX_ROOTFS_DIR} \
            --disableplugin=subscription-manager \
            --disablerepo="*" \
            --enablerepo="ubi-10-for-$ARCH_SUFFIX-baseos-rpms" \
            --setopt=install_weak_deps=0 \
            --setopt=keepcache=0 \
            --setopt=tsflags=nodocs \
        ca-certificates \
        tzdata && \
    mkdir -p \
            ${ZBX_ROOTFS_DIR}${ZABBIX_USER_HOME_DIR} \
            ${ZBX_ROOTFS_DIR}${ZABBIX_USER_HOME_DIR}/bridge-adapter \
            ${ZBX_ROOTFS_DIR}${ZABBIX_USER_HOME_DIR}/enc \
            ${ZBX_ROOTFS_DIR}/run/zabbix/secrets/tls \
            ${ZBX_ROOTFS_DIR}/run/zabbix/secrets_internal \
            ${ZBX_ROOTFS_DIR}${ZABBIX_USER_HOME_DIR}/ssl/certs \
            ${ZBX_ROOTFS_DIR}${ZABBIX_USER_HOME_DIR}/ssl/keys \
            ${ZBX_ROOTFS_DIR}${ZABBIX_USER_HOME_DIR}/ssl/ssl_ca && \
    chown --quiet -R 1997:0 \
        ${ZBX_ROOTFS_DIR}${ZABBIX_USER_HOME_DIR} \
        ${ZBX_ROOTFS_DIR}/run/zabbix/secrets_internal && \
    chmod -R g=u \
        ${ZBX_ROOTFS_DIR}${ZABBIX_USER_HOME_DIR} \
        ${ZBX_ROOTFS_DIR}/run/zabbix/secrets_internal

FROM runtime-base

ARG MAJOR_VERSION
ARG RELEASE
ARG ZBX_VERSION
ARG ZBX_PRODUCT_REF=${ZBX_VERSION}
ARG ZBX_PRODUCT_REVISION=unknown

ENV ZBX_VERSION=${ZBX_VERSION} \
    ZABBIX_USER_HOME_DIR=/var/lib/zabbix \
    ZABBIX_CONF_DIR=/etc/zabbix

LABEL description="Zabbix Bridge Adapter enables the Zabbix mobile application to communicate with an on-premise Zabbix installation" \
      maintainer="alexey.pustovalov@zabbix.com" \
      name="zabbix/zabbix-bridge-adapter-80" \
      release="${RELEASE}" \
      run="podman run --name zabbix-bridge-adapter --network zabbix-backend -d registry.connect.redhat.com/zabbix/zabbix-bridge-adapter-80:${ZBX_VERSION}" \
      summary="Zabbix Bridge Adapter" \
      url="https://www.zabbix.com/" \
      vendor="Zabbix SIA" \
      version="${ZBX_VERSION}" \
      io.k8s.description="Zabbix Bridge Adapter enables the Zabbix mobile application to communicate with an on-premise Zabbix installation" \
      io.k8s.display-name="Zabbix Bridge Adapter" \
      io.openshift.expose-services="10054:10054" \
      io.openshift.tags="zabbix,zabbix-bridge-adapter" \
      org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix Bridge Adapter enables the Zabbix mobile application to communicate with an on-premise Zabbix installation" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="LicenseRef-Zabbix-Mobile-App-EULA" \
      com.zabbix.eula.url="https://www.zabbix.com/mobile_app_eula" \
      com.zabbix.product.ref="${ZBX_PRODUCT_REF}" \
      com.zabbix.product.revision="${ZBX_PRODUCT_REVISION}" \
      org.opencontainers.image.source="https://github.com/zabbix/zabbix-docker" \
      org.opencontainers.image.title="Zabbix Bridge Adapter" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix SIA" \
      org.opencontainers.image.version="${ZBX_VERSION}"

COPY --from=runtime-prep /tmp/zabbix-${ZBX_VERSION}-output/ /
COPY --from=builder --chown=0:0 /tmp/zabbix-${ZBX_VERSION}-output/ /

RUN ["/usr/sbin/zabbix_bridge_adapter", "-V"]

EXPOSE 10054/tcp 8081/tcp

WORKDIR ${ZABBIX_USER_HOME_DIR}

USER 1997

ENTRYPOINT ["/usr/bin/docker-entrypoint"]

HEALTHCHECK --interval=10s --timeout=3s --start-period=90s --start-interval=2s --retries=5 \
    CMD ["/usr/bin/http-healthcheck", "/ready"]

CMD ["/usr/sbin/zabbix_bridge_adapter", "-c", "/etc/zabbix/zabbix_bridge_adapter.conf"]
