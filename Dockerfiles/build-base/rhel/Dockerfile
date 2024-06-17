# syntax=docker/dockerfile:1
FROM registry.access.redhat.com/ubi8/ubi-minimal:8.10

ARG MAJOR_VERSION=6.0
ARG RELEASE=31
ARG ZBX_VERSION=${MAJOR_VERSION}.31

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
    PATH=/usr/local/go/bin:$PATH

LABEL description="Prepared environment to build Zabbix components" \
      maintainer="alexey.pustovalov@zabbix.com" \
      name="zabbix/zabbix-build-base-60" \
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
      org.label-schema.name="zabbix-build-base-rhel" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.vendor="Zabbix SIA"

COPY ["licenses", "/licenses"]

RUN --mount=type=tmpfs,target=/var/lib/dnf/ \
    --mount=type=bind,target=/run/secrets/,src=secrets/ \
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
            libmodbus-devel \
            libxml2-devel \
            openssl-devel \
            openldap-devel \
            make \
            mysql-devel \
            net-snmp-devel \
            OpenIPMI-devel \
            openldap-devel \
            sqlite-devel \
            java-17-openjdk-devel \
            git \
            gettext \
            unixODBC-devel \
            zlib-devel" && \
    curl --tlsv1.2 -sSf -L https://dl.fedoraproject.org/pub/epel/epel-release-latest-8.noarch.rpm -o /tmp/epel-release-latest-8.noarch.rpm && \
    rpm -ivh /tmp/epel-release-latest-8.noarch.rpm && \
    rm -rf /tmp/epel-release-latest-8.noarch.rpm && \
    microdnf -y module enable \
        --disablerepo "*" \
        --enablerepo "rhel-8-for-x86_64-appstream-rpms" \
        --setopt=install_weak_deps=0 \
        --setopt=keepcache=0 \
        mysql && \
    microdnf -y install \
            --disablerepo "*" \
            --enablerepo "ubi-8-baseos-rpms" \
            --enablerepo "ubi-8-appstream-rpms" \
            --enablerepo "rhel-8-for-x86_64-baseos-rpms" \
            --enablerepo "rhel-8-for-x86_64-appstream-rpms" \
            --enablerepo "codeready-builder-for-rhel-8-x86_64-rpms" \
            --enablerepo "epel" \
            --setopt=install_weak_deps=0 \
            --setopt=keepcache=0 \
            --best \
            --setopt=tsflags=nodocs \
        ${INSTALL_PKGS} && \
    microdnf -y clean all
