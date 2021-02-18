#!/bin/bash

set -o pipefail

set +e

# Script trace mode
if [ "${DEBUG_MODE}" == "true" ]; then
    set -o xtrace
fi

# Default directories
# User 'zabbix' home directory
ZABBIX_USER_HOME_DIR="/var/lib/zabbix"
# Configuration files directory
ZABBIX_ETC_DIR="/etc/zabbix"

# usage: file_env VAR [DEFAULT]
# as example: file_env 'MYSQL_PASSWORD' 'zabbix'
#    (will allow for "$MYSQL_PASSWORD_FILE" to fill in the value of "$MYSQL_PASSWORD" from a file)
# unsets the VAR_FILE afterwards and just leaving VAR
file_env() {
    local var="$1"
    local fileVar="${var}_FILE"
    local defaultValue="${2:-}"

    if [ "${!var:-}" ] && [ "${!fileVar:-}" ]; then
        echo "**** Both variables $var and $fileVar are set (but are exclusive)"
        exit 1
    fi

    local val="$defaultValue"

    if [ "${!var:-}" ]; then
        val="${!var}"
        echo "** Using ${var} variable from ENV"
    elif [ "${!fileVar:-}" ]; then
        if [ ! -f "${!fileVar}" ]; then
            echo "**** Secret file \"${!fileVar}\" is not found"
            exit 1
        fi
        val="$(< "${!fileVar}")"
        echo "** Using ${var} variable from secret file"
    fi
    export "$var"="$val"
    unset "$fileVar"
}

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

    local masklist=("DBPassword TLSPSKIdentity")

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

# Check prerequisites for PostgreSQL database
check_variables_postgresql() {
    file_env POSTGRES_USER
    file_env POSTGRES_PASSWORD

    : ${DB_SERVER_HOST:="postgres-server"}
    : ${DB_SERVER_PORT:="5432"}

    DB_SERVER_ROOT_USER=${POSTGRES_USER:-"postgres"}
    DB_SERVER_ROOT_PASS=${POSTGRES_PASSWORD:-""}

    DB_SERVER_ZBX_USER=${POSTGRES_USER:-"zabbix"}
    DB_SERVER_ZBX_PASS=${POSTGRES_PASSWORD:-"zabbix"}

    : ${DB_SERVER_SCHEMA:="public"}

    DB_SERVER_DBNAME=${POSTGRES_DB:-"zabbix"}
}

check_db_connect_postgresql() {
    echo "********************"
    echo "* DB_SERVER_HOST: ${DB_SERVER_HOST}"
    echo "* DB_SERVER_PORT: ${DB_SERVER_PORT}"
    echo "* DB_SERVER_DBNAME: ${DB_SERVER_DBNAME}"
    echo "* DB_SERVER_SCHEMA: ${DB_SERVER_SCHEMA}"
    if [ "${DEBUG_MODE}" == "true" ]; then
        echo "* DB_SERVER_ZBX_USER: ${DB_SERVER_ZBX_USER}"
        echo "* DB_SERVER_ZBX_PASS: ${DB_SERVER_ZBX_PASS}"
    fi
    echo "********************"

    if [ -n "${DB_SERVER_ZBX_PASS}" ]; then
        export PGPASSWORD="${DB_SERVER_ZBX_PASS}"
    fi

    WAIT_TIMEOUT=5

    if [ -n "${DB_SERVER_SCHEMA}" ]; then
        PGOPTIONS="--search_path=${DB_SERVER_SCHEMA}"
        export PGOPTIONS
    fi

    while true :
    do
        psql --host ${DB_SERVER_HOST} --port ${DB_SERVER_PORT} --username ${DB_SERVER_ROOT_USER} --list --quiet 1>/dev/null 2>&1 && break
        psql --host ${DB_SERVER_HOST} --port ${DB_SERVER_PORT} --username ${DB_SERVER_ROOT_USER} --list --dbname ${DB_SERVER_DBNAME} --quiet 1>/dev/null 2>&1 && break

        echo "**** PostgreSQL server is not available. Waiting $WAIT_TIMEOUT seconds..."
        sleep $WAIT_TIMEOUT
    done

    unset PGPASSWORD
    unset PGOPTIONS
}

