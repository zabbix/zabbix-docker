#!/usr/bin/env bash

set -euo pipefail

readonly ENTRYPOINT_LIBS="/usr/lib/docker-entrypoint"
source "${ENTRYPOINT_LIBS}/bootstrap.sh"

source "${ENTRYPOINT_LIBS}/sqlite3.sh"
source "${ENTRYPOINT_LIBS}/proxy-config.sh"

readonly ZBX_PROXY_CONFIG="${ZABBIX_CONF_DIR}/zabbix_proxy.conf"

update_config() {
    local proxy_name="${ZBX_HOSTNAME:-zabbix-proxy-sqlite3}"
    local node_name=""

    info "** Preparing Zabbix proxy configuration file"

    [[ -f "$ZBX_PROXY_CONFIG" ]] || error "Missing configuration file: $ZBX_PROXY_CONFIG"

    update_config_var "${ZBX_PROXY_CONFIG}" "DBHost"

    : "${ZBX_USE_NODE_NAME_AS_DB_NAME:=false}"
    if [ "${ZBX_USE_NODE_NAME_AS_DB_NAME,,}" = "false" ]; then
        update_config_var "${ZBX_PROXY_CONFIG}" "DBName" "${ZABBIX_USER_HOME_DIR}/db_data/${proxy_name}.sqlite"
        export ZBX_DB_NAME="${ZABBIX_USER_HOME_DIR}/db_data/${proxy_name}.sqlite"
    else
        node_name="$(uname -n)"
        update_config_var "${ZBX_PROXY_CONFIG}" "DBName" "${ZABBIX_USER_HOME_DIR}/db_data/${node_name}.sqlite"
        export ZBX_DB_NAME="${ZABBIX_USER_HOME_DIR}/db_data/${node_name}.sqlite"
    fi

    update_config_var "${ZBX_PROXY_CONFIG}" "DBUser"
    update_config_var "${ZBX_PROXY_CONFIG}" "DBPort"
    update_config_var "${ZBX_PROXY_CONFIG}" "DBPassword"

    proxy_config "$proxy_name"
}

prepare_service() {
    info "** Preparing Zabbix proxy"

    update_config
    create_db_schema "/usr/share/doc/zabbix-proxy-sqlite3/create.sql.gz"
    clear_zbx_env
}

#################################################

if [ $# -eq 0 ]; then
    set -- /usr/sbin/zabbix_proxy
elif [ "${1#-}" != "$1" ]; then
    set -- /usr/sbin/zabbix_proxy "$@"
fi

if [ "${1:-}" = "/usr/sbin/zabbix_proxy" ]; then
    prepare_service
fi

exec "$@"

#################################################
