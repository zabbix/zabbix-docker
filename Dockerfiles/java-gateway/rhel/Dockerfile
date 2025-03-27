# syntax=docker/dockerfile:1
ARG OS_BASE_IMAGE=registry.access.redhat.com/ubi9/ubi-minimal:9.5
ARG MAJOR_VERSION=7.2
ARG RELEASE=5
ARG ZBX_VERSION=${MAJOR_VERSION}.5
ARG BUILD_BASE_IMAGE=zabbix-build-mysql:rhel-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} AS builder

FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION
ARG RELEASE
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} \
    ZABBIX_USER_HOME_DIR="/var/lib/zabbix" \
    ZABBIX_CONF_DIR="/etc/zabbix"

LABEL description="Zabbix Java Gateway performs native support for monitoring JMX applications" \
      maintainer="alexey.pustovalov@zabbix.com" \
      name="zabbix/zabbix-java-gateway-72" \
      release="${RELEASE}" \
      run="docker run --name zabbix-java-gateway -p 10052:10052 -d registry.connect.redhat.com/zabbix/zabbix-java-gateway-72:${ZBX_VERSION}" \
      summary="Zabbix Java Gateway" \
      url="https://www.zabbix.com/" \
      vendor="Zabbix SIA" \
      version="${MAJOR_VERSION}" \
      io.k8s.description="Zabbix Java Gateway performs native support for monitoring JMX applications" \
      io.k8s.display-name="Zabbix Java Gateway" \
      io.openshift.expose-services="10052:10052" \
      io.openshift.tags="gateway,zabbix-java,java" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.description="Zabbix Java Gateway performs native support for monitoring JMX applications" \
      org.label-schema.docker.cmd="docker run --name zabbix-java-gateway -p 10052:10052 -d registry.connect.redhat.com/zabbix/zabbix-java-gateway-72:${ZBX_VERSION}" \
      org.label-schema.license="AGPL v3.0" \
      org.label-schema.name="zabbix-java-gateway-rhel" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.vcs-url="${ZBX_SOURCES}" \
      org.label-schema.vendor="Zabbix SIA" \
      org.label-schema.version="${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY ["licenses", "/licenses"]
COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}-output/java_gateway/sbin/", "/usr/sbin/zabbix_java/"]

RUN --mount=type=tmpfs,target=/var/lib/dnf/ \
    --mount=type=tmpfs,target=/var/cache/yum/ \
    set -eux && \
    INSTALL_PKGS="bash \
            findutils \
            shadow-utils \
            java-17-openjdk-headless" && \
    microdnf -y install \
            --disableplugin=subscription-manager \
            --disablerepo "*" \
            --enablerepo "ubi-9-baseos-rpms" \
            --enablerepo "ubi-9-appstream-rpms" \
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
    mkdir -p ${ZABBIX_CONF_DIR}/ && \
    mkdir -p /usr/sbin/zabbix_java/ && \
    mkdir -p /usr/sbin/zabbix_java/ext_lib/ && \
    rm -rf /usr/sbin/zabbix_java/lib/logback.xml && \
    mv /usr/sbin/zabbix_java/lib/logback-console.xml ${ZABBIX_CONF_DIR}/zabbix_java_gateway_logback.xml && \
    chown --quiet -R zabbix:root ${ZABBIX_CONF_DIR}/ /usr/sbin/zabbix_java/ && \
    chgrp -R 0 ${ZABBIX_CONF_DIR}/ /usr/sbin/zabbix_java/ && \
    chmod -R g=u ${ZABBIX_CONF_DIR}/ /usr/sbin/zabbix_java/

EXPOSE 10052/TCP

WORKDIR ${ZABBIX_USER_HOME_DIR}

COPY ["conf/usr/sbin/zabbix_java_gateway", "/usr/sbin/"]
COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENTRYPOINT ["docker-entrypoint.sh"]

USER 1997

CMD ["/usr/sbin/zabbix_java_gateway"]
