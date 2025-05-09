#!/bin/bash

set -o pipefail

set +e

# Script trace mode
if [ "${DEBUG_MODE,,}" == "true" ]; then
    set -o xtrace
fi

# Internal directory for TLS related files, used when TLS*File specified as plain text values
ZABBIX_INTERNAL_ENC_DIR="${ZABBIX_USER_HOME_DIR}/enc_internal"

: ${DB_CHARACTER_SET:="utf8mb4"}
: ${DB_CHARACTER_COLLATE:="utf8mb4_bin"}

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

    # Remove configuration parameter definition in case of unset or empty parameter value
    if [ -z "$var_value" ]; then
        sed -i -e "/^$var_name=/d" "$config_path"
        echo "removed"
        return
    fi

    # Remove value from configuration parameter in case of set to double quoted parameter value
    if [[ "$var_value" == '""' ]]; then
        if [ "$(grep -E "^$var_name=" $config_path)" ]; then
            sed -i -e "/^$var_name=/s/=.*/=/" "$config_path"
        else
            sed -i -e "/^[#;] $var_name=/s/.*/&\n$var_name=/" "$config_path"
        fi
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

    if [ "$(grep -E "^$var_name=$var_value$" $config_path)" ]; then
        echo "exists"
    elif [ "$(grep -E "^$var_name=" $config_path)" ] && [ "$is_multiple" != "true" ]; then
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

file_process_from_env() {
    local var_name=$1
    local file_name=$2
    local var_value=$3

    if [ ! -z "$var_value" ]; then
        echo -n "$var_value" > "${ZABBIX_INTERNAL_ENC_DIR}/$var_name"
        file_name="${ZABBIX_INTERNAL_ENC_DIR}/${var_name}"
    fi

    if [ -n "$var_value" ]; then
        export "$var_name"="$file_name"
    fi
    # Remove variable with plain text data
    unset "${var_name%%FILE}"
}

# Check prerequisites for MySQL database
check_variables_mysql() {
    if [ -n "${DB_SERVER_SOCKET}" ]; then
        mysql_connect_args="-S ${DB_SERVER_SOCKET}"
    else
        : ${DB_SERVER_HOST:="mysql-server"}
        : ${DB_SERVER_PORT:="3306"}
        mysql_connect_args="-h ${DB_SERVER_HOST} -P ${DB_SERVER_PORT}"
    fi

    USE_DB_ROOT_USER=false
    CREATE_ZBX_DB_USER=false
    file_env MYSQL_USER
    file_env MYSQL_PASSWORD

    file_env MYSQL_ROOT_USER
    file_env MYSQL_ROOT_PASSWORD

    if [ ! -n "${MYSQL_USER}" ] && [ "${MYSQL_RANDOM_ROOT_PASSWORD,,}" == "true" ]; then
        echo "**** Impossible to use MySQL server because of unknown Zabbix user and random 'root' password"
        exit 1
    fi

    if [ ! -n "${MYSQL_USER}" ] && [ ! -n "${MYSQL_ROOT_PASSWORD}" ] && [ "${MYSQL_ALLOW_EMPTY_PASSWORD,,}" != "true" ]; then
        echo "*** Impossible to use MySQL server because 'root' password is not defined and it is not empty"
        exit 1
    fi

    if [ "${MYSQL_ALLOW_EMPTY_PASSWORD,,}" == "true" ] || [ -n "${MYSQL_ROOT_PASSWORD}" ]; then
        USE_DB_ROOT_USER=true
        DB_SERVER_ROOT_USER=${MYSQL_ROOT_USER:-"root"}
        DB_SERVER_ROOT_PASS=${MYSQL_ROOT_PASSWORD:-""}
    fi

    [ -n "${MYSQL_USER}" ] && [ "${USE_DB_ROOT_USER}" == "true" ] && CREATE_ZBX_DB_USER=true

    # If root password is not specified use provided credentials
    : ${DB_SERVER_ROOT_USER:=${MYSQL_USER}}
    [ "${MYSQL_ALLOW_EMPTY_PASSWORD,,}" == "true" ] || DB_SERVER_ROOT_PASS=${DB_SERVER_ROOT_PASS:-${MYSQL_PASSWORD}}

    DB_SERVER_ZBX_USER=${MYSQL_USER:-"zabbix"}
    DB_SERVER_ZBX_PASS=${MYSQL_PASSWORD:-"zabbix"}

    DB_SERVER_DBNAME=${MYSQL_DATABASE:-"zabbix"}

}

