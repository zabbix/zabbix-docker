#!/bin/bash

set -o pipefail

set +e

# Script trace mode
if [ "${DEBUG_MODE,,}" == "true" ]; then
    set -o xtrace
fi

# Internal directory for TLS related files, used when TLS*File specified as plain text values
ZABBIX_INTERNAL_ENC_DIR="${ZABBIX_USER_HOME_DIR}/enc_internal"

# Default Zabbix installation name
# Used only by Zabbix web-interface
: ${ZBX_SERVER_NAME:="Zabbix docker"}
# Default Zabbix server port number
: ${ZBX_SERVER_PORT:="10051"}

# Default timezone for web interface
: ${PHP_TZ:="Europe/Riga"}

# Default user settings
: ${DAEMON_USER:="apache"}
: ${DAEMON_GROUP:="apache"}

# DefaultRuntimeDir configuration option value
export APACHE_RUN_DIR="/tmp/apache2"

# Default directories
# Apache main configuration file
HTTPD_CONF_FILE="/etc/apache2/httpd.conf"
# Apache additional configuration files directory
APACHE_SITES_DIR="/etc/apache2/conf.d"
# Directory with SSL certificate files for Apache
APACHE_SSL_CONFIG_DIR="/etc/ssl/apache2"
# PHP-FPM configuration file
PHP_CONFIG_FILE="/etc/php84/php-fpm.d/zabbix.conf"

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

file_process_from_env() {
    local var_name=$1
    local file_name=$2
    local var_value=$3

    if [ ! -z "$var_value" ]; then
        echo -n "$var_value" > "${ZABBIX_INTERNAL_ENC_DIR}/$var_name"
        file_name="${ZABBIX_INTERNAL_ENC_DIR}/${var_name}"
    fi

    export "$var_name"="$file_name"

    # Remove variable with plain text data
    unset "${var_name%%FILE}"
}

