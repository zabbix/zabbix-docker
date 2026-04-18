#!/usr/bin/env bash

set -euo pipefail

readonly ENTRYPOINT_LIBS="/usr/lib/docker-entrypoint"
source "${ENTRYPOINT_LIBS}/bootstrap.sh"

source "${ENTRYPOINT_LIBS}/mysql.sh"
source "${ENTRYPOINT_LIBS}/proxy-config.sh"

readonly ZBX_PROXY_CONFIG="${ZABBIX_CONF_DIR}/zabbix_proxy.conf"

update_config() {
    local proxy_name="${ZBX_HOSTNAME:-zabbix-proxy-mysql}"

    info "** Preparing Zabbix proxy configuration file"

    [[ -f "$ZBX_PROXY_CONFIG" ]] || error "Missing configuration file: $ZBX_PROXY_CONFIG"

    if [ -n "${ZBX_DBTLSCONNECT:-}" ]; then
        update_config_var "${ZBX_PROXY_CONFIG}" "DBTLSConnect" "${ZBX_DBTLSCONNECT:-}"
        update_config_var "${ZBX_PROXY_CONFIG}" "DBTLSCAFile" "${ZBX_DBTLSCAFILE:-}"
        update_config_var "${ZBX_PROXY_CONFIG}" "DBTLSCertFile" "${ZBX_DBTLSCERTFILE:-}"
        update_config_var "${ZBX_PROXY_CONFIG}" "DBTLSKeyFile" "${ZBX_DBTLSKEYFILE:-}"
        update_config_var "${ZBX_PROXY_CONFIG}" "DBTLSCipher" "${ZBX_DBTLSCIPHER:-}"
        update_config_var "${ZBX_PROXY_CONFIG}" "DBTLSCipher13" "${ZBX_DBTLSCIPHER13:-}"
    fi

    if [ -z "${DB_SERVER_SOCKET:-}" ]; then
        update_config_var "${ZBX_PROXY_CONFIG}" "DBHost" "${DB_SERVER_HOST:-}"
        update_config_var "${ZBX_PROXY_CONFIG}" "DBPort" "${DB_SERVER_PORT:-}"
    else
        update_config_var "${ZBX_PROXY_CONFIG}" "DBHost"
        update_config_var "${ZBX_PROXY_CONFIG}" "DBPort"
    fi

    update_config_var "${ZBX_PROXY_CONFIG}" "DBSocket" "${DB_SERVER_SOCKET:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "DBName" "${DB_SERVER_DBNAME:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "DBSchema" "${DB_SERVER_SCHEMA:-}"

    if [ -n "${ZBX_VAULTDBPATH:-}" ] && [ -n "${ZBX_VAULTURL:-}" ]; then
        update_config_var "${ZBX_PROXY_CONFIG}" "VaultDBPath" "${ZBX_VAULTDBPATH:-}"
        update_config_var "${ZBX_PROXY_CONFIG}" "VaultURL" "${ZBX_VAULTURL:-}"
        update_config_var "${ZBX_PROXY_CONFIG}" "DBUser"
        update_config_var "${ZBX_PROXY_CONFIG}" "DBPassword"
    else
        update_config_var "${ZBX_PROXY_CONFIG}" "VaultDBPath"
        update_config_var "${ZBX_PROXY_CONFIG}" "VaultURL"
        update_config_var "${ZBX_PROXY_CONFIG}" "DBUser" "${DB_SERVER_ZBX_USER:-}"
        update_config_var "${ZBX_PROXY_CONFIG}" "DBPassword" "${DB_SERVER_ZBX_PASS:-}"
    fi

    proxy_config "$proxy_name"
}

prepare_database() {
    info "** Preparing database"

    check_db_variables "zabbix_proxy"
    check_db_connect
    create_db_user
    create_db_database
    create_db_schema "/usr/share/doc/zabbix-proxy-mysql/create.sql.gz"
}

prepare_service() {
    info "** Preparing Zabbix proxy"

    prepare_database
    update_config
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

if [ "${1:-}" = "init_db_only" ]; then
    prepare_database
else
    exec "$@"
fi

#################################################
