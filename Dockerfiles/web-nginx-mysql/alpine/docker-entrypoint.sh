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
: ${DAEMON_USER:="nginx"}
: ${DAEMON_GROUP:="nginx"}

# Default directories
# Nginx main configuration file
NGINX_CONF_FILE="/etc/nginx/nginx.conf"
# Nginx virtual hosts configuration directory
NGINX_CONFD_DIR="/etc/nginx/http.d"
# Directory with SSL certificate files for Nginx
NGINX_SSL_CONFIG_DIR="/etc/ssl/nginx"
# PHP-FPM configuration file
PHP_CONFIG_FILE="/etc/php84/php-fpm.d/zabbix.conf"

escape_spec_char() {
    local var_value=$1

    var_value="${var_value//\\/\\\\}"
    var_value="${var_value//./\\.}"

    echo "$var_value"
}

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

# Check prerequisites for MySQL database
check_variables() {
    if [ -n "${DB_SERVER_SOCKET}" ]; then
        mysql_connect_args="-S ${DB_SERVER_SOCKET}"
        DB_SERVER_HOST="localhost"
    else
        : ${DB_SERVER_HOST:="mysql-server"}
        : ${DB_SERVER_PORT:="3306"}
        mysql_connect_args="-h ${DB_SERVER_HOST} -P ${DB_SERVER_PORT}"
    fi

    file_env MYSQL_USER
    file_env MYSQL_PASSWORD

    DB_SERVER_ZBX_USER=${MYSQL_USER:-"zabbix"}
    DB_SERVER_ZBX_PASS=${MYSQL_PASSWORD:-"zabbix"}

    DB_SERVER_DBNAME=${MYSQL_DATABASE:-"zabbix"}

}

db_tls_params() {
    local result=""

    if [ "${ZBX_DB_ENCRYPTION,,}" == "true" ]; then
        result="--ssl"

        if [ -n "${ZBX_DB_CA_FILE}" ]; then
            result="${result} --ssl-ca=${ZBX_DB_CA_FILE}"
        fi

        if [ -n "${ZBX_DB_KEY_FILE}" ]; then
            result="${result} --ssl-key=${ZBX_DB_KEY_FILE}"
        fi

        if [ -n "${ZBX_DB_CERT_FILE}" ]; then
            result="${result} --ssl-cert=${ZBX_DB_CERT_FILE}"
        fi
    fi

    echo $result
}

check_db_connect() {
    echo "********************"
    echo "* DB_SERVER_HOST: ${DB_SERVER_HOST}"
    echo "* DB_SERVER_PORT: ${DB_SERVER_PORT}"
    if [ -n "${DB_SERVER_SOCKET}" ]; then
        echo "* DB_SERVER_SOCKET: ${DB_SERVER_SOCKET}"
    fi
    echo "* DB_SERVER_DBNAME: ${DB_SERVER_DBNAME}"
    if [ "${DEBUG_MODE,,}" == "true" ]; then
        echo "* DB_SERVER_ZBX_USER: ${DB_SERVER_ZBX_USER}"
        echo "* DB_SERVER_ZBX_PASS: ${DB_SERVER_ZBX_PASS}"
    fi
    echo "********************"

    WAIT_TIMEOUT=5

    ssl_opts="$(db_tls_params)"

    export MYSQL_PWD="${DB_SERVER_ZBX_PASS}"

    while [ ! "$(mariadb-admin ping $mysql_connect_args -u ${DB_SERVER_ZBX_USER} \
                --silent --skip-ssl-verify-server-cert --connect_timeout=10 $ssl_opts)" ]; do
        echo "**** MySQL server is not available. Waiting $WAIT_TIMEOUT seconds..."
        sleep $WAIT_TIMEOUT
    done

    unset MYSQL_PWD
}

