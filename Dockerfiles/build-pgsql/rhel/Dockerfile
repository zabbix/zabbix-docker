# syntax=docker/dockerfile:1
ARG MAJOR_VERSION=6.4
ARG RELEASE=13
ARG ZBX_VERSION=${MAJOR_VERSION}.13
ARG BUILD_BASE_IMAGE=zabbix-build-base:rhel-${ZBX_VERSION}

FROM ${BUILD_BASE_IMAGE} AS builder

ARG MAJOR_VERSION
ARG RELEASE
ARG ZBX_VERSION
ARG ZBX_SOURCES=https://git.zabbix.com/scm/zbx/zabbix.git

LABEL description="Zabbix build base for PostgreSQL based images" \
      maintainer="alexey.pustovalov@zabbix.com" \
      name="zabbix/zabbix-build-pgsql-64" \
      release="${RELEASE}" \
      summary="Zabbix build base (PostgreSQL)" \
      url="https://www.zabbix.com/" \
      vendor="Zabbix LLC" \
      version="${MAJOR_VERSION}" \
      io.k8s.description="Zabbix build base for PostgreSQL based images" \
      io.k8s.display-name="Zabbix build base (PostgreSQL)" \
      io.openshift.expose-services="" \
      io.openshift.tags="zabbix,build,pgsql" \
      org.label-schema.description="Zabbix build base for PostgreSQL based images" \
      org.label-schema.name="zabbix-build-pgsql-rhel" \
      org.label-schema.url="https://zabbix.com/" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.vendor="Zabbix LLC"

ENV MONGODB_PLUGIN_SOURCES=https://git.zabbix.com/scm/ap/mongodb.git MONGODB_PLUGIN_VERSION=${ZBX_VERSION} \
    POSTGRESQL_PLUGIN_SOURCES=https://git.zabbix.com/scm/ap/postgresql.git POSTGRESQL_PLUGIN_VERSION=${ZBX_VERSION}

RUN --mount=type=cache,target=/root/.cache/go-build/ \
    --mount=type=cache,target=/root/go/ \
    --mount=type=bind,source=conf/chromedp_no_sandbox.patch,target=/tmp/chromedp_no_sandbox.patch \
    set -eux && \
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
            --enable-agent \
            --enable-agent2 \
            --enable-ipv6 \
            --enable-java \
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
            --with-postgresql \
            --with-ssh \
            --with-unixodbc \
            --silent && \
    make -j"$(nproc)" -s dbschema && \
    make -j"$(nproc)" -s && \
    make -j"$(nproc)" -s gettext && \
    cat database/postgresql/schema.sql > database/postgresql/create.sql && \
    cat database/postgresql/images.sql >> database/postgresql/create.sql && \
    cat database/postgresql/data.sql >> database/postgresql/create.sql && \
    gzip -c database/postgresql/create.sql > database/postgresql/create_server.sql.gz && \
    rm -rf database/postgresql/create.sql && \
    cat database/postgresql/schema.sql > database/postgresql/create.sql && \
    gzip -c database/postgresql/create.sql > database/postgresql/create_proxy.sql.gz && \
    rm -rf database/postgresql/create.sql && \
    mkdir /tmp/fonts/ && \
    curl --tlsv1.2 -sSf -L "https://noto-website.storage.googleapis.com/pkgs/NotoSansCJKjp-hinted.zip" -o /tmp/fonts/NotoSansCJKjp-hinted.zip && \
    unzip /tmp/fonts/NotoSansCJKjp-hinted.zip -d /tmp/fonts/ && \
    cp /tmp/fonts/NotoSansCJKjp-Regular.otf /tmp/zabbix-${ZBX_VERSION}/ui/assets/fonts/NotoSansCJKjp-Regular.ttf && \
    cp /tmp/fonts/LICENSE_OFL.txt /tmp/zabbix-${ZBX_VERSION}/ui/assets/fonts/ && \
    rm -f /tmp/zabbix-${ZBX_VERSION}/ui/assets/fonts/DejaVuSans.ttf && \
    sed -i -r "s/(define\(.*_FONT_NAME.*)DejaVuSans/\1NotoSansCJKjp-Regular/" /tmp/zabbix-${ZBX_VERSION}/ui/include/defines.inc.php && \
    rm -rf /tmp/fonts/ && \
    chmod o+r /tmp/zabbix-${ZBX_VERSION}/ui/assets/fonts/* && \
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
    strip /tmp/postgresql_plugin/zabbix-agent2-plugin-postgresql
