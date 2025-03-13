#!/bin/bash

set -o pipefail

set +e

# Script trace mode
if [ "${DEBUG_MODE,,}" == "true" ]; then
    set -o xtrace
fi

#Enable PostgreSQL timescaleDB feature:
: ${ENABLE_TIMESCALEDB:="false"}

# Default directories
# Internal directory for TLS related files, used when TLS*File specified as plain text values
ZABBIX_INTERNAL_ENC_DIR="${ZABBIX_USER_HOME_DIR}/enc_internal"

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

# Check prerequisites for PostgreSQL database
check_variables_postgresql() {
    : ${DB_SERVER_HOST="postgres-server"}
    : ${DB_SERVER_PORT:="5432"}

    file_env POSTGRES_USER
    file_env POSTGRES_PASSWORD

    DB_SERVER_ROOT_USER=${POSTGRES_USER:-"postgres"}
    DB_SERVER_ROOT_PASS=${POSTGRES_PASSWORD:-""}

    DB_SERVER_ZBX_USER=${POSTGRES_USER:-"zabbix"}
    DB_SERVER_ZBX_PASS=${POSTGRES_PASSWORD:-"zabbix"}

    : ${DB_SERVER_SCHEMA:="public"}

    DB_SERVER_DBNAME=${POSTGRES_DB:-"zabbix"}

    : ${POSTGRES_USE_IMPLICIT_SEARCH_PATH:="false"}

    if [ -n "${DB_SERVER_HOST}" ]; then
        psql_connect_args="--host ${DB_SERVER_HOST} --port ${DB_SERVER_PORT}"
    else
        psql_connect_args="--port ${DB_SERVER_PORT}"
    fi
}

