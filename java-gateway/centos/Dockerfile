FROM centos:centos7 as builder

ARG YUM_FLAGS_COMMON="-y"
ARG YUM_FLAGS_DEV="${YUM_FLAGS_COMMON}"
    
ARG MAJOR_VERSION=4.2
ARG ZBX_VERSION=${MAJOR_VERSION}.6
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ENV TERM=xterm ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} \
    ZBX_TYPE=java-gateway

RUN set -eux && \
    yum ${YUM_FLAGS_COMMON} makecache && \
    yum ${YUM_FLAGS_DEV} install \
            autoconf \
            automake \
            java-1.8.0-openjdk-devel \
            make \
            git \
            gcc  && \
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
    rm -rf src/zabbix_java/lib/*.xml

FROM centos:centos7
LABEL maintainer="Alexey Pustovalov <alexey.pustovalov@zabbix.com>"

ARG BUILD_DATE
ARG VCS_REF

ARG YUM_FLAGS_COMMON="-y"
ARG YUM_FLAGS_PERSISTENT="${YUM_FLAGS_COMMON}"

ARG MAJOR_VERSION=4.2
ARG ZBX_VERSION=${MAJOR_VERSION}.6
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ENV TERM=xterm ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} \
    ZBX_TYPE=java-gateway ZBX_DB_TYPE=none ZBX_OPT_TYPE=none

LABEL org.label-schema.name="zabbix-${ZBX_TYPE}-centos" \
      org.label-schema.vendor="Zabbix LLC" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.description="Zabbix Java Gateway performs native support for monitoring JMX applications" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.license="GPL v2.0" \
      org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.version="${ZBX_VERSION}" \
      org.label-schema.vcs-url="${ZBX_SOURCES}" \
      org.label-schema.docker.cmd="docker run --name zabbix-${ZBX_TYPE} --link zabbix-server:zabbix-server -p 10052:10052 -d zabbix-${ZBX_TYPE}:centos-${ZBX_VERSION}"

STOPSIGNAL SIGTERM

COPY --from=builder /tmp/zabbix-${ZBX_VERSION}/src/zabbix_java/bin/* /usr/sbin/zabbix_java/bin/
COPY --from=builder /tmp/zabbix-${ZBX_VERSION}/src/zabbix_java/lib/* /usr/sbin/zabbix_java/lib/

RUN set -eux && \
    groupadd --system zabbix && \
    adduser --system --shell /sbin/nologin \
            -g zabbix \
            -d /var/lib/zabbix/ \
        zabbix && \
    mkdir -p /etc/zabbix/ && \
    chown --quiet -R zabbix:root /etc/zabbix && \
    mkdir -p /usr/sbin/zabbix_java/ && \
    yum ${YUM_FLAGS_COMMON} makecache && \
    yum ${YUM_FLAGS_PERSISTENT} install \
            java-1.8.0-openjdk-headless && \
    yum ${YUM_FLAGS_PERSISTENT} clean all && \
    rm -rf /var/cache/yum/

EXPOSE 10052/TCP

WORKDIR /var/lib/zabbix

VOLUME ["/usr/sbin/zabbix_java/ext_lib"]

COPY ["conf/etc/zabbix/zabbix_java_gateway_logback.xml", "/etc/zabbix/"]
COPY ["conf/usr/sbin/zabbix_java_gateway", "/usr/sbin/"]
COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENTRYPOINT ["docker-entrypoint.sh"]
