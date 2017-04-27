FROM ubuntu:trusty
LABEL maintainer "Alexey Pustovalov <alexey.pustovalov@zabbix.com>"

ARG APT_FLAGS_COMMON="-qq -y"
ARG APT_FLAGS_PERSISTANT="${APT_FLAGS_COMMON} --no-install-recommends"
ARG APT_FLAGS_DEV="${APT_FLAGS_COMMON} --no-install-recommends"
ARG DB_TYPE=mysql
ENV LANG=en_US.UTF-8 LANGUAGE=en_US:en LC_ALL=en_US.UTF-8 DEBIAN_FRONTEND=noninteractive TERM=xterm

ADD snmptrapfmt_1.14+nmu1ubuntu2_amd64.deb /tmp/

RUN locale-gen $LC_ALL && \
    echo "#!/bin/sh\nexit 0" > /usr/sbin/policy-rc.d && \
    DISTRIB_CODENAME=$(/bin/bash -c 'source /etc/lsb-release && echo $DISTRIB_CODENAME') && \
    echo "deb http://us.archive.ubuntu.com/ubuntu/ $DISTRIB_CODENAME multiverse" >> /etc/apt/sources.list && \
    addgroup --system --quiet zabbix && \
    adduser --quiet \
            --system --disabled-login \
            --ingroup zabbix \
            --home /var/lib/zabbix/ \
        zabbix && \
    apt-get ${APT_FLAGS_COMMON} update && \
    apt-get ${APT_FLAGS_PERSISTANT} install \
            supervisor \
            snmpd \
            snmp-mibs-downloader && \
    dpkg -i /tmp/snmptrapfmt_1.14+nmu1ubuntu2_amd64.deb && \
    mkdir -p /var/lib/zabbix && \
    mkdir -p /var/lib/zabbix/snmptraps && \
    mkdir -p /var/lib/zabbix/mibs && \
    chown --quiet -R zabbix:root /var/lib/zabbix && \
    echo "disableAuthorization yes" >> "/etc/snmp/snmptrapd.conf" && \
    echo "traphandle default /usr/sbin/snmptrapfmthdlr" >> "/etc/snmp/snmptrapd.conf" && \
    sed -i \
            -e "/^VARFMT=/s/=.*/=\"%n %v \"/" \
            -e '/^LOGFMT=/s/=.*/=\"$x ZBXTRAP $R $G $S $e $*\"/' \
            -e "/^LOGFILE=/s/=.*/=\"\/var\/lib\/zabbix\/snmptraps\/snmptraps.log\"/" \
        "/etc/snmp/snmptrapfmt.conf" && \
    rm -rf /tmp/snmptrapfmt_1.14+nmu1ubuntu2_amd64.deb && \
    apt-get ${APT_FLAGS_COMMON} autoremove && \
    apt-get ${APT_FLAGS_COMMON} clean && \
    rm -rf /var/lib/apt/lists/*

EXPOSE 162/UDP

WORKDIR /var/lib/zabbix/snmptraps/

VOLUME ["/var/lib/zabbix/snmptraps", "/var/lib/zabbix/mibs"]

ADD conf/etc/supervisor/ /etc/supervisor/
ADD conf/etc/logrotate.d/zabbix_snmptraps /etc/logrotate.d/

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/supervisord.conf"]