prepare_web_server() {
    if [ "$(id -u)" == '0' ]; then
        sed -i -e "/^[#;] user/s/.*/user ${DAEMON_USER};/" "$NGINX_CONF_FILE"
    fi

    if [ ! -f "/proc/net/if_inet6" ]; then
        sed -i '/listen \[::\]/d' "$ZABBIX_CONF_DIR/nginx.conf"
        sed -i '/allow ::1/d' "$ZABBIX_CONF_DIR/nginx.conf"
        sed -i '/listen \[::\]/d' "$ZABBIX_CONF_DIR/nginx_ssl.conf"
        sed -i '/allow ::1/d' "$ZABBIX_CONF_DIR/nginx_ssl.conf"
    fi

    echo "** Adding Zabbix virtual host (HTTP)"
    if [ -f "$ZABBIX_CONF_DIR/nginx.conf" ]; then
        ln -sfT "$ZABBIX_CONF_DIR/nginx.conf" "$NGINX_CONFD_DIR/nginx.conf"
    else
        echo "**** Impossible to enable HTTP virtual host"
    fi

    if [ -f "$NGINX_SSL_CONFIG_DIR/ssl.crt" ] && [ -f "$NGINX_SSL_CONFIG_DIR/ssl.key" ] && [ -f "$NGINX_SSL_CONFIG_DIR/dhparam.pem" ]; then
        echo "** Enable SSL support for Nginx"
        if [ -f "$ZABBIX_CONF_DIR/nginx_ssl.conf" ]; then
            ln -sfT "$ZABBIX_CONF_DIR/nginx_ssl.conf" "$NGINX_CONFD_DIR/nginx_ssl.conf"
        else
            echo "**** Impossible to enable HTTPS virtual host"
        fi
    else
        echo "**** Impossible to enable SSL support for Nginx. Certificates are missed."
    fi

    FCGI_READ_TIMEOUT=$(expr ${ZBX_MAXEXECUTIONTIME} + 1)
    sed -i \
        -e "s/{FCGI_READ_TIMEOUT}/${FCGI_READ_TIMEOUT}/g" \
    "$ZABBIX_CONF_DIR/nginx.conf"

    : ${HTTP_INDEX_FILE:="index.php"}
    sed -i \
        -e "s/{HTTP_INDEX_FILE}/${HTTP_INDEX_FILE}/g" \
    "$ZABBIX_CONF_DIR/nginx.conf"

    if [ -f "$ZABBIX_CONF_DIR/nginx_ssl.conf" ]; then
        sed -i \
            -e "s/{FCGI_READ_TIMEOUT}/${FCGI_READ_TIMEOUT}/g" \
        "$ZABBIX_CONF_DIR/nginx_ssl.conf"

        sed -i \
            -e "s/{HTTP_INDEX_FILE}/${HTTP_INDEX_FILE}/g" \
        "$ZABBIX_CONF_DIR/nginx_ssl.conf"
    fi

    : ${ENABLE_WEB_ACCESS_LOG:="true"}

    if [ "${ENABLE_WEB_ACCESS_LOG,,}" == "false" ]; then
        sed -ri \
            -e 's!^(\s*access_log).+\;!\1 off\;!g' \
            "$NGINX_CONF_FILE"
        sed -ri \
            -e 's!^(\s*access_log).+\;!\1 off\;!g' \
            "$ZABBIX_CONF_DIR/nginx.conf"
        sed -ri \
            -e 's!^(\s*access_log).+\;!\1 off\;!g' \
            "$ZABBIX_CONF_DIR/nginx_ssl.conf"
    fi

    : ${EXPOSE_WEB_SERVER_INFO:="on"}

    [[ "${EXPOSE_WEB_SERVER_INFO}" != "off" ]] && EXPOSE_WEB_SERVER_INFO="on"

    export EXPOSE_WEB_SERVER_INFO=${EXPOSE_WEB_SERVER_INFO}
    sed -i \
        -e "s/{EXPOSE_WEB_SERVER_INFO}/${EXPOSE_WEB_SERVER_INFO}/g" \
    "$NGINX_CONF_FILE"

    [ ! -z "${WEB_REAL_IP_FROM}" ] && WEB_REAL_IP_FROM="set_real_ip_from ${WEB_REAL_IP_FROM};"
    WEB_REAL_IP_FROM=$(escape_spec_char "$WEB_REAL_IP_FROM")
    [ ! -z "${WEB_REAL_IP_HEADER}" ] && WEB_REAL_IP_HEADER="real_ip_header ${WEB_REAL_IP_HEADER};"
    WEB_REAL_IP_HEADER=$(escape_spec_char "$WEB_REAL_IP_HEADER")

    sed -i \
        -e "s/{WEB_REAL_IP_FROM}/${WEB_REAL_IP_FROM}/g" \
    "$ZABBIX_CONF_DIR/nginx.conf"

    sed -i \
        -e "s/{WEB_REAL_IP_FROM}/${WEB_REAL_IP_FROM}/g" \
    "$ZABBIX_CONF_DIR/nginx_ssl.conf"

    sed -i \
        -e "s/{WEB_REAL_IP_HEADER}/${WEB_REAL_IP_HEADER}/g" \
    "$ZABBIX_CONF_DIR/nginx.conf"

    sed -i \
        -e "s/{WEB_REAL_IP_HEADER}/${WEB_REAL_IP_HEADER}/g" \
    "$ZABBIX_CONF_DIR/nginx_ssl.conf"
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

    export DB_SERVER_TYPE="MYSQL"
    test -z "${DB_SERVER_HOST}" || export DB_SERVER_HOST
    test -z "${DB_SERVER_PORT}" || export DB_SERVER_PORT
    test -z "${DB_SERVER_SOCKET}" || export DB_SERVER_SOCKET
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

echo "** Deploying Zabbix web-interface (Nginx) with MySQL database"

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
