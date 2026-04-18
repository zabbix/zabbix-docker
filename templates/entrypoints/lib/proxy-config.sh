# shellcheck shell=bash

source "${ENTRYPOINT_LIBS}/bootstrap.sh"
source "${ENTRYPOINT_LIBS}/openssl.sh"

# Default Zabbix server host
: ${ZBX_SERVER_HOST:="zabbix-server"}

# Internal directory for TLS related files, used when TLS*File specified as plain text values
readonly ZABBIX_INTERNAL_ENC_DIR="${ZABBIX_USER_HOME_DIR}/enc_internal"

proxy_config() {
    local default_hostname="${1:-}"

    [[ -f "$ZBX_PROXY_CONFIG" ]] || error "Missing configuration file: $ZBX_PROXY_CONFIG"

    update_config_var "${ZBX_PROXY_CONFIG}" "ProxyMode" "${ZBX_PROXYMODE:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "Server" "${ZBX_SERVER_HOST}"

    if [ -z "${ZBX_HOSTNAME:-}" ] && [ -n "${ZBX_HOSTNAMEITEM:-}" ]; then
        update_config_var "${ZBX_PROXY_CONFIG}" "Hostname" ""
        update_config_var "${ZBX_PROXY_CONFIG}" "HostnameItem" "${ZBX_HOSTNAMEITEM:-}"
    else
        update_config_var "${ZBX_PROXY_CONFIG}" "Hostname" "${ZBX_HOSTNAME:-$default_hostname}"
        update_config_var "${ZBX_PROXY_CONFIG}" "HostnameItem" "${ZBX_HOSTNAMEITEM:-}"
    fi

    update_config_var "${ZBX_PROXY_CONFIG}" "ListenIP" "${ZBX_LISTENIP:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "ListenPort" "${ZBX_LISTENPORT:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "ListenBacklog" "${ZBX_LISTENBACKLOG:-}"

    update_config_var "${ZBX_PROXY_CONFIG}" "SourceIP" "${ZBX_SOURCEIP:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "LogType" "console"
    update_config_var "${ZBX_PROXY_CONFIG}" "LogFile"
    update_config_var "${ZBX_PROXY_CONFIG}" "LogFileSize"
    update_config_var "${ZBX_PROXY_CONFIG}" "PidFile"

    update_config_var "${ZBX_PROXY_CONFIG}" "DebugLevel" "${ZBX_DEBUGLEVEL:-}"

    update_config_var "${ZBX_PROXY_CONFIG}" "EnableRemoteCommands" "${ZBX_ENABLEREMOTECOMMANDS:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "LogRemoteCommands" "${ZBX_LOGREMOTECOMMANDS:-}"

    update_config_var "${ZBX_PROXY_CONFIG}" "ProxyLocalBuffer" "${ZBX_PROXYLOCALBUFFER:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "ProxyOfflineBuffer" "${ZBX_PROXYOFFLINEBUFFER:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "ProxyConfigFrequency" "${ZBX_PROXYCONFIGFREQUENCY:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "DataSenderFrequency" "${ZBX_DATASENDERFREQUENCY:-}"

    update_config_var "${ZBX_PROXY_CONFIG}" "StatsAllowedIP" "${ZBX_STATSALLOWEDIP:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "StartPreprocessors" "${ZBX_STARTPREPROCESSORS:-}"

    update_config_var "${ZBX_PROXY_CONFIG}" "StartPollers" "${ZBX_STARTPOLLERS:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "StartIPMIPollers" "${ZBX_STARTIPMIPOLLERS:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "StartPollersUnreachable" "${ZBX_STARTPOLLERSUNREACHABLE:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "StartTrappers" "${ZBX_STARTTRAPPERS:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "StartPingers" "${ZBX_STARTPINGERS:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "StartDiscoverers" "${ZBX_STARTDISCOVERERS:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "StartHistoryPollers" "${ZBX_STARTHISTORYPOLLERS:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "StartHTTPPollers" "${ZBX_STARTHTTPPOLLERS:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "StartODBCPollers" "${ZBX_STARTODBCPOLLERS:-}"

    : "${ZBX_JAVAGATEWAY_ENABLE:=false}"
    if [ "${ZBX_JAVAGATEWAY_ENABLE,,}" = "true" ]; then
        update_config_var "${ZBX_PROXY_CONFIG}" "JavaGateway" "${ZBX_JAVAGATEWAY:-zabbix-java-gateway}"
        update_config_var "${ZBX_PROXY_CONFIG}" "JavaGatewayPort" "${ZBX_JAVAGATEWAYPORT:-}"
        update_config_var "${ZBX_PROXY_CONFIG}" "StartJavaPollers" "${ZBX_STARTJAVAPOLLERS:-5}"
    else
        update_config_var "${ZBX_PROXY_CONFIG}" "JavaGateway"
        update_config_var "${ZBX_PROXY_CONFIG}" "JavaGatewayPort"
        update_config_var "${ZBX_PROXY_CONFIG}" "StartJavaPollers"
    fi

    update_config_var "${ZBX_PROXY_CONFIG}" "StartVMwareCollectors" "${ZBX_STARTVMWARECOLLECTORS:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "VMwareFrequency" "${ZBX_VMWAREFREQUENCY:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "VMwarePerfFrequency" "${ZBX_VMWAREPERFFREQUENCY:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "VMwareCacheSize" "${ZBX_VMWARECACHESIZE:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "VMwareTimeout" "${ZBX_VMWARETIMEOUT:-}"

    : "${ZBX_ENABLE_SNMP_TRAPS:=false}"
    if [ "${ZBX_ENABLE_SNMP_TRAPS,,}" = "true" ]; then
        update_config_var "${ZBX_PROXY_CONFIG}" "SNMPTrapperFile" "${ZABBIX_USER_HOME_DIR}/snmptraps/snmptraps.log"
        update_config_var "${ZBX_PROXY_CONFIG}" "StartSNMPTrapper" "1"
    else
        update_config_var "${ZBX_PROXY_CONFIG}" "SNMPTrapperFile"
        update_config_var "${ZBX_PROXY_CONFIG}" "StartSNMPTrapper"
    fi

    update_config_var "${ZBX_PROXY_CONFIG}" "HousekeepingFrequency" "${ZBX_HOUSEKEEPINGFREQUENCY:-}"

    update_config_var "${ZBX_PROXY_CONFIG}" "CacheSize" "${ZBX_CACHESIZE:-}"

    update_config_var "${ZBX_PROXY_CONFIG}" "StartDBSyncers" "${ZBX_STARTDBSYNCERS:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "HistoryCacheSize" "${ZBX_HISTORYCACHESIZE:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "HistoryIndexCacheSize" "${ZBX_HISTORYINDEXCACHESIZE:-}"

    update_config_var "${ZBX_PROXY_CONFIG}" "Timeout" "${ZBX_TIMEOUT:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "TrapperTimeout" "${ZBX_TRAPPERTIMEOUT:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "UnreachablePeriod" "${ZBX_UNREACHABLEPERIOD:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "UnavailableDelay" "${ZBX_UNAVAILABLEDELAY:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "UnreachableDelay" "${ZBX_UNREACHABLEDELAY:-}"

    update_config_var "${ZBX_PROXY_CONFIG}" "ExternalScripts" "/usr/lib/zabbix/externalscripts"

    update_config_var "${ZBX_PROXY_CONFIG}" "FpingLocation" "${ZBX_FPINGLOCATION:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "Fping6Location" '""'

    update_config_var "${ZBX_PROXY_CONFIG}" "SSHKeyLocation" "${ZABBIX_USER_HOME_DIR}/ssh_keys"
    update_config_var "${ZBX_PROXY_CONFIG}" "LogSlowQueries" "${ZBX_LOGSLOWQUERIES:-}"

    update_config_var "${ZBX_PROXY_CONFIG}" "SSLCertLocation" "${ZABBIX_USER_HOME_DIR}/ssl/certs/"
    update_config_var "${ZBX_PROXY_CONFIG}" "SSLKeyLocation" "${ZABBIX_USER_HOME_DIR}/ssl/keys/"
    update_config_var "${ZBX_PROXY_CONFIG}" "SSLCALocation" "${ZABBIX_USER_HOME_DIR}/ssl/ssl_ca/"
    update_config_var "${ZBX_PROXY_CONFIG}" "LoadModulePath" "${ZABBIX_USER_HOME_DIR}/modules/"
    update_config_multiple_var "${ZBX_PROXY_CONFIG}" "LoadModule" "${ZBX_LOADMODULE:-}"

    update_config_var "${ZBX_PROXY_CONFIG}" "TLSConnect" "${ZBX_TLSCONNECT:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "TLSAccept" "${ZBX_TLSACCEPT:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "${ZBX_PROXY_CONFIG}" "TLSCAFile" "${ZBX_TLSCAFILE:-}" "${ZBX_TLSCA:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "${ZBX_PROXY_CONFIG}" "TLSCRLFile" "${ZBX_TLSCRLFILE:-}" "${ZBX_TLSCRL:-}"

    update_config_var "${ZBX_PROXY_CONFIG}" "TLSServerCertIssuer" "${ZBX_TLSSERVERCERTISSUER:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "TLSServerCertSubject" "${ZBX_TLSSERVERCERTSUBJECT:-}"

    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "${ZBX_PROXY_CONFIG}" "TLSCertFile" "${ZBX_TLSCERTFILE:-}" "${ZBX_TLSCERT:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "TLSCipherAll" "${ZBX_TLSCIPHERALL:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "TLSCipherAll13" "${ZBX_TLSCIPHERALL13:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "TLSCipherCert" "${ZBX_TLSCIPHERCERT:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "TLSCipherCert13" "${ZBX_TLSCIPHERCERT13:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "TLSCipherPSK" "${ZBX_TLSCIPHERPSK:-}"
    update_config_var "${ZBX_PROXY_CONFIG}" "TLSCipherPSK13" "${ZBX_TLSCIPHERPSK13:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "${ZBX_PROXY_CONFIG}" "TLSKeyFile" "${ZBX_TLSKEYFILE:-}" "${ZBX_TLSKEY:-}"

    update_config_var "${ZBX_PROXY_CONFIG}" "TLSPSKIdentity" "${ZBX_TLSPSKIDENTITY:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "${ZBX_PROXY_CONFIG}" "TLSPSKFile" "${ZBX_TLSPSKFILE:-}" "${ZBX_TLSPSK:-}"

    if [ "$(id -u)" -ne 0 ]; then
        update_config_var "${ZBX_PROXY_CONFIG}" "User" "$(id -un)"
    else
        update_config_var "${ZBX_PROXY_CONFIG}" "AllowRoot" "1"
    fi

    openssl_rehash "${ZABBIX_USER_HOME_DIR}/ssl/ssl_ca/"
}