psql_query() {
    query=$1
    db=$2

    local result=""

    if [ -n "${DB_SERVER_ZBX_PASS}" ]; then
        export PGPASSWORD="${DB_SERVER_ZBX_PASS}"
    fi

    if [ -n "${DB_SERVER_SCHEMA}" ]; then
        PGOPTIONS="--search_path=${DB_SERVER_SCHEMA}"
        export PGOPTIONS
    fi

    result=$(psql --no-align --quiet --tuples-only --host "${DB_SERVER_HOST}" --port "${DB_SERVER_PORT}" \
             --username "${DB_SERVER_ROOT_USER}" --command "$query" --dbname "$db" 2>/dev/null);

    unset PGPASSWORD
    unset PGOPTIONS

    echo $result
}

create_db_database_postgresql() {
    DB_EXISTS=$(psql_query "SELECT 1 AS result FROM pg_database WHERE datname='${DB_SERVER_DBNAME}'" "${DB_SERVER_DBNAME}")

    if [ -z ${DB_EXISTS} ]; then
        echo "** Database '${DB_SERVER_DBNAME}' does not exist. Creating..."

        if [ -n "${DB_SERVER_ZBX_PASS}" ]; then
            export PGPASSWORD="${DB_SERVER_ZBX_PASS}"
        fi

        if [ -n "${DB_SERVER_SCHEMA}" ]; then
            PGOPTIONS="--search_path=${DB_SERVER_SCHEMA}"
            export PGOPTIONS
        fi

        createdb --host "${DB_SERVER_HOST}" --port "${DB_SERVER_PORT}" --username "${DB_SERVER_ROOT_USER}" \
                 --owner "${DB_SERVER_ZBX_USER}" --lc-ctype "en_US.utf8" --lc-collate "en_US.utf8" "${DB_SERVER_DBNAME}"

        unset PGPASSWORD
        unset PGOPTIONS
    else
        echo "** Database '${DB_SERVER_DBNAME}' already exists. Please be careful with database owner!"
    fi

    psql_query "CREATE SCHEMA IF NOT EXISTS ${DB_SERVER_SCHEMA}" "${DB_SERVER_DBNAME}" 1>/dev/null
}

