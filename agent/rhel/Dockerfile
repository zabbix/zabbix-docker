FROM registry.access.redhat.com/ubi8/ubi
MAINTAINER Alexey Pustovalov <alexey.pustovalov@zabbix.com>

ARG DNF_FLAGS_COMMON="-y"
ARG DNF_FLAGS_PERSISTENT="${DNF_FLAGS_COMMON}"
ARG DNF_FLAGS_DEV="${DNF_FLAGS_COMMON}"

ARG BUILD_DATE
ARG VCS_REF

ARG MAJOR_VERSION=4.4
ARG RELEASE=7
ARG ZBX_VERSION=${MAJOR_VERSION}.7

ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ENV ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES}

LABEL name="zabbix/zabbix-agent-44" \
      maintainer="alexey.pustovalov@zabbix.com" \
      vendor="Zabbix LLC" \
      version="${MAJOR_VERSION}" \
      release="${RELEASE}" \
      summary="Zabbix agent" \
      description="Zabbix agent is deployed on a monitoring target to actively monitor local resources and applications" \
      url="https://www.zabbix.com/" \
      run="docker run --name zabbix-agent --link zabbix-server:zabbix-server -p 10050:10050 -d registry.connect.redhat.com/zabbix/zabbix-agent-44:${ZBX_VERSION}" \
      io.k8s.description="Zabbix agent is deployed on a monitoring target to actively monitor local resources and applications" \
      io.k8s.display-name="Zabbix Agent" \
      io.openshift.expose-services="10050:10050" \
      io.openshift.tags="zabbix,zabbix-agent" \
      org.label-schema.name="zabbix-agent-rhel" \
      org.label-schema.vendor="Zabbix LLC" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.description="Zabbix agent is deployed on a monitoring target to actively monitor local resources and applications" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.license="GPL v2.0" \
      org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.version="${ZBX_VERSION}" \
      org.label-schema.vcs-url="${ZBX_SOURCES}" \
      org.label-schema.docker.cmd="docker run --name zabbix-agent --link zabbix-server:zabbix-server -p 10050:10050 -d registry.connect.redhat.com/zabbix/zabbix-agent-44:${ZBX_VERSION}"

STOPSIGNAL SIGTERM

### add licenses to this directory
COPY ["licenses", "/licenses"]

ENV TINI_VERSION v0.18.0
COPY ["conf/sbin/tini","/sbin/tini"]

RUN set -eux && INSTALL_PKGS="bash \
            iputils \
            pcre \
            libcurl" && \
    REPOLIST="rhel-8-for-x86_64-baseos-rpms,rhel-8-for-x86_64-appstream-rpms" && \
    dnf ${DNF_FLAGS_COMMON} update-minimal --disablerepo "*" --enablerepo rhel-8-for-x86_64-baseos-rpms --setopt=tsflags=nodocs \
        --security --sec-severity=Important --sec-severity=Critical && \
    dnf ${DNF_FLAGS_COMMON} install --disablerepo "*" --enablerepo "${REPOLIST}" --setopt=tsflags=nodocs ${INSTALL_PKGS} && \
    groupadd -g 1995 --system zabbix && \
    adduser -r --shell /sbin/nologin \
            -g zabbix -G dialout -G root \
            -d /var/lib/zabbix/ -u 1997 \
        zabbix && \
    mkdir -p /etc/zabbix && \
    mkdir -p /etc/zabbix/zabbix_agentd.d && \
    mkdir -p /var/lib/zabbix && \
    mkdir -p /var/lib/zabbix/enc && \
    mkdir -p /var/lib/zabbix/modules && \
    chmod +x /sbin/tini && \
    dnf ${DNF_FLAGS_COMMON} clean all && \
    rm -rf /var/cache/yum /var/lib/yum/yumdb/* /usr/lib/udev/hwdb.d/* && \
    rm -rf /var/cache/dnf /etc/udev/hwdb.bin /root/.pki

ARG MAJOR_VERSION=4.4
ARG ZBX_VERSION=${MAJOR_VERSION}.7
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

RUN set -eux && REPOLIST="rhel-8-for-x86_64-baseos-rpms,rhel-8-for-x86_64-appstream-rpms" && \
    INSTALL_PKGS="autoconf \
            automake \
            libcurl-devel \
            openssl-devel \
            openldap-devel \
            gcc \
            pcre-devel \
            make \
            git" && \
    dnf ${DNF_FLAGS_COMMON} install --disablerepo "*" --enablerepo "${REPOLIST}" --setopt=tsflags=nodocs ${INSTALL_PKGS} && \
    cd /tmp/ && \
    git clone ${ZBX_SOURCES} --branch ${ZBX_VERSION} --depth 1 --single-branch zabbix-${ZBX_VERSION} && \
    cd /tmp/zabbix-${ZBX_VERSION} && \
    zabbix_revision=`git rev-parse --short HEAD` && \
    sed -i "s/{ZABBIX_REVISION}/$zabbix_revision/g" include/version.h && \
    ./bootstrap.sh && \
    export CFLAGS="-fPIC -pie -Wl,-z,relro -Wl,-z,now" && \
    ./configure \
            --datadir=/usr/lib \
            --libdir=/usr/lib/zabbix \
            --prefix=/usr \
            --sysconfdir=/etc/zabbix \
            --prefix=/usr \
            --enable-agent \
            --with-libcurl \
            --with-ldap \
            --with-openssl \
            --enable-ipv6 \
            --silent && \
    make -j"$(nproc)" -s && \
    cp /tmp/zabbix-${ZBX_VERSION}/src/zabbix_agent/zabbix_agentd /usr/sbin/zabbix_agentd && \
    cp /tmp/zabbix-${ZBX_VERSION}/src/zabbix_get/zabbix_get /usr/bin/zabbix_get && \
    cp /tmp/zabbix-${ZBX_VERSION}/src/zabbix_sender/zabbix_sender /usr/bin/zabbix_sender && \
    cp /tmp/zabbix-${ZBX_VERSION}/conf/zabbix_agentd.conf /etc/zabbix/zabbix_agentd.conf && \
    cd /tmp/ && \
    rm -rf /tmp/zabbix-${ZBX_VERSION}/ && \
    dnf ${DNF_FLAGS_COMMON} history undo `dnf -q history | sed -n 3p |column -t | cut -d' ' -f1` && \
    dnf ${DNF_FLAGS_COMMON} clean all && \
    chown --quiet -R zabbix:root /etc/zabbix/ /var/lib/zabbix/ && \
    chgrp -R 0 /etc/zabbix/ /var/lib/zabbix/ && \
    chmod -R g=u /etc/zabbix/ /var/lib/zabbix/ && \
    rm -rf /var/cache/yum /var/lib/yum/yumdb/* /usr/lib/udev/hwdb.d/* && \
    rm -rf /var/cache/dnf /etc/udev/hwdb.bin /root/.pki

EXPOSE 10050/TCP

WORKDIR /var/lib/zabbix

VOLUME ["/var/lib/zabbix/enc"]

COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENTRYPOINT ["/sbin/tini", "--", "/usr/bin/docker-entrypoint.sh"]

USER 1997

CMD ["/usr/sbin/zabbix_agentd", "--foreground", "-c", "/etc/zabbix/zabbix_agentd.conf"]
