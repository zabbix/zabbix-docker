FROM registry.access.redhat.com/ubi8/ubi
MAINTAINER Alexey Pustovalov <alexey.pustovalov@zabbix.com>

ARG MAJOR_VERSION=6.0
ARG RELEASE=
ARG ZBX_VERSION=${MAJOR_VERSION}

ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ENV ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES}

LABEL name="zabbix/zabbix-agent2-trunk" \
      maintainer="alexey.pustovalov@zabbix.com" \
      vendor="Zabbix LLC" \
      version="${MAJOR_VERSION}" \
      release="${RELEASE}" \
      summary="Zabbix agent" \
      description="Zabbix agent 2 is deployed on a monitoring target to actively monitor local resources and applications" \
      url="https://www.zabbix.com/" \
      run="docker run --name zabbix-agent2 --link zabbix-server:zabbix-server -p 10050:10050 -d registry.connect.redhat.com/zabbix/zabbix-agent2-trunk:${ZBX_VERSION}" \
      io.k8s.description="Zabbix agent 2 is deployed on a monitoring target to actively monitor local resources and applications" \
      io.k8s.display-name="Zabbix Agent 2" \
      io.openshift.expose-services="10050:10050" \
      io.openshift.tags="zabbix,zabbix-agent" \
      org.label-schema.name="zabbix-agent2-rhel" \
      org.label-schema.vendor="Zabbix LLC" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.description="Zabbix agent 2 is deployed on a monitoring target to actively monitor local resources and applications" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.license="GPL v2.0" \
      org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.version="${ZBX_VERSION}" \
      org.label-schema.vcs-url="${ZBX_SOURCES}" \
      org.label-schema.docker.cmd="docker run --name zabbix-agent2 --link zabbix-server:zabbix-server -p 10050:10050 -d registry.connect.redhat.com/zabbix/zabbix-agent2-trunk:${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY ["licenses", "/licenses"]

RUN set -eux && INSTALL_PKGS="bash \
            tini \
            tzdata \
            iputils \
            pcre \
            libcurl" && \
    REPOLIST="ubi-8-baseos,ubi-8-appstream,epel" && \
    dnf -y update-minimal --disablerepo "*" --enablerepo "ubi-8-baseos" --setopt=tsflags=nodocs \
        --security --sec-severity=Important --setopt=install_weak_deps=False --sec-severity=Critical && \
    dnf -y install https://dl.fedoraproject.org/pub/epel/epel-release-latest-8.noarch.rpm && \
    dnf -y install --disablerepo "*" --enablerepo "${REPOLIST}" --setopt=install_weak_deps=False --best \
        --setopt=tsflags=nodocs ${INSTALL_PKGS} && \
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
    dnf -y clean all && \
    rm -rf /var/cache/yum /var/lib/yum/yumdb/* /usr/lib/udev/hwdb.d/* && \
    rm -rf /var/cache/dnf /etc/udev/hwdb.bin /root/.pki

RUN set -eux && REPOLIST="ubi-8-baseos,ubi-8-appstream" && \
    PATH=/usr/local/go/bin:$PATH && \
    INSTALL_PKGS="autoconf \
            automake \
            libcurl-devel \
            openssl-devel \
            openldap-devel \
            gcc \
            pcre-devel \
            make \
            git \
            wget \
            tar \
            binutils" && \
    dnf -y install --disablerepo "*" --enablerepo "${REPOLIST}" --setopt=install_weak_deps=False --best \
        --setopt=tsflags=nodocs ${INSTALL_PKGS} && \
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
