# syntax=docker/dockerfile:1.6
ARG OS_BASE_IMAGE=registry.access.redhat.com/ubi10/ubi-minimal:10.2
ARG MAJOR_VERSION=8.0
ARG RELEASE=0
ARG ZBX_VERSION=${MAJOR_VERSION}
ARG BUILD_BASE_IMAGE=zabbix-build-mysql:rhel-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} AS builder

FROM builder AS entrypoint-builder

RUN --mount=from=entrypoints,target=/tmp/entrypoints,ro \
    --mount=type=cache,target=/root/.cache/go-build \
    cd /tmp/entrypoints && \
    CGO_ENABLED=0 go build -buildvcs=false -trimpath -ldflags="-s -w -buildid=" \
            -o /tmp/docker-entrypoint \
        ./cmd/web-service

FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION
ARG RELEASE
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
    ZBX_SOURCES=${ZBX_SOURCES} \
    ZABBIX_USER_HOME_DIR=/var/lib/zabbix \
    ZABBIX_CONF_DIR=/etc/zabbix \
    PATH=/usr/lib64/chromium-browser:$PATH

LABEL description="Zabbix web service for performing various tasks using headless web browser" \
      maintainer="alexey.pustovalov@zabbix.com" \
      name="zabbix/zabbix-web-service-trunk" \
      release="${RELEASE}" \
      run="docker run --name zabbix-web-service --link zabbix-server:zabbix-server -p 10053:10053 -d registry.connect.redhat.com/zabbix/zabbix-web-service-trunk:${ZBX_VERSION}" \
      summary="Zabbix web service" \
      url="https://www.zabbix.com/" \
      vendor="Zabbix SIA" \
      version="${MAJOR_VERSION}" \
      io.k8s.description="Zabbix web service for performing various tasks using headless web browser" \
      io.k8s.display-name="Zabbix web service" \
      io.openshift.expose-services="10053:10053" \
      io.openshift.tags="zabbix,zabbix-web-service" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.description="Zabbix web service for performing various tasks using headless web browser" \
      org.label-schema.docker.cmd="docker run --name zabbix-web-service --link zabbix-server:zabbix-server -p 10053:10053 -d registry.connect.redhat.com/zabbix/zabbix-web-service-trunk:${ZBX_VERSION}" \
      org.label-schema.license="AGPL v3.0" \
      org.label-schema.name="zabbix-web-service-rhel" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.vcs-url="${ZBX_SOURCES}" \
      org.label-schema.vendor="Zabbix SIA" \
      org.label-schema.version="${ZBX_VERSION}"


COPY licenses /licenses/
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/web_service/sbin/zabbix_web_service /usr/sbin/zabbix_web_service
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}-output/web_service/conf/ ${ZABBIX_CONF_DIR}/
COPY --chmod=0755 --from=entrypoint-builder /tmp/docker-entrypoint /usr/bin/docker-entrypoint
COPY --from=entrypoints /licenses/go-THIRD-PARTY-NOTICES.txt /licenses/

RUN --mount=type=tmpfs,target=/var/lib/dnf/ \
    --mount=type=tmpfs,target=/var/cache/yum/ \
    set -eux && \
    INSTALL_PKGS="shadow-utils \
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
    mkdir -p ${ZABBIX_CONF_DIR} \
        ${ZABBIX_USER_HOME_DIR} \
        ${ZABBIX_USER_HOME_DIR}/enc \
        ${ZABBIX_USER_HOME_DIR}/enc_internal && \
    chown --quiet -R zabbix:0 ${ZABBIX_USER_HOME_DIR}/ && \
    chmod -R g=u ${ZABBIX_USER_HOME_DIR}/ && \
    /usr/sbin/zabbix_web_service -V

EXPOSE 10053/tcp

WORKDIR ${ZABBIX_USER_HOME_DIR}

USER 1997

ENTRYPOINT ["/usr/bin/docker-entrypoint"]

CMD ["/usr/sbin/zabbix_web_service", "-c", "/etc/zabbix/zabbix_web_service.conf"]