db_tls_params() {
    local result=""

    if [ -n "${ZBX_DBTLSCONNECT}" ]; then
        result="--ssl"

        if [ "${ZBX_DBTLSCONNECT}" != "required" ]; then
            result="${result} --ssl-verify-server-cert"
        fi

        if [ -n "${ZBX_DBTLSCAFILE}" ]; then
            result="${result} --ssl-ca=${ZBX_DBTLSCAFILE}"
        fi

        if [ -n "${ZBX_DBTLSKEYFILE}" ]; then
            result="${result} --ssl-key=${ZBX_DBTLSKEYFILE}"
        fi

        if [ -n "${ZBX_DBTLSCERTFILE}" ]; then
            result="${result} --ssl-cert=${ZBX_DBTLSCERTFILE}"
        fi
    fi

    echo $result
}

check_db_connect_mysql() {
    echo "********************"
    if [ ! -n "${DB_SERVER_SOCKET}" ]; then
        echo "* DB_SERVER_HOST: ${DB_SERVER_HOST}"
        echo "* DB_SERVER_PORT: ${DB_SERVER_PORT}"
    else
        echo "* DB_SERVER_SOCKET: ${DB_SERVER_SOCKET}"
    fi
    echo "* DB_SERVER_DBNAME: ${DB_SERVER_DBNAME}"
    if [ "${DEBUG_MODE,,}" == "true" ]; then
        if [ "${USE_DB_ROOT_USER}" == "true" ]; then
            echo "* DB_SERVER_ROOT_USER: ${DB_SERVER_ROOT_USER}"
            echo "* DB_SERVER_ROOT_PASS: ${DB_SERVER_ROOT_PASS}"
        fi
        echo "* DB_SERVER_ZBX_USER: ${DB_SERVER_ZBX_USER}"
        echo "* DB_SERVER_ZBX_PASS: ${DB_SERVER_ZBX_PASS}"
    fi
    echo "********************"

    WAIT_TIMEOUT=5

    ssl_opts="$(db_tls_params)"

    export MYSQL_PWD="${DB_SERVER_ROOT_PASS}"

    while [ ! "$(mariadb-admin ping $mysql_connect_args -u ${DB_SERVER_ROOT_USER} \
                --silent --skip-ssl-verify-server-cert --connect_timeout=10 $ssl_opts)" ]; do
        echo "**** MySQL server is not available. Waiting $WAIT_TIMEOUT seconds..."
        sleep $WAIT_TIMEOUT
    done

    unset MYSQL_PWD
}

mysql_query() {
    query=$1
    local result=""

    ssl_opts="$(db_tls_params)"

    export MYSQL_PWD="${DB_SERVER_ROOT_PASS}"

    result=$(mariadb --silent --skip-column-names --skip-ssl-verify-server-cert $mysql_connect_args \
             -u ${DB_SERVER_ROOT_USER} -e "$query" $ssl_opts)

    unset MYSQL_PWD

    echo $result
}

exec_sql_file() {
    sql_script=$1

    local command="cat"

    ssl_opts="$(db_tls_params)"

    export MYSQL_PWD="${DB_SERVER_ROOT_PASS}"

    if [ "${sql_script: -3}" == ".gz" ]; then
        command="zcat"
    fi

    $command "$sql_script" | mariadb --silent --skip-column-names --skip-ssl-verify-server-cert \
            --default-character-set=${DB_CHARACTER_SET} \
            $mysql_connect_args \
            -u ${DB_SERVER_ROOT_USER} $ssl_opts  \
            ${DB_SERVER_DBNAME} 1>/dev/null

    unset MYSQL_PWD
}

