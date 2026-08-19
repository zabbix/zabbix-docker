# syntax=docker/dockerfile:1.6
ARG MAJOR_VERSION=6.0
ARG ZBX_VERSION=${MAJOR_VERSION}.48
ARG BUILD_BASE_IMAGE=zabbix-build-base:ol-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} AS builder
ARG BUILD_JOBS

ARG MAJOR_VERSION
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ARG GIT_BRANCH

ARG ZBX_PLUGINS_VERSION=${GIT_BRANCH:-$ZBX_VERSION}
ARG CFLAGS
ARG LDFLAGS

LABEL org.opencontainers.image.authors="Alexey Pustovalov <alexey.pustovalov@zabbix.com>" \
      org.opencontainers.image.description="Zabbix build base for SQLite3 based images" \
      org.opencontainers.image.documentation="https://www.zabbix.com/documentation/${MAJOR_VERSION}/manual/installation/containers" \
      org.opencontainers.image.licenses="GPL v2.0" \
      org.opencontainers.image.source="${ZBX_SOURCES}" \
      org.opencontainers.image.title="Zabbix build base (SQLite3)" \
      org.opencontainers.image.url="https://zabbix.com/" \
      org.opencontainers.image.vendor="Zabbix SIA" \
      org.opencontainers.image.version="${ZBX_VERSION}"

ENV ZBX_SOURCES_DIR=/tmp/zabbix-${ZBX_VERSION} ZBX_OUTPUT_DIR=/tmp/zabbix-${ZBX_VERSION}-output \
    DB_TYPE=sqlite3 \
    CFLAGS=${CFLAGS:-"-O2 -fPIE -fstack-protector-strong -D_FORTIFY_SOURCE=2 -Wformat -Wformat-security -pipe"} \
    LDFLAGS=${LDFLAGS:-"-pie -Wl,-z,relro,-z,now,-z,defs -Wl,--as-needed"} \
    MONGODB_PLUGIN_SOURCES=https://git.zabbix.com/scm/ap/mongodb.git MONGODB_PLUGIN_VERSION=${ZBX_PLUGINS_VERSION} \
    POSTGRESQL_PLUGIN_SOURCES=https://git.zabbix.com/scm/ap/postgresql.git POSTGRESQL_PLUGIN_VERSION=${ZBX_PLUGINS_VERSION} \
    MSSQL_PLUGIN_SOURCES=https://git.zabbix.com/scm/ap/mssql.git MSSQL_PLUGIN_VERSION=${ZBX_PLUGINS_VERSION} \
    EMBER_PLUS_PLUGIN_SOURCES=https://git.zabbix.com/scm/ap/ember-plus.git EMBER_PLUS_PLUGIN_VERSION=${ZBX_PLUGINS_VERSION}