# Check prerequisites for PostgreSQL database
check_variables() {
    file_env POSTGRES_USER
    file_env POSTGRES_PASSWORD

    : ${DB_SERVER_HOST="postgres-server"}
    : ${DB_SERVER_PORT:="5432"}

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

check_db_connect() {
    echo "********************"
    if [ -n "${DB_SERVER_HOST}" ]; then
        echo "* DB_SERVER_HOST: ${DB_SERVER_HOST}"
        echo "* DB_SERVER_PORT: ${DB_SERVER_PORT}"
    else
        echo "* DB_SERVER_HOST: Using DB socket"
        echo "* DB_SERVER_PORT: ${DB_SERVER_PORT}"
    fi
    echo "* DB_SERVER_PORT: ${DB_SERVER_PORT}"
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

    while [ ! "$(psql $psql_connect_args --username ${DB_SERVER_ZBX_USER} --dbname ${DB_SERVER_DBNAME} --list --quiet 2>/dev/null)" ]; do
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

prepare_web_server() {
    if [ "$(id -u)" == '0' ]; then
        export APACHE_RUN_USER=${DAEMON_USER}
    else
        export APACHE_RUN_USER=$(id -n -u)
    fi
    export APACHE_RUN_GROUP=${DAEMON_GROUP}

    echo "** Adding Zabbix virtual host (HTTP)"
    if [ -f "$ZABBIX_CONF_DIR/apache.conf" ]; then
        ln -sfT "$ZABBIX_CONF_DIR/apache.conf" "$APACHE_SITES_DIR/zabbix.conf"
    else
        echo "**** Impossible to enable HTTP virtual host"
    fi

    if [ -f "$APACHE_SSL_CONFIG_DIR/ssl.crt" ] && [ -f "$APACHE_SSL_CONFIG_DIR/ssl.key" ]; then
        echo "** Adding Zabbix virtual host (HTTPS)"
        if [ -f "$ZABBIX_CONF_DIR/apache_ssl.conf" ]; then
            ln -sfT "$ZABBIX_CONF_DIR/apache_ssl.conf" "$APACHE_SITES_DIR/zabbix_ssl.conf"
        else
            echo "**** Impossible to enable HTTPS virtual host"
        fi
    else
        echo "**** Impossible to enable SSL support for Apache2. Certificates are missed."
    fi

    export HTTP_INDEX_FILE=${HTTP_INDEX_FILE:="index.php"}

    : ${ENABLE_WEB_ACCESS_LOG:="true"}
    export APACHE_CUSTOM_LOG="/proc/self/fd/1"
    if [ "${ENABLE_WEB_ACCESS_LOG,,}" == "false" ]; then
        export APACHE_CUSTOM_LOG="/dev/null"
    fi

    : ${EXPOSE_WEB_SERVER_INFO:="on"}
    export APACHE_SERVER_TOKENS="OS"
    export APACHE_SERVER_SIGNATURE="On"
    if [ "${EXPOSE_WEB_SERVER_INFO}" == "off" ]; then
        export APACHE_SERVER_TOKENS="Prod"
        export APACHE_SERVER_SIGNATURE="Off"
    fi

    [ -z "${WEB_REAL_IP_FROM}" ] && sed -i '/WEB_REAL_IP_FROM/d' "$ZABBIX_CONF_DIR/apache.conf"
    [ -z "${WEB_REAL_IP_FROM}" ] && sed -i '/WEB_REAL_IP_FROM/d' "$ZABBIX_CONF_DIR/apache_ssl.conf"
    [ -z "${WEB_REAL_IP_HEADER}" ] && sed -i '/WEB_REAL_IP_HEADER/d' "$ZABBIX_CONF_DIR/apache.conf"
    [ -z "${WEB_REAL_IP_HEADER}" ] && sed -i '/WEB_REAL_IP_HEADER/d' "$ZABBIX_CONF_DIR/apache_ssl.conf"

    mkdir -p "${APACHE_RUN_DIR}"
}

prepare_zbx_php_config() {
    echo "** Preparing PHP configuration"

    export PHP_FPM_PM=${PHP_FPM_PM:-"dynamic"}
    export PHP_FPM_PM_MAX_CHILDREN=${PHP_FPM_PM_MAX_CHILDREN:-"50"}
    export PHP_FPM_PM_START_SERVERS=${PHP_FPM_PM_START_SERVERS:-"5"}
    export PHP_FPM_PM_MIN_SPARE_SERVERS=${PHP_FPM_PM_MIN_SPARE_SERVERS:-"5"}
    export PHP_FPM_PM_MAX_SPARE_SERVERS=${PHP_FPM_PM_MAX_SPARE_SERVERS:-"35"}
    export PHP_FPM_PM_MAX_REQUESTS=${PHP_FPM_PM_MAX_REQUESTS:-"0"}

    if [ "$(id -u)" == '0' ]; then
        echo "user = ${DAEMON_USER}" >> "$PHP_CONFIG_FILE"
        echo "group = ${DAEMON_GROUP}" >> "$PHP_CONFIG_FILE"
        echo "listen.owner = ${DAEMON_USER}" >> "$PHP_CONFIG_FILE"
        echo "listen.group = ${DAEMON_GROUP}" >> "$PHP_CONFIG_FILE"
    fi

    : ${ZBX_DENY_GUI_ACCESS:="false"}
    export ZBX_DENY_GUI_ACCESS=${ZBX_DENY_GUI_ACCESS,,}
    export ZBX_GUI_ACCESS_IP_RANGE=${ZBX_GUI_ACCESS_IP_RANGE:-"['127.0.0.1']"}
    export ZBX_GUI_WARNING_MSG=${ZBX_GUI_WARNING_MSG:-"Zabbix is under maintenance."}

    export ZBX_MAXEXECUTIONTIME=${ZBX_MAXEXECUTIONTIME:-"600"}
    export ZBX_MEMORYLIMIT=${ZBX_MEMORYLIMIT:-"128M"}
    export ZBX_POSTMAXSIZE=${ZBX_POSTMAXSIZE:-"16M"}
    export ZBX_UPLOADMAXFILESIZE=${ZBX_UPLOADMAXFILESIZE:-"2M"}
    export ZBX_MAXINPUTTIME=${ZBX_MAXINPUTTIME:-"300"}
    export PHP_TZ=${PHP_TZ}

    export DB_SERVER_TYPE="POSTGRESQL"
    export DB_SERVER_HOST=${DB_SERVER_HOST}
    export DB_SERVER_PORT=${DB_SERVER_PORT}
    export DB_SERVER_DBNAME=${DB_SERVER_DBNAME}
    export DB_SERVER_SCHEMA=${DB_SERVER_SCHEMA}
    export DB_SERVER_USER=${DB_SERVER_ZBX_USER}
    export DB_SERVER_PASS=${DB_SERVER_ZBX_PASS}
    export ZBX_SERVER_HOST=${ZBX_SERVER_HOST}
    export ZBX_SERVER_PORT=${ZBX_SERVER_PORT}
    export ZBX_SERVER_NAME=${ZBX_SERVER_NAME}

    : ${ZBX_DB_ENCRYPTION:="false"}
    export ZBX_DB_ENCRYPTION=${ZBX_DB_ENCRYPTION,,}
    export ZBX_DB_KEY_FILE=${ZBX_DB_KEY_FILE}
    export ZBX_DB_CERT_FILE=${ZBX_DB_CERT_FILE}
    export ZBX_DB_CA_FILE=${ZBX_DB_CA_FILE}
    : ${ZBX_DB_VERIFY_HOST:="false"}
    export ZBX_DB_VERIFY_HOST=${ZBX_DB_VERIFY_HOST,,}

    export ZBX_VAULT=${ZBX_VAULT}
    export ZBX_VAULTURL=${ZBX_VAULTURL}
    export ZBX_VAULTDBPATH=${ZBX_VAULTDBPATH}
    export VAULT_TOKEN=${VAULT_TOKEN}
    export ZBX_VAULTCERTFILE=${ZBX_VAULTCERTFILE}
    export ZBX_VAULTKEYFILE=${ZBX_VAULTKEYFILE}

    : ${DB_DOUBLE_IEEE754:="true"}
    export DB_DOUBLE_IEEE754=${DB_DOUBLE_IEEE754,,}

    export ZBX_HISTORYSTORAGEURL=${ZBX_HISTORYSTORAGEURL}
    export ZBX_HISTORYSTORAGETYPES=${ZBX_HISTORYSTORAGETYPES:-"[]"}

    export ZBX_SSO_SETTINGS=${ZBX_SSO_SETTINGS:-""}
    export ZBX_SSO_SP_KEY=${ZBX_SSO_SP_KEY}
    export ZBX_SSO_SP_CERT=${ZBX_SSO_SP_CERT}
    export ZBX_SSO_IDP_CERT=${ZBX_SSO_IDP_CERT}

    : ${ZBX_ALLOW_HTTP_AUTH:="true"}
    export ZBX_ALLOW_HTTP_AUTH=${ZBX_ALLOW_HTTP_AUTH}

    : ${ZBX_SERVER_TLS_ACTIVE:="0"}
    export ZBX_SERVER_TLS_ACTIVE=${ZBX_SERVER_TLS_ACTIVE}
    file_process_from_env "ZBX_SERVER_TLS_CAFILE" "${ZBX_SERVER_TLS_CAFILE}" "${ZBX_SERVER_TLS_CA}"
    file_process_from_env "ZBX_SERVER_TLS_KEYFILE" "${ZBX_SERVER_TLS_KEYFILE}" "${ZBX_SERVER_TLS_KEY}"
    file_process_from_env "ZBX_SERVER_TLS_CERTFILE" "${ZBX_SERVER_TLS_CERTFILE}" "${ZBX_SERVER_TLS_CERT}"
    export ZBX_SERVER_TLS_CERT_ISSUER=${ZBX_SERVER_TLS_CERT_ISSUER}
    export ZBX_SERVER_TLS_CERT_SUBJECT=${ZBX_SERVER_TLS_CERT_SUBJECT}
}

prepare_zbx_config() {
    if [ -n "${ZBX_SESSION_NAME}" ]; then
        cp "$ZABBIX_WWW_ROOT/include/defines.inc.php" "/tmp/defines.inc.php_tmp"
        sed "/ZBX_SESSION_NAME/s/'[^']*'/'${ZBX_SESSION_NAME}'/2" "/tmp/defines.inc.php_tmp" > "$ZABBIX_WWW_ROOT/include/defines.inc.php"
        rm -f "/tmp/defines.inc.php_tmp"
    fi
}

#################################################

echo "** Deploying Zabbix web-interface (Apache) with PostgreSQL database"

check_variables
check_db_connect
prepare_zbx_php_config
prepare_web_server
prepare_zbx_config

echo "########################################################"

if [ "$1" != "" ]; then
    echo "** Executing '$@'"
    exec "$@"
elif [ -f "/usr/bin/supervisord" ]; then
    echo "** Executing supervisord"
    exec /usr/bin/supervisord -c /etc/supervisor/supervisord.conf
else
    echo "Unknown instructions. Exiting..."
    exit 1
fi

#################################################
