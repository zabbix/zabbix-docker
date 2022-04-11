# syntax=docker/dockerfile:1
FROM ubuntu:focal

ARG MAJOR_VERSION=6.0
ARG ZBX_VERSION=${MAJOR_VERSION}.3

ENV TERM=xterm \
    ZBX_VERSION=${ZBX_VERSION}  \
    PATH=/usr/local/go/bin:$PATH

LABEL org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix build base image contains all required packages to build Zabbix images" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="GPL v2.0" \
      org.opencontainers.image.title="Zabbix build base" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix LLC" \
      org.opencontainers.image.version="${ZBX_VERSION}"

RUN set -eux && \
    INSTALL_PKGS="bash \
            autoconf \
            automake \
            binutils \
            gcc \
            libc6-dev \
            libcurl4-openssl-dev \
            libevent-dev \
            libmodbus-dev \
            libldap2-dev \
            libmysqlclient-dev \
            libopenipmi-dev \
            libpcre2-dev \
            libsnmp-dev \
            libsqlite3-dev \
            libpq-dev \
            libssh-dev \
            libxml2-dev \
            make \
            openjdk-8-jdk-headless \
            pkg-config \
            gettext \
            git \
            gnupg \
            curl \
            unixodbc-dev" && \
    apt-get -y update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y \
            --no-install-recommends install \
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
        armv7l) \
            url='https://dl.google.com/go/go1.17.8.linux-armv6l.tar.gz'; \
            sha256='3287ca2fe6819fa87af95182d5942bf4fa565aff8f145812c6c70c0466ce25ae'; \
            ;; \
        s390x) \
            url='https://dl.google.com/go/go1.17.8.linux-s390x.tar.gz'; \
            sha256='3fac23801644a2f93a1643acecd5a94a5ea05d88e19467092fb6e64205710f61'; \
            ;; \
        ppc64le) \
            url='https://dl.google.com/go/go1.17.8.linux-ppc64le.tar.gz'; \
            sha256='2077dd2fc57a74b0630b0c239ae4e3114607311778effd43fcfe5174133ee188'; \
            ;; \
        *) echo "Unknown ARCH_SUFFIX=${ARCH_SUFFIX-}"; exit 1 ;; \
    esac; \
    curl "$url.asc" --output go.tgz.asc && \
    curl "$url" --output go.tgz && \
    echo "$sha256 *go.tgz" | sha256sum -c - && \
    GNUPGHOME="$(mktemp -d)"; export GNUPGHOME; \
    gpg --batch --keyserver keyserver.ubuntu.com --recv-keys 'EB4C 1BFD 4F04 2F6D DDCC EC91 7721 F63B D38B 4796' && \
    gpg --batch --verify go.tgz.asc go.tgz && \
    gpgconf --kill all && \
    rm -rf "$GNUPGHOME" go.tgz.asc && \
    tar -C /usr/local -xzf go.tgz && \
    rm go.tgz && \
    apt-get -y autoremove && \
    apt-get -y clean && \
    rm -rf /var/lib/apt/lists/*
