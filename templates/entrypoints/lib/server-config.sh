source "${ENTRYPOINT_LIBS}/bootstrap.sh"
source "${ENTRYPOINT_LIBS}/openssl.sh"

# Internal directory for TLS related files, used when TLS*File specified as plain text values
readonly ZABBIX_INTERNAL_ENC_DIR="${ZABBIX_USER_HOME_DIR}/enc_internal"

server_config() {
    [[ -f "$ZBX_SERVER_CONFIG" ]] || error "Missing configuration file: $ZBX_SERVER_CONFIG"

    update_config_var "${ZBX_SERVER_CONFIG}" "ListenIP" "${ZBX_LISTENIP:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "ListenPort" "${ZBX_LISTENPORT:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "ListenBacklog" "${ZBX_LISTENBACKLOG:-}"

    update_config_var "${ZBX_SERVER_CONFIG}" "SourceIP" "${ZBX_SOURCEIP:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "LogType" "console"
    update_config_var "${ZBX_SERVER_CONFIG}" "LogFile"
    update_config_var "${ZBX_SERVER_CONFIG}" "LogFileSize"
    update_config_var "${ZBX_SERVER_CONFIG}" "PidFile"

    update_config_var "${ZBX_SERVER_CONFIG}" "DebugLevel" "${ZBX_DEBUGLEVEL:-}"

    if [ -n "${ZBX_DBTLSCONNECT:-}" ]; then
        update_config_var "${ZBX_SERVER_CONFIG}" "DBTLSConnect" "${ZBX_DBTLSCONNECT:-}"
        update_config_var "${ZBX_SERVER_CONFIG}" "DBTLSCAFile" "${ZBX_DBTLSCAFILE:-}"
        update_config_var "${ZBX_SERVER_CONFIG}" "DBTLSCertFile" "${ZBX_DBTLSCERTFILE:-}"
        update_config_var "${ZBX_SERVER_CONFIG}" "DBTLSKeyFile" "${ZBX_DBTLSKEYFILE:-}"
        update_config_var "${ZBX_SERVER_CONFIG}" "DBTLSCipher" "${ZBX_DBTLSCIPHER:-}"
        update_config_var "${ZBX_SERVER_CONFIG}" "DBTLSCipher13" "${ZBX_DBTLSCIPHER13:-}"
    else
        update_config_var "${ZBX_SERVER_CONFIG}" "DBTLSConnect"
        update_config_var "${ZBX_SERVER_CONFIG}" "DBTLSCAFile"
        update_config_var "${ZBX_SERVER_CONFIG}" "DBTLSCertFile"
        update_config_var "${ZBX_SERVER_CONFIG}" "DBTLSKeyFile"
        update_config_var "${ZBX_SERVER_CONFIG}" "DBTLSCipher"
        update_config_var "${ZBX_SERVER_CONFIG}" "DBTLSCipher13"
    fi

    update_config_var "${ZBX_SERVER_CONFIG}" "DBSocket" "${DB_SERVER_SOCKET:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "DBName" "${DB_SERVER_DBNAME:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "DBSchema" "${DB_SERVER_SCHEMA:-}"

    if [ -n "${ZBX_VAULT:-}" ] && [ -n "${ZBX_VAULTURL:-}" ]; then
        update_config_var "${ZBX_SERVER_CONFIG}" "Vault" "${ZBX_VAULT:-}"
        update_config_var "${ZBX_SERVER_CONFIG}" "VaultTLSCertFile" "${ZBX_VAULTTLSCERTFILE:-}"
        update_config_var "${ZBX_SERVER_CONFIG}" "VaultTLSKeyFile" "${ZBX_VAULTTLSKEYFILE:-}"
        update_config_var "${ZBX_SERVER_CONFIG}" "VaultPrefix" "${ZBX_VAULTPREFIX:-}"
        update_config_var "${ZBX_SERVER_CONFIG}" "VaultURL" "${ZBX_VAULTURL:-}"
        update_config_var "${ZBX_SERVER_CONFIG}" "VaultDBPath" "${ZBX_VAULTDBPATH:-}"

        if [ -n "${ZBX_VAULTDBPATH:-}" ]; then
            update_config_var "${ZBX_SERVER_CONFIG}" "DBUser"
            update_config_var "${ZBX_SERVER_CONFIG}" "DBPassword"
        else
            update_config_var "${ZBX_SERVER_CONFIG}" "DBUser" "${DB_SERVER_ZBX_USER:-}"
            update_config_var "${ZBX_SERVER_CONFIG}" "DBPassword" "${DB_SERVER_ZBX_PASS:-}"
        fi
    else
        update_config_var "${ZBX_SERVER_CONFIG}" "Vault"
        update_config_var "${ZBX_SERVER_CONFIG}" "VaultTLSCertFile"
        update_config_var "${ZBX_SERVER_CONFIG}" "VaultTLSKeyFile"
        update_config_var "${ZBX_SERVER_CONFIG}" "VaultPrefix"
        update_config_var "${ZBX_SERVER_CONFIG}" "VaultURL"
        update_config_var "${ZBX_SERVER_CONFIG}" "VaultDBPath"

        update_config_var "${ZBX_SERVER_CONFIG}" "DBUser" "${DB_SERVER_ZBX_USER:-}"
        update_config_var "${ZBX_SERVER_CONFIG}" "DBPassword" "${DB_SERVER_ZBX_PASS:-}"
    fi

    update_config_var "${ZBX_SERVER_CONFIG}" "AllowUnsupportedDBVersions" "${ZBX_ALLOWUNSUPPORTEDDBVERSIONS:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "MaxConcurrentChecksPerPoller" "${ZBX_MAXCONCURRENTCHECKSPERPOLLER:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "EnableGlobalScripts" "${ZBX_ENABLEGLOBALSCRIPTS:-}"

    update_config_var "${ZBX_SERVER_CONFIG}" "StartReportWriters" "${ZBX_STARTREPORTWRITERS:-}"
    : "${ZBX_WEBSERVICEURL:=http://zabbix-web-service:10053/report}"
    update_config_var "${ZBX_SERVER_CONFIG}" "WebServiceURL" "${ZBX_WEBSERVICEURL:-}"

    update_config_var "${ZBX_SERVER_CONFIG}" "HistoryStorageURL" "${ZBX_HISTORYSTORAGEURL:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "HistoryStorageTypes" "${ZBX_HISTORYSTORAGETYPES:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "HistoryStorageDateIndex" "${ZBX_HISTORYSTORAGEDATEINDEX:-}"

    update_config_var "${ZBX_SERVER_CONFIG}" "StatsAllowedIP" "${ZBX_STATSALLOWEDIP:-}"

    update_config_var "${ZBX_SERVER_CONFIG}" "StartPollers" "${ZBX_STARTPOLLERS:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "StartIPMIPollers" "${ZBX_STARTIPMIPOLLERS:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "StartPollersUnreachable" "${ZBX_STARTPOLLERSUNREACHABLE:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "StartTrappers" "${ZBX_STARTTRAPPERS:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "StartPingers" "${ZBX_STARTPINGERS:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "StartDiscoverers" "${ZBX_STARTDISCOVERERS:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "StartHistoryPollers" "${ZBX_STARTHISTORYPOLLERS:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "StartHTTPAgentPollers" "${ZBX_STARTHTTPAGENTPOLLERS:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "StartHTTPPollers" "${ZBX_STARTHTTPPOLLERS:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "StartODBCPollers" "${ZBX_STARTODBCPOLLERS:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "StartSNMPPollers" "${ZBX_STARTSNMPPOLLERS:-}"

    update_config_var "${ZBX_SERVER_CONFIG}" "StartConnectors" "${ZBX_STARTCONNECTORS:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "StartPreprocessors" "${ZBX_STARTPREPROCESSORS:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "StartTimers" "${ZBX_STARTTIMERS:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "StartEscalators" "${ZBX_STARTESCALATORS:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "StartAgentPollers" "${ZBX_STARTAGENTPOLLERS:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "StartAlerters" "${ZBX_STARTALERTERS:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "StartLLDProcessors" "${ZBX_STARTLLDPROCESSORS:-}"

    : "${ZBX_JAVAGATEWAY_ENABLE:=false}"
    if [ "${ZBX_JAVAGATEWAY_ENABLE,,}" = "true" ]; then
        update_config_var "${ZBX_SERVER_CONFIG}" "JavaGateway" "${ZBX_JAVAGATEWAY:-zabbix-java-gateway}"
        update_config_var "${ZBX_SERVER_CONFIG}" "JavaGatewayPort" "${ZBX_JAVAGATEWAYPORT:-}"
        update_config_var "${ZBX_SERVER_CONFIG}" "StartJavaPollers" "${ZBX_STARTJAVAPOLLERS:-5}"
    else
        update_config_var "${ZBX_SERVER_CONFIG}" "JavaGateway"
        update_config_var "${ZBX_SERVER_CONFIG}" "JavaGatewayPort"
        update_config_var "${ZBX_SERVER_CONFIG}" "StartJavaPollers"
    fi

    update_config_var "${ZBX_SERVER_CONFIG}" "StartVMwareCollectors" "${ZBX_STARTVMWARECOLLECTORS:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "VMwareFrequency" "${ZBX_VMWAREFREQUENCY:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "VMwarePerfFrequency" "${ZBX_VMWAREPERFFREQUENCY:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "VMwareCacheSize" "${ZBX_VMWARECACHESIZE:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "VMwareTimeout" "${ZBX_VMWARETIMEOUT:-}"

    : "${ZBX_ENABLE_SNMP_TRAPS:=false}"
    if [ "${ZBX_ENABLE_SNMP_TRAPS,,}" = "true" ]; then
        update_config_var "${ZBX_SERVER_CONFIG}" "SNMPTrapperFile" "${ZABBIX_USER_HOME_DIR}/snmptraps/snmptraps.log"
        update_config_var "${ZBX_SERVER_CONFIG}" "StartSNMPTrapper" "1"
    else
        update_config_var "${ZBX_SERVER_CONFIG}" "SNMPTrapperFile"
        update_config_var "${ZBX_SERVER_CONFIG}" "StartSNMPTrapper"
    fi

    update_config_var "${ZBX_SERVER_CONFIG}" "SocketDir" "/tmp/"

    update_config_var "${ZBX_SERVER_CONFIG}" "HousekeepingFrequency" "${ZBX_HOUSEKEEPINGFREQUENCY:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "MaxHousekeeperDelete" "${ZBX_MAXHOUSEKEEPERDELETE:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "ProblemHousekeepingFrequency" "${ZBX_PROBLEMHOUSEKEEPINGFREQUENCY:-}"

    update_config_var "${ZBX_SERVER_CONFIG}" "CacheSize" "${ZBX_CACHESIZE:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "CacheUpdateFrequency" "${ZBX_CACHEUPDATEFREQUENCY:-}"

    update_config_var "${ZBX_SERVER_CONFIG}" "StartDBSyncers" "${ZBX_STARTDBSYNCERS:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "HistoryCacheSize" "${ZBX_HISTORYCACHESIZE:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "HistoryIndexCacheSize" "${ZBX_HISTORYINDEXCACHESIZE:-}"

    update_config_var "${ZBX_SERVER_CONFIG}" "TrendCacheSize" "${ZBX_TRENDCACHESIZE:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "TrendFunctionCacheSize" "${ZBX_TRENDFUNCTIONCACHESIZE:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "ValueCacheSize" "${ZBX_VALUECACHESIZE:-}"

    update_config_var "${ZBX_SERVER_CONFIG}" "Timeout" "${ZBX_TIMEOUT:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "TrapperTimeout" "${ZBX_TRAPPERTIMEOUT:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "UnreachablePeriod" "${ZBX_UNREACHABLEPERIOD:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "UnavailableDelay" "${ZBX_UNAVAILABLEDELAY:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "UnreachableDelay" "${ZBX_UNREACHABLEDELAY:-}"

    update_config_var "${ZBX_SERVER_CONFIG}" "AlertScriptsPath" "/usr/lib/zabbix/alertscripts"
    update_config_var "${ZBX_SERVER_CONFIG}" "ExternalScripts" "/usr/lib/zabbix/externalscripts"

    if [ -n "${ZBX_EXPORTFILESIZE:-}" ]; then
        update_config_var "${ZBX_SERVER_CONFIG}" "ExportDir" "${ZABBIX_USER_HOME_DIR}/export/"
        update_config_var "${ZBX_SERVER_CONFIG}" "ExportFileSize" "${ZBX_EXPORTFILESIZE:-}"
        update_config_var "${ZBX_SERVER_CONFIG}" "ExportType" "${ZBX_EXPORTTYPE:-}"
    else
        update_config_var "${ZBX_SERVER_CONFIG}" "ExportDir"
        update_config_var "${ZBX_SERVER_CONFIG}" "ExportFileSize"
        update_config_var "${ZBX_SERVER_CONFIG}" "ExportType"
    fi

    update_config_var "${ZBX_SERVER_CONFIG}" "FpingLocation" "${ZBX_FPINGLOCATION:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "Fping6Location" '""'

    update_config_var "${ZBX_SERVER_CONFIG}" "SSHKeyLocation" "${ZABBIX_USER_HOME_DIR}/ssh_keys"
    update_config_var "${ZBX_SERVER_CONFIG}" "LogSlowQueries" "${ZBX_LOGSLOWQUERIES:-}"

    update_config_var "${ZBX_SERVER_CONFIG}" "StartProxyPollers" "${ZBX_STARTPROXYPOLLERS:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "ProxyConfigFrequency" "${ZBX_PROXYCONFIGFREQUENCY:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "ProxyDataFrequency" "${ZBX_PROXYDATAFREQUENCY:-}"

    update_config_var "${ZBX_SERVER_CONFIG}" "SSLCertLocation" "${ZABBIX_USER_HOME_DIR}/ssl/certs/"
    update_config_var "${ZBX_SERVER_CONFIG}" "SSLKeyLocation" "${ZABBIX_USER_HOME_DIR}/ssl/keys/"
    update_config_var "${ZBX_SERVER_CONFIG}" "SSLCALocation" "${ZABBIX_USER_HOME_DIR}/ssl/ssl_ca/"
    update_config_var "${ZBX_SERVER_CONFIG}" "LoadModulePath" "${ZABBIX_USER_HOME_DIR}/modules/"
    update_config_multiple_var "${ZBX_SERVER_CONFIG}" "LoadModule" "${ZBX_LOADMODULE:-}"

    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "${ZBX_SERVER_CONFIG}" "TLSCAFile" "${ZBX_TLSCAFILE:-}" "${ZBX_TLSCA:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "${ZBX_SERVER_CONFIG}" "TLSCRLFile" "${ZBX_TLSCRLFILE:-}" "${ZBX_TLSCRL:-}"

    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "${ZBX_SERVER_CONFIG}" "TLSCertFile" "${ZBX_TLSCERTFILE:-}" "${ZBX_TLSCERT:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "TLSCipherAll" "${ZBX_TLSCIPHERALL:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "TLSCipherAll13" "${ZBX_TLSCIPHERALL13:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "TLSCipherCert" "${ZBX_TLSCIPHERCERT:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "TLSCipherCert13" "${ZBX_TLSCIPHERCERT13:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "TLSCipherPSK" "${ZBX_TLSCIPHERPSK:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "TLSCipherPSK13" "${ZBX_TLSCIPHERPSK13:-}"
    file_process_from_env "${ZABBIX_INTERNAL_ENC_DIR}" "${ZBX_SERVER_CONFIG}" "TLSKeyFile" "${ZBX_TLSKEYFILE:-}" "${ZBX_TLSKEY:-}"

    update_config_var "${ZBX_SERVER_CONFIG}" "ServiceManagerSyncFrequency" "${ZBX_SERVICEMANAGERSYNCFREQUENCY:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "AllowSoftwareUpdateCheck" "${ZBX_ALLOWSOFTWAREUPDATECHECK:-}"

    update_config_var "${ZBX_SERVER_CONFIG}" "SMSDevices" "${ZBX_SMSDEVICES:-}"

    if [ "${ZBX_AUTOHANODENAME:-}" = "fqdn" ] && [ -z "${ZBX_HANODENAME:-}" ]; then
        update_config_var "${ZBX_SERVER_CONFIG}" "HANodeName" "$(hostname -f)"
    elif [ "${ZBX_AUTOHANODENAME:-}" = "hostname" ] && [ -z "${ZBX_HANODENAME:-}" ]; then
        update_config_var "${ZBX_SERVER_CONFIG}" "HANodeName" "$(hostname)"
    else
        update_config_var "${ZBX_SERVER_CONFIG}" "HANodeName" "${ZBX_HANODENAME:-}"
    fi

    : "${ZBX_NODEADDRESSPORT:=10051}"
    if [ "${ZBX_AUTONODEADDRESS:-}" = "fqdn" ] && [ -z "${ZBX_NODEADDRESS:-}" ]; then
        update_config_var "${ZBX_SERVER_CONFIG}" "NodeAddress" "$(hostname -f):${ZBX_NODEADDRESSPORT}"
    elif [ "${ZBX_AUTONODEADDRESS:-}" = "hostname" ] && [ -z "${ZBX_NODEADDRESS:-}" ]; then
        update_config_var "${ZBX_SERVER_CONFIG}" "NodeAddress" "$(hostname):${ZBX_NODEADDRESSPORT}"
    else
        update_config_var "${ZBX_SERVER_CONFIG}" "NodeAddress" "${ZBX_NODEADDRESS:-}"
    fi

    update_config_var "${ZBX_SERVER_CONFIG}" "WebDriverURL" "${ZBX_WEBDRIVERURL:-}"
    update_config_var "${ZBX_SERVER_CONFIG}" "StartBrowserPollers" "${ZBX_STARTBROWSERPOLLERS:-}"

    if [ "$(id -u)" -ne 0 ]; then
        export ZBX_USER="$(id -un)"
    else
        export ZBX_ALLOWROOT=1
    fi

    openssl_rehash "${ZABBIX_USER_HOME_DIR}/ssl/ssl_ca/"
}