RUN --mount=from=sources,target=/tmp/src,ro \
    --mount=type=bind,source=patches/,target=/tmp/patches,ro \
    set -eux && \
    mkdir -p \
        ${ZBX_OUTPUT_DIR}/agent/sbin \
        ${ZBX_OUTPUT_DIR}/agent/conf \
        ${ZBX_OUTPUT_DIR}/agent2/sbin \
        ${ZBX_OUTPUT_DIR}/agent2/conf \
        ${ZBX_OUTPUT_DIR}/agent2/conf/zabbix_agent2.d/plugins.d \
        ${ZBX_OUTPUT_DIR}/proxy/sbin \
        ${ZBX_OUTPUT_DIR}/proxy/conf \
        ${ZBX_OUTPUT_DIR}/proxy/database/${DB_TYPE} \
        ${ZBX_OUTPUT_DIR}/java_gateway/sbin/bin \
        ${ZBX_OUTPUT_DIR}/java_gateway/sbin/lib \
        ${ZBX_OUTPUT_DIR}/java_gateway/conf \
        ${ZBX_OUTPUT_DIR}/web/conf \
        ${ZBX_OUTPUT_DIR}/general/bin && \
    if [ -f /tmp/src/bootstrap.sh ]; then \
        mkdir -p "${ZBX_SOURCES_DIR}" && \
        cp -a /tmp/src/. "${ZBX_SOURCES_DIR}/"; \
    else \
        git -c advice.detachedHead=false clone \
            "${ZBX_SOURCES}" \
            --branch "${GIT_BRANCH:-$ZBX_VERSION}" \
            --depth 1 \
            --single-branch \
            "${ZBX_SOURCES_DIR}"; \
    fi && \
    cd "${ZBX_SOURCES_DIR}" && \
    for patch_filename in /tmp/patches/*.patch; do \
        [ -f "$patch_filename" ] || continue; \
        patch -p1 < "$patch_filename"; \
    done && \
    if git rev-parse --short HEAD >/dev/null 2>&1; then \
        zabbix_revision="$(git rev-parse --short HEAD)"; \
    else \
        zabbix_revision="unknown"; \
    fi && \
    sed -i "s/{ZABBIX_REVISION}/$zabbix_revision/g" include/version.h && \
    sed -i "s/{ZABBIX_REVISION}/$zabbix_revision/g" src/go/pkg/version/version.go && \
    sed -i "s/{ZABBIX_REVISION}/$zabbix_revision/g" src/zabbix_java/src/com/zabbix/gateway/GeneralInformation.java

WORKDIR ${ZBX_SOURCES_DIR}

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    --mount=from=config_templates,target=/tmp/conf,ro \
    set -eux && \
    cd "${ZBX_SOURCES_DIR}" && \
    ./bootstrap.sh && \
    ./configure \
        --datadir=/usr/lib \
        --libdir=/usr/lib/zabbix \
        --prefix=/usr \
        --sysconfdir=/etc/zabbix \
        --enable-ipv6 \
        --enable-agent \
        --enable-agent2 \
        --enable-java \
        --enable-proxy \
        --with-ldap \
        --with-libcurl \
        --with-libmodbus \
        --with-libpcre2 \
        --with-libxml2 \
        --with-${DB_TYPE} \
        --with-net-snmp \
        --with-openipmi \
        --with-openssl \
        --with-ssh \
        --with-unixodbc \
        --silent && \
    jobs="${BUILD_JOBS:-$(nproc)}" && \
    make -j"${jobs}" -s dbschema && \
    make -j"${jobs}" -s && \
    make -j"${jobs}" -s gettext && \
    cat "database/${DB_TYPE}/schema.sql" > "database/${DB_TYPE}/create.sql" && \
    gzip -c < "database/${DB_TYPE}/create.sql" > "${ZBX_OUTPUT_DIR}/proxy/database/${DB_TYPE}/create.sql.gz" && \
    strip src/zabbix_agent/zabbix_agentd && \
    strip src/zabbix_proxy/zabbix_proxy && \
    strip src/zabbix_get/zabbix_get && \
    strip src/zabbix_sender/zabbix_sender && \
    strip src/go/bin/zabbix_agent2 || true && \
    mv src/zabbix_agent/zabbix_agentd "${ZBX_OUTPUT_DIR}/agent/sbin/" && \
    cp conf/zabbix_agentd.conf "${ZBX_OUTPUT_DIR}/agent/conf/" && \
    cp -a conf/zabbix_agentd "${ZBX_OUTPUT_DIR}/agent/conf/" && \
    mv src/go/bin/zabbix_agent2 "${ZBX_OUTPUT_DIR}/agent2/sbin/" && \
    cp src/go/conf/zabbix_agent2.conf "${ZBX_OUTPUT_DIR}/agent2/conf/" && \
    cp -a src/go/conf/zabbix_agent2.d "${ZBX_OUTPUT_DIR}/agent2/conf/" && \
    mv src/zabbix_proxy/zabbix_proxy "${ZBX_OUTPUT_DIR}/proxy/sbin/" && \
    cp conf/zabbix_proxy.conf ${ZBX_OUTPUT_DIR}/proxy/conf/ && \
    cp -a src/zabbix_java/bin/. "${ZBX_OUTPUT_DIR}/java_gateway/sbin/bin/" && \
    cp -a src/zabbix_java/lib/*.jar "${ZBX_OUTPUT_DIR}/java_gateway/sbin/lib/" && \
    cp src/zabbix_java/lib/logback.xml "${ZBX_OUTPUT_DIR}/java_gateway/conf/logback.xml.dist" && \
    cp src/zabbix_java/lib/logback-console.xml "${ZBX_OUTPUT_DIR}/java_gateway/conf/logback-console.xml.dist" && \
    cp -a /tmp/conf/java-gateway/. "${ZBX_OUTPUT_DIR}/java_gateway/conf/" && \
    mv src/zabbix_get/zabbix_get "${ZBX_OUTPUT_DIR}/general/bin/" && \
    mv src/zabbix_sender/zabbix_sender "${ZBX_OUTPUT_DIR}/general/bin/" && \
    cp -a /tmp/conf/web/. "${ZBX_OUTPUT_DIR}/web/conf/" && \
    make -s distclean

ADD --chmod=0644 --checksum=sha256:68a3fc98800b2a27b371f2fb79991daf3633bd89309d4ffaa6946fd587f375b5 \
    https://raw.githubusercontent.com/notofonts/noto-cjk/main/Sans/OTF/Japanese/NotoSansCJKjp-Regular.otf \
    ui/assets/fonts/NotoSansCJKjp-Regular.otf
ADD --chmod=0644 --checksum=sha256:6a73f9541c2de74158c0e7cf6b0a58ef774f5a780bf191f2d7ec9cc53efe2bf2 \
    https://raw.githubusercontent.com/notofonts/noto-cjk/main/Sans/LICENSE \
    ui/assets/fonts/OFL.txt

RUN set -eux && \
    sed -i -r 's/(define\(.*_FONT_NAME.*)DejaVuSans/\1NotoSansCJKjp-Regular/' ui/include/defines.inc.php && \
    sed -i -r "s/(define\(.*ZBX_SESSION_NAME.*)'zbx_sessionid'/\1getenv('ZBX_SESSION_NAME') ?: 'zbx_sessionid'/" ui/include/defines.inc.php && \
    sed -i -r 's/\.ttf/\.otf/' ui/include/graphs.inc.php && \
    rm -f ui/assets/fonts/DejaVuSans.ttf && \
    chmod o+r ui/assets/fonts/*

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    set -eux && \
    mkdir -p "${ZBX_OUTPUT_DIR}/agent2/sbin/zabbix-agent2-plugin" && \
    mkdir -p "${ZBX_OUTPUT_DIR}/agent2/conf/zabbix_agent2.d/plugins.d" && \
    build_plugin() { \
        name="$1"; \
        repo="$2"; \
        version="$3"; \
        binary="$4"; \
        conf="$5"; \
        outname="$6"; \
        workdir="/tmp/${name}-plugin-${version}"; \
        rm -rf "$workdir"; \
        git -c advice.detachedHead=false clone "$repo" \
            --branch "$version" \
            --depth 1 \
            --single-branch \
            "$workdir"; \
        cd "$workdir"; \
        make; \
        strip "$binary" || true; \
        mv "$binary" "${ZBX_OUTPUT_DIR}/agent2/sbin/zabbix-agent2-plugin/${outname}"; \
        cp "$conf" "${ZBX_OUTPUT_DIR}/agent2/conf/zabbix_agent2.d/plugins.d/"; \
        make clean; \
    }; \
    build_plugin mongodb "${MONGODB_PLUGIN_SOURCES}" "${MONGODB_PLUGIN_VERSION}" \
        zabbix-agent2-plugin-mongodb mongodb.conf mongodb && \
    build_plugin postgresql "${POSTGRESQL_PLUGIN_SOURCES}" "${POSTGRESQL_PLUGIN_VERSION}" \
        zabbix-agent2-plugin-postgresql postgresql.conf postgresql && \
    build_plugin mssql "${MSSQL_PLUGIN_SOURCES}" "${MSSQL_PLUGIN_VERSION}" \
        zabbix-agent2-plugin-mssql mssql.conf mssql && \
    build_plugin ember "${EMBER_PLUS_PLUGIN_SOURCES}" "${EMBER_PLUS_PLUGIN_VERSION}" \
        zabbix-agent2-plugin-ember-plus ember.conf ember-plus
