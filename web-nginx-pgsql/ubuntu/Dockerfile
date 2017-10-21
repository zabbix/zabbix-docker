FROM ubuntu:trusty
LABEL maintainer "Alexey Pustovalov <alexey.pustovalov@zabbix.com>"

ARG APT_FLAGS_COMMON="-qq -y"
ARG APT_FLAGS_PERSISTANT="${APT_FLAGS_COMMON} --no-install-recommends"
ARG APT_FLAGS_DEV="${APT_FLAGS_COMMON} --no-install-recommends"
ARG DB_TYPE=postgresql
ENV LANG=en_US.UTF-8 LANGUAGE=en_US:en LC_ALL=en_US.UTF-8 DEBIAN_FRONTEND=noninteractive TERM=xterm

RUN DISTRIB_CODENAME=$(/bin/bash -c 'source /etc/lsb-release && echo $DISTRIB_CODENAME') && \
    locale-gen $LC_ALL && \
    echo "#!/bin/sh\nexit 0" > /usr/sbin/policy-rc.d && \
    echo "deb http://us.archive.ubuntu.com/ubuntu/ $DISTRIB_CODENAME multiverse" >> /etc/apt/sources.list && \
    addgroup --system --quiet zabbix && \
    adduser --quiet \
            --system --disabled-login \
            --ingroup zabbix \
            --home /var/lib/zabbix/ \
            --no-create-home \
        zabbix && \
    mkdir -p /etc/zabbix && \
    mkdir -p /etc/zabbix/web && \
    chown --quiet -R zabbix:root /etc/zabbix && \
    apt-get ${APT_FLAGS_COMMON} update && \
    apt-get ${APT_FLAGS_COMMON} install \
            wget 1>/dev/null && \
    wget -q http://nginx.org/keys/nginx_signing.key -O- | apt-key add - && \
    echo "deb http://nginx.org/packages/ubuntu/ $DISTRIB_CODENAME nginx" >> /etc/apt/sources.list.d/nginx.list && \
    apt-get ${APT_FLAGS_COMMON} update && \
    apt-get ${APT_FLAGS_PERSISTANT} install \
            supervisor \
            postgresql-client \
            nginx \
            php5-fpm \
            php5-pgsql \
            php5-gd \
            php5-json \
            php5-ldap \
            ttf-dejavu-core 1>/dev/null && \
    apt-get ${APT_FLAGS_COMMON} autoremove && \
    apt-get ${APT_FLAGS_COMMON} clean && \
    mkdir -p /var/lib/php5 && \
    chown --quiet -R www-data:www-data /var/lib/php5 && \
    rm -rf /var/cache/nginx/* && \
    rm -rf /var/lib/apt/lists/*

ARG MAJOR_VERSION=3.4
ARG ZBX_VERSION=${MAJOR_VERSION}.3
ARG ZBX_SOURCES=svn://svn.zabbix.com/tags/${ZBX_VERSION}/
ENV ZBX_VERSION=${ZBX_VERSION} ZBX_SOURCES=${ZBX_SOURCES} DB_TYPE=${DB_TYPE}

ADD conf/tmp/font-config /tmp/font-config

RUN apt-get ${APT_FLAGS_COMMON} update && \
    apt-get ${APT_FLAGS_DEV} install \
            subversion \
            patch \
            gettext 1>/dev/null && \
    mkdir -p /var/lib/locales/supported.d/ && \
    echo "en_US.UTF-8 UTF-8" > /var/lib/locales/supported.d/local && \
    echo "ru_RU.UTF-8 UTF-8" >> /var/lib/locales/supported.d/local && \
    echo "ja_JP.UTF-8 UTF-8" >> /var/lib/locales/supported.d/local && \
    echo "cs_CZ.UTF-8 UTF-8" >> /var/lib/locales/supported.d/local && \
    echo "ko_KR.UTF-8 UTF-8" >> /var/lib/locales/supported.d/local && \
    echo "it_IT.UTF-8 UTF-8" >> /var/lib/locales/supported.d/local && \
    echo "pt_BR.UTF-8 UTF-8" >> /var/lib/locales/supported.d/local && \
    echo "sk_SK.UTF-8 UTF-8" >> /var/lib/locales/supported.d/local && \
    echo "uk_UA.UTF-8 UTF-8" >> /var/lib/locales/supported.d/local && \
    echo "fr_FR.UTF-8 UTF-8" >> /var/lib/locales/supported.d/local && \
    echo "pl_PL.UTF-8 UTF-8" >> /var/lib/locales/supported.d/local && \
    echo "zh_CN.UTF-8 UTF-8" >> /var/lib/locales/supported.d/local && \
    dpkg-reconfigure locales 1>/dev/null && \
    cd /usr/share/ && \
    svn --quiet export ${ZBX_SOURCES}/frontends/php/ zabbix 1>/dev/null && \
    cd /usr/share/zabbix/ && \
    patch -p3 < /tmp/font-config && \
    rm /tmp/font-config && \
    rm -f conf/zabbix.conf.php && \
    rm -rf tests && \
    rm /usr/share/zabbix/fonts/DejaVuSans.ttf && \
    ./locale/make_mo.sh 2>/dev/null && \
    update-alternatives --install /usr/share/zabbix/fonts/graphfont.ttf \
            zabbix-frontend-font /usr/share/fonts/truetype/ttf-dejavu/DejaVuSans.ttf 10 && \
    apt-get ${APT_FLAGS_COMMON} purge \
            subversion \
            patch \
            gettext 1>/dev/null && \
    apt-get ${APT_FLAGS_COMMON} autoremove && \
    apt-get ${APT_FLAGS_COMMON} clean && \
    rm -rf /var/lib/apt/lists/*

EXPOSE 80/TCP 443/TCP

WORKDIR /usr/share/zabbix

VOLUME ["/etc/ssl/nginx"]

ADD conf/etc/supervisor/ /etc/supervisor/
ADD conf/etc/zabbix/nginx.conf /etc/zabbix/
ADD conf/etc/zabbix/nginx_ssl.conf /etc/zabbix/
ADD conf/etc/zabbix/web/zabbix.conf.php /etc/zabbix/web/
ADD conf/etc/nginx/nginx.conf /etc/nginx/
ADD conf/etc/php5/fpm/conf.d/99-zabbix.ini /etc/php5/fpm/conf.d/
ADD run_zabbix_component.sh /

ENTRYPOINT ["/bin/bash"]

CMD ["/run_zabbix_component.sh", "frontend", "postgresql", "nginx"]
