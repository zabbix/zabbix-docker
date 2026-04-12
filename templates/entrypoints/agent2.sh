#!/usr/bin/env bash

set -euo pipefail

readonly ENTRYPOINT_LIBS="/usr/lib/docker-entrypoint"
source "${ENTRYPOINT_LIBS}/bootstrap.sh"

# Default Zabbix server host
: "${ZBX_SERVER_HOST=zabbix-server}"
# Default Zabbix server port number
: "${ZBX_SERVER_PORT=10051}"

readonly ZBX_AGENT_CONFIG="${ZABBIX_CONF_DIR}/zabbix_agent2.conf"
readonly ZABBIX_INTERNAL_ENC_DIR="${ZABBIX_USER_HOME_DIR}/enc_internal"

update_config() {
    : "${ZBX_PASSIVESERVERS=}"
    : "${ZBX_ACTIVESERVERS=}"

    local server="${ZBX_SERVER_HOST}"

    if [ -n "${ZBX_SERVER_PORT}" ] && [ "${ZBX_SERVER_PORT}" != "10051" ]; then
        server="${server}:${ZBX_SERVER_PORT}"
    fi

    if [ -n "${ZBX_SERVER_HOST}" ]; then
        if [ -n "${ZBX_PASSIVESERVERS}" ]; then
            ZBX_PASSIVESERVERS="${ZBX_SERVER_HOST},${ZBX_PASSIVESERVERS}"
        else
            ZBX_PASSIVESERVERS="${ZBX_SERVER_HOST}"
        fi

        if [ -n "${ZBX_ACTIVESERVERS}" ]; then
            ZBX_ACTIVESERVERS="${server},${ZBX_ACTIVESERVERS}"
        else
            ZBX_ACTIVESERVERS="${server}"
        fi
    fi

    : "${ZBX_PASSIVE_ALLOW:=true}"
    if [ "${ZBX_PASSIVE_ALLOW,,}" = "true" ] && [ -n "${ZBX_PASSIVESERVERS}" ]; then
        info "** Using '${ZBX_PASSIVESERVERS}' servers for passive checks"
        update_config_var "${ZBX_AGENT_CONFIG}" "Server" "${ZBX_PASSIVESERVERS}"
    else
        update_config_var "${ZBX_AGENT_CONFIG}" "Server"
    fi

    : "${ZBX_ACTIVE_ALLOW:=true}"
    if [ "${ZBX_ACTIVE_ALLOW,,}" = "true" ] && [ -n "${ZBX_ACTIVESERVERS}" ]; then
        info "** Using '${ZBX_ACTIVESERVERS}' servers for active checks"
        update_config_var "${ZBX_AGENT_CONFIG}" "ServerActive" "${ZBX_ACTIVESERVERS}"
    else
        update_config_var "${ZBX_AGENT_CONFIG}" "ServerActive"
    fi

    update_config_var "${ZBX_AGENT_CONFIG}" "PidFile"
    update_config_var "${ZBX_AGENT_CONFIG}" "LogType" "console"
    update_config_var "${ZBX_AGENT_CONFIG}" "LogFile"
    update_config_var "${ZBX_AGENT_CONFIG}" "LogFileSize"
    update_config_var "${ZBX_AGENT_CONFIG}" "DebugLevel" "${ZBX_DEBUGLEVEL:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "SourceIP"

    update_config_var "${ZBX_AGENT_CONFIG}" "ListenPort" "${ZBX_LISTENPORT:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "ListenIP" "${ZBX_LISTENIP:-}"

    update_config_var "${ZBX_AGENT_CONFIG}" "HeartbeatFrequency" "${ZBX_HEARTBEAT_FREQUENCY:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "ForceActiveChecksOnStart" "${ZBX_FORCEACTIVECHECKSONSTART:-}"

    : "${ZBX_ENABLEPERSISTENTBUFFER:=false}"
    if [ "${ZBX_ENABLEPERSISTENTBUFFER,,}" = "true" ]; then
        update_config_var "${ZBX_AGENT_CONFIG}" "EnablePersistentBuffer" "1"
        update_config_var "${ZBX_AGENT_CONFIG}" "PersistentBufferFile" "${ZABBIX_USER_HOME_DIR}/buffer/agent2.db"
        update_config_var "${ZBX_AGENT_CONFIG}" "PersistentBufferPeriod" "${ZBX_PERSISTENTBUFFERPERIOD:-}"
    else
        update_config_var "${ZBX_AGENT_CONFIG}" "EnablePersistentBuffer" "0"
        update_config_var "${ZBX_AGENT_CONFIG}" "PersistentBufferFile"
        update_config_var "${ZBX_AGENT_CONFIG}" "PersistentBufferPeriod"
    fi

    : "${ZBX_ENABLESTATUSPORT:=false}"
    if [ "${ZBX_ENABLESTATUSPORT,,}" = "true" ]; then
        update_config_var "${ZBX_AGENT_CONFIG}" "StatusPort" "31999"
    else
        update_config_var "${ZBX_AGENT_CONFIG}" "StatusPort"
    fi

    update_config_var "${ZBX_AGENT_CONFIG}" "HostInterface" "${ZBX_HOSTINTERFACE:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "HostInterfaceItem" "${ZBX_HOSTINTERFACEITEM:-}"

    update_config_var "${ZBX_AGENT_CONFIG}" "Hostname" "${ZBX_HOSTNAME:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "HostnameItem" "${ZBX_HOSTNAMEITEM:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "HostMetadata" "${ZBX_METADATA:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "HostMetadataItem" "${ZBX_METADATAITEM:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "RefreshActiveChecks" "${ZBX_REFRESHACTIVECHECKS:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "BufferSend" "${ZBX_BUFFERSEND:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "BufferSize" "${ZBX_BUFFERSIZE:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "Timeout" "${ZBX_TIMEOUT:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "Include" "${ZABBIX_CONF_DIR}/zabbix_agent2.d/plugins.d/*.conf"
    update_config_var "${ZBX_AGENT_CONFIG}" "Include" "${ZABBIX_CONF_DIR}/zabbix_agentd.d/*.conf" "true"
    update_config_var "${ZBX_AGENT_CONFIG}" "UserParameterDir" "${ZABBIX_USER_HOME_DIR}/user_scripts"
    update_config_var "${ZBX_AGENT_CONFIG}" "UnsafeUserParameters" "${ZBX_UNSAFEUSERPARAMETERS:-}"

    update_config_var "${ZBX_AGENT_CONFIG}" "TLSConnect" "${ZBX_TLSCONNECT:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "TLSAccept" "${ZBX_TLSACCEPT:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "${ZBX_AGENT_CONFIG}" "TLSCAFile" "${ZBX_TLSCAFILE:-}" "${ZBX_TLSCA:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "${ZBX_AGENT_CONFIG}" "TLSCRLFile" "${ZBX_TLSCRLFILE:-}" "${ZBX_TLSCRL:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "TLSServerCertIssuer" "${ZBX_TLSSERVERCERTISSUER:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "TLSServerCertSubject" "${ZBX_TLSSERVERCERTSUBJECT:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "${ZBX_AGENT_CONFIG}" "TLSCertFile" "${ZBX_TLSCERTFILE:-}" "${ZBX_TLSCERT:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "${ZBX_AGENT_CONFIG}" "TLSKeyFile" "${ZBX_TLSKEYFILE:-}" "${ZBX_TLSKEY:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "TLSPSKIdentity" "${ZBX_TLSPSKIDENTITY:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "${ZBX_AGENT_CONFIG}" "TLSPSKFile" "${ZBX_TLSPSKFILE:-}" "${ZBX_TLSPSK:-}"

    update_config_multiple_var "${ZBX_AGENT_CONFIG}" "DenyKey" "${ZBX_DENYKEY:-}"
    update_config_multiple_var "${ZBX_AGENT_CONFIG}" "AllowKey" "${ZBX_ALLOWKEY:-}"
}

update_plugin_config() {
    info "** Preparing Zabbix agent 2 plugin configuration files"

    local plugin_config_dir="${ZABBIX_CONF_DIR}/zabbix_agent2.d/plugins.d"
    local plugin_bin_dir="/usr/sbin/zabbix-agent2-plugin"

    update_config_var "$plugin_config_dir/mongodb.conf" "Plugins.MongoDB.System.Path" "$plugin_bin_dir/mongodb"
    update_config_var "$plugin_config_dir/postgresql.conf" "Plugins.PostgreSQL.System.Path" "$plugin_bin_dir/postgresql"
    update_config_var "$plugin_config_dir/mssql.conf" "Plugins.MSSQL.System.Path" "$plugin_bin_dir/mssql"
    update_config_var "$plugin_config_dir/ember.conf" "Plugins.EmberPlus.System.Path" "$plugin_bin_dir/ember-plus"
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

if [ "${1:-}" = "/usr/sbin/zabbix_agent2" ]; then
    prepare_service
fi

exec "$@"

#################################################