create_db_user_mysql() {
    [ "${CREATE_ZBX_DB_USER}" == "true" ] || return

    echo "** Creating '${DB_SERVER_ZBX_USER}' user in MySQL database"

    USER_EXISTS=$(mysql_query "SELECT 1 FROM mysql.user WHERE user = '${DB_SERVER_ZBX_USER}' AND host = '%'")

    if [ -z "$USER_EXISTS" ]; then
        mysql_query "CREATE USER '${DB_SERVER_ZBX_USER}'@'%' IDENTIFIED BY '${DB_SERVER_ZBX_PASS}'" 1>/dev/null
    else
        mysql_query "ALTER USER ${DB_SERVER_ZBX_USER} IDENTIFIED BY '${DB_SERVER_ZBX_PASS}';" 1>/dev/null
    fi

    mysql_query "GRANT ALL PRIVILEGES ON $DB_SERVER_DBNAME. * TO '${DB_SERVER_ZBX_USER}'@'%'" 1>/dev/null
}

create_db_database_mysql() {
    DB_EXISTS=$(mysql_query "SELECT SCHEMA_NAME FROM information_schema.SCHEMATA WHERE SCHEMA_NAME='${DB_SERVER_DBNAME}'")

    if [ -z ${DB_EXISTS} ]; then
        echo "** Database '${DB_SERVER_DBNAME}' does not exist. Creating..."
        mysql_query "CREATE DATABASE ${DB_SERVER_DBNAME} CHARACTER SET ${DB_CHARACTER_SET} COLLATE ${DB_CHARACTER_COLLATE}" 1>/dev/null
        # better solution?
        mysql_query "GRANT ALL PRIVILEGES ON $DB_SERVER_DBNAME. * TO '${DB_SERVER_ZBX_USER}'@'%'" 1>/dev/null
    else
        echo "** Database '${DB_SERVER_DBNAME}' already exists. Please be careful with database COLLATE!"
    fi
}

apply_db_scripts() {
    db_scripts=$1

    for sql_script in $db_scripts; do
        [ -e "$sql_script" ] || continue
        echo "** Processing additional '$sql_script' SQL script"

        exec_sql_file "$sql_script"
    done
}

create_db_schema_mysql() {
    DBVERSION_TABLE_EXISTS=$(mysql_query "SELECT 1 FROM information_schema.tables WHERE table_schema='${DB_SERVER_DBNAME}' and table_name = 'dbversion'")

    if [ -n "${DBVERSION_TABLE_EXISTS}" ]; then
        echo "** Table '${DB_SERVER_DBNAME}.dbversion' already exists."
        ZBX_DB_VERSION=$(mysql_query "SELECT mandatory FROM ${DB_SERVER_DBNAME}.dbversion")
    fi

    if [ -z "${ZBX_DB_VERSION}" ]; then
        echo "** Creating '${DB_SERVER_DBNAME}' schema in MySQL"

        exec_sql_file "/usr/share/doc/zabbix-server-mysql/create.sql.gz"

        apply_db_scripts "${ZABBIX_USER_HOME_DIR}/dbscripts/*.sql"
    fi
}

