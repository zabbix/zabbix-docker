# syntax=docker/dockerfile:1.6
ARG OS_BASE_IMAGE=registry.access.redhat.com/ubi10/ubi-minimal:10.2
ARG MAJOR_VERSION=7.4
ARG RELEASE=14
ARG ZBX_VERSION=${MAJOR_VERSION}.14
ARG BUILD_BASE_IMAGE=zabbix-build-common:rhel-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} AS builder

FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION
ARG RELEASE
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ARG ZBX_PRODUCT_REF=${ZBX_VERSION}
ARG ZBX_PRODUCT_REVISION=unknown

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
    ZBX_SOURCES=${ZBX_SOURCES} \
    ZABBIX_USER_HOME_DIR=/var/lib/zabbix \
    ZABBIX_CONF_DIR=/etc/zabbix

LABEL description="Zabbix Java Gateway performs native support for monitoring JMX applications" \
      maintainer="alexey.pustovalov@zabbix.com" \
      name="zabbix/zabbix-java-gateway-74" \
      release="${RELEASE}" \
      run="docker run --name zabbix-java-gateway -p 10052:10052 -d registry.connect.redhat.com/zabbix/zabbix-java-gateway-74:${ZBX_VERSION}" \
      summary="Zabbix Java Gateway" \
      url="https://www.zabbix.com/" \
      vendor="Zabbix SIA" \
      version="${MAJOR_VERSION}" \
      io.k8s.description="Zabbix Java Gateway performs native support for monitoring JMX applications" \
      io.k8s.display-name="Zabbix Java Gateway" \
      io.openshift.expose-services="10052:10052" \
      io.openshift.tags="gateway,zabbix-java,java" \
      org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
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


COPY ["licenses", "/licenses"]
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/java_gateway/sbin/ /usr/sbin/zabbix_java/
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/java_gateway/conf/ ${ZABBIX_CONF_DIR}/

RUN --mount=type=tmpfs,target=/var/lib/dnf/ \
    --mount=type=tmpfs,target=/var/cache/yum/ \
    set -eux && \
    INSTALL_PKGS="bash \
            shadow-utils \
            java-21-openjdk-headless" && \
    microdnf -y install \
            --disableplugin=subscription-manager \
            --disablerepo="*" \
            --enablerepo="ubi-10-baseos-rpms" \
            --enablerepo="ubi-10-appstream-rpms" \
            --setopt=install_weak_deps=0 \
            --setopt=keepcache=0 \
            --best \
            --setopt=tsflags=nodocs \
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
    microdnf -y remove shadow-utils && \
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
