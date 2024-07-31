# syntax=docker/dockerfile:1
ARG MAJOR_VERSION=7.0
ARG ZBX_VERSION=${MAJOR_VERSION}.2
ARG BUILD_BASE_IMAGE=zabbix-build-base:ubuntu-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} AS builder

ARG MAJOR_VERSION
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ARG GIT_BRANCH

ARG ZBX_PLUGINS_VERSION=${GIT_BRANCH:-$ZBX_VERSION}

LABEL org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix build base for PostgreSQL based images" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="AGPL v3.0" \
      org.opencontainers.image.source="${ZBX_SOURCES}" \
      org.opencontainers.image.title="Zabbix build base (PostgreSQL)" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix SIA" \
      org.opencontainers.image.version="${ZBX_VERSION}"

ENV ZBX_SOURCES_DIR=/tmp/zabbix-${ZBX_VERSION} ZBX_OUTPUT_DIR=/tmp/zabbix-${ZBX_VERSION}-output \
    DB_TYPE=postgresql \
    MONGODB_PLUGIN_SOURCES=https://git.zabbix.com/scm/ap/mongodb.git MONGODB_PLUGIN_VERSION=${ZBX_PLUGINS_VERSION} \
    POSTGRESQL_PLUGIN_SOURCES=https://git.zabbix.com/scm/ap/postgresql.git POSTGRESQL_PLUGIN_VERSION=${ZBX_PLUGINS_VERSION} \
    MSSQL_PLUGIN_SOURCES=https://git.zabbix.com/scm/ap/mssql.git MSSQL_PLUGIN_VERSION=${ZBX_PLUGINS_VERSION} \
    EMBER_PLUS_PLUGIN_SOURCES=https://git.zabbix.com/scm/ap/ember-plus.git EMBER_PLUS_PLUGIN_VERSION=${ZBX_PLUGINS_VERSION}

