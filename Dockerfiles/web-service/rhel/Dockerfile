# syntax=docker/dockerfile:1.6
ARG OS_BASE_IMAGE=registry.access.redhat.com/ubi10/ubi-minimal:10.2
ARG MAJOR_VERSION=7.4
ARG RELEASE=13
ARG ZBX_VERSION=${MAJOR_VERSION}.13
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
    ZABBIX_CONF_DIR=/etc/zabbix \
    PATH=/usr/lib64/chromium-browser:$PATH

LABEL description="Zabbix web service for performing various tasks using headless web browser" \
      maintainer="alexey.pustovalov@zabbix.com" \
      name="zabbix/zabbix-web-service-74" \
      release="${RELEASE}" \
      run="docker run --name zabbix-web-service --link zabbix-server:zabbix-server -p 10053:10053 -d registry.connect.redhat.com/zabbix/zabbix-web-service-74:${ZBX_VERSION}" \
      summary="Zabbix web service" \
      url="https://www.zabbix.com/" \
      vendor="Zabbix SIA" \
      version="${MAJOR_VERSION}" \
      io.k8s.description="Zabbix web service for performing various tasks using headless web browser" \
      io.k8s.display-name="Zabbix web service" \
      io.openshift.expose-services="10053:10053" \
      io.openshift.tags="zabbix,zabbix-web-service" \
      org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix web service for performing various tasks using headless web browser" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="AGPL-3.0-only" \
      com.zabbix.product.source="${ZBX_SOURCES}" \
      com.zabbix.product.ref="${ZBX_PRODUCT_REF}" \
      com.zabbix.product.revision="${ZBX_PRODUCT_REVISION}" \
      org.opencontainers.image.source="https://github.com/zabbix/zabbix-docker" \
      org.opencontainers.image.title="Zabbix web service" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix SIA" \
      org.opencontainers.image.version="${ZBX_VERSION}"


COPY licenses /licenses/
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/web_service/sbin/zabbix_web_service /usr/sbin/zabbix_web_service
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/web_service/conf/ ${ZABBIX_CONF_DIR}/

RUN --mount=type=tmpfs,target=/var/lib/dnf/ \
    --mount=type=tmpfs,target=/var/cache/yum/ \
    set -eux && \
    INSTALL_PKGS="bash \
            shadow-utils \
            chromium-headless" && \
    curl --tlsv1.2 -sSf -L https://dl.fedoraproject.org/pub/epel/epel-release-latest-10.noarch.rpm -o /tmp/epel-release-latest-10.noarch.rpm && \
    rpm -ivh /tmp/epel-release-latest-10.noarch.rpm && \
    rm -rf /tmp/epel-release-latest-10.noarch.rpm && \
    ARCH_SUFFIX="$(arch)"; \
    microdnf -y install \
            --disablerepo="*" \
            --enablerepo="ubi-10-baseos-rpms" \
            --enablerepo="ubi-10-appstream-rpms" \
            --enablerepo="rhel-10-for-$ARCH_SUFFIX-baseos-rpms" \
            --enablerepo="rhel-10-for-$ARCH_SUFFIX-appstream-rpms" \
            --enablerepo="epel" \
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
    mkdir -p ${ZABBIX_CONF_DIR} \
        ${ZABBIX_USER_HOME_DIR} \
        ${ZABBIX_USER_HOME_DIR}/enc \
        ${ZABBIX_USER_HOME_DIR}/enc_internal && \
    chown --quiet -R zabbix:0 ${ZABBIX_USER_HOME_DIR}/ && \
    chmod -R g=u ${ZABBIX_USER_HOME_DIR}/ && \
    /usr/sbin/zabbix_web_service -V

RUN --mount=from=entrypoints,target=/tmp/entrypoints,ro \
    set -eux && \
    install -d /usr/lib/docker-entrypoint && \
    install -m 0755 /tmp/entrypoints/web-service.sh /usr/bin/docker-entrypoint.sh && \
    install -m 0644 /tmp/entrypoints/lib/*.sh /usr/lib/docker-entrypoint/

EXPOSE 10053/tcp

WORKDIR ${ZABBIX_USER_HOME_DIR}

USER 1997

ENTRYPOINT ["docker-entrypoint.sh"]

CMD ["/usr/sbin/zabbix_web_service", "-c", "/etc/zabbix/zabbix_web_service.conf"]
