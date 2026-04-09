#!/usr/bin/env bash

set -euo pipefail

readonly ENTRYPOINT_LIBS="/usr/lib/docker-entrypoint"
source "${ENTRYPOINT_LIBS}/bootstrap.sh"

# Default Zabbix server host
: "${ZBX_SERVER_HOST=zabbix-server}"
# Default Zabbix server port number
: "${ZBX_SERVER_PORT=10051}"

readonly ZBX_AGENT_CONFIG=$ZABBIX_CONF_DIR/zabbix_agentd.conf

readonly ZABBIX_INTERNAL_ENC_DIR="${ZABBIX_USER_HOME_DIR}/enc_internal"

update_config() {
    : "${ZBX_PASSIVESERVERS:=}"
    : "${ZBX_ACTIVESERVERS:=}"

    [[ -f "$ZBX_AGENT_CONFIG" ]] || error "Missing configuration file: $ZBX_AGENT_CONFIG"

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
    update_config_var "${ZBX_AGENT_CONFIG}" "LogRemoteCommands" "${ZBX_LOGREMOTECOMMANDS:-}"

    update_config_var "${ZBX_AGENT_CONFIG}" "ListenPort" "${ZBX_LISTENPORT:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "ListenIP" "${ZBX_LISTENIP:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "ListenBacklog" "${ZBX_LISTENBACKLOG:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "StartAgents" "${ZBX_STARTAGENTS:-}"

    update_config_var "${ZBX_AGENT_CONFIG}" "HeartbeatFrequency" "${ZBX_HEARTBEAT_FREQUENCY:-}"

    update_config_var "${ZBX_AGENT_CONFIG}" "HostInterface" "${ZBX_HOSTINTERFACE:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "HostInterfaceItem" "${ZBX_HOSTINTERFACEITEM:-}"

    update_config_var "${ZBX_AGENT_CONFIG}" "Hostname" "${ZBX_HOSTNAME:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "HostnameItem" "${ZBX_HOSTNAMEITEM:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "HostMetadata" "${ZBX_METADATA:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "HostMetadataItem" "${ZBX_METADATAITEM:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "RefreshActiveChecks" "${ZBX_REFRESHACTIVECHECKS:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "BufferSend" "${ZBX_BUFFERSEND:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "BufferSize" "${ZBX_BUFFERSIZE:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "MaxLinesPerSecond" "${ZBX_MAXLINESPERSECOND:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "Timeout" "${ZBX_TIMEOUT:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "Include" "${ZABBIX_CONF_DIR}/zabbix_agentd.d/*.conf"
    update_config_var "${ZBX_AGENT_CONFIG}" "UserParameterDir" "${ZABBIX_USER_HOME_DIR}/user_scripts"
    update_config_var "${ZBX_AGENT_CONFIG}" "UnsafeUserParameters" "${ZBX_UNSAFEUSERPARAMETERS:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "LoadModulePath" "${ZABBIX_USER_HOME_DIR}/modules/"
    update_config_multiple_var "${ZBX_AGENT_CONFIG}" "LoadModule" "${ZBX_LOADMODULE:-}"

    update_config_var "${ZBX_AGENT_CONFIG}" "TLSConnect" "${ZBX_TLSCONNECT:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "TLSAccept" "${ZBX_TLSACCEPT:-}"

    file_process_from_env "${ZBX_AGENT_CONFIG}" "${ZABBIX_INTERNAL_ENC_DIR}" "TLSCAFile" "${ZBX_TLSCAFILE:-}" "${ZBX_TLSCA:-}"
    file_process_from_env "${ZBX_AGENT_CONFIG}" "${ZABBIX_INTERNAL_ENC_DIR}" "TLSCRLFile" "${ZBX_TLSCRLFILE:-}" "${ZBX_TLSCRL:-}"

    update_config_var "${ZBX_AGENT_CONFIG}" "TLSServerCertIssuer" "${ZBX_TLSSERVERCERTISSUER:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "TLSServerCertSubject" "${ZBX_TLSSERVERCERTSUBJECT:-}"

    file_process_from_env "${ZBX_AGENT_CONFIG}" "${ZABBIX_INTERNAL_ENC_DIR}" "TLSCertFile" "${ZBX_TLSCERTFILE:-}" "${ZBX_TLSCERT:-}"

    update_config_var "${ZBX_AGENT_CONFIG}" "TLSCipherAll" "${ZBX_TLSCIPHERALL:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "TLSCipherAll13" "${ZBX_TLSCIPHERALL13:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "TLSCipherCert" "${ZBX_TLSCIPHERCERT:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "TLSCipherCert13" "${ZBX_TLSCIPHERCERT13:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "TLSCipherPSK" "${ZBX_TLSCIPHERPSK:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "TLSCipherPSK13" "${ZBX_TLSCIPHERPSK13:-}"

    file_process_from_env "${ZBX_AGENT_CONFIG}" "${ZABBIX_INTERNAL_ENC_DIR}" "TLSKeyFile" "${ZBX_TLSKEYFILE:-}" "${ZBX_TLSKEY:-}"
    update_config_var "${ZBX_AGENT_CONFIG}" "TLSPSKIdentity" "${ZBX_TLSPSKIDENTITY:-}"
    file_process_from_env "${ZBX_AGENT_CONFIG}" "${ZABBIX_INTERNAL_ENC_DIR}" "TLSPSKFile" "${ZBX_TLSPSKFILE:-}" "${ZBX_TLSPSK:-}"

    update_config_multiple_var "${ZBX_AGENT_CONFIG}" "DenyKey" "${ZBX_DENYKEY:-}"
    update_config_multiple_var "${ZBX_AGENT_CONFIG}" "AllowKey" "${ZBX_ALLOWKEY:-}"

    if [ "$(id -u)" -ne 0 ]; then
        update_config_var "${ZBX_AGENT_CONFIG}" "User" "$(id -un)"
    else
        update_config_var "${ZBX_AGENT_CONFIG}" "AllowRoot" "1"
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

if [ "${1:-}" = "/usr/sbin/zabbix_agentd" ]; then
    prepare_service
fi

exec "$@"

#################################################