RUN --mount=type=cache,target=/root/.cache/go-build/ \
    --mount=type=cache,target=/root/go/ \
    --mount=type=bind,source=conf/chromedp_no_sandbox.patch,target=/tmp/chromedp_no_sandbox.patch \
    set -eux && \
    cd /tmp/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/agent/sbin/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/agent/conf/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/agent2/sbin/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/agent2/conf/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/agent2/conf/zabbix_agent2.d/plugins.d/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/proxy/sbin/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/proxy/conf/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/proxy/database/${DB_TYPE}/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/server/database/${DB_TYPE}/option-patches/with-compression/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/server/database/${DB_TYPE}/option-patches/without-compression/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/server/sbin/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/server/conf/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/server/database/${DB_TYPE}/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/java_gateway/sbin/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/web_service/sbin/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/web_service/conf/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/general/sbin/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/general/bin/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/general/conf/ && \
    git -c advice.detachedHead=false clone ${ZBX_SOURCES} --branch ${GIT_BRANCH:-$ZBX_VERSION} --depth 1 --single-branch ${ZBX_SOURCES_DIR} && \
    cd ${ZBX_SOURCES_DIR} && \
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
            --enable-agent \
            --enable-agent2 \
            --enable-ipv6 \
            --enable-proxy \
            --enable-server \
            --enable-webservice \
            --with-ldap \
            --with-libcurl \
            --with-libmodbus \
            --with-libpcre2 \
            --with-libxml2 \
            --with-net-snmp \
            --with-openipmi \
            --with-openssl \
            --with-${DB_TYPE} \
            --with-ssh \
            --with-unixodbc \
            --enable-java \
            --silent && \
    make -j"$(nproc)" -s dbschema && \
    make -j"$(nproc)" -s && \
    make -j"$(nproc)" -s gettext && \
    cat database/${DB_TYPE}/schema.sql > database/${DB_TYPE}/create.sql && \
    gzip -c database/${DB_TYPE}/create.sql > ${ZBX_OUTPUT_DIR}/proxy/database/${DB_TYPE}/create.sql.gz && \
    cat database/${DB_TYPE}/images.sql >> database/${DB_TYPE}/create.sql && \
    cat database/${DB_TYPE}/data.sql >> database/${DB_TYPE}/create.sql && \
    gzip -c database/${DB_TYPE}/create.sql > ${ZBX_OUTPUT_DIR}/server/database/${DB_TYPE}/create.sql.gz && \
    rm -rf database/${DB_TYPE}/create.sql && \
    cp -R ${ZBX_SOURCES_DIR}/database/postgresql/timescaledb/schema.sql ${ZBX_OUTPUT_DIR}/server/database/${DB_TYPE}/timescaledb.sql && \
    cp -R ${ZBX_SOURCES_DIR}/database/postgresql/timescaledb/option-patches/with-compression/*.sql ${ZBX_OUTPUT_DIR}/server/database/${DB_TYPE}/option-patches/with-compression/ && \
    cp -R ${ZBX_SOURCES_DIR}/database/postgresql/timescaledb/option-patches/without-compression/*.sql ${ZBX_OUTPUT_DIR}/server/database/${DB_TYPE}/option-patches/without-compression/ && \
    mkdir /tmp/fonts/ && \
    curl --tlsv1.2 -sSf -L "https://noto-website.storage.googleapis.com/pkgs/NotoSansCJKjp-hinted.zip" -o /tmp/fonts/NotoSansCJKjp-hinted.zip && \
    unzip /tmp/fonts/NotoSansCJKjp-hinted.zip -d /tmp/fonts/ && \
    cp /tmp/fonts/NotoSansCJKjp-Regular.otf ${ZBX_SOURCES_DIR}/ui/assets/fonts/NotoSansCJKjp-Regular.ttf && \
    cp /tmp/fonts/LICENSE_OFL.txt ${ZBX_SOURCES_DIR}/ui/assets/fonts/ && \
    rm -f ${ZBX_SOURCES_DIR}/ui/assets/fonts/DejaVuSans.ttf && \
    sed -i -r "s/(define\(.*_FONT_NAME.*)DejaVuSans/\1NotoSansCJKjp-Regular/" ${ZBX_SOURCES_DIR}/ui/include/defines.inc.php && \
    rm -rf /tmp/fonts/ && \
    chmod o+r ${ZBX_SOURCES_DIR}/ui/assets/fonts/* && \
    strip ${ZBX_SOURCES_DIR}/src/zabbix_agent/zabbix_agentd && \
    strip ${ZBX_SOURCES_DIR}/src/zabbix_server/zabbix_server && \
    strip ${ZBX_SOURCES_DIR}/src/zabbix_proxy/zabbix_proxy && \
    strip ${ZBX_SOURCES_DIR}/src/go/bin/zabbix_agent2 && \
    strip ${ZBX_SOURCES_DIR}/src/zabbix_get/zabbix_get && \
    strip ${ZBX_SOURCES_DIR}/src/zabbix_sender/zabbix_sender && \
    strip ${ZBX_SOURCES_DIR}/src/go/bin/zabbix_web_service && \
    cp ${ZBX_SOURCES_DIR}/src/zabbix_agent/zabbix_agentd ${ZBX_OUTPUT_DIR}/agent/sbin/ && \
    cp ${ZBX_SOURCES_DIR}/conf/zabbix_agentd.conf ${ZBX_OUTPUT_DIR}/agent/conf/ && \
    cp -R ${ZBX_SOURCES_DIR}/conf/zabbix_agentd/ ${ZBX_OUTPUT_DIR}/agent/conf/ && \
    cp ${ZBX_SOURCES_DIR}/src/go/bin/zabbix_agent2 ${ZBX_OUTPUT_DIR}/agent2/sbin/ && \
    cp ${ZBX_SOURCES_DIR}/src/go/conf/zabbix_agent2.conf ${ZBX_OUTPUT_DIR}/agent2/conf/ && \
    cp -R ${ZBX_SOURCES_DIR}/src/go/conf/zabbix_agent2.d/ ${ZBX_OUTPUT_DIR}/agent2/conf/ && \
    cp ${ZBX_SOURCES_DIR}/src/zabbix_server/zabbix_server ${ZBX_OUTPUT_DIR}/server/sbin/ && \
    cp ${ZBX_SOURCES_DIR}/conf/zabbix_server.conf ${ZBX_OUTPUT_DIR}/server/conf/ && \
    cp ${ZBX_SOURCES_DIR}/src/zabbix_proxy/zabbix_proxy ${ZBX_OUTPUT_DIR}/proxy/sbin/ && \
    cp ${ZBX_SOURCES_DIR}/conf/zabbix_proxy.conf ${ZBX_OUTPUT_DIR}/proxy/conf/ && \
    cp -R ${ZBX_SOURCES_DIR}/src/zabbix_java/bin/ ${ZBX_OUTPUT_DIR}/java_gateway/sbin/ && \
    cp -R ${ZBX_SOURCES_DIR}/src/zabbix_java/lib/ ${ZBX_OUTPUT_DIR}/java_gateway/sbin/ && \
    cp ${ZBX_SOURCES_DIR}/src/go/bin/zabbix_web_service ${ZBX_OUTPUT_DIR}/web_service/sbin/ && \
    cp ${ZBX_SOURCES_DIR}/src/go/conf/zabbix_web_service.conf ${ZBX_OUTPUT_DIR}/web_service/conf/ && \
    cp ${ZBX_SOURCES_DIR}/src/zabbix_get/zabbix_get ${ZBX_OUTPUT_DIR}/general/bin/ && \
    cp ${ZBX_SOURCES_DIR}/src/zabbix_sender/zabbix_sender ${ZBX_OUTPUT_DIR}/general/bin/ && \
    make -s distclean && \
    cd /tmp/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/agent2/sbin/zabbix-agent2-plugin/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/agent2/conf/zabbix_agent2.d/plugins.d/ && \
    git -c advice.detachedHead=false clone ${MONGODB_PLUGIN_SOURCES} --branch ${MONGODB_PLUGIN_VERSION} --depth 1 --single-branch /tmp/mongodb-plugin-${MONGODB_PLUGIN_VERSION} && \
    cd /tmp/mongodb-plugin-${MONGODB_PLUGIN_VERSION} && \
    make && \
    strip /tmp/mongodb-plugin-${MONGODB_PLUGIN_VERSION}/zabbix-agent2-plugin-mongodb && \
    cp /tmp/mongodb-plugin-${MONGODB_PLUGIN_VERSION}/zabbix-agent2-plugin-mongodb ${ZBX_OUTPUT_DIR}/agent2/sbin/zabbix-agent2-plugin/mongodb && \
    cp /tmp/mongodb-plugin-${MONGODB_PLUGIN_VERSION}/mongodb.conf ${ZBX_OUTPUT_DIR}/agent2/conf/zabbix_agent2.d/plugins.d/ && \
    cd /tmp/ && \
    git -c advice.detachedHead=false clone ${POSTGRESQL_PLUGIN_SOURCES} --branch ${POSTGRESQL_PLUGIN_VERSION} --depth 1 --single-branch /tmp/postgresql-plugin-${POSTGRESQL_PLUGIN_VERSION} && \
    cd /tmp/postgresql-plugin-${POSTGRESQL_PLUGIN_VERSION} && \
    make && \
    strip /tmp/postgresql-plugin-${POSTGRESQL_PLUGIN_VERSION}/zabbix-agent2-plugin-postgresql && \
    cp /tmp/postgresql-plugin-${POSTGRESQL_PLUGIN_VERSION}/zabbix-agent2-plugin-postgresql ${ZBX_OUTPUT_DIR}/agent2/sbin/zabbix-agent2-plugin/postgresql && \
    cp /tmp/postgresql-plugin-${POSTGRESQL_PLUGIN_VERSION}/postgresql.conf ${ZBX_OUTPUT_DIR}/agent2/conf/zabbix_agent2.d/plugins.d/ && \
    cd /tmp/ && \
    git -c advice.detachedHead=false clone ${MSSQL_PLUGIN_SOURCES} --branch ${MSSQL_PLUGIN_VERSION} --depth 1 --single-branch /tmp/mssql-plugin-${MSSQL_PLUGIN_VERSION} && \
    cd /tmp/mssql-plugin-${MSSQL_PLUGIN_VERSION} && \
    make && \
    strip /tmp/mssql-plugin-${MSSQL_PLUGIN_VERSION}/zabbix-agent2-plugin-mssql && \
    cp /tmp/mssql-plugin-${MSSQL_PLUGIN_VERSION}/zabbix-agent2-plugin-mssql ${ZBX_OUTPUT_DIR}/agent2/sbin/zabbix-agent2-plugin/mssql && \
    cp /tmp/mssql-plugin-${MSSQL_PLUGIN_VERSION}/mssql.conf ${ZBX_OUTPUT_DIR}/agent2/conf/zabbix_agent2.d/plugins.d/ && \
    cd /tmp/ && \
    git -c advice.detachedHead=false clone ${EMBER_PLUS_PLUGIN_SOURCES} --branch ${EMBER_PLUS_PLUGIN_VERSION} --depth 1 --single-branch /tmp/ember-plugin-${EMBER_PLUS_PLUGIN_VERSION} && \
    cd /tmp/ember-plugin-${EMBER_PLUS_PLUGIN_VERSION} && \
    make && \
    strip /tmp/ember-plugin-${EMBER_PLUS_PLUGIN_VERSION}/zabbix-agent2-plugin-ember-plus && \
    cp /tmp/ember-plugin-${EMBER_PLUS_PLUGIN_VERSION}/zabbix-agent2-plugin-ember-plus ${ZBX_OUTPUT_DIR}/agent2/sbin/zabbix-agent2-plugin/ember-plus && \
    cp /tmp/ember-plugin-${EMBER_PLUS_PLUGIN_VERSION}/ember.conf ${ZBX_OUTPUT_DIR}/agent2/conf/zabbix_agent2.d/plugins.d/
