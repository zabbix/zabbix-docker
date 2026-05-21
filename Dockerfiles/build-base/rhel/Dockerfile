# syntax=docker/dockerfile:1.6
ARG OS_BASE_IMAGE=registry.access.redhat.com/ubi10/ubi-minimal:10.2
FROM ${OS_BASE_IMAGE}

ARG MAJOR_VERSION=8.0
ARG RELEASE=0
ARG ZBX_VERSION=${MAJOR_VERSION}

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
    GOPATH=/root/go \
    GOCACHE=/root/.cache/go-build \
    GOMODCACHE=/root/go/pkg/mod \
    GOTELEMETRY=off \
    GOFLAGS=-buildvcs=false

LABEL description="Prepared environment to build Zabbix components" \
      maintainer="alexey.pustovalov@zabbix.com" \
      name="zabbix/zabbix-build-base-trunk" \
      release="${RELEASE}" \
      summary="Zabbix build base" \
      url="https://www.zabbix.com/" \
      vendor="Zabbix SIA" \
      version="${MAJOR_VERSION}" \
      io.k8s.description="Prepared environment to build Zabbix components" \
      io.k8s.display-name="Zabbix build base" \
      io.openshift.expose-services="" \
      io.openshift.tags="zabbix,build" \
      org.label-schema.description="Prepared environment to build Zabbix components" \
      org.label-schema.license="AGPL v3.0" \
      org.label-schema.name="zabbix-build-base-rhel" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.vendor="Zabbix SIA"

COPY licenses /licenses/

RUN --mount=type=tmpfs,target=/var/lib/dnf/ \
    --mount=type=tmpfs,target=/var/cache/yum/ \
#    --mount=type=bind,target=/run/secrets/,src=secrets/ \
    set -eux && \
    INSTALL_PKGS="autoconf \
            automake \
            bash \
            gcc \
            go-toolset \
            patch \
            c-ares-devel \
            pcre2-devel \
            libcurl-devel \
            libevent-devel \
            libssh-devel \
            libmodbus-devel \
            libxml2-devel \
            openssl-devel \
            make \
            mysql8.4-devel \
            net-snmp-devel \
            OpenIPMI-devel \
            openldap-devel \
            sqlite-devel \
            libpq-devel \
            java-21-openjdk-devel \
            git \
            gettext \
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