check_db_connect_postgresql() {
    echo "********************"
    if [ -n "${DB_SERVER_HOST}" ]; then
        echo "* DB_SERVER_HOST: ${DB_SERVER_HOST}"
        echo "* DB_SERVER_PORT: ${DB_SERVER_PORT}"
    else
        echo "* DB_SERVER_HOST: Using DB socket"
        echo "* DB_SERVER_PORT: ${DB_SERVER_PORT}"
    fi
    echo "* DB_SERVER_DBNAME: ${DB_SERVER_DBNAME}"
    echo "* DB_SERVER_SCHEMA: ${DB_SERVER_SCHEMA}"
    if [ "${DEBUG_MODE,,}" == "true" ]; then
        echo "* DB_SERVER_ZBX_USER: ${DB_SERVER_ZBX_USER}"
        echo "* DB_SERVER_ZBX_PASS: ${DB_SERVER_ZBX_PASS}"
    fi
    echo "********************"

    if [ -n "${DB_SERVER_ZBX_PASS}" ]; then
        export PGPASSWORD="${DB_SERVER_ZBX_PASS}"
    fi

    WAIT_TIMEOUT=5

    if [ "${POSTGRES_USE_IMPLICIT_SEARCH_PATH,,}" == "false" ] && [ -n "${DB_SERVER_SCHEMA}" ]; then
        PGOPTIONS="--search_path=${DB_SERVER_SCHEMA}"
        export PGOPTIONS
    fi

    if [ -n "${ZBX_DBTLSCONNECT}" ]; then
        PGSSLMODE=${ZBX_DBTLSCONNECT//_/-}
        export PGSSLMODE=${PGSSLMODE//required/require}
        export PGSSLROOTCERT=${ZBX_DBTLSCAFILE}
        export PGSSLCERT=${ZBX_DBTLSCERTFILE}
        export PGSSLKEY=${ZBX_DBTLSKEYFILE}
    fi

    while true :
    do
        psql $psql_connect_args --username ${DB_SERVER_ROOT_USER} --list --quiet 1>/dev/null 2>&1 && break
        psql $psql_connect_args --username ${DB_SERVER_ROOT_USER} --list --dbname ${DB_SERVER_DBNAME} --quiet 1>/dev/null 2>&1 && break

        echo "**** PostgreSQL server is not available. Waiting $WAIT_TIMEOUT seconds..."
        sleep $WAIT_TIMEOUT
    done

    unset PGPASSWORD
    unset PGOPTIONS
    unset PGSSLMODE
    unset PGSSLROOTCERT
    unset PGSSLCERT
    unset PGSSLKEY
}

psql_query() {
    query=$1
    db=$2

    local result=""

    if [ -n "${DB_SERVER_ZBX_PASS}" ]; then
        export PGPASSWORD="${DB_SERVER_ZBX_PASS}"
    fi

    if [ "${POSTGRES_USE_IMPLICIT_SEARCH_PATH,,}" == "false" ] && [ -n "${DB_SERVER_SCHEMA}" ]; then
        PGOPTIONS="--search_path=${DB_SERVER_SCHEMA}"
        export PGOPTIONS
    fi

    if [ -n "${ZBX_DBTLSCONNECT}" ]; then
        PGSSLMODE=${ZBX_DBTLSCONNECT//_/-}
        export PGSSLMODE=${PGSSLMODE//required/require}
        export PGSSLROOTCERT=${ZBX_DBTLSCAFILE}
        export PGSSLCERT=${ZBX_DBTLSCERTFILE}
        export PGSSLKEY=${ZBX_DBTLSKEYFILE}
    fi

    result=$(psql --no-align --quiet --tuples-only $psql_connect_args \
             --username "${DB_SERVER_ROOT_USER}" --command "$query" --dbname "$db" 2>/dev/null);

    unset PGPASSWORD
    unset PGOPTIONS
    unset PGSSLMODE
    unset PGSSLROOTCERT
    unset PGSSLCERT
    unset PGSSLKEY

    echo $result
}

exec_sql_file() {
    sql_script=$1

    local command="cat"

    if [ -n "${DB_SERVER_ZBX_PASS}" ]; then
        export PGPASSWORD="${DB_SERVER_ZBX_PASS}"
    fi

    if [ "${POSTGRES_USE_IMPLICIT_SEARCH_PATH,,}" == "false" ] && [ -n "${DB_SERVER_SCHEMA}" ]; then
        PGOPTIONS="--search_path=${DB_SERVER_SCHEMA}"
        export PGOPTIONS
    fi

    if [ -n "${ZBX_DBTLSCONNECT}" ]; then
        PGSSLMODE=${ZBX_DBTLSCONNECT//_/-}
        export PGSSLMODE=${PGSSLMODE//required/require}
        export PGSSLROOTCERT=${ZBX_DBTLSCAFILE}
        export PGSSLCERT=${ZBX_DBTLSCERTFILE}
        export PGSSLKEY=${ZBX_DBTLSKEYFILE}
    fi

    if [ "${sql_script: -3}" == ".gz" ]; then
        command="zcat"
    fi

    $command $sql_script | psql --quiet \
        $psql_connect_args \
        --username "${DB_SERVER_ZBX_USER}" --dbname "${DB_SERVER_DBNAME}" 1>/dev/null || exit 1

    unset PGPASSWORD
    unset PGOPTIONS
    unset PGSSLMODE
    unset PGSSLROOTCERT
    unset PGSSLCERT
    unset PGSSLKEY
}

create_db_database_postgresql() {
    DB_EXISTS=$(psql_query "SELECT 1 AS result FROM pg_database WHERE datname='${DB_SERVER_DBNAME}'" "${DB_SERVER_DBNAME}")

    if [ -z ${DB_EXISTS} ]; then
        echo "** Database '${DB_SERVER_DBNAME}' does not exist. Creating..."

        if [ -n "${DB_SERVER_ZBX_PASS}" ]; then
            export PGPASSWORD="${DB_SERVER_ZBX_PASS}"
        fi

        if [ "${POSTGRES_USE_IMPLICIT_SEARCH_PATH,,}" == "false" ] && [ -n "${DB_SERVER_SCHEMA}" ]; then
            PGOPTIONS="--search_path=${DB_SERVER_SCHEMA}"
            export PGOPTIONS
        fi

        if [ -n "${ZBX_DBTLSCONNECT}" ]; then
            PGSSLMODE=${ZBX_DBTLSCONNECT//_/-}
            export PGSSLMODE=${PGSSLMODE//required/require}
            export PGSSLROOTCERT=${ZBX_DBTLSCAFILE}
            export PGSSLCERT=${ZBX_DBTLSCERTFILE}
            export PGSSLKEY=${ZBX_DBTLSKEYFILE}
        fi

        createdb $psql_connect_args --username "${DB_SERVER_ROOT_USER}" \
                 --owner "${DB_SERVER_ZBX_USER}" --lc-ctype "en_US.utf8" --lc-collate "en_US.utf8" "${DB_SERVER_DBNAME}"

        unset PGPASSWORD
        unset PGOPTIONS
        unset PGSSLMODE
        unset PGSSLROOTCERT
        unset PGSSLCERT
        unset PGSSLKEY
    else
        echo "** Database '${DB_SERVER_DBNAME}' already exists. Please be careful with database owner!"
    fi

    psql_query "CREATE SCHEMA IF NOT EXISTS ${DB_SERVER_SCHEMA}" "${DB_SERVER_DBNAME}" 1>/dev/null
}

apply_db_scripts() {
    db_scripts=$1

    for sql_script in $db_scripts; do
        [ -e "$sql_script" ] || continue
        echo "** Processing additional '$sql_script' SQL script"

        exec_sql_file "$sql_script"
    done
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

        if [ "${ENABLE_TIMESCALEDB,,}" == "true" ]; then
            psql_query "CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;" "${DB_SERVER_DBNAME}"
        fi

        exec_sql_file "/usr/share/doc/zabbix-server-postgresql/create.sql.gz"

        if [ "${ENABLE_TIMESCALEDB,,}" == "true" ]; then
            exec_sql_file "/usr/share/doc/zabbix-server-postgresql/timescaledb.sql"
        fi

        apply_db_scripts "${ZABBIX_USER_HOME_DIR}/dbscripts/*.sql"
    fi
}

update_zbx_config() {
    export ZBX_DB_HOST="${DB_SERVER_HOST}"
    export ZBX_DB_PORT="${DB_SERVER_PORT}"

    export ZBX_DB_NAME="${DB_SERVER_DBNAME}"
    export ZBX_DB_SCHEMA="${DB_SERVER_SCHEMA}"

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

    for env_var in $(env | grep -E "^(ZABBIX|DB|POSTGRES)_"); do
        unset "${env_var%%=*}"
    done
}

prepare_db() {
    echo "** Preparing database"

    check_variables_postgresql
    check_db_connect_postgresql
    create_db_database_postgresql
    create_db_schema_postgresql
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
