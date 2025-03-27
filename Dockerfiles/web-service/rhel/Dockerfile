# syntax=docker/dockerfile:1
ARG OS_BASE_IMAGE=registry.access.redhat.com/ubi9/ubi-minimal:9.5
ARG MAJOR_VERSION=7.2
ARG RELEASE=5
ARG ZBX_VERSION=${MAJOR_VERSION}.5
ARG BUILD_BASE_IMAGE=zabbix-build-mysql:rhel-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} AS builder

ARG OS_BASE_IMAGE=registry.access.redhat.com/ubi9/ubi-minimal:9.5
FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION
ARG RELEASE
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} \
    ZABBIX_USER_HOME_DIR="/var/lib/zabbix" \
    ZABBIX_CONF_DIR="/etc/zabbix" \
    PATH=/usr/lib64/chromium-browser:$PATH

LABEL description="Zabbix web service for performing various tasks using headless web browser" \
      maintainer="alexey.pustovalov@zabbix.com" \
      name="zabbix/zabbix-web-service-72" \
      release="${RELEASE}" \
      run="docker run --name zabbix-web-service --link zabbix-server:zabbix-server -p 10053:10053 -d registry.connect.redhat.com/zabbix/zabbix-web-service-72:${ZBX_VERSION}" \
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
      org.label-schema.docker.cmd="docker run --name zabbix-web-service --link zabbix-server:zabbix-server -p 10053:10053 -d registry.connect.redhat.com/zabbix/zabbix-web-service-72:${ZBX_VERSION}" \
      org.label-schema.license="AGPL v3.0" \
      org.label-schema.name="zabbix-web-service-rhel" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.vcs-url="${ZBX_SOURCES}" \
      org.label-schema.vendor="Zabbix SIA" \
      org.label-schema.version="${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY ["licenses", "/licenses"]
COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}-output/web_service/sbin/zabbix_web_service", "/usr/sbin/zabbix_web_service"]
COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}-output/web_service/conf/", "${ZABBIX_CONF_DIR}/"]

RUN --mount=type=tmpfs,target=/var/lib/dnf/ \
    --mount=type=tmpfs,target=/var/cache/yum/ \
#    --mount=type=bind,target=/run/secrets/,src=secrets/ \
    set -eux && \
    INSTALL_PKGS="bash \
            shadow-utils \
            chromium-headless" && \
    curl --tlsv1.2 -sSf -L https://dl.fedoraproject.org/pub/epel/epel-release-latest-9.noarch.rpm -o /tmp/epel-release-latest-9.noarch.rpm && \
    rpm -ivh /tmp/epel-release-latest-9.noarch.rpm && \
    rm -rf /tmp/epel-release-latest-9.noarch.rpm && \
    ARCH_SUFFIX="$(arch)"; \
    microdnf -y install \
            --disablerepo "*" \
            --enablerepo "ubi-9-baseos-rpms" \
            --enablerepo "ubi-9-appstream-rpms" \
            --enablerepo "rhel-9-for-$ARCH_SUFFIX-appstream-rpms" \
            --enablerepo "epel" \
            --setopt=install_weak_deps=0 \
            --setopt=keepcache=0 \
            --best \
            --setopt=tsflags=nodocs \
        ${INSTALL_PKGS} && \
    microdnf -y install \
            --disableplugin=subscription-manager \
            --disablerepo "*" \
            --enablerepo "ubi-9-baseos-rpms" \
            --enablerepo "ubi-9-appstream-rpms" \
            --setopt=install_weak_deps=0 \
            --setopt=keepcache=0 \
            --best \
            --setopt=tsflags=nodocs \
            nss nss-sysinit && \
    groupadd --system --gid 1995 zabbix && \
    useradd \
            --system --comment "Zabbix monitoring system" \
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
    /usr/sbin/zabbix_web_service -V

EXPOSE 10053/TCP

WORKDIR ${ZABBIX_USER_HOME_DIR}

COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENTRYPOINT ["docker-entrypoint.sh"]

USER 1997

CMD ["/usr/sbin/zabbix_web_service", "-c", "/etc/zabbix/zabbix_web_service.conf"]
