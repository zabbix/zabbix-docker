# syntax=docker/dockerfile:1
FROM ubuntu:focal

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
    apt-get -y update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y --no-install-recommends install \
            bash \
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
            libpcre3-dev \
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
            tar \
            wget \
            unixodbc-dev && \
    ARCH_SUFFIX="$(dpkg --print-architecture)"; arch="${arch##*-}"; \
    case "$ARCH_SUFFIX" in \
        'amd64') \
            url='https://dl.google.com/go/go1.17.1.linux-amd64.tar.gz'; \
            sha256='dab7d9c34361dc21ec237d584590d72500652e7c909bf082758fb63064fca0ef'; \
            ;; \
        'armhf') \
            url='https://dl.google.com/go/go1.17.1.linux-armv6l.tar.gz'; \
            sha256='ed3e4dbc9b80353f6482c441d65b51808290e94ff1d15d56da5f4a7be7353758'; \
            ;; \
        'arm64') \
            url='https://dl.google.com/go/go1.17.1.linux-arm64.tar.gz'; \
            sha256='53b29236fa03ed862670a5e5e2ab2439a2dc288fe61544aa392062104ac0128c'; \
            ;; \
        'ppc64el') \
            url='https://dl.google.com/go/go1.17.1.linux-ppc64le.tar.gz'; \
            sha256='df4fa945512c3b472cf3d2dcb2e4ae5b34819607bc63f3223f5bc0c17b637dd0'; \
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
