#!/bin/bash

set -o pipefail

set +e

# Script trace mode
if [ "${DEBUG_MODE}" == "true" ]; then
    set -o xtrace
fi

# Default directories
# Configuration files directory
ZABBIX_ETC_DIR="/etc/zabbix"

prepare_java_gateway_config() {
    echo "** Preparing Zabbix Java Gateway log configuration file"

    ZBX_GATEWAY_CONFIG=$ZABBIX_ETC_DIR/zabbix_java_gateway_logback.xml

    if [ -n "${ZBX_DEBUGLEVEL}" ]; then
        echo "Updating $ZBX_GATEWAY_CONFIG 'DebugLevel' parameter: '${ZBX_DEBUGLEVEL}'... updated"
        if [ -f "$ZBX_GATEWAY_CONFIG" ]; then
            sed -i -e "/^.*<root level=/s/=.*/=\"${ZBX_DEBUGLEVEL}\">/" "$ZBX_GATEWAY_CONFIG"
        else
            echo "**** Zabbix Java Gateway log configuration file '$ZBX_GATEWAY_CONFIG' not found"
        fi
    fi
}

prepare_java_gateway() {
    echo "** Preparing Zabbix Java Gateway"

    prepare_java_gateway_config
}

#################################################

if [ "$1" == '/usr/sbin/zabbix_java_gateway' ]; then
    prepare_java_gateway
fi

exec "$@"


#################################################
