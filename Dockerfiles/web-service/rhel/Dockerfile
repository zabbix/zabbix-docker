# syntax=docker/dockerfile:1
ARG MAJOR_VERSION=6.0
ARG RELEASE=0
ARG ZBX_VERSION=${MAJOR_VERSION}.0
ARG BUILD_BASE_IMAGE=zabbix-build-mysql:rhel-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} as builder

FROM registry.access.redhat.com/ubi8/ubi-minimal

ARG MAJOR_VERSION
ARG RELEASE
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES}

LABEL description="Zabbix web service for performing various tasks using headless web browser" \
      maintainer="alexey.pustovalov@zabbix.com" \
      name="zabbix/zabbix-web-service-60" \
      release="${RELEASE}" \
      run="docker run --name zabbix-web-service --link zabbix-server:zabbix-server -p 10053:10053 -d registry.connect.redhat.com/zabbix/zabbix-web-service-60:${ZBX_VERSION}" \
      summary="Zabbix web service" \
      url="https://www.zabbix.com/" \
      vendor="Zabbix LLC" \
      version="${MAJOR_VERSION}" \
      io.k8s.description="Zabbix web service for performing various tasks using headless web browser" \
      io.k8s.display-name="Zabbix web service" \
      io.openshift.expose-services="10053:10053" \
      io.openshift.tags="zabbix,zabbix-web-service" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.description="Zabbix web service for performing various tasks using headless web browser" \
      org.label-schema.docker.cmd="docker run --name zabbix-web-service --link zabbix-server:zabbix-server -p 10053:10053 -d registry.connect.redhat.com/zabbix/zabbix-web-service-60:${ZBX_VERSION}" \
      org.label-schema.license="GPL v2.0" \
      org.label-schema.name="zabbix-web-service-rhel" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.vcs-url="${ZBX_SOURCES}" \
      org.label-schema.vendor="Zabbix LLC" \
      org.label-schema.version="${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY ["licenses", "/licenses"]
COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}/src/go/bin/zabbix_web_service", "/usr/sbin/zabbix_web_service"]
COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}/src/go/conf/zabbix_web_service.conf", "/etc/zabbix/zabbix_web_service.conf"]

RUN set -eux && \
    INSTALL_PKGS="bash \
            shadow-utils \
            chromium" && \
    curl -sSL -o /tmp/epel-release-latest-8.noarch.rpm https://dl.fedoraproject.org/pub/epel/epel-release-latest-8.noarch.rpm && \
    rpm -ivh /tmp/epel-release-latest-8.noarch.rpm && \
    rm -rf /tmp/epel-release-latest-8.noarch.rpm && \
    microdnf -y install \
            --disablerepo "*" \
            --enablerepo "ubi-8-baseos" \
            --enablerepo "ubi-8-appstream" \
            --enablerepo "rhel-8-for-x86_64-baseos-rpms" \
            --enablerepo "rhel-8-for-x86_64-appstream-rpms" \
            --enablerepo "epel" \
            --setopt=install_weak_deps=0 \
            --best \
            --setopt=tsflags=nodocs \
        ${INSTALL_PKGS} && \
    groupadd --system --gid 1995 zabbix && \
    useradd \
            --system --comment "Zabbix monitoring system" \
            -g zabbix -G root \
            --uid 1997 \
            --shell /sbin/nologin \
            --home-dir /var/lib/zabbix/ \
        zabbix && \
    mkdir -p /etc/zabbix && \
    mkdir -p /var/lib/zabbix && \
    mkdir -p /var/lib/zabbix/enc && \
    chown --quiet -R zabbix:root /etc/zabbix/ /var/lib/zabbix/ && \
    chgrp -R 0 /etc/zabbix/ /var/lib/zabbix/ && \
    chmod -R g=u /etc/zabbix/ /var/lib/zabbix/ && \
    microdnf -y clean all && \
    rm -rf /var/cache/yum /var/lib/yum/yumdb/* /usr/lib/udev/hwdb.d/* && \
    rm -rf /var/cache/dnf /etc/udev/hwdb.bin /root/.pki

EXPOSE 10053/TCP

WORKDIR /var/lib/zabbix

COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENTRYPOINT ["docker-entrypoint.sh"]

USER 1997

CMD ["/usr/sbin/zabbix_web_service", "-c", "/etc/zabbix/zabbix_web_service.conf"]
