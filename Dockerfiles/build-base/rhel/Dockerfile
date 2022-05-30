# syntax=docker/dockerfile:1
FROM registry.access.redhat.com/ubi8/ubi-minimal

ARG MAJOR_VERSION=6.0
ARG RELEASE=5
ARG ZBX_VERSION=${MAJOR_VERSION}.5

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION} \
    PATH=/usr/local/go/bin:$PATH

LABEL description="Prepared environment to build Zabbix components" \
      maintainer="alexey.pustovalov@zabbix.com" \
      name="zabbix/zabbix-build-base-60" \
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
            wget \
            tar \
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
    ARCH_SUFFIX="$(arch)"; \
    case "$ARCH_SUFFIX" in \
        x86_64) \
            url='https://dl.google.com/go/go1.17.8.linux-amd64.tar.gz'; \
            sha256='980e65a863377e69fd9b67df9d8395fd8e93858e7a24c9f55803421e453f4f99'; \
            ;; \
        aarch64) \
            url='https://dl.google.com/go/go1.17.8.linux-arm64.tar.gz'; \
            sha256='57a9171682e297df1a5bd287be056ed0280195ad079af90af16dcad4f64710cb'; \
            ;; \
        ppc64le) \
            url='https://dl.google.com/go/go1.17.8.linux-ppc64le.tar.gz'; \
            sha256='2077dd2fc57a74b0630b0c239ae4e3114607311778effd43fcfe5174133ee188'; \
            ;; \
        *) echo "Unknown ARCH_SUFFIX=${ARCH_SUFFIX-}"; exit 1 ;; \
    esac; \
    wget -O go.tgz.asc "$url.asc" && \
    wget -O go.tgz "$url" --progress=dot:giga && \
    echo "$sha256 *go.tgz" | sha256sum -c - && \
    GNUPGHOME="$(mktemp -d)"; export GNUPGHOME; \
    gpg --batch --keyserver keyserver.ubuntu.com --recv-keys 'EB4C 1BFD 4F04 2F6D DDCC EC91 7721 F63B D38B 4796' && \
    gpg --batch --verify go.tgz.asc go.tgz && \
    gpgconf --kill all && \
    rm -rf "$GNUPGHOME" go.tgz.asc && \
    tar -C /usr/local -xzf go.tgz && \
    rm go.tgz && \
    microdnf -y clean all && \
    rm -rf /var/cache/yum /var/lib/yum/yumdb/* /usr/lib/udev/hwdb.d/* && \
    rm -rf /var/cache/dnf /etc/udev/hwdb.bin /root/.pki