create_db_schema_postgresql() {
    DBVERSION_TABLE_EXISTS=$(psql_query "SELECT 1 FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON n.oid = 
                                         c.relnamespace WHERE  n.nspname = '$DB_SERVER_SCHEMA' AND c.relname = 'dbversion'" "${DB_SERVER_DBNAME}")

    if [ -n "${DBVERSION_TABLE_EXISTS}" ]; then
        echo "** Table '${DB_SERVER_DBNAME}.dbversion' already exists."
        ZBX_DB_VERSION=$(psql_query "SELECT mandatory FROM ${DB_SERVER_SCHEMA}.dbversion" "${DB_SERVER_DBNAME}")
    fi

    if [ -z "${ZBX_DB_VERSION}" ]; then
        echo "** Creating '${DB_SERVER_DBNAME}' schema in PostgreSQL"

        if [ -n "${DB_SERVER_ZBX_PASS}" ]; then
            export PGPASSWORD="${DB_SERVER_ZBX_PASS}"
        fi

        if [ -n "${DB_SERVER_SCHEMA}" ]; then
            PGOPTIONS="--search_path=${DB_SERVER_SCHEMA}"
            export PGOPTIONS
        fi

        zcat /usr/share/doc/zabbix-server-postgresql/create.sql.gz | psql --quiet \
                --host "${DB_SERVER_HOST}" --port "${DB_SERVER_PORT}" \
                --username "${DB_SERVER_ZBX_USER}" --dbname "${DB_SERVER_DBNAME}" 1>/dev/null || exit 1

        unset PGPASSWORD
        unset PGOPTIONS
    fi
}

update_zbx_config() {
    local type=$1
    local db_type=$2

    echo "** Preparing Zabbix server configuration file"

    ZBX_CONFIG=$ZABBIX_ETC_DIR/zabbix_server.conf

    update_config_var $ZBX_CONFIG "ListenIP" "${ZBX_LISTENIP}"
    update_config_var $ZBX_CONFIG "ListenPort" "${ZBX_LISTENPORT}"

    update_config_var $ZBX_CONFIG "SourceIP" "${ZBX_SOURCEIP}"
    update_config_var $ZBX_CONFIG "LogType" "console"
    update_config_var $ZBX_CONFIG "LogFile"
    update_config_var $ZBX_CONFIG "LogFileSize"
    update_config_var $ZBX_CONFIG "PidFile"

    update_config_var $ZBX_CONFIG "DebugLevel" "${ZBX_DEBUGLEVEL}"

    update_config_var $ZBX_CONFIG "DBHost" "${DB_SERVER_HOST}"
    update_config_var $ZBX_CONFIG "DBName" "${DB_SERVER_DBNAME}"
    update_config_var $ZBX_CONFIG "DBSchema" "${DB_SERVER_SCHEMA}"
    update_config_var $ZBX_CONFIG "DBUser" "${DB_SERVER_ZBX_USER}"
    update_config_var $ZBX_CONFIG "DBPort" "${DB_SERVER_PORT}"
    update_config_var $ZBX_CONFIG "DBPassword" "${DB_SERVER_ZBX_PASS}"

    update_config_var $ZBX_CONFIG "DBSocket" "${DB_SERVER_SOCKET}"

    update_config_var $ZBX_CONFIG "StartPollers" "${ZBX_STARTPOLLERS}"
    update_config_var $ZBX_CONFIG "StartIPMIPollers" "${ZBX_IPMIPOLLERS}"
    update_config_var $ZBX_CONFIG "StartPollersUnreachable" "${ZBX_STARTPOLLERSUNREACHABLE}"
    update_config_var $ZBX_CONFIG "StartTrappers" "${ZBX_STARTTRAPPERS}"
    update_config_var $ZBX_CONFIG "StartPingers" "${ZBX_STARTPINGERS}"
    update_config_var $ZBX_CONFIG "StartDiscoverers" "${ZBX_STARTDISCOVERERS}"
    update_config_var $ZBX_CONFIG "StartHTTPPollers" "${ZBX_STARTHTTPPOLLERS}"

    update_config_var $ZBX_CONFIG "StartTimers" "${ZBX_STARTTIMERS}"
    update_config_var $ZBX_CONFIG "StartEscalators" "${ZBX_STARTESCALATORS}"

    : ${ZBX_JAVAGATEWAY_ENABLE:="false"}
    if [ "${ZBX_JAVAGATEWAY_ENABLE}" == "true" ]; then
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
    if [ "${ZBX_ENABLE_SNMP_TRAPS}" == "true" ]; then
        update_config_var $ZBX_CONFIG "SNMPTrapperFile" "${ZABBIX_USER_HOME_DIR}/snmptraps/snmptraps.log"
        update_config_var $ZBX_CONFIG "StartSNMPTrapper" "1"
    else
        update_config_var $ZBX_CONFIG "SNMPTrapperFile"
        update_config_var $ZBX_CONFIG "StartSNMPTrapper"
    fi

    update_config_var $ZBX_CONFIG "HousekeepingFrequency" "${ZBX_HOUSEKEEPINGFREQUENCY}"

    update_config_var $ZBX_CONFIG "MaxHousekeeperDelete" "${ZBX_MAXHOUSEKEEPERDELETE}"
    update_config_var $ZBX_CONFIG "SenderFrequency" "${ZBX_SENDERFREQUENCY}"

    update_config_var $ZBX_CONFIG "CacheSize" "${ZBX_CACHESIZE}"

    update_config_var $ZBX_CONFIG "CacheUpdateFrequency" "${ZBX_CACHEUPDATEFREQUENCY}"

    update_config_var $ZBX_CONFIG "StartDBSyncers" "${ZBX_STARTDBSYNCERS}"
    update_config_var $ZBX_CONFIG "HistoryCacheSize" "${ZBX_HISTORYCACHESIZE}"
    update_config_var $ZBX_CONFIG "HistoryIndexCacheSize" "${ZBX_HISTORYINDEXCACHESIZE}"

    update_config_var $ZBX_CONFIG "TrendCacheSize" "${ZBX_TRENDCACHESIZE}"
    update_config_var $ZBX_CONFIG "ValueCacheSize" "${ZBX_VALUECACHESIZE}"

    update_config_var $ZBX_CONFIG "Timeout" "${ZBX_TIMEOUT}"
    update_config_var $ZBX_CONFIG "TrapperTimeout" "${ZBX_TRAPPERIMEOUT}"
    update_config_var $ZBX_CONFIG "UnreachablePeriod" "${ZBX_UNREACHABLEPERIOD}"
    update_config_var $ZBX_CONFIG "UnavailableDelay" "${ZBX_UNAVAILABLEDELAY}"
    update_config_var $ZBX_CONFIG "UnreachableDelay" "${ZBX_UNREACHABLEDELAY}"

    update_config_var $ZBX_CONFIG "AlertScriptsPath" "/usr/lib/zabbix/alertscripts"
    update_config_var $ZBX_CONFIG "ExternalScripts" "/usr/lib/zabbix/externalscripts"

    update_config_var $ZBX_CONFIG "FpingLocation" "/usr/bin/fping"
    update_config_var $ZBX_CONFIG "Fping6Location" "/usr/bin/fping6"

    update_config_var $ZBX_CONFIG "SSHKeyLocation" "$ZABBIX_USER_HOME_DIR/ssh_keys"
    update_config_var $ZBX_CONFIG "LogSlowQueries" "${ZBX_LOGSLOWQUERIES}"

    update_config_var $ZBX_CONFIG "StartProxyPollers" "${ZBX_STARTPROXYPOLLERS}"
    update_config_var $ZBX_CONFIG "ProxyConfigFrequency" "${ZBX_PROXYCONFIGFREQUENCY}"
    update_config_var $ZBX_CONFIG "ProxyDataFrequency" "${ZBX_PROXYDATAFREQUENCY}"

    update_config_var $ZBX_CONFIG "SSLCertLocation" "$ZABBIX_USER_HOME_DIR/ssl/certs/"
    update_config_var $ZBX_CONFIG "SSLKeyLocation" "$ZABBIX_USER_HOME_DIR/ssl/keys/"
    update_config_var $ZBX_CONFIG "SSLCALocation" "$ZABBIX_USER_HOME_DIR/ssl/ssl_ca/"
    update_config_var $ZBX_CONFIG "LoadModulePath" "$ZABBIX_USER_HOME_DIR/modules/"
    update_config_multiple_var $ZBX_CONFIG "LoadModule" "${ZBX_LOADMODULE}"

    update_config_var $ZBX_CONFIG "TLSCAFile" "${ZBX_TLSCAFILE}"
    update_config_var $ZBX_CONFIG "TLSCRLFile" "${ZBX_TLSCRLFILE}"

    update_config_var $ZBX_CONFIG "TLSCertFile" "${ZBX_TLSCERTFILE}"
    update_config_var $ZBX_CONFIG "TLSKeyFile" "${ZBX_TLSKEYFILE}"

    if [ "$(id -u)" != '0' ]; then
        update_config_var $ZBX_CONFIG "User" "$(whoami)"
    else
        update_config_var $ZBX_CONFIG "AllowRoot" "1"
    fi
}

prepare_server() {
    echo "** Preparing Zabbix server"

    check_variables_postgresql
    check_db_connect_postgresql
    create_db_database_postgresql
    create_db_schema_postgresql

    update_zbx_config
}

#################################################

if [ "${1#-}" != "$1" ]; then
    set -- /usr/sbin/zabbix_server "$@"
fi

if [ "$1" == '/usr/sbin/zabbix_server' ]; then
    prepare_server
fi

exec "$@"

#################################################
