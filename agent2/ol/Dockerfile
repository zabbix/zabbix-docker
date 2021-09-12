FROM oraclelinux:8-slim

LABEL org.opencontainers.image.title="Zabbix agent 2" \
      org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.vendor="Zabbix LLC" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.description="Zabbix agent 2 is deployed on a monitoring target to actively monitor local resources and applications" \
      org.opencontainers.image.licenses="GPL v2.0"

STOPSIGNAL SIGTERM

COPY ["conf/etc/yum.repos.d/oracle-epel-ol8.repo", "/etc/yum.repos.d/oracle-epel-ol8.repo"]

RUN set -eux && \
    groupadd --system --gid 1995 zabbix && \
    useradd \
            --system --comment "Zabbix monitoring system" \
            -g zabbix -G root \
            --uid 1997 \
            --shell /sbin/nologin \
            --home-dir /var/lib/zabbix/ \
        zabbix && \
    mkdir -p /etc/zabbix && \
    mkdir -p /etc/zabbix/zabbix_agentd.d && \
    mkdir -p /var/lib/zabbix && \
    mkdir -p /var/lib/zabbix/enc && \
    mkdir -p /var/lib/zabbix/modules && \
    mkdir -p /var/lib/zabbix/buffer && \
    mkdir -p /etc/yum/vars/ && \
    echo "" > /etc/yum/vars/ociregion && \
    echo "oracle.com" > /etc/yum/vars/ocidomain && \
    INSTALL_PKGS="bash \
            tini \
            tzdata \
            iputils \
            pcre \
            libcurl \
            openssl-libs" && \
    microdnf -y --disablerepo="*" --enablerepo="ol8_baseos_latest" \
        --enablerepo="ol8_appstream" \
        --enablerepo="ol8_developer_EPEL" \
        install --setopt=install_weak_deps=0 --best --nodocs ${INSTALL_PKGS} && \
    microdnf -y clean all && \
    rm -rf /var/cache/yum /var/lib/yum/yumdb/* /usr/lib/udev/hwdb.d/* && \
    rm -rf /var/cache/dnf /etc/udev/hwdb.bin /root/.pki

ARG MAJOR_VERSION=6.0
ARG ZBX_VERSION=${MAJOR_VERSION}
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES}

LABEL org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.version="${ZBX_VERSION}" \
      org.opencontainers.image.source="${ZBX_SOURCES}"

RUN set -eux && REPOLIST="ol8_baseos_latest,ol8_appstream" && \
    PATH=/usr/local/go/bin:$PATH && \
    INSTALL_PKGS="autoconf \
            automake \
            pcre-devel \
            libcurl-devel \
            make \
            openssl-devel \
            openldap-devel \
            git \
            wget \
            tar \
            gcc" && \
    microdnf -y install --setopt=install_weak_deps=0 --best \
        --nodocs dnf && \
    dnf -y install --disablerepo "*" --enablerepo "${REPOLIST}" --setopt=install_weak_deps=False --best \
        --setopt=tsflags=nodocs ${INSTALL_PKGS} && \
    ARCH_SUFFIX="$(arch)"; \
    case "$ARCH_SUFFIX" in \
        x86_64) \
            url='https://dl.google.com/go/go1.17.1.linux-amd64.tar.gz'; \
            sha256='dab7d9c34361dc21ec237d584590d72500652e7c909bf082758fb63064fca0ef'; \
            ;; \
        aarch64) \
            url='https://dl.google.com/go/go1.17.1.linux-arm64.tar.gz'; \
            sha256='53b29236fa03ed862670a5e5e2ab2439a2dc288fe61544aa392062104ac0128c'; \
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
    cd /tmp/ && \
    git -c advice.detachedHead=false clone ${ZBX_SOURCES} --branch master --depth 1 --single-branch zabbix-${ZBX_VERSION} && \
    cd /tmp/zabbix-${ZBX_VERSION} && \
    zabbix_revision=`git rev-parse --short HEAD` && \
    sed -i "s/{ZABBIX_REVISION}/$zabbix_revision/g" src/go/pkg/version/version.go && \
    ./bootstrap.sh && \
    export CFLAGS="-fPIC -pie -Wl,-z,relro -Wl,-z,now" && \
    export GOPATH=/tmp/zabbix-${ZBX_VERSION}/go && \
    ./configure \
            --datadir=/usr/lib \
            --libdir=/usr/lib/zabbix \
            --prefix=/usr \
            --sysconfdir=/etc/zabbix \
            --prefix=/usr \
            --with-openssl \
            --enable-ipv6 \
            --enable-agent2 \
            --enable-agent \
            --silent && \
    make -j"$(nproc)" -s && \
    cp /tmp/zabbix-${ZBX_VERSION}/src/go/bin/zabbix_agent2 /usr/sbin/zabbix_agent2 && \
    cp /tmp/zabbix-${ZBX_VERSION}/src/zabbix_get/zabbix_get /usr/bin/zabbix_get && \
    cp /tmp/zabbix-${ZBX_VERSION}/src/zabbix_sender/zabbix_sender /usr/bin/zabbix_sender && \
    cp /tmp/zabbix-${ZBX_VERSION}/src/go/conf/zabbix_agent2.conf /etc/zabbix/zabbix_agent2.conf && \
    strip /usr/sbin/zabbix_agent2 && \
    strip /usr/bin/zabbix_get && \
    strip /usr/bin/zabbix_sender && \
    cd /tmp/ && \
    rm -rf /tmp/zabbix-${ZBX_VERSION}/ && \
    rm -rf /usr/local/go/ && \
    dnf -y history undo last && \
    dnf -y clean all && \
    microdnf -y remove dnf && \
    chown --quiet -R zabbix:root /etc/zabbix/ /var/lib/zabbix/ && \
    chgrp -R 0 /etc/zabbix/ /var/lib/zabbix/ && \
    chmod -R g=u /etc/zabbix/ /var/lib/zabbix/ && \
    rm -rf /var/cache/yum /var/lib/yum/yumdb/* /usr/lib/udev/hwdb.d/* && \
    rm -rf /var/cache/dnf /etc/udev/hwdb.bin /root/.pki

EXPOSE 10050/TCP 31999/TCP

WORKDIR /var/lib/zabbix

VOLUME ["/var/lib/zabbix/enc"]

COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENTRYPOINT ["/usr/bin/tini", "--", "/usr/bin/docker-entrypoint.sh"]

USER 1997

CMD ["/usr/sbin/zabbix_agent2", "--foreground", "-c", "/etc/zabbix/zabbix_agent2.conf"]
