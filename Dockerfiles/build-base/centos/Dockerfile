# syntax=docker/dockerfile:1
FROM centos:8

ARG MAJOR_VERSION=6.0
ARG ZBX_VERSION=${MAJOR_VERSION}
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} \
    PATH=/usr/local/go/bin:$PATH

LABEL org.opencontainers.image.title="Zabbix build base" \
      org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.vendor="Zabbix LLC" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.description="Zabbix build base image contains all required packages to build Zabbix images" \
      org.opencontainers.image.licenses="GPL v2.0" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.version="${ZBX_VERSION}" \
      org.opencontainers.image.source="${ZBX_SOURCES}"

RUN set -eux && \
    sed -i 's/enabled=0/enabled=1/g' /etc/yum.repos.d/CentOS-Linux-PowerTools.repo && \
    dnf --quiet makecache && \
    dnf -y install epel-release && \
    dnf -y module enable mysql && \
    INSTALL_PKGS="autoconf \
            automake \
            bash \
            gcc \
            pcre-devel \
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
    dnf -y install --setopt=tsflags=nodocs --setopt=install_weak_deps=False --best \
        ${INSTALL_PKGS} && \
    ARCH_SUFFIX="$(arch)"; \
    case "$ARCH_SUFFIX" in \
        x86_64) \
            url='https://dl.google.com/go/go1.17.2.linux-amd64.tar.gz'; \
            sha256='f242a9db6a0ad1846de7b6d94d507915d14062660616a61ef7c808a76e4f1676'; \
            ;; \
        aarch64) \
            url='https://dl.google.com/go/go1.17.2.linux-arm64.tar.gz'; \
            sha256='a5a43c9cdabdb9f371d56951b14290eba8ce2f9b0db48fb5fc657943984fd4fc'; \
            ;; \
        ppc64le) \
            url='https://golang.org/dl/go1.17.2.linux-ppc64le.tar.gz'; \
            sha256='12e2dc7e0ffeebe77083f267ef6705fec1621cdf2ed6489b3af04a13597ed68d'; \
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
    rm go.tgz
