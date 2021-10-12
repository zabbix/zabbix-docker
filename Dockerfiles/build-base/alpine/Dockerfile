# syntax=docker/dockerfile:1
FROM alpine:3.12

ARG MAJOR_VERSION=5.0
ARG ZBX_VERSION=${MAJOR_VERSION}.16
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} \
    PATH=/usr/lib/jvm/default-jvm/bin:/usr/local/go/bin:$PATH \
    JAVA_HOME=/usr/lib/jvm/default-jvm

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
    apk add --no-cache --clean-protected \
            bash \
            autoconf \
            automake \
            coreutils \
            curl-dev \
            libevent-dev \
            libssh-dev \
            sqlite-dev \
            libxml2-dev \
            mysql-dev \
            net-snmp-dev \
            openipmi-dev \
            openldap-dev \
            openssl-dev \
            pcre-dev \
            postgresql-dev \
            openjdk8 \
            gettext \
            git \
            g++ \
            go \
            make \
            openjdk8-jre-base \
            unixodbc-dev && \
    rm -rf /var/cache/apk/*

ARG GOLANG_VERSION=1.17.2
ENV PATH=/usr/local/go/bin:$PATH

RUN set -eux && \
    apk add --no-cache musl-dev \
        openssl-dev \
        gnupg \
        ca-certificates && \
    update-ca-certificates && \
    url='https://dl.google.com/go/go1.17.2.src.tar.gz' && \
    sha256='2255eb3e4e824dd7d5fcdc2e7f84534371c186312e546fb1086a34c17752f431' && \
    wget -O go.tgz.asc "$url.asc" && \
    wget -O go.tgz "$url" && \
    echo "$sha256 *go.tgz" | sha256sum -c - && \
    GNUPGHOME="$(mktemp -d)"; export GNUPGHOME && \
    gpg --batch --keyserver keyserver.ubuntu.com --recv-keys 'EB4C 1BFD 4F04 2F6D DDCC EC91 7721 F63B D38B 4796' && \
    gpg --batch --verify go.tgz.asc go.tgz && \
    gpgconf --kill all && \
    rm -rf "$GNUPGHOME" go.tgz.asc && \
    tar -C /usr/local -xzf go.tgz && \
    rm go.tgz && \
    cd /usr/local/go/src && \
    export GOROOT_BOOTSTRAP="$(go env GOROOT)" && \
    ./make.bash && \
    apk del go && \
    rm -rf /var/cache/apk/*
