FROM centos:centos7
LABEL maintainer="Alexey Pustovalov <alexey.pustovalov@zabbix.com>"

ARG BUILD_DATE
ARG VCS_REF

ARG YUM_FLAGS_COMMON="--quiet -y"
ARG YUM_FLAGS_PERSISTANT="${YUM_FLAGS_COMMON}"
ARG YUM_FLAGS_DEV="${YUM_FLAGS_COMMON}"
ENV TERM=xterm

LABEL org.label-schema.name="zabbix-agent-centos" \
      org.label-schema.vendor="Zabbix LLC" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.description="Zabbix agent is deployed on a monitoring target to actively monitor local resources and applications" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.license="GPL 2.0"

STOPSIGNAL SIGTERM

RUN groupadd --system zabbix && \
    adduser -r --shell /sbin/nologin \
            -g zabbix \
            -d /var/lib/zabbix/ \
        zabbix && \
    mkdir -p /etc/zabbix && \
    mkdir -p /etc/zabbix/zabbix_agentd.d && \
    mkdir -p /var/lib/zabbix && \
    mkdir -p /var/lib/zabbix/enc && \
    mkdir -p /var/lib/zabbix/modules && \
    chown --quiet -R zabbix:root /var/lib/zabbix && \
    yum ${YUM_FLAGS_COMMON} makecache && \
    yum ${YUM_FLAGS_DEV} install epel-release && \
    yum ${YUM_FLAGS_PERSISTANT} install \
            libldap \
            openssl-libs \
            supervisor && \
    yum ${YUM_FLAGS_PERSISTANT} clean all && \
    rm -rf /var/cache/yum/

ARG MAJOR_VERSION=4.0
ARG ZBX_VERSION=${MAJOR_VERSION}
ARG ZBX_SOURCES=svn://svn.zabbix.com/trunk/
ENV ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES}

LABEL org.label-schema.usage="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.label-schema.version="${ZBX_VERSION}" \
      org.label-schema.vcs-url="${ZBX_SOURCES}" \
      org.label-schema.docker.cmd="docker run --name zabbix-agent --link zabbix-server:zabbix-server -p 10050:10050 -d zabbix-agent:centos-${ZBX_VERSION}"

RUN yum --quiet makecache && \
    yum ${YUM_FLAGS_DEV} install \
            autoconf \
            automake \
            make \
            openssl-devel \
            openldap-devel \
            subversion \
            gcc && \
    cd /tmp/ && \
    svn --quiet export ${ZBX_SOURCES} zabbix-${ZBX_VERSION} 1>/dev/null && \
    cd /tmp/zabbix-${ZBX_VERSION} && \
    zabbix_revision=`set -o pipefail && svn info ${ZBX_SOURCES} |grep "Last Changed Rev"|awk '{print $4;}'` && \
    sed -i "s/{ZABBIX_REVISION}/$zabbix_revision/g" include/version.h && \
    ./bootstrap.sh 1>/dev/null && \
    export CFLAGS="-fPIC -pie -Wl,-z,relro -Wl,-z,now" && \
    ./configure \
            --datadir=/usr/lib \
            --libdir=/usr/lib/zabbix \
            --prefix=/usr \
            --sysconfdir=/etc/zabbix \
            --prefix=/usr \
            --enable-agent \
            --with-ldap \
            --with-openssl \
            --enable-ipv6 \
            --silent && \
    make -j"$(nproc)" -s 1>/dev/null && \
    cp src/zabbix_agent/zabbix_agentd /usr/sbin/zabbix_agentd && \
    cp src/zabbix_get/zabbix_get /usr/bin/zabbix_get && \
    cp src/zabbix_sender/zabbix_sender /usr/bin/zabbix_sender && \
    cp conf/zabbix_agentd.conf /etc/zabbix/zabbix_agentd.conf && \
    chown --quiet -R zabbix:root /etc/zabbix && \
    cd /tmp/ && \
    rm -rf /tmp/zabbix-${ZBX_VERSION}/ && \
    yum ${YUM_FLAGS_COMMON} history undo `yum history | sed -n 4p |column -t | cut -d' ' -f1` 1>/dev/null && \
    yum ${YUM_FLAGS_COMMON} clean all && \
    rm -rf /etc/udev/hwdb.bin && \
    rm -rf /var/cache/yum && \
    rm -rf /var/lib/yum/yumdb/* && \
    rm -rf /usr/lib/udev/hwdb.d/* && \
    rm -rf /root/.pki && \
    rm -rf /root/.subversion 

EXPOSE 10050/TCP

WORKDIR /var/lib/zabbix

VOLUME ["/etc/zabbix/zabbix_agentd.d", "/var/lib/zabbix/enc", "/var/lib/zabbix/modules"]

COPY ["conf/etc/supervisor/", "/etc/supervisor/"]
COPY ["docker-entrypoint.sh", "/usr/bin/"]

ENTRYPOINT ["docker-entrypoint.sh"]

CMD ["agentd", "none"]
