FROM ubuntu:bionic

LABEL org.opencontainers.image.title="Zabbix Java Gateway" \
      org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.vendor="Zabbix LLC" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.description="Zabbix Java Gateway performs native support for monitoring JMX applications" \
      org.opencontainers.image.licenses="GPL v2.0"
    
STOPSIGNAL SIGTERM

RUN set -eux && \
    echo "#!/bin/sh\nexit 101" > /usr/sbin/policy-rc.d && \
    groupadd --system --gid 1995 zabbix && \
    useradd \
            --system --comment "Zabbix monitoring system" \
            -g zabbix -G root \
            --uid 1997 \
            --shell /sbin/nologin \
            --home-dir /var/lib/zabbix/ \
        zabbix && \
    mkdir -p /etc/zabbix/ && \
    mkdir -p /usr/sbin/zabbix_java/ && \
    mkdir -p /usr/sbin/zabbix_java/ext_lib/ && \
    apt-get -y update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y --no-install-recommends install \
            openjdk-8-jre-headless && \
    apt-get -y autoremove && \
    apt-get -y clean && \
    rm -rf /var/lib/apt/lists/*

ARG MAJOR_VERSION=master
ARG ZBX_VERSION=${MAJOR_VERSION}
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

ENV TERM=xterm ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES}

LABEL org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.version="${ZBX_VERSION}" \
      org.opencontainers.image.source="${ZBX_SOURCES}"

COPY ["conf/etc/", "/etc/"]

RUN set -eux && \
    apt-get -y update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y --no-install-recommends install \
            autoconf \
            automake \
            libc6-dev \
            make \
            openjdk-8-jdk \
            pkg-config \
            git \
            gcc && \
    cd /tmp/ && \
    git clone ${ZBX_SOURCES} --branch ${ZBX_VERSION} --depth 1 --single-branch zabbix-${ZBX_VERSION} && \
    cd /tmp/zabbix-${ZBX_VERSION} && \
    zabbix_revision=`git rev-parse --short HEAD` && \
    sed -i "s/{ZABBIX_REVISION}/$zabbix_revision/g" include/version.h && \
    sed -i "s/{ZABBIX_REVISION}/$zabbix_revision/g" src/zabbix_java/src/com/zabbix/gateway/GeneralInformation.java && \
    ./bootstrap.sh && \
    ./configure \
            --datadir=/usr/lib \
            --libdir=/usr/lib/zabbix \
            --sysconfdir=/etc/zabbix \
            --prefix=/usr \
            --enable-java \
            --silent && \
    make -j"$(nproc)" -s && \
    mkdir -p /usr/sbin/zabbix_java/ && \
    cp -r src/zabbix_java/bin /usr/sbin/zabbix_java/ && \
    cp -r src/zabbix_java/lib /usr/sbin/zabbix_java/ && \
    rm -rf /usr/sbin/zabbix_java/lib/*.xml && \
    cd /tmp/ && \
    rm -rf /tmp/zabbix-${ZBX_VERSION}/ && \
    DEBIAN_FRONTEND=noninteractive apt-get -y purge \
            autoconf \
            automake \
            libc6-dev \
            make \
            openjdk-8-jdk \
            pkg-config \
            git \
            gcc && \
    apt-get -y autoremove && \
    apt-get -y clean && \
    chown --quiet -R zabbix:root /etc/zabbix/ /usr/sbin/zabbix_java/ && \
    chgrp -R 0 /etc/zabbix/ /usr/sbin/zabbix_java/ && \
    chmod -R g=u /etc/zabbix/ /usr/sbin/zabbix_java/ && \
    rm -rf /var/lib/apt/lists/*

EXPOSE 10052/TCP
    
WORKDIR /var/lib/zabbix
    
COPY ["conf/usr/sbin/zabbix_java_gateway", "/usr/sbin/"]
COPY ["docker-entrypoint.sh", "/usr/bin/"]
    
ENTRYPOINT ["docker-entrypoint.sh"]
    
USER 1997
    
CMD ["/usr/sbin/zabbix_java_gateway"]
