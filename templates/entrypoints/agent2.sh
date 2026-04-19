#!/usr/bin/env bash

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

    if [ "${ZBX_ENABLEPERSISTENTBUFFER:-}" == "true" ]; then
        export ZBX_ENABLEPERSISTENTBUFFER=1
    else
        unset ZBX_ENABLEPERSISTENTBUFFER
        unset ZBX_PERSISTENTBUFFERFILE
    fi

    if [ "${ZBX_ENABLESTATUSPORT:-}" == "true" ]; then
        export ZBX_STATUSPORT=${ZBX_STATUSPORT:=31999}
    else
        unset ZBX_STATUSPORT
    fi

    update_config_multiple_var "${ZABBIX_CONF_DIR}/zabbix_agent2_item_keys.conf" "DenyKey" "${ZBX_DENYKEY:-}"
    update_config_multiple_var "${ZABBIX_CONF_DIR}/zabbix_agent2_item_keys.conf" "AllowKey" "${ZBX_ALLOWKEY:-}"

    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "ZBX_TLSCAFILE" "${ZBX_TLSCAFILE:-}" "${ZBX_TLSCA:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "ZBX_TLSCRLFILE" "${ZBX_TLSCRLFILE:-}" "${ZBX_TLSCRL:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "ZBX_TLSCERTFILE" "${ZBX_TLSCERTFILE:-}" "${ZBX_TLSCERT:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "ZBX_TLSKEYFILE" "${ZBX_TLSKEYFILE:-}" "${ZBX_TLSKEY:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "ZBX_TLSPSKFILE" "${ZBX_TLSPSKFILE:-}" "${ZBX_TLSPSK:-}"
}

update_plugin_config() {
    info "** Preparing Zabbix agent 2 plugin configuration files"

    local plugin_config_dir="${ZABBIX_CONF_DIR}/zabbix_agent2.d/plugins.d"
    local plugin_bin_dir="/usr/sbin/zabbix-agent2-plugin"

    update_config_var "$plugin_config_dir/mongodb.conf" "Plugins.MongoDB.System.Path" "$plugin_bin_dir/mongodb"
    update_config_var "$plugin_config_dir/postgresql.conf" "Plugins.PostgreSQL.System.Path" "$plugin_bin_dir/postgresql"
    update_config_var "$plugin_config_dir/mssql.conf" "Plugins.MSSQL.System.Path" "$plugin_bin_dir/mssql"
    update_config_var "$plugin_config_dir/ember.conf" "Plugins.EmberPlus.System.Path" "$plugin_bin_dir/ember-plus"

    if command -v nvidia-smi >/dev/null 2>&1
    then
        update_config_var "$plugin_config_dir/nvidia.conf" "Plugins.NVIDIA.System.Path" "$plugin_bin_dir/nvidia-gpu"
    fi
}

prepare_service() {
    info "** Preparing Zabbix agent 2"

    update_config
    update_plugin_config
    clear_zbx_env
}

#################################################

if [ $# -eq 0 ]; then
    set -- /usr/sbin/zabbix_agent2
elif [ "${1#-}" != "$1" ]; then
    set -- /usr/sbin/zabbix_agent2 "$@"
fi

if [ "${1:-}" = '/usr/sbin/zabbix_agent2' ]; then
    prepare_service
fi

exec "$@"

#################################################