update_zbx_config() {
    test -z "${DB_SERVER_SOCKET}" || export ZBX_DB_SOCKET="${DB_SERVER_SOCKET}"
    test -z "${DB_SERVER_HOST}" || export ZBX_DB_HOST="${DB_SERVER_HOST}"
    test -z "${DB_SERVER_PORT}" || export ZBX_DB_PORT="${DB_SERVER_PORT}"

    export ZBX_DB_NAME="${DB_SERVER_DBNAME}"

    if [ -n "${ZBX_VAULT}" ] && [ -n "${ZBX_VAULTURL}" ] && [ ! -n "${ZBX_VAULTDBPATH}" ]; then
        export ZBX_DB_USER="${DB_SERVER_ZBX_USER}"
        export ZBX_DB_PASSWORD="${DB_SERVER_ZBX_PASS}"
    elif [ ! -n "${ZBX_VAULT}" ] && [ ! -n "${ZBX_VAULTURL}" ]; then
        export ZBX_DB_USER="${DB_SERVER_ZBX_USER}"
        export ZBX_DB_PASSWORD="${DB_SERVER_ZBX_PASS}"
    else
        unset ZBX_DB_USER
        unset ZBX_DB_PASSWORD
    fi

    : ${ZBX_ENABLE_SNMP_TRAPS:="false"}
    [[ "${ZBX_ENABLE_SNMP_TRAPS,,}" == "true" ]] && export ZBX_STARTSNMPTRAPPER=1
    unset ZBX_ENABLE_SNMP_TRAPS

    update_config_multiple_var "${ZABBIX_CONF_DIR}/zabbix_server_modules.conf" "LoadModule" "${ZBX_LOADMODULE}"

    file_process_from_env "ZBX_TLSCAFILE" "${ZBX_TLSCAFILE}" "${ZBX_TLSCA}"
    file_process_from_env "ZBX_TLSCRLFILE" "${ZBX_TLSCRLFILE}" "${ZBX_TLSCRL}"

    file_process_from_env "ZBX_TLSCERTFILE" "${ZBX_TLSCERTFILE}" "${ZBX_TLSCERT}"
    file_process_from_env "ZBX_TLSKEYFILE" "${ZBX_TLSKEYFILE}" "${ZBX_TLSKEY}"

    if [ "${ZBX_AUTOHANODENAME}" == 'fqdn' ] && [ ! -n "${ZBX_HANODENAME}" ]; then
        export ZBX_HANODENAME="$(hostname -f)"
    elif [ "${ZBX_AUTOHANODENAME}" == 'hostname' ] && [ ! -n "${ZBX_HANODENAME}" ]; then
        export ZBX_HANODENAME="$(hostname)"
    fi

    : ${ZBX_NODEADDRESSPORT:="10051"}
    if [ "${ZBX_AUTONODEADDRESS}" == 'fqdn' ] && [ ! -n "${ZBX_NODEADDRESS}" ]; then
        export ZBX_NODEADDRESS="$(hostname -f):${ZBX_NODEADDRESSPORT}"
    elif [ "${ZBX_AUTONODEADDRESS}" == 'hostname' ] && [ ! -n "${ZBX_NODEADDRESS}" ]; then
        export ZBX_NODEADDRESS="$(hostname):${ZBX_NODEADDRESSPORT}"
    fi

    if [ "$(id -u)" != '0' ]; then
        export ZBX_USER="$(whoami)"
    else
        export ZBX_ALLOWROOT=1
    fi

    command -v openssl >/dev/null 2>&1 && openssl rehash -v "${ZBX_SSLCALOCATION}" 1>/dev/null
}

clear_zbx_env() {
    [[ "${ZBX_CLEAR_ENV}" == "false" ]] && return

    for env_var in $(env | grep -E "^(ZABBIX|DB|MYSQL)_"); do
        unset "${env_var%%=*}"
    done
}

prepare_db() {
    echo "** Preparing database"

    check_variables_mysql
    check_db_connect_mysql
    create_db_user_mysql
    create_db_database_mysql
    create_db_schema_mysql
}

prepare_server() {
    echo "** Preparing Zabbix server"

    prepare_db
    update_zbx_config
    clear_zbx_env
}

#################################################

if [ "${1#-}" != "$1" ]; then
    set -- /usr/sbin/zabbix_server "$@"
fi

if [ "$1" == '/usr/sbin/zabbix_server' ]; then
    prepare_server
fi

if [ "$1" == "init_db_only" ]; then
    prepare_db
else
    exec "$@"
fi

#################################################
