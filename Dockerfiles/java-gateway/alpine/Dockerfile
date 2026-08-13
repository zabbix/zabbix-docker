# syntax=docker/dockerfile:1.6
ARG OS_BASE_IMAGE=alpine:3.24
ARG MAJOR_VERSION=8.0
ARG ZBX_VERSION=${MAJOR_VERSION}
ARG BUILD_BASE_IMAGE=zabbix-build-mysql:alpine-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} AS builder

FROM builder AS entrypoint-builder

RUN --mount=from=entrypoints,target=/tmp/entrypoints,ro \
    --mount=type=cache,target=/root/.cache/go-build \
    cd /tmp/entrypoints && \
    CGO_ENABLED=0 go build -buildvcs=false -trimpath -ldflags="-s -w -buildid=" \
            -o /tmp/docker-entrypoint \
        ./cmd/java-gateway

FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
    ZBX_SOURCES=${ZBX_SOURCES} \
    PATH=${PATH}:/usr/lib/jvm/default-jvm/bin/ JAVA_HOME=/usr/lib/jvm/default-jvm \
    ZABBIX_USER_HOME_DIR=/var/lib/zabbix \
    ZABBIX_CONF_DIR=/etc/zabbix

LABEL org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix Java Gateway performs native support for monitoring JMX applications" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="AGPL v3.0" \
      org.opencontainers.image.source="${ZBX_SOURCES}" \
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

COPY --chmod=0755 --from=entrypoint-builder /tmp/docker-entrypoint /usr/bin/docker-entrypoint
COPY --from=entrypoints /licenses/go-THIRD-PARTY-NOTICES.txt /licenses/

EXPOSE 10052/tcp

USER 1997

WORKDIR ${ZABBIX_USER_HOME_DIR}

ENTRYPOINT ["/usr/bin/docker-entrypoint"]

CMD ["java-gateway"]
