#!/bin/bash

set -euo pipefail

readonly ENTRYPOINT_LIBS="/usr/lib/docker-entrypoint"
source "${ENTRYPOINT_LIBS}/bootstrap.sh"

source "${ENTRYPOINT_LIBS}/proxy-config.sh"

update_config() {
    : "${ZBX_USE_NODE_NAME_AS_DB_NAME:=false}"
    if [ "${ZBX_USE_NODE_NAME_AS_DB_NAME,,}" = "true" ]; then
        local node_name
        node_name="$(uname -n)"
        export ZBX_DB_NAME="${ZABBIX_USER_HOME_DIR}/db_data/${node_name}.sqlite"
    else
        export ZBX_DB_NAME="${ZABBIX_USER_HOME_DIR}/db_data/${ZBX_HOSTNAME:-zabbix-proxy-sqlite3}.sqlite"
    fi
    unset ZBX_USE_NODE_NAME_AS_DB_NAME

    proxy_config
}

prepare_service() {
    info "** Preparing Zabbix proxy"

    update_config
    clear_zbx_env
}

#################################################

if [ $# -eq 0 ]; then
    set -- /usr/sbin/zabbix_proxy
elif [ "${1#-}" != "$1" ]; then
    set -- /usr/sbin/zabbix_proxy "$@"
fi

if [ "${1:-}" = '/usr/sbin/zabbix_proxy' ]; then
    prepare_service
fi

exec "$@"

#################################################
