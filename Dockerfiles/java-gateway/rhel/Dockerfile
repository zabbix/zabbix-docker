# syntax=docker/dockerfile:1
ARG MAJOR_VERSION=6.4
ARG RELEASE=11
ARG ZBX_VERSION=${MAJOR_VERSION}.11
ARG BUILD_BASE_IMAGE=zabbix-build-mysql:rhel-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} AS builder

FROM registry.access.redhat.com/ubi9/ubi-minimal:9.3

ARG MAJOR_VERSION
ARG RELEASE
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES}

LABEL description="Zabbix Java Gateway performs native support for monitoring JMX applications" \
      maintainer="alexey.pustovalov@zabbix.com" \
      name="zabbix/zabbix-java-gateway-64" \
      release="${RELEASE}" \
      run="docker run --name zabbix-java-gateway -p 10052:10052 -d registry.connect.redhat.com/zabbix/zabbix-java-gateway-64:${ZBX_VERSION}" \
      summary="Zabbix Java Gateway" \
      url="https://www.zabbix.com/" \
      vendor="Zabbix LLC" \
      version="${MAJOR_VERSION}" \
      io.k8s.description="Zabbix Java Gateway performs native support for monitoring JMX applications" \
      io.k8s.display-name="Zabbix Java Gateway" \
      io.openshift.expose-services="10052:10052" \
      io.openshift.tags="gateway,zabbix-java,java" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.description="Zabbix Java Gateway performs native support for monitoring JMX applications" \
      org.label-schema.docker.cmd="docker run --name zabbix-java-gateway -p 10052:10052 -d registry.connect.redhat.com/zabbix/zabbix-java-gateway-64:${ZBX_VERSION}" \
      org.label-schema.license="GPL v2.0" \
      org.label-schema.name="zabbix-java-gateway-rhel" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.vcs-url="${ZBX_SOURCES}" \
      org.label-schema.vendor="Zabbix LLC" \
      org.label-schema.version="${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY ["licenses", "/licenses"]
COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}/src/zabbix_java/bin", "/usr/sbin/zabbix_java/bin"]
COPY --from=builder ["/tmp/zabbix-${ZBX_VERSION}/src/zabbix_java/lib", "/usr/sbin/zabbix_java/lib"]

RUN --mount=type=tmpfs,target=/var/lib/dnf/ \
    set -eux && \
    INSTALL_PKGS="bash \
            findutils \
            shadow-utils \
            java-17-openjdk-headless" && \
    microdnf -y install \
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
            --home-dir /var/lib/zabbix/ \
        zabbix && \
    mkdir -p /etc/zabbix/ && \
    mkdir -p /usr/sbin/zabbix_java/ && \
    mkdir -p /usr/sbin/zabbix_java/ext_lib/ && \
    rm -rf /usr/sbin/zabbix_java/lib/logback.xml && \
    mv /usr/sbin/zabbix_java/lib/logback-console.xml /etc/zabbix/zabbix_java_gateway_logback.xml && \
    chown --quiet -R zabbix:root /etc/zabbix/ /usr/sbin/zabbix_java/ && \
    chgrp -R 0 /etc/zabbix/ /usr/sbin/zabbix_java/ && \
    chmod -R g=u /etc/zabbix/ /usr/sbin/zabbix_java/ && \
    microdnf -y clean all

EXPOSE 10052/TCP

WORKDIR /var/lib/zabbix

COPY ["conf/usr/sbin/zabbix_java_gateway", "/usr/sbin/"]
COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENTRYPOINT ["docker-entrypoint.sh"]

USER 1997

CMD ["/usr/sbin/zabbix_java_gateway"]
