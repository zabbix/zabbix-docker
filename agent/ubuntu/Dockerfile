FROM ubuntu:bionic as builder

ARG APT_FLAGS_COMMON="-y"
ARG APT_FLAGS_DEV="${APT_FLAGS_COMMON} --no-install-recommends"

ARG MAJOR_VERSION=4.2
ARG ZBX_VERSION=${MAJOR_VERSION}.8
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ENV LANG=en_US.UTF-8 LANGUAGE=en_US:en LC_ALL=en_US.UTF-8 TERM=xterm \
    ZBX_TYPE=agent ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES}

RUN set -eux && \
    apt-get ${APT_FLAGS_COMMON} update && \
    DEBIAN_FRONTEND=noninteractive apt-get ${APT_FLAGS_DEV} install locales && \
    locale-gen $LC_ALL && \
    DEBIAN_FRONTEND=noninteractive apt-get ${APT_FLAGS_DEV} install \
            autoconf \
            automake \
            libcurl4-openssl-dev \
            libc6-dev \
            libldap2-dev \
            libpcre3-dev \
            libssl-dev \
            make \
            pkg-config \
            git \
            ca-certificates \
            gcc && \
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
            --sysconfdir=/etc/zabbix \
            --prefix=/usr \
            --enable-agent \
            --with-libcurl \
            --with-ldap \
            --with-openssl \
            --enable-ipv6 \
            --silent && \
    make -j"$(nproc)" -s

FROM ubuntu:bionic
LABEL maintainer="Alexey Pustovalov <alexey.pustovalov@zabbix.com>"

ARG BUILD_DATE
ARG VCS_REF

ARG APT_FLAGS_COMMON="-y"
ARG APT_FLAGS_PERSISTENT="${APT_FLAGS_COMMON} --no-install-recommends"

ARG MAJOR_VERSION=4.2
ARG ZBX_VERSION=${MAJOR_VERSION}.8
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ENV LANG=en_US.UTF-8 LANGUAGE=en_US:en LC_ALL=en_US.UTF-8 TERM=xterm \
    ZBX_TYPE=agent ZBX_DB_TYPE=none ZBX_OPT_TYPE=none \
    ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES}
ENV TINI_VERSION v0.18.0

LABEL org.label-schema.name="zabbix-${ZBX_TYPE}-ubuntu" \
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
      org.label-schema.docker.cmd="docker run --name zabbix-${ZBX_TYPE} --link zabbix-server:zabbix-server -p 10050:10050 -d zabbix-${ZBX_TYPE}:ubuntu-${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY --from=builder /tmp/zabbix-${ZBX_VERSION}/src/zabbix_agent/zabbix_agentd /usr/sbin/zabbix_agentd
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}/src/zabbix_get/zabbix_get /usr/bin/zabbix_get
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}/src/zabbix_sender/zabbix_sender /usr/bin/zabbix_sender
COPY --from=builder  /tmp/zabbix-${ZBX_VERSION}/conf/zabbix_agentd.conf /etc/zabbix/zabbix_agentd.conf
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /sbin/tini

RUN set -eux && \
    apt-get ${APT_FLAGS_COMMON} update && \
    DEBIAN_FRONTEND=noninteractive apt-get ${APT_FLAGS_PERSISTENT} install locales && \
    locale-gen $LC_ALL && \
    echo "#!/bin/sh\nexit 0" > /usr/sbin/policy-rc.d && \
    addgroup --system --quiet zabbix && \
    adduser --quiet \
            --system --disabled-login \
            --ingroup zabbix \
            --home /var/lib/zabbix/ \
        zabbix && \
    mkdir -p /etc/zabbix && \
    mkdir -p /etc/zabbix/zabbix_agentd.d && \
    mkdir -p /var/lib/zabbix && \
    mkdir -p /var/lib/zabbix/enc && \
    mkdir -p /var/lib/zabbix/modules && \
    chown --quiet -R zabbix:root /var/lib/zabbix && \
    apt-get ${APT_FLAGS_COMMON} update && \
    DEBIAN_FRONTEND=noninteractive apt-get ${APT_FLAGS_PERSISTENT} install \
            libpcre3 \
            libssl1.1 \
            libcurl4 \
            libldap-2.4 && \
    apt-get ${APT_FLAGS_COMMON} autoremove && \
    apt-get ${APT_FLAGS_COMMON} clean && \
    rm -rf /var/lib/apt/lists/*  && \
    chmod +x /sbin/tini

EXPOSE 10050/TCP

WORKDIR /var/lib/zabbix

VOLUME ["/etc/zabbix/zabbix_agentd.d", "/var/lib/zabbix/enc", "/var/lib/zabbix/modules"]

COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENTRYPOINT ["/sbin/tini", "--", "/usr/bin/docker-entrypoint.sh"]
