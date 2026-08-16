# syntax=docker/dockerfile:1.6
ARG OS_BASE_IMAGE=registry.access.redhat.com/ubi10/ubi-minimal:10.2
FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION=7.4
ARG RELEASE=13
ARG ZBX_VERSION=${MAJOR_VERSION}.13

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
    GOPATH=/root/go \
    GOCACHE=/root/.cache/go-build \
    GOMODCACHE=/root/go/pkg/mod \
    GOTELEMETRY=off \
    GOFLAGS=-buildvcs=false

LABEL description="Prepared environment to build Zabbix components" \
      maintainer="alexey.pustovalov@zabbix.com" \
      name="zabbix/zabbix-build-base-74" \
      release="${RELEASE}" \
      summary="Zabbix build base" \
      url="https://www.zabbix.com/" \
      vendor="Zabbix SIA" \
      version="${MAJOR_VERSION}" \
      io.k8s.description="Prepared environment to build Zabbix components" \
      io.k8s.display-name="Zabbix build base" \
      io.openshift.expose-services="" \
      io.openshift.tags="zabbix,build" \
      org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Prepared environment to build Zabbix components" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="AGPL-3.0-only" \
      org.opencontainers.image.source="https://github.com/zabbix/zabbix-docker" \
      org.opencontainers.image.title="Zabbix build base" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix SIA" \
      org.opencontainers.image.version="${ZBX_VERSION}"

COPY licenses /licenses/

RUN --mount=type=tmpfs,target=/var/lib/dnf/ \
    --mount=type=tmpfs,target=/var/cache/yum/ \
#    --mount=type=bind,target=/run/secrets/,src=secrets/ \
    set -eux && \
    INSTALL_PKGS="bash \
            autoconf \
            automake \
            make \
            gcc \
            go-toolset \
            java-21-openjdk-devel \
            gettext \
            git \
            patch \
            c-ares-devel \
            libcurl-devel \
            libevent-devel \
            libmodbus-devel \
            libssh-devel \
            libxml2-devel \
            mysql8.4-devel \
            net-snmp-devel \
            OpenIPMI-devel \
            openldap-devel \
            openssl-devel \
            pcre2-devel \
            postgresql18-private-devel \
            postgresql18-server-devel \
            sqlite-devel \
            unixODBC-devel \
            zlib-devel" && \
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
            --enablerepo="codeready-builder-for-rhel-10-$ARCH_SUFFIX-rpms" \
            --enablerepo="epel" \
            --setopt=install_weak_deps=0 \
            --setopt=keepcache=0 \
            --best \
            --setopt=tsflags=nodocs \
        ${INSTALL_PKGS} && \
    go telemetry off
