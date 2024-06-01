# syntax=docker/dockerfile:1
ARG MAJOR_VERSION=5.0
ARG RELEASE=42
ARG ZBX_VERSION=${MAJOR_VERSION}.42
ARG BUILD_BASE_IMAGE=zabbix-build-base:rhel-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} AS builder

ARG MAJOR_VERSION
ARG RELEASE
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git
ARG GIT_BRANCH

LABEL description="Zabbix build base for SQLite3 based images" \
      maintainer="alexey.pustovalov@zabbix.com" \
      name="zabbix/zabbix-build-sqlite3-50" \
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
      org.label-schema.name="zabbix-build-sqlite3-rhel" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.vendor="Zabbix SIA"

ENV ZBX_SOURCES_DIR=/tmp/zabbix-${ZBX_VERSION} ZBX_OUTPUT_DIR=/tmp/zabbix-${ZBX_VERSION}-output \
    DB_TYPE=sqlite3

RUN --mount=type=cache,target=/root/.cache/go-build/ \
    --mount=type=cache,target=/root/go/ \
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
    mkdir -p ${ZBX_OUTPUT_DIR}/agent/sbin/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/agent/conf/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/agent2/sbin/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/agent2/conf/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/proxy/sbin/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/proxy/conf/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/java_gateway/sbin/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/general/sbin/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/general/bin/ && \
    mkdir -p ${ZBX_OUTPUT_DIR}/general/conf/ && \
    git -c advice.detachedHead=false clone ${ZBX_SOURCES} --branch ${GIT_BRANCH:-$ZBX_VERSION} --depth 1 --single-branch ${ZBX_SOURCES_DIR} && \
    cd ${ZBX_SOURCES_DIR} && \
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
            --with-ldap \
            --with-libcurl \
            --with-libxml2 \
            --with-net-snmp \
            --with-openipmi \
            --with-openssl \
            --with-sqlite3 \
            --with-ssh \
            --with-unixodbc \
            $additional_components \
            --silent && \
    make -j"$(nproc)" -s dbschema && \
    make -j"$(nproc)" -s && \
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
    strip ${ZBX_SOURCES_DIR}/src/zabbix_proxy/zabbix_proxy && \
    strip ${ZBX_SOURCES_DIR}/src/go/bin/zabbix_agent2 && \
    strip ${ZBX_SOURCES_DIR}/src/zabbix_get/zabbix_get && \
    strip ${ZBX_SOURCES_DIR}/src/zabbix_sender/zabbix_sender && \
    cp ${ZBX_SOURCES_DIR}/src/zabbix_agent/zabbix_agentd ${ZBX_OUTPUT_DIR}/agent/sbin/ && \
    cp ${ZBX_SOURCES_DIR}/conf/zabbix_agentd.conf ${ZBX_OUTPUT_DIR}/agent/conf/ && \
    cp -R ${ZBX_SOURCES_DIR}/conf/zabbix_agentd/ ${ZBX_OUTPUT_DIR}/agent/conf/ && \
    cp ${ZBX_SOURCES_DIR}/src/go/bin/zabbix_agent2 ${ZBX_OUTPUT_DIR}/agent2/sbin/ && \
    cp ${ZBX_SOURCES_DIR}/src/go/conf/zabbix_agent2.conf ${ZBX_OUTPUT_DIR}/agent2/conf/ && \
    cp ${ZBX_SOURCES_DIR}/src/zabbix_proxy/zabbix_proxy ${ZBX_OUTPUT_DIR}/proxy/sbin/ && \
    cp ${ZBX_SOURCES_DIR}/conf/zabbix_proxy.conf ${ZBX_OUTPUT_DIR}/proxy/conf/ && \
    cp -R ${ZBX_SOURCES_DIR}/src/zabbix_java/bin/ ${ZBX_OUTPUT_DIR}/java_gateway/sbin/ && \
    cp -R ${ZBX_SOURCES_DIR}/src/zabbix_java/lib/ ${ZBX_OUTPUT_DIR}/java_gateway/sbin/ && \
    cp ${ZBX_SOURCES_DIR}/src/zabbix_get/zabbix_get ${ZBX_OUTPUT_DIR}/general/bin/ && \
    cp ${ZBX_SOURCES_DIR}/src/zabbix_sender/zabbix_sender ${ZBX_OUTPUT_DIR}/general/bin/ && \
    make -s distclean
