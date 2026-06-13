# syntax=docker/dockerfile:1.6
ARG MAJOR_VERSION=7.4
ARG RELEASE=11
ARG ZBX_VERSION=${MAJOR_VERSION}.11
ARG BUILD_BASE_IMAGE=zabbix-build-base:rhel-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} AS builder
ARG BUILD_JOBS

ARG MAJOR_VERSION
ARG RELEASE
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ARG GIT_BRANCH

ARG ZBX_PLUGINS_VERSION=${GIT_BRANCH:-$ZBX_VERSION}
ARG CFLAGS
ARG LDFLAGS

LABEL description="Zabbix build base for SQLite3 based images" \
      maintainer="alexey.pustovalov@zabbix.com" \
      name="zabbix/zabbix-build-sqlite3-74" \
      release="${RELEASE}" \
      summary="Zabbix build base (SQLite3)" \
      url="https://www.zabbix.com/" \
      vendor="Zabbix SIA" \
      version="${MAJOR_VERSION}" \
      io.k8s.description="Zabbix build base for SQLite3 based images" \
      io.k8s.display-name="Zabbix build base (SQLite3)" \
      io.openshift.expose-services="" \
      io.openshift.tags="zabbix,build,sqlite3" \
      org.label-schema.description="Zabbix build base for SQLite3 based images" \
      org.label-schema.license="AGPL v3.0" \
      org.label-schema.name="zabbix-build-sqlite3-rhel" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.vendor="Zabbix SIA"

ENV ZBX_SOURCES_DIR=/tmp/zabbix-${ZBX_VERSION} ZBX_OUTPUT_DIR=/tmp/zabbix-${ZBX_VERSION}-output \
    DB_TYPE=sqlite3 \
    CFLAGS=${CFLAGS:-"-O2 -fPIE -fstack-protector-strong -D_FORTIFY_SOURCE=2 -Wformat -Wformat-security -pipe"} \
    LDFLAGS=${LDFLAGS:-"-pie -Wl,-z,relro,-z,now,-z,defs -Wl,--as-needed"} \
    MONGODB_PLUGIN_SOURCES=https://git.zabbix.com/scm/ap/mongodb.git MONGODB_PLUGIN_VERSION=${ZBX_PLUGINS_VERSION} \
    POSTGRESQL_PLUGIN_SOURCES=https://git.zabbix.com/scm/ap/postgresql.git POSTGRESQL_PLUGIN_VERSION=${ZBX_PLUGINS_VERSION} \
    MSSQL_PLUGIN_SOURCES=https://git.zabbix.com/scm/ap/mssql.git MSSQL_PLUGIN_VERSION=${ZBX_PLUGINS_VERSION} \
    EMBER_PLUS_PLUGIN_SOURCES=https://git.zabbix.com/scm/ap/ember-plus.git EMBER_PLUS_PLUGIN_VERSION=${ZBX_PLUGINS_VERSION} \
    NVIDIA_GPU_PLUGIN_SOURCES=https://git.zabbix.com/scm/ap/nvidia-gpu.git NVIDIA_GPU_PLUGIN_VERSION=${ZBX_PLUGINS_VERSION}

RUN --mount=from=sources,target=/tmp/src,z \
    --mount=type=bind,source=patches/,target=/tmp/patches,z \
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
        ${ZBX_OUTPUT_DIR}/java_gateway/sbin \
        ${ZBX_OUTPUT_DIR}/web/conf \
        ${ZBX_OUTPUT_DIR}/general/sbin \
        ${ZBX_OUTPUT_DIR}/general/bin \
        ${ZBX_OUTPUT_DIR}/general/conf && \
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
    --mount=from=config_templates,target=/tmp/conf,z \
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
        --with-ares \
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
    gzip -c "database/${DB_TYPE}/create.sql" > "${ZBX_OUTPUT_DIR}/proxy/database/${DB_TYPE}/create.sql.gz" && \
    strip src/zabbix_agent/zabbix_agentd && \
    strip src/zabbix_proxy/zabbix_proxy && \
    strip src/zabbix_get/zabbix_get && \
    strip src/zabbix_sender/zabbix_sender && \
    strip src/go/bin/zabbix_agent2 || true && \
    cp src/zabbix_agent/zabbix_agentd "${ZBX_OUTPUT_DIR}/agent/sbin/" && \
    cp conf/zabbix_agentd.conf "${ZBX_OUTPUT_DIR}/agent/conf/zabbix_agentd.conf.dist" && \
    cp -a conf/zabbix_agentd/. "${ZBX_OUTPUT_DIR}/agent/conf/" && \
    cp -a /tmp/conf/agent/. "${ZBX_OUTPUT_DIR}/agent/conf/" && \
    cp src/go/bin/zabbix_agent2 "${ZBX_OUTPUT_DIR}/agent2/sbin/" && \
    cp src/go/conf/zabbix_agent2.conf "${ZBX_OUTPUT_DIR}/agent2/conf/zabbix_agent2.conf.dist" && \
    cp -a src/go/conf/zabbix_agent2.d/. "${ZBX_OUTPUT_DIR}/agent2/conf/zabbix_agent2.d/" && \
    cp -a /tmp/conf/agent2/. "${ZBX_OUTPUT_DIR}/agent2/conf/" && \
    cp src/zabbix_proxy/zabbix_proxy "${ZBX_OUTPUT_DIR}/proxy/sbin/" && \
    cp conf/zabbix_proxy.conf "${ZBX_OUTPUT_DIR}/proxy/conf/zabbix_proxy.conf.dist" && \
    cp -a /tmp/conf/proxy/. "${ZBX_OUTPUT_DIR}/proxy/conf/" && \
    cp -a src/zabbix_java/bin/. "${ZBX_OUTPUT_DIR}/java_gateway/sbin/bin/" && \
    cp -a src/zabbix_java/lib/. "${ZBX_OUTPUT_DIR}/java_gateway/sbin/lib/" && \
    cp src/zabbix_get/zabbix_get "${ZBX_OUTPUT_DIR}/general/bin/" && \
    cp src/zabbix_sender/zabbix_sender "${ZBX_OUTPUT_DIR}/general/bin/" && \
    cp -a /tmp/conf/web/. "${ZBX_OUTPUT_DIR}/web/conf/" && \
    make -s distclean

RUN set -eux && \
    cd "${ZBX_SOURCES_DIR}" && \
    curl --retry 5 --retry-all-errors --retry-delay 2 --tlsv1.2 -sSf -L \
        "https://raw.githubusercontent.com/notofonts/noto-cjk/main/Sans/OTF/Japanese/NotoSansCJKjp-Regular.otf" \
        -o ui/assets/fonts/NotoSansCJKjp-Regular.otf && \
    curl --retry 5 --retry-all-errors --retry-delay 2 --tlsv1.2 -sSf -L \
        "https://raw.githubusercontent.com/notofonts/noto-cjk/main/Sans/LICENSE" \
        -o ui/assets/fonts/OFL.txt && \
    sed -i -r 's/(define\(.*_FONT_NAME.*)DejaVuSans/\1NotoSansCJKjp-Regular/' ui/include/defines.inc.php && \
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
        cp "$binary" "${ZBX_OUTPUT_DIR}/agent2/sbin/zabbix-agent2-plugin/${outname}"; \
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
        zabbix-agent2-plugin-ember-plus ember.conf ember-plus && \
    build_plugin nvidia-gpu "${NVIDIA_GPU_PLUGIN_SOURCES}" "${NVIDIA_GPU_PLUGIN_VERSION}" \
        zabbix-agent2-plugin-nvidia-gpu nvidia.conf nvidia-gpu
