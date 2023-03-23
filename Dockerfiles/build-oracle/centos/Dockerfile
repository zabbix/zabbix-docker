# syntax=docker/dockerfile:1
ARG MAJOR_VERSION=6.4
ARG ZBX_VERSION=${MAJOR_VERSION}.0
ARG BUILD_BASE_IMAGE=zabbix-build-base:centos-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} as builder

ARG MAJOR_VERSION
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

LABEL org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix build base for MySQL based images" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="GPL v2.0" \
      org.opencontainers.image.source="${ZBX_SOURCES}" \
      org.opencontainers.image.title="Zabbix build base (MySQL)" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix LLC" \
      org.opencontainers.image.version="${ZBX_VERSION}"

ENV MONGODB_PLUGIN_SOURCES=https://git.zabbix.com/scm/ap/mongodb.git MONGODB_PLUGIN_VERSION=${ZBX_VERSION} \
    POSTGRESQL_PLUGIN_SOURCES=https://git.zabbix.com/scm/ap/postgresql.git POSTGRESQL_PLUGIN_VERSION=${ZBX_VERSION} \
    LD_LIBRARY_PATH=/usr/lib/oracle/19.18/client64/lib

COPY ["conf/chromedp_no_sandbox.patch", "/tmp/chromedp_no_sandbox.patch"]
RUN wget -O /basic.rpm https://download.oracle.com/otn_software/linux/instantclient/1918000/oracle-instantclient19.18-basic-19.18.0.0.0-2.x86_64.rpm &&\
    wget -O /sqlplus.rpm https://download.oracle.com/otn_software/linux/instantclient/1918000/oracle-instantclient19.18-sqlplus-19.18.0.0.0-2.x86_64.rpm &&\
    wget -O /devel.rpm https://download.oracle.com/otn_software/linux/instantclient/1918000/oracle-instantclient19.18-devel-19.18.0.0.0-2.x86_64.rpm &&\
    wget -O /odbc.rpm https://download.oracle.com/otn_software/linux/instantclient/1918000/oracle-instantclient19.18-odbc-19.18.0.0.0-2.x86_64.rpm  &&\
    wget -O /jdbc.rpm https://download.oracle.com/otn_software/linux/instantclient/1918000/oracle-instantclient19.18-jdbc-19.18.0.0.0-2.x86_64.rpm &&\
    wget -O /basiclite.rpm https://download.oracle.com/otn_software/linux/instantclient/1918000/oracle-instantclient19.18-basiclite-19.18.0.0.0-2.x86_64.rpm &&\
    rpm2cpio basic.rpm | cpio -idmv && \
    rpm2cpio sqlplus.rpm | cpio -idmv && \
    rpm2cpio devel.rpm | cpio -idmv && \
    rpm2cpio odbc.rpm | cpio -idmv && \
    rpm2cpio jdbc.rpm | cpio -idmv && \
    rpm2cpio basiclite.rpm | cpio -idmv && \
    set -eux && \
    ARCH_SUFFIX="$(arch)"; \
    case "$ARCH_SUFFIX" in \
        x86_64) \
            additional_components='--enable-java'; \
            ;; \
        aarch64) \
            additional_components='--enable-java'; \
            ;; \
        ppc64le) \
            additional_components=''; \
            ;; \
        *) echo "Unknown ARCH_SUFFIX=${ARCH_SUFFIX-}"; exit 1 ;; \
    esac; \
    cd /tmp/ && \
    git -c advice.detachedHead=false clone ${ZBX_SOURCES} --branch ${ZBX_VERSION} --depth 1 --single-branch /tmp/zabbix-${ZBX_VERSION} && \
    cd /tmp/zabbix-${ZBX_VERSION} && \
    patch -p1 < /tmp/chromedp_no_sandbox.patch && \
    zabbix_revision=`git rev-parse --short HEAD` && \
    sed -i "s/{ZABBIX_REVISION}/$zabbix_revision/g" include/version.h && \
    sed -i "s/{ZABBIX_REVISION}/$zabbix_revision/g" src/go/pkg/version/version.go && \
    sed -i "s/{ZABBIX_REVISION}/$zabbix_revision/g" src/zabbix_java/src/com/zabbix/gateway/GeneralInformation.java && \
    ./bootstrap.sh && \
    export CFLAGS="-fPIC -pie -Wl,-z,relro,-z,now,-z,defs" && \
    export CFLAGS="$CFLAGS -D_FORTIFY_SOURCE=2 -fexceptions -O2 -pipe" && \
    ./configure \
            --datadir=/usr/lib \
            --libdir=/usr/lib/zabbix \
            --prefix=/usr \
            --sysconfdir=/etc/zabbix \
            --enable-ipv6 \
            --enable-agent \
            --enable-agent2 \
            --enable-proxy \
            --enable-server \
            --enable-webservice \
            --with-ldap \
            --with-libcurl \
            --with-libpcre2 \
            --with-libxml2 \
            --with-oracle \
            --with-oracle-include=/usr/include/oracle/19.18/client64 \
            --with-oracle-lib=/usr/lib/oracle/19.18/client64/lib \
            --with-net-snmp \
            --with-openipmi \
            --with-openssl \
            --with-ssh \
            --with-unixodbc \
            $additional_components \
            --silent && \
    make -j"$(nproc)" -s dbschema && \
    make -j"$(nproc)" -s && \
    make -j"$(nproc)" -s gettext && \
    cat database/oracle/schema.sql > database/oracle/create.sql && \
    cat database/oracle/images.sql >> database/oracle/create.sql && \
    cat database/oracle/data.sql >> database/oracle/create.sql && \
    gzip -c database/oracle/create.sql > database/oracle/create_server.sql.gz && \
    rm -rf database/oracle/create.sql && \
    cat database/oracle/schema.sql > database/oracle/create.sql && \
    gzip -c database/oracle/create.sql > database/oracle/create_proxy.sql.gz && \
    rm -rf database/oracle/create.sql && \
    strip /tmp/zabbix-${ZBX_VERSION}/src/zabbix_agent/zabbix_agentd && \
    strip /tmp/zabbix-${ZBX_VERSION}/src/zabbix_server/zabbix_server && \
    strip /tmp/zabbix-${ZBX_VERSION}/src/zabbix_proxy/zabbix_proxy && \
    strip /tmp/zabbix-${ZBX_VERSION}/src/go/bin/zabbix_agent2 && \
    strip /tmp/zabbix-${ZBX_VERSION}/src/zabbix_get/zabbix_get && \
    strip /tmp/zabbix-${ZBX_VERSION}/src/zabbix_sender/zabbix_sender && \
    strip /tmp/zabbix-${ZBX_VERSION}/src/go/bin/zabbix_web_service && \
    cd /tmp/ && \
    git -c advice.detachedHead=false clone ${MONGODB_PLUGIN_SOURCES} --branch ${MONGODB_PLUGIN_VERSION} --depth 1 --single-branch /tmp/mongodb_plugin && \
    cd /tmp/mongodb_plugin && \
    make && \
    strip /tmp/mongodb_plugin/zabbix-agent2-plugin-mongodb && \
    cd /tmp/ && \
    git -c advice.detachedHead=false clone ${POSTGRESQL_PLUGIN_SOURCES} --branch ${POSTGRESQL_PLUGIN_VERSION} --depth 1 --single-branch /tmp/postgresql_plugin && \
    cd /tmp/postgresql_plugin && \
    make && \
    strip /tmp/postgresql_plugin/zabbix-agent2-plugin-postgresql && \
    rm -rf $(go env GOCACHE) /*.rpm && \
    mkdir -p /usr/include/oracle/19.18/client64/rdbms/public/ && \
    cp -R /usr/include/oracle/19.18/client64/*.h /usr/include/oracle/19.18/client64/rdbms/public/ && \
    rm -rf /usr/include/oracle/19.18/client64/*.h

