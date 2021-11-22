#!/bin/bash

set -o pipefail

set +e

# Script trace mode
if [ "${DEBUG_MODE,,}" == "true" ]; then
    set -o xtrace
fi

# Default Zabbix server host
: ${ZBX_SERVER_HOST:="zabbix-server"}
# Default Zabbix server port number
: ${ZBX_SERVER_PORT:="10051"}

# Default directories
# User 'zabbix' home directory
ZABBIX_USER_HOME_DIR="/var/lib/zabbix"
# Configuration files directory
ZABBIX_ETC_DIR="/etc/zabbix"

escape_spec_char() {
    local var_value=$1

    var_value="${var_value//\\/\\\\}"
    var_value="${var_value//[$'\n']/}"
    var_value="${var_value//\//\\/}"
    var_value="${var_value//./\\.}"
    var_value="${var_value//\*/\\*}"
    var_value="${var_value//^/\\^}"
    var_value="${var_value//\$/\\\$}"
    var_value="${var_value//\&/\\\&}"
    var_value="${var_value//\[/\\[}"
    var_value="${var_value//\]/\\]}"

    echo "$var_value"
}

update_config_var() {
    local config_path=$1
    local var_name=$2
    local var_value=$3
    local is_multiple=$4

    local masklist=("TLSPSKIdentity")

    if [ ! -f "$config_path" ]; then
        echo "**** Configuration file '$config_path' does not exist"
        return
    fi

    if [[ " ${masklist[@]} " =~ " $var_name " ]] && [ ! -z "$var_value" ]; then
        echo -n "** Updating '$config_path' parameter \"$var_name\": '****'. Enable DEBUG_MODE to view value ..."
    else
        echo -n "** Updating '$config_path' parameter \"$var_name\": '$var_value'..."
    fi

    # Remove configuration parameter definition in case of unset parameter value
    if [ -z "$var_value" ]; then
        sed -i -e "/^$var_name=/d" "$config_path"
        echo "removed"
        return
    fi

    # Remove value from configuration parameter in case of double quoted parameter value
    if [ "$var_value" == '""' ]; then
        sed -i -e "/^$var_name=/s/=.*/=/" "$config_path"
        echo "undefined"
        return
    fi

    # Use full path to a file for TLS related configuration parameters
    if [[ $var_name =~ ^TLS.*File$ ]] && [[ ! $var_value =~ ^/.+$ ]]; then
        var_value=$ZABBIX_USER_HOME_DIR/enc/$var_value
    fi

    # Escaping characters in parameter value and name
    var_value=$(escape_spec_char "$var_value")
    var_name=$(escape_spec_char "$var_name")

    if [ "$(grep -E "^$var_name=" $config_path)" ] && [ "$is_multiple" != "true" ]; then
        sed -i -e "/^$var_name=/s/=.*/=$var_value/" "$config_path"
        echo "updated"
    elif [ "$(grep -Ec "^# $var_name=" $config_path)" -gt 1 ]; then
        sed -i -e  "/^[#;] $var_name=$/i\\$var_name=$var_value" "$config_path"
        echo "added first occurrence"
    else
        sed -i -e "/^[#;] $var_name=/s/.*/&\n$var_name=$var_value/" "$config_path"
        echo "added"
    fi

}

update_config_multiple_var() {
    local config_path=$1
    local var_name=$2
    local var_value=$3

    var_value="${var_value%\"}"
    var_value="${var_value#\"}"

    local IFS=,
    local OPT_LIST=($var_value)

    for value in "${OPT_LIST[@]}"; do
        update_config_var $config_path $var_name $value true
    done
}

update_zbx_config() {
    echo "** Preparing Zabbix proxy configuration file"

    ZBX_CONFIG=$ZABBIX_ETC_DIR/zabbix_proxy.conf

    update_config_var $ZBX_CONFIG "ProxyMode" "${ZBX_PROXYMODE}"
    update_config_var $ZBX_CONFIG "Server" "${ZBX_SERVER_HOST}"
    update_config_var $ZBX_CONFIG "ServerPort" "${ZBX_SERVER_PORT}"
    if [ -z "${ZBX_HOSTNAME}" ] && [ -n "${ZBX_HOSTNAMEITEM}" ]; then
        update_config_var $ZBX_CONFIG "Hostname" ""
        update_config_var $ZBX_CONFIG "HostnameItem" "${ZBX_HOSTNAMEITEM}"
    else
        update_config_var $ZBX_CONFIG "Hostname" "${ZBX_HOSTNAME:-"zabbix-proxy-sqlite3"}"
        update_config_var $ZBX_CONFIG "HostnameItem" "${ZBX_HOSTNAMEITEM}"
    fi

    update_config_var $ZBX_CONFIG "ListenIP" "${ZBX_LISTENIP}"
    update_config_var $ZBX_CONFIG "ListenPort" "${ZBX_LISTENPORT}"
    update_config_var $ZBX_CONFIG "ListenBacklog" "${ZBX_LISTENBACKLOG}"

    update_config_var $ZBX_CONFIG "SourceIP" "${ZBX_SOURCEIP}"
    update_config_var $ZBX_CONFIG "LogType" "console"
    update_config_var $ZBX_CONFIG "LogFile"
    update_config_var $ZBX_CONFIG "LogFileSize"
    update_config_var $ZBX_CONFIG "PidFile"

    update_config_var $ZBX_CONFIG "DebugLevel" "${ZBX_DEBUGLEVEL}"

    update_config_var $ZBX_CONFIG "EnableRemoteCommands" "${ZBX_ENABLEREMOTECOMMANDS}"
    update_config_var $ZBX_CONFIG "LogRemoteCommands" "${ZBX_LOGREMOTECOMMANDS}"

    update_config_var $ZBX_CONFIG "DBHost"
    update_config_var $ZBX_CONFIG "DBName" "/var/lib/zabbix/db_data/${ZBX_HOSTNAME:-"zabbix-proxy-sqlite3"}.sqlite"
    update_config_var $ZBX_CONFIG "DBUser"
    update_config_var $ZBX_CONFIG "DBPort"
    update_config_var $ZBX_CONFIG "DBPassword"

    if [ -n "${VAULT_TOKEN}" ] && [ -n "${ZBX_VAULTURL}" ]; then
        update_config_var $ZBX_CONFIG "VaultDBPath" "${ZBX_VAULTDBPATH}"
        update_config_var $ZBX_CONFIG "VaultURL" "${ZBX_VAULTURL}"
    else
        update_config_var $ZBX_CONFIG "VaultDBPath"
        update_config_var $ZBX_CONFIG "VaultURL"
    fi

    update_config_var $ZBX_CONFIG "ProxyLocalBuffer" "${ZBX_PROXYLOCALBUFFER}"
    update_config_var $ZBX_CONFIG "ProxyOfflineBuffer" "${ZBX_PROXYOFFLINEBUFFER}"
    update_config_var $ZBX_CONFIG "HeartbeatFrequency" "${ZBX_PROXYHEARTBEATFREQUENCY}"
    update_config_var $ZBX_CONFIG "ConfigFrequency" "${ZBX_CONFIGFREQUENCY}"
    update_config_var $ZBX_CONFIG "DataSenderFrequency" "${ZBX_DATASENDERFREQUENCY}"

    update_config_var $ZBX_CONFIG "StatsAllowedIP" "${ZBX_STATSALLOWEDIP}"
    update_config_var $ZBX_CONFIG "StartPreprocessors" "${ZBX_STARTPREPROCESSORS}"

    update_config_var $ZBX_CONFIG "StartPollers" "${ZBX_STARTPOLLERS}"
    update_config_var $ZBX_CONFIG "StartIPMIPollers" "${ZBX_IPMIPOLLERS}"
    update_config_var $ZBX_CONFIG "StartPollersUnreachable" "${ZBX_STARTPOLLERSUNREACHABLE}"
    update_config_var $ZBX_CONFIG "StartTrappers" "${ZBX_STARTTRAPPERS}"
    update_config_var $ZBX_CONFIG "StartPingers" "${ZBX_STARTPINGERS}"
    update_config_var $ZBX_CONFIG "StartDiscoverers" "${ZBX_STARTDISCOVERERS}"
    update_config_var $ZBX_CONFIG "StartHistoryPollers" "${ZBX_STARTHISTORYPOLLERS}"
    update_config_var $ZBX_CONFIG "StartHTTPPollers" "${ZBX_STARTHTTPPOLLERS}"

    : ${ZBX_JAVAGATEWAY_ENABLE:="false"}
    if [ "${ZBX_JAVAGATEWAY_ENABLE,,}" == "true" ]; then
        update_config_var $ZBX_CONFIG "JavaGateway" "${ZBX_JAVAGATEWAY:-"zabbix-java-gateway"}"
        update_config_var $ZBX_CONFIG "JavaGatewayPort" "${ZBX_JAVAGATEWAYPORT}"
        update_config_var $ZBX_CONFIG "StartJavaPollers" "${ZBX_STARTJAVAPOLLERS:-"5"}"
    else
        update_config_var $ZBX_CONFIG "JavaGateway"
        update_config_var $ZBX_CONFIG "JavaGatewayPort"
        update_config_var $ZBX_CONFIG "StartJavaPollers"
    fi

    update_config_var $ZBX_CONFIG "StartVMwareCollectors" "${ZBX_STARTVMWARECOLLECTORS}"
    update_config_var $ZBX_CONFIG "VMwareFrequency" "${ZBX_VMWAREFREQUENCY}"
    update_config_var $ZBX_CONFIG "VMwarePerfFrequency" "${ZBX_VMWAREPERFFREQUENCY}"
    update_config_var $ZBX_CONFIG "VMwareCacheSize" "${ZBX_VMWARECACHESIZE}"
    update_config_var $ZBX_CONFIG "VMwareTimeout" "${ZBX_VMWARETIMEOUT}"

    : ${ZBX_ENABLE_SNMP_TRAPS:="false"}
    if [ "${ZBX_ENABLE_SNMP_TRAPS,,}" == "true" ]; then
        update_config_var $ZBX_CONFIG "SNMPTrapperFile" "${ZABBIX_USER_HOME_DIR}/snmptraps/snmptraps.log"
        update_config_var $ZBX_CONFIG "StartSNMPTrapper" "1"
    else
        update_config_var $ZBX_CONFIG "SNMPTrapperFile"
        update_config_var $ZBX_CONFIG "StartSNMPTrapper"
    fi

    update_config_var $ZBX_CONFIG "HousekeepingFrequency" "${ZBX_HOUSEKEEPINGFREQUENCY}"

    update_config_var $ZBX_CONFIG "CacheSize" "${ZBX_CACHESIZE}"

    update_config_var $ZBX_CONFIG "StartDBSyncers" "${ZBX_STARTDBSYNCERS}"
    update_config_var $ZBX_CONFIG "HistoryCacheSize" "${ZBX_HISTORYCACHESIZE}"
    update_config_var $ZBX_CONFIG "HistoryIndexCacheSize" "${ZBX_HISTORYINDEXCACHESIZE}"

    update_config_var $ZBX_CONFIG "Timeout" "${ZBX_TIMEOUT}"
    update_config_var $ZBX_CONFIG "TrapperTimeout" "${ZBX_TRAPPERTIMEOUT}"
    update_config_var $ZBX_CONFIG "UnreachablePeriod" "${ZBX_UNREACHABLEPERIOD}"
    update_config_var $ZBX_CONFIG "UnavailableDelay" "${ZBX_UNAVAILABLEDELAY}"
    update_config_var $ZBX_CONFIG "UnreachableDelay" "${ZBX_UNREACHABLEDELAY}"

    update_config_var $ZBX_CONFIG "AlertScriptsPath" "/usr/lib/zabbix/alertscripts"
    update_config_var $ZBX_CONFIG "ExternalScripts" "/usr/lib/zabbix/externalscripts"

    update_config_var $ZBX_CONFIG "FpingLocation" "/usr/bin/fping"
    update_config_var $ZBX_CONFIG "Fping6Location" "/usr/bin/fping6"

    update_config_var $ZBX_CONFIG "SSHKeyLocation" "$ZABBIX_USER_HOME_DIR/ssh_keys"
    update_config_var $ZBX_CONFIG "LogSlowQueries" "${ZBX_LOGSLOWQUERIES}"

    update_config_var $ZBX_CONFIG "SSLCertLocation" "$ZABBIX_USER_HOME_DIR/ssl/certs/"
    update_config_var $ZBX_CONFIG "SSLKeyLocation" "$ZABBIX_USER_HOME_DIR/ssl/keys/"
    update_config_var $ZBX_CONFIG "SSLCALocation" "$ZABBIX_USER_HOME_DIR/ssl/ssl_ca/"
    update_config_var $ZBX_CONFIG "LoadModulePath" "$ZABBIX_USER_HOME_DIR/modules/"
    update_config_multiple_var $ZBX_CONFIG "LoadModule" "${ZBX_LOADMODULE}"

    update_config_var $ZBX_CONFIG "TLSConnect" "${ZBX_TLSCONNECT}"
    update_config_var $ZBX_CONFIG "TLSAccept" "${ZBX_TLSACCEPT}"
    update_config_var $ZBX_CONFIG "TLSCAFile" "${ZBX_TLSCAFILE}"
    update_config_var $ZBX_CONFIG "TLSCRLFile" "${ZBX_TLSCRLFILE}"
    update_config_var $ZBX_CONFIG "TLSServerCertIssuer" "${ZBX_TLSSERVERCERTISSUER}"
    update_config_var $ZBX_CONFIG "TLSServerCertSubject" "${ZBX_TLSSERVERCERTSUBJECT}"

    update_config_var $ZBX_CONFIG "TLSCertFile" "${ZBX_TLSCERTFILE}"
    update_config_var $ZBX_CONFIG "TLSCipherAll" "${ZBX_TLSCIPHERALL}"
    update_config_var $ZBX_CONFIG "TLSCipherAll13" "${ZBX_TLSCIPHERALL13}"
    update_config_var $ZBX_CONFIG "TLSCipherCert" "${ZBX_TLSCIPHERCERT}"
    update_config_var $ZBX_CONFIG "TLSCipherCert13" "${ZBX_TLSCIPHERCERT13}"
    update_config_var $ZBX_CONFIG "TLSCipherPSK" "${ZBX_TLSCIPHERPSK}"
    update_config_var $ZBX_CONFIG "TLSCipherPSK13" "${ZBX_TLSCIPHERPSK13}"
    update_config_var $ZBX_CONFIG "TLSKeyFile" "${ZBX_TLSKEYFILE}"

    update_config_var $ZBX_CONFIG "TLSPSKIdentity" "${ZBX_TLSPSKIDENTITY}"
    update_config_var $ZBX_CONFIG "TLSPSKFile" "${ZBX_TLSPSKFILE}"

    if [ "$(id -u)" != '0' ]; then
        update_config_var $ZBX_CONFIG "User" "$(whoami)"
    else
        update_config_var $ZBX_CONFIG "AllowRoot" "1"
    fi
}

prepare_proxy() {
    echo "Preparing Zabbix proxy"

    update_zbx_config
}

#################################################

if [ "${1#-}" != "$1" ]; then
    set -- /usr/sbin/zabbix_proxy "$@"
fi

if [ "$1" == '/usr/sbin/zabbix_proxy' ]; then
    prepare_proxy
fi

exec "$@"

#################################################
