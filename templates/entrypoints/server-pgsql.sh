#!/usr/bin/env bash

set -euo pipefail

readonly ENTRYPOINT_LIBS="/usr/lib/docker-entrypoint"
source "${ENTRYPOINT_LIBS}/bootstrap.sh"

source "${ENTRYPOINT_LIBS}/pgsql.sh"
source "${ENTRYPOINT_LIBS}/server-config.sh"

readonly ZBX_SERVER_CONFIG="$ZABBIX_CONF_DIR/zabbix_server.conf"

update_config() {
    info "** Preparing Zabbix server configuration file"

    [[ -f "$ZBX_SERVER_CONFIG" ]] || error "Missing configuration file: $ZBX_SERVER_CONFIG"

    if [ -n "${DB_SERVER_HOST:-}" ]; then
        update_config_var "$ZBX_SERVER_CONFIG" "DBHost" "${DB_SERVER_HOST}"
    else
        update_config_var "$ZBX_SERVER_CONFIG" "DBHost" '""'
    fi

    update_config_var "$ZBX_SERVER_CONFIG" "DBPort" "${DB_SERVER_PORT:-}"

    server_config
}

prepare_database() {
    info "** Preparing database"

    check_db_variables "zabbix"
    check_db_connect
    create_db_database
    create_db_schema "/usr/share/doc/zabbix-server-postgresql/create.sql.gz"
}

prepare_service() {
    info "** Preparing Zabbix server"

    prepare_database
    update_config
    clear_zbx_env
}

#################################################

if [ $# -eq 0 ]; then
    set -- /usr/sbin/zabbix_server
elif [ "${1#-}" != "$1" ]; then
    set -- /usr/sbin/zabbix_server "$@"
fi

if [ "${1:-}" = '/usr/sbin/zabbix_server' ]; then
    prepare_service
fi

if [ "${1:-}" = "init_db_only" ]; then
    prepare_database
else
    exec "$@"
fi

#################################################
