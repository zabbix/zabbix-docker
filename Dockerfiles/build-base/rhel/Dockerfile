# syntax=docker/dockerfile:1
FROM registry.access.redhat.com/ubi9/ubi-minimal:9.3

ARG MAJOR_VERSION=6.4
ARG RELEASE=11
ARG ZBX_VERSION=${MAJOR_VERSION}.11

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
    PATH=/usr/local/go/bin:$PATH

LABEL description="Prepared environment to build Zabbix components" \
      maintainer="alexey.pustovalov@zabbix.com" \
      name="zabbix/zabbix-build-base-64" \
      release="${RELEASE}" \
      summary="Zabbix build base" \
      url="https://www.zabbix.com/" \
      vendor="Zabbix LLC" \
      version="${MAJOR_VERSION}" \
      io.k8s.description="Prepared environment to build Zabbix components" \
      io.k8s.display-name="Zabbix build base" \
      io.openshift.expose-services="" \
      io.openshift.tags="zabbix,build" \
      org.label-schema.description="Prepared environment to build Zabbix components" \
      org.label-schema.name="zabbix-build-base-rhel" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.vendor="Zabbix LLC"

COPY ["licenses", "/licenses"]

RUN --mount=type=tmpfs,target=/var/lib/dnf/ \
    set -eux && \
    INSTALL_PKGS="autoconf \
            automake \
            bash \
            gcc \
            go-toolset \
            patch \
            pcre2-devel \
            libcurl-devel \
            libevent-devel \
            libssh-devel \
            libxml2-devel \
            openssl-devel \
            openldap-devel \
            make \
            mysql-devel \
            net-snmp-devel \
            OpenIPMI-devel \
            openldap-devel \
            sqlite-devel \
            postgresql-devel \
            java-17-openjdk-devel \
            git \
            gettext \
            unixODBC-devel" && \
    ARCH_SUFFIX="$(arch)"; \
    microdnf -y install \
            --disablerepo "*" \
            --enablerepo "ubi-9-baseos-rpms" \
            --enablerepo "ubi-9-appstream-rpms" \
            --enablerepo "rhel-9-for-$ARCH_SUFFIX-baseos-rpms" \
            --enablerepo "rhel-9-for-$ARCH_SUFFIX-appstream-rpms" \
            --enablerepo "codeready-builder-for-rhel-9-$ARCH_SUFFIX-rpms" \
            --setopt=install_weak_deps=0 \
            --setopt=keepcache=0 \
            --best \
            --setopt=tsflags=nodocs \
        ${INSTALL_PKGS} && \
    microdnf -y clean all
