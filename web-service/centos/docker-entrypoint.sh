#!/bin/bash

set -o pipefail

set +e

# Script trace mode
if [ "${DEBUG_MODE,,}" == "true" ]; then
    set -o xtrace
fi

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

    if [ ! -f "$config_path" ]; then
        echo "**** Configuration file '$config_path' does not exist"
        return
    fi

    echo -n "** Updating '$config_path' parameter \"$var_name\": '$var_value'..."

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
    if [[ $var_name =~ ^TLS.*File$ ]]; then
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
    elif [ "$(grep -Ec "^[#;] $var_name=" $config_path)" -gt 0 ]; then
        sed -i -e "/^[#;] $var_name=/s/.*/&\n$var_name=$var_value/" "$config_path"
        echo "added"
    else
        sed -i -e '$a\' -e "$var_name=$var_value" "$config_path"
        echo "added at the end"
    fi

}

prepare_zbx_web_service_config() {
    echo "** Preparing Zabbix web service configuration file"
    ZBX_CONFIG=$ZABBIX_ETC_DIR/zabbix_web_service.conf

    update_config_var $ZBX_CONFIG "LogType" "console"
    update_config_var $ZBX_CONFIG "LogFile"
    update_config_var $ZBX_CONFIG "LogFileSize"
    update_config_var $ZBX_CONFIG "DebugLevel" "${ZBX_DEBUGLEVEL}"

    : ${ZBX_ALLOWEDIP:="zabbix-server"}
    update_config_var $ZBX_CONFIG "AllowedIP" "${ZBX_ALLOWEDIP}"

    update_config_var $ZBX_CONFIG "ListenPort" "${ZBX_LISTENPORT}"

    update_config_var $ZBX_CONFIG "Timeout" "${ZBX_TIMEOUT}"

    update_config_var $ZBX_CONFIG "TLSAccept" "${ZBX_TLSACCEPT}"
    update_config_var $ZBX_CONFIG "TLSCAFile" "${ZBX_TLSCAFILE}"

    update_config_var $ZBX_CONFIG "TLSCertFile" "${ZBX_TLSCERTFILE}"
    update_config_var $ZBX_CONFIG "TLSKeyFile" "${ZBX_TLSKEYFILE}"
}

prepare_web_service() {
    echo "** Preparing Zabbix web service"
    prepare_zbx_web_service_config
}

#################################################

if [ "$1" == '/usr/sbin/zabbix_web_service' ]; then
    prepare_web_service
fi

exec "$@"

#################################################
