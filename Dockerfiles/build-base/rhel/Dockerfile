# syntax=docker/dockerfile:1
FROM registry.access.redhat.com/ubi8/ubi-minimal

ARG MAJOR_VERSION=6.2
ARG RELEASE=4
ARG ZBX_VERSION=${MAJOR_VERSION}.4

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
    PATH=/usr/local/go/bin:$PATH

LABEL description="Prepared environment to build Zabbix components" \
      maintainer="alexey.pustovalov@zabbix.com" \
      name="zabbix/zabbix-build-base-62" \
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

RUN set -eux && \
    INSTALL_PKGS="autoconf \
            automake \
            bash \
            gcc \
            go-toolset \
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
            java-1.8.0-openjdk-devel \
            git \
            gettext \
            unixODBC-devel" && \
    microdnf -y module enable mysql && \
    microdnf -y install \
            --disablerepo "*" \
            --enablerepo "ubi-8-baseos" \
            --enablerepo "ubi-8-appstream" \
            --enablerepo "rhel-8-for-x86_64-baseos-rpms" \
            --enablerepo "rhel-8-for-x86_64-appstream-rpms" \
            --enablerepo "codeready-builder-for-rhel-8-x86_64-rpms" \
            --setopt=install_weak_deps=0 \
            --best \
            --setopt=tsflags=nodocs \
        ${INSTALL_PKGS} && \
    microdnf -y clean all && \
    rm -rf /var/cache/yum /var/lib/yum/yumdb/* /usr/lib/udev/hwdb.d/* && \
    rm -rf /var/cache/dnf /etc/udev/hwdb.bin /root/.pki
