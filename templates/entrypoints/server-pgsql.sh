#!/bin/bash

set -euo pipefail

readonly ENTRYPOINT_LIBS="/usr/lib/docker-entrypoint"
source "${ENTRYPOINT_LIBS}/bootstrap.sh"

source "${ENTRYPOINT_LIBS}/pgsql.sh"
source "${ENTRYPOINT_LIBS}/server-config.sh"

update_config() {
    [ -n "${DB_SERVER_HOST:-}" ] && export ZBX_DB_HOST="${DB_SERVER_HOST}"
    [ -n "${DB_SERVER_PORT:-}" ] && export ZBX_DB_PORT="${DB_SERVER_PORT}"

    export ZBX_DB_NAME="${DB_SERVER_DBNAME}"
    export ZBX_DB_SCHEMA="${DB_SERVER_SCHEMA}"

    if [ -n "${ZBX_VAULT:-}" ] && [ -n "${ZBX_VAULTURL:-}" ] && [ -z "${ZBX_VAULTDBPATH:-}" ]; then
        export ZBX_DB_USER="${DB_SERVER_ZBX_USER}"
        export ZBX_DB_PASSWORD="${DB_SERVER_ZBX_PASS}"
    elif [ -z "${ZBX_VAULT:-}" ] && [ -z "${ZBX_VAULTURL:-}" ]; then
        export ZBX_DB_USER="${DB_SERVER_ZBX_USER}"
        export ZBX_DB_PASSWORD="${DB_SERVER_ZBX_PASS}"
    else
        unset ZBX_DB_USER
        unset ZBX_DB_PASSWORD
    fi

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
