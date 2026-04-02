#!/bin/bash

set -euo pipefail

readonly ENTRYPOINT_LIBS="/usr/lib/docker-entrypoint"
source "${ENTRYPOINT_LIBS}/bootstrap.sh"

# Default Zabbix server host
: "${ZBX_SERVER_HOST=zabbix-server}"
# Default Zabbix server port number
: "${ZBX_SERVER_PORT=10051}"

readonly ZABBIX_INTERNAL_ENC_DIR="${ZABBIX_USER_HOME_DIR}/enc_internal"

update_config() {
    : "${ZBX_PASSIVESERVERS:=}"
    : "${ZBX_ACTIVESERVERS:=}"

    local server="${ZBX_SERVER_HOST}"

    if [ -n "$ZBX_SERVER_PORT" ] && [ "$ZBX_SERVER_PORT" != "10051" ]; then
        server="${server}:${ZBX_SERVER_PORT}"
    fi

    if [ -n "$ZBX_SERVER_HOST" ]; then
        if [ -n "$ZBX_PASSIVESERVERS" ]; then
            ZBX_PASSIVESERVERS="${ZBX_SERVER_HOST},${ZBX_PASSIVESERVERS}"
        else
            ZBX_PASSIVESERVERS="${ZBX_SERVER_HOST}"
        fi

        if [ -n "$ZBX_ACTIVESERVERS" ]; then
            ZBX_ACTIVESERVERS="${server},${ZBX_ACTIVESERVERS}"
        else
            ZBX_ACTIVESERVERS="${server}"
        fi
    fi

    : "${ZBX_PASSIVE_ALLOW:=true}"
    if [ "${ZBX_PASSIVE_ALLOW,,}" = "true" ] && [ -n "$ZBX_PASSIVESERVERS" ]; then
        info "** Using '$ZBX_PASSIVESERVERS' servers for passive checks"
        export ZBX_PASSIVESERVERS
    else
        unset ZBX_PASSIVESERVERS
    fi

    : "${ZBX_ACTIVE_ALLOW:=true}"
    if [ "${ZBX_ACTIVE_ALLOW,,}" = "true" ] && [ -n "$ZBX_ACTIVESERVERS" ]; then
        info "** Using '$ZBX_ACTIVESERVERS' servers for active checks"
        export ZBX_ACTIVESERVERS
    else
        unset ZBX_ACTIVESERVERS
    fi

    unset ZBX_SERVER_HOST
    unset ZBX_SERVER_PORT

    update_config_multiple_var "${ZABBIX_CONF_DIR}/zabbix_agentd_item_keys.conf" "DenyKey" "${ZBX_DENYKEY:-}"
    update_config_multiple_var "${ZABBIX_CONF_DIR}/zabbix_agentd_item_keys.conf" "AllowKey" "${ZBX_ALLOWKEY:-}"

    update_config_multiple_var "${ZABBIX_CONF_DIR}/zabbix_agentd_modules.conf" "LoadModule" "${ZBX_LOADMODULE:-}"

    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "ZBX_TLSCAFILE" "${ZBX_TLSCAFILE:-}" "${ZBX_TLSCA:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "ZBX_TLSCRLFILE" "${ZBX_TLSCRLFILE:-}" "${ZBX_TLSCRL:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "ZBX_TLSCERTFILE" "${ZBX_TLSCERTFILE:-}" "${ZBX_TLSCERT:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "ZBX_TLSKEYFILE" "${ZBX_TLSKEYFILE:-}" "${ZBX_TLSKEY:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "ZBX_TLSPSKFILE" "${ZBX_TLSPSKFILE:-}" "${ZBX_TLSPSK:-}"

    if [ "$(id -u)" -ne 0 ]; then
        export ZBX_USER="$(id -un)"
    else
        export ZBX_ALLOWROOT=1
    fi
}

prepare_service() {
    info "** Preparing Zabbix agent"

    update_config
    clear_zbx_env
}

#################################################

if [ $# -eq 0 ]; then
    set -- /usr/sbin/zabbix_agentd
elif [ "${1#-}" != "$1" ]; then
    set -- /usr/sbin/zabbix_agentd "$@"
fi

if [ "${1:-}" = '/usr/sbin/zabbix_agentd' ]; then
    prepare_service
fi

exec "$@"

#################################################
