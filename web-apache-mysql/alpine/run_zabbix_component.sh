#!/bin/bash

set +e

# Script trace mode
if [ "${DEBUG_MODE}" == "true" ]; then
    set -o xtrace
fi

# Type of Zabbix component
# Possible values: [server, proxy, agent, web, dev]
zbx_type="$1"
# Type of Zabbix database
# Possible values: [mysql, postgresql]
zbx_db_type="$2"
# Type of web-server. Valid only with zbx_type = web
# Possible values: [apache, nginx]
zbx_opt_type="$3"

# Default Zabbix installation name
# Used only by Zabbix web-interface
ZBX_SERVER_NAME=${ZBX_SERVER_NAME:-"Zabbix docker"}
# Default Zabbix server host
ZBX_SERVER_HOST=${ZBX_SERVER_HOST:-"zabbix-server"}
# Default Zabbix server port number
ZBX_SERVER_PORT=${ZBX_SERVER_PORT:-"10051"}

# Default timezone for web interface
PHP_TZ=${PHP_TZ:-"Europe/Riga"}

# Default directories
# User 'zabbix' home directory
ZABBIX_USER_HOME_DIR="/var/lib/zabbix"
# Configuration files directory
ZABBIX_ETC_DIR="/etc/zabbix"
# Web interface www-root directory
ZBX_FRONTEND_PATH="/usr/share/zabbix"

prepare_system() {
    local type=$1
    local web_server=$2

    echo "** Preparing the system"

    if [ "$type" != "dev" ]; then
        return
    fi
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

    echo -n "** Updating '$config_path' parameter \"$var_name\": '$var_value'... "

    # Remove configuration parameter definition in case of unset parameter value
    if [ -z "$var_value" ]; then
        sed -i -e "/$var_name=/d" "$config_path"
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

    # Escaping "/" character in parameter value
    var_value=${var_value//\//\\/}

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

# Check prerequisites for MySQL database
check_variables_mysql() {
    local type=$1

    DB_SERVER_HOST=${DB_SERVER_HOST:-"mysql-server"}
    DB_SERVER_PORT=${DB_SERVER_PORT:-"3306"}
    USE_DB_ROOT_USER=false
    CREATE_ZBX_DB_USER=false

    if [ ! -n "${MYSQL_USER}" ] && [ "${MYSQL_RANDOM_ROOT_PASSWORD}" == "true" ]; then
        echo "**** Impossible to use MySQL server because of unknown Zabbix user and random 'root' password"
        exit 1
    fi

    if [ ! -n "${MYSQL_USER}" ] && [ ! -n "${MYSQL_ROOT_PASSWORD}" ] && [ "${MYSQL_ALLOW_EMPTY_PASSWORD}" != "true" ]; then
        echo "*** Impossible to use MySQL server because 'root' password is not defined and not empty"
        exit 1
    fi

    if [ "${MYSQL_ALLOW_EMPTY_PASSWORD}" == "true" ] || [ -n "${MYSQL_ROOT_PASSWORD}" ]; then
        USE_DB_ROOT_USER=true
        DB_SERVER_ROOT_USER="root"
        DB_SERVER_ROOT_PASS=${MYSQL_ROOT_PASSWORD:-""}
    fi

    [ -n "${MYSQL_USER}" ] || CREATE_ZBX_DB_USER=true

    # If root password is not specified use provided credentials
    DB_SERVER_ROOT_USER=${DB_SERVER_ROOT_USER:-${MYSQL_USER}}
    [ "${MYSQL_ALLOW_EMPTY_PASSWORD}" == "true" ] || DB_SERVER_ROOT_PASS=${DB_SERVER_ROOT_PASS:-${MYSQL_PASSWORD}}
    DB_SERVER_ZBX_USER=${MYSQL_USER:-"zabbix"}
    DB_SERVER_ZBX_PASS=${MYSQL_PASSWORD:-"zabbix"}

    if [ "$type" == "proxy" ]; then
        DB_SERVER_DBNAME=${MYSQL_DATABASE:-"zabbix_proxy"}
    else
        DB_SERVER_DBNAME=${MYSQL_DATABASE:-"zabbix"}
    fi
}

# Check prerequisites for PostgreSQL database
check_variables_postgresql() {
    local type=$1

    DB_SERVER_HOST=${DB_SERVER_HOST:-"postgres-server"}
    DB_SERVER_PORT=${DB_SERVER_PORT:-"5432"}
    CREATE_ZBX_DB_USER=${CREATE_ZBX_DB_USER:-"false"}

    DB_SERVER_ROOT_USER=${POSTGRES_USER:-"postgres"}
    DB_SERVER_ROOT_PASS=${POSTGRES_PASSWORD:-""}

    DB_SERVER_ZBX_USER=${POSTGRES_USER:-"zabbix"}
    DB_SERVER_ZBX_PASS=${POSTGRES_PASSWORD:-"zabbix"}

    if [ "$type" == "proxy" ]; then
        DB_SERVER_DBNAME=${POSTGRES_DB:-"zabbix_proxy"}
    else
        DB_SERVER_DBNAME=${POSTGRES_DB:-"zabbix"}
    fi
}

check_db_connect_mysql() {
    echo "********************"
    echo "* DB_SERVER_HOST: ${DB_SERVER_HOST}"
    echo "* DB_SERVER_PORT: ${DB_SERVER_PORT}"
    echo "* DB_SERVER_DBNAME: ${DB_SERVER_DBNAME}"
    if [ "${USE_DB_ROOT_USER}" == "true" ]; then
        echo "* DB_SERVER_ROOT_USER: ${DB_SERVER_ROOT_USER}"
        echo "* DB_SERVER_ROOT_PASS: ${DB_SERVER_ROOT_PASS}"
    fi
    echo "* DB_SERVER_ZBX_USER: ${DB_SERVER_ZBX_USER}"
    echo "* DB_SERVER_ZBX_PASS: ${DB_SERVER_ZBX_PASS}"
    echo "********************"

    WAIT_TIMEOUT=5

    while [ ! "$(mysqladmin ping -h ${DB_SERVER_HOST} -P ${DB_SERVER_PORT} -u ${DB_SERVER_ROOT_USER} \
                --password="${DB_SERVER_ROOT_PASS}" --silent --connect_timeout=10)" ]; do
        echo "**** MySQL server is not available. Waiting $WAIT_TIMEOUT seconds..."
        sleep $WAIT_TIMEOUT
    done
}

check_db_connect_postgresql() {
    echo "********************"
    echo "* DB_SERVER_HOST: ${DB_SERVER_HOST}"
    echo "* DB_SERVER_PORT: ${DB_SERVER_PORT}"
    echo "* DB_SERVER_DBNAME: ${DB_SERVER_DBNAME}"
    if [ "${USE_DB_ROOT_USER}" == "true" ]; then
        echo "* DB_SERVER_ROOT_USER: ${DB_SERVER_ROOT_USER}"
        echo "* DB_SERVER_ROOT_PASS: ${DB_SERVER_ROOT_PASS}"
    else
        DB_SERVER_ROOT_USER=${DB_SERVER_ZBX_USER}
        DB_SERVER_ROOT_PASS=${DB_SERVER_ZBX_PASS}
    fi
    echo "* DB_SERVER_ZBX_USER: ${DB_SERVER_ZBX_USER}"
    echo "* DB_SERVER_ZBX_PASS: ${DB_SERVER_ZBX_PASS}"
    echo "********************"

    if [ -n "${DB_SERVER_ZBX_PASS}" ]; then
        export PGPASSWORD="${DB_SERVER_ZBX_PASS}"
    fi

    WAIT_TIMEOUT=5

    while [ ! "$(psql -h ${DB_SERVER_HOST} -p ${DB_SERVER_PORT} -U ${DB_SERVER_ROOT_USER} -l -q 2>/dev/null)" ]; do
        echo "**** PostgreSQL server is not available. Waiting $WAIT_TIMEOUT seconds..."
        sleep $WAIT_TIMEOUT
    done

    unset PGPASSWORD
}


mysql_query() {
    query=$1
    local result=""

    result=$(mysql --silent --skip-column-names -h ${DB_SERVER_HOST} -P ${DB_SERVER_PORT} \
             -u ${DB_SERVER_ROOT_USER} --password="${DB_SERVER_ROOT_PASS}" -e "$query")

    echo $result
}

psql_query() {
    query=$1
    db=$2

    local result=""

    if [ -n "${DB_SERVER_ZBX_PASS}" ]; then
        export PGPASSWORD="${DB_SERVER_ZBX_PASS}"
    fi

    result=$(psql -A -q -t  -h ${DB_SERVER_HOST} -p ${DB_SERVER_PORT} \
             -U ${DB_SERVER_ROOT_USER} -c "$query" $db 2>/dev/null);

    unset PGPASSWORD

    echo $result
}

create_db_user_mysql() {
    [ "${CREATE_ZBX_DB_USER}" == "true" ] || return

    echo "** Creating '${DB_SERVER_ZBX_USER}' user in MySQL database"

    USER_EXISTS=$(mysql_query "SELECT 1 FROM mysql.user WHERE user = '${DB_SERVER_ZBX_USER}' AND host = '%'")

    if [ -z "$USER_EXISTS" ]; then
        mysql_query "CREATE USER '${DB_SERVER_ZBX_USER}'@'%' IDENTIFIED BY '${DB_SERVER_ZBX_PASS}'" 1>/dev/null
    else
        mysql_query "SET PASSWORD FOR '${DB_SERVER_ZBX_USER}'@'%' = PASSWORD('${DB_SERVER_ZBX_PASS}');" 1>/dev/null
    fi

    mysql_query "GRANT ALL PRIVILEGES ON $DB_SERVER_DBNAME. * TO '${DB_SERVER_ZBX_USER}'@'%'" 1>/dev/null
}

create_db_user_postgresql() {
    [ "${CREATE_ZBX_DB_USER}" == "true" ] || return

    echo "** Creating '${DB_SERVER_ZBX_USER}' user in PostgreSQL database"

    USER_EXISTS=$(psql_query "SELECT 1 FROM pg_roles WHERE rolname='${DB_SERVER_ZBX_USER}'")

    if [ -z "$USER_EXISTS" ]; then
        psql_query "CREATE USER ${DB_SERVER_ZBX_USER} WITH PASSWORD '${DB_SERVER_ZBX_PASS}'" 1>/dev/null
    else
        psql_query "ALTER USER ${DB_SERVER_ZBX_USER} WITH ENCRYPTED PASSWORD '${DB_SERVER_ZBX_PASS}'" 1>/dev/null
    fi
}

create_db_database_mysql() {
    DB_EXISTS=$(mysql_query "SELECT SCHEMA_NAME FROM information_schema.SCHEMATA WHERE SCHEMA_NAME='${DB_SERVER_DBNAME}'")

    if [ -z ${DB_EXISTS} ]; then
        echo "** Database '${DB_SERVER_DBNAME}' does not exist. Creating..."
        mysql_query "CREATE DATABASE ${DB_SERVER_DBNAME} CHARACTER SET utf8 COLLATE utf8_bin" 1>/dev/null
        # better solution?
        mysql_query "GRANT ALL PRIVILEGES ON $DB_SERVER_DBNAME. * TO '${DB_SERVER_ZBX_USER}'@'%'" 1>/dev/null
    else
        echo "** Database '${DB_SERVER_DBNAME}' already exists. Please be careful with database COLLATE!"
    fi
}

create_db_database_postgresql() {
    DB_EXISTS=$(psql_query "SELECT 1 AS result FROM pg_database WHERE datname='${DB_SERVER_DBNAME}'")

    if [ -z ${DB_EXISTS} ]; then
        echo "** Database '${DB_SERVER_DBNAME}' does not exist. Creating..."
        psql_query "CREATE DATABASE ${DB_SERVER_DBNAME} WITH OWNER ${DB_SERVER_ZBX_USER} ENCODING='UTF8' LC_CTYPE='en_US.utf8' LC_COLLATE='en_US.utf8'" 1>/dev/null
    else
        echo "** Database '${DB_SERVER_DBNAME}' already exists. Please be careful with database owner!"
    fi
}

create_db_schema_mysql() {
    local type=$1

    DBVERSION_TABLE_EXISTS=$(mysql_query "SELECT 1 FROM information_schema.tables WHERE table_schema='${DB_SERVER_DBNAME}' and table_name = 'dbversion'")

    if [ -n "${DBVERSION_TABLE_EXISTS}" ]; then
        echo "** Table '${DB_SERVER_DBNAME}.dbversion' already exists."
        ZBX_DB_VERSION=$(mysql_query "SELECT mandatory FROM ${DB_SERVER_DBNAME}.dbversion")
    fi

    if [ -z "${ZBX_DB_VERSION}" ]; then
        echo "** Creating '${DB_SERVER_DBNAME}' schema in MySQL"

        cat /usr/share/doc/zabbix-$type-mysql/schema.sql | mysql --silent --skip-column-names \
                    -h ${DB_SERVER_HOST} -P ${DB_SERVER_PORT} \
                    -u ${DB_SERVER_ROOT_USER} --password="${DB_SERVER_ROOT_PASS}"  \
                    ${DB_SERVER_DBNAME} 1>/dev/null
        if [ "$type" == "server" ]; then
            echo "** Fill the schema with initial data"
            cat /usr/share/doc/zabbix-$type-mysql/images.sql | mysql --silent --skip-column-names \
                    -h ${DB_SERVER_HOST} -P ${DB_SERVER_PORT} \
                    -u ${DB_SERVER_ROOT_USER} --password="${DB_SERVER_ROOT_PASS}" \
                    ${DB_SERVER_DBNAME} 1>/dev/null
            cat /usr/share/doc/zabbix-$type-mysql/data.sql | mysql --silent --skip-column-names \
                    -h ${DB_SERVER_HOST} -P ${DB_SERVER_PORT} \
                    -u ${DB_SERVER_ROOT_USER} --password="${DB_SERVER_ROOT_PASS}" \
                    ${DB_SERVER_DBNAME} 1>/dev/null
        fi
    fi
}

create_db_schema_postgresql() {
    local type=$1

    DBVERSION_TABLE_EXISTS=$(psql_query "SELECT 1 FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON n.oid = 
                                         c.relnamespace WHERE  n.nspname = 'public' AND c.relname = 'dbversion'" "${DB_SERVER_DBNAME}")

    if [ -n "${DBVERSION_TABLE_EXISTS}" ]; then
        echo "** Table '${DB_SERVER_DBNAME}.dbversion' already exists."
        ZBX_DB_VERSION=$(psql_query "SELECT mandatory FROM public.dbversion" "${DB_SERVER_DBNAME}")
    fi

    if [ -z "${ZBX_DB_VERSION}" ]; then
        echo "** Creating '${DB_SERVER_DBNAME}' schema in PostgreSQL"

        if [ -n "${DB_SERVER_ZBX_PASS}" ]; then
            export PGPASSWORD="${DB_SERVER_ZBX_PASS}"
        fi

        cat /usr/share/doc/zabbix-$type-postgresql/schema.sql | psql -q \
                -h ${DB_SERVER_HOST} -p ${DB_SERVER_PORT} \
                -U ${DB_SERVER_ZBX_USER} ${DB_SERVER_DBNAME} 1>/dev/null
        if [ "$type" == "server" ]; then
            echo "** Fill the schema with initial data"
            cat /usr/share/doc/zabbix-$type-postgresql/images.sql | psql -q \
                -h ${DB_SERVER_HOST} -p ${DB_SERVER_PORT} \
                -U ${DB_SERVER_ZBX_USER} ${DB_SERVER_DBNAME} 1>/dev/null
            cat /usr/share/doc/zabbix-$type-postgresql/data.sql | psql -q \
                -h ${DB_SERVER_HOST} -p ${DB_SERVER_PORT} \
                -U ${DB_SERVER_ZBX_USER} ${DB_SERVER_DBNAME} 1>/dev/null
        fi

        unset PGPASSWORD
    fi
}

prepare_web_server_apache() {
    if [ -d "/etc/apache2/sites-available" ]; then
        APACHE_SITES_DIR=/etc/apache2/sites-available
    elif [ -d "/etc/apache2/conf.d" ]; then
        APACHE_SITES_DIR=/etc/apache2/conf.d
    else
        echo "**** Apache is not available"
        exit 1
    fi

    if [ -f "/usr/sbin/a2dissite" ]; then
        echo "** Disable default site"
        /usr/sbin/a2dissite 000-default 1>/dev/null
        rm -rf "$APACHE_SITES_DIR/*"
    elif [ -f "/etc/apache2/conf.d/default.conf" ]; then
        echo "** Disable default site"
        rm -f "/etc/apache2/conf.d/default.conf"
    fi

    echo "** Adding Zabbix virtual host (HTTP)"
    if [ -f "$ZABBIX_ETC_DIR/apache.conf" ]; then
        ln -s "$ZABBIX_ETC_DIR/apache.conf" "$APACHE_SITES_DIR/zabbix.conf"
        if [ -f "/usr/sbin/a2dissite" ]; then
            /usr/sbin/a2ensite zabbix.conf 1>/dev/null
        fi
    else
        echo "**** Impossible to enable HTTP virtual host"
    fi

    if [ -f "/etc/apache2/conf.d/ssl.conf" ]; then
        rm -f "/etc/apache2/conf.d/ssl.conf"
    fi

    if [ -f "/etc/ssl/apache2/ssl.crt" ] && [ -f "/etc/ssl/apache2/ssl.key" ]; then
        echo "** Enable SSL support for Apache2"
        if [ -f "/usr/sbin/a2enmod" ]; then
            /usr/sbin/a2enmod ssl 1>/dev/null
        fi

        echo "** Adding Zabbix virtual host (HTTPS)"
        if [ -f "$ZABBIX_ETC_DIR/apache_ssl.conf" ]; then
            ln -s "$ZABBIX_ETC_DIR/apache_ssl.conf" "$APACHE_SITES_DIR/zabbix_ssl.conf"
            if [ -f "/usr/sbin/a2dissite" ]; then
                /usr/sbin/a2ensite zabbix_ssl.conf 1>/dev/null
            fi
        else
            echo "**** Impossible to enable HTTPS virtual host"
        fi
    else
        echo "**** Impossible to enable SSL support for Apache2. Certificates are missed."
    fi

    # Change Apache2 logging to stdout and stderr
    if [ -f "/etc/apache2/apache2.conf" ]; then
        sed -ri \
            -e 's!^(\s*CustomLog)\s+\S+!\1 /proc/self/fd/1!g' \
            -e 's!^(\s*ErrorLog)\s+\S+!\1 /proc/self/fd/2!g' \
            "/etc/apache2/apache2.conf"
    fi

    if [ -f "/etc/apache2/httpd.conf" ]; then
        sed -ri \
            -e 's!^(\s*CustomLog)\s+\S+!\1 /proc/self/fd/1!g' \
            -e 's!^(\s*ErrorLog)\s+\S+!\1 /proc/self/fd/2!g' \
            "/etc/apache2/httpd.conf"
    fi

    if [ -f "/etc/apache2/conf-available/other-vhosts-access-log.conf" ]; then
        sed -ri \
            -e 's!^(\s*CustomLog)\s+\S+!\1 /proc/self/fd/1!g' \
            -e 's!^(\s*ErrorLog)\s+\S+!\1 /proc/self/fd/2!g' \
            "/etc/apache2/conf-available/other-vhosts-access-log.conf"
    fi

    if [ -f "/etc/apache2/conf.d/mpm.conf" ]; then
        sed -ri \
            -e 's!^(\s*PidFile)\s+\S+!\1 "/var/run/httpd.pid"!g' \
            "/etc/apache2/conf.d/mpm.conf"
    fi

    if [ -f "/var/run/apache2/apache2.pid" ]; then
        rm -f "/var/run/apache2/apache2.pid"
    fi
}

prepare_web_server_nginx() {
    NGINX_CONFD_DIR="/etc/nginx/conf.d"
    NGINX_SSL_CONFIG="/etc/ssl/nginx"
    PHP_SESSIONS_DIR="/var/lib/php5"

    echo "** Disable default vhosts"
    rm -f $NGINX_CONFD_DIR/*.conf

    echo "** Adding Zabbix virtual host (HTTP)"
    if [ -f "$ZABBIX_ETC_DIR/nginx.conf" ]; then
        ln -s "$ZABBIX_ETC_DIR/nginx.conf" "$NGINX_CONFD_DIR"
    else
        echo "**** Impossible to enable HTTP virtual host"
    fi

    if [ -f "$NGINX_SSL_CONFIG/ssl.crt" ] && [ -f "$NGINX_SSL_CONFIG/ssl.key" ] && [ -f "$NGINX_SSL_CONFIG/dhparam.pem" ]; then
        echo "** Enable SSL support for Nginx"
        if [ -f "$ZABBIX_ETC_DIR/nginx_ssl.conf" ]; then
            ln -s "$ZABBIX_ETC_DIR/nginx_ssl.conf" "$NGINX_CONFD_DIR"
        else
            echo "**** Impossible to enable HTTPS virtual host"
        fi
    else
        echo "**** Impossible to enable SSL support for Nginx. Certificates are missed."
    fi

    if [ -d "/var/log/nginx/" ]; then
        ln -sf /dev/fd/2 /var/log/nginx/error.log
    fi

    ln -sf /dev/fd/2 /var/log/php5-fpm.log
}

clear_deploy() {
    local type=$1
    echo "** Cleaning the system"

    [ "$type" != "dev" ] && return
}

update_zbx_config() {
    local type=$1
    local db_type=$2

    echo "** Preparing Zabbix $type configuration file"

    ZBX_CONFIG=$ZABBIX_ETC_DIR/zabbix_$type.conf

    if [ "$type" == "proxy" ]; then
        update_config_var $ZBX_CONFIG "ProxyMode" "${ZBX_PROXYMODE}"
        update_config_var $ZBX_CONFIG "Server" "${ZBX_SERVER_HOST}"
        update_config_var $ZBX_CONFIG "ServerPort" "${ZBX_SERVER_PORT}"
        update_config_var $ZBX_CONFIG "Hostname" "${ZBX_HOSTNAME:-"zabbix-proxy-"$db_type}"
        update_config_var $ZBX_CONFIG "HostnameItem" "${ZBX_HOSTNAMEITEM}"
    fi

    if [ $type == "proxy" ] && [ "${ZBX_ADD_SERVER}" = "true" ]; then
        update_config_var $ZBX_CONFIG "ListenPort" "10061"
    else
        update_config_var $ZBX_CONFIG "ListenPort"
    fi
    update_config_var $ZBX_CONFIG "SourceIP" "${ZBX_SOURCEIP}"
    update_config_var $ZBX_CONFIG "LogType" "console"
    update_config_var $ZBX_CONFIG "LogFile"
    update_config_var $ZBX_CONFIG "LogFileSize"
    update_config_var $ZBX_CONFIG "PidFile"

    update_config_var $ZBX_CONFIG "DebugLevel" "${ZBX_DEBUGLEVEL}"

    if [ "$db_type" == "sqlite3" ]; then
        update_config_var $ZBX_CONFIG "DBHost"
        update_config_var $ZBX_CONFIG "DBName" "/var/lib/zabbix/zabbix_proxy_db"
        update_config_var $ZBX_CONFIG "DBUser"
        update_config_var $ZBX_CONFIG "DBPort"
        update_config_var $ZBX_CONFIG "DBPassword"
    else
        update_config_var $ZBX_CONFIG "DBHost" "${DB_SERVER_HOST}"
        update_config_var $ZBX_CONFIG "DBName" "${DB_SERVER_DBNAME}"
        update_config_var $ZBX_CONFIG "DBUser" "${DB_SERVER_ZBX_USER}"
        update_config_var $ZBX_CONFIG "DBPort" "${DB_SERVER_PORT}"
        update_config_var $ZBX_CONFIG "DBPassword" "${DB_SERVER_ZBX_PASS}"
    fi

    if [ "$type" == "proxy" ]; then
        update_config_var $ZBX_CONFIG "ProxyLocalBuffer" "${ZBX_PROXYLOCALBUFFER}"
        update_config_var $ZBX_CONFIG "ProxyOfflineBuffer" "${ZBX_PROXYOFFLINEBUFFER}"
        update_config_var $ZBX_CONFIG "HeartbeatFrequency" "${ZBX_PROXYHEARTBEATFREQUENCY}"
        update_config_var $ZBX_CONFIG "ConfigFrequency" "${ZBX_CONFIGFREQUENCY}"
        update_config_var $ZBX_CONFIG "DataSenderFrequency" "${ZBX_DATASENDERFREQUENCY}"
    fi

    update_config_var $ZBX_CONFIG "StartPollers" "${ZBX_STARTPOLLERS}"
    update_config_var $ZBX_CONFIG "StartIPMIPollers" "${ZBX_IPMIPOLLERS}"
    update_config_var $ZBX_CONFIG "StartPollersUnreachable" "${ZBX_STARTPOLLERSUNREACHABLE}"
    update_config_var $ZBX_CONFIG "StartTrappers" "${ZBX_STARTTRAPPERS}"
    update_config_var $ZBX_CONFIG "StartPingers" "${ZBX_STARTPINGERS}"
    update_config_var $ZBX_CONFIG "StartDiscoverers" "${ZBX_STARTDISCOVERERS}"
    update_config_var $ZBX_CONFIG "StartHTTPPollers" "${ZBX_STARTHTTPPOLLERS}"

    if [ "$type" == "server" ]; then
        update_config_var $ZBX_CONFIG "StartTimers" "${ZBX_STARTTIMERS}"
        update_config_var $ZBX_CONFIG "StartEscalators" "${ZBX_STARTESCALATORS}"
    fi

    ZBX_JAVAGATEWAY_ENABLE=${ZBX_JAVAGATEWAY_ENABLE:-"false"}
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

    ZBX_ENABLE_SNMP_TRAPS=${ZBX_ENABLE_SNMP_TRAPS:-"false"}
    if [ "${ZBX_ENABLE_SNMP_TRAPS}" == "true" ]; then
        update_config_var $ZBX_CONFIG "SNMPTrapperFile" "${ZABBIX_USER_HOME_DIR}/snmptraps/snmptraps.log"
        update_config_var $ZBX_CONFIG "StartSNMPTrapper" "1"
    else
        update_config_var $ZBX_CONFIG "SNMPTrapperFile"
        update_config_var $ZBX_CONFIG "StartSNMPTrapper"
    fi

    update_config_var $ZBX_CONFIG "HousekeepingFrequency" "${ZBX_HOUSEKEEPINGFREQUENCY}"
    if [ "$type" == "server" ]; then
        update_config_var $ZBX_CONFIG "MaxHousekeeperDelete" "${ZBX_MAXHOUSEKEEPERDELETE}"
        update_config_var $ZBX_CONFIG "SenderFrequency" "${ZBX_SENDERFREQUENCY}"
    fi

    update_config_var $ZBX_CONFIG "CacheSize" "${ZBX_CACHESIZE}"

    if [ "$type" == "server" ]; then
        update_config_var $ZBX_CONFIG "CacheUpdateFrequency" "${ZBX_CACHEUPDATEFREQUENCY}"
    fi

    update_config_var $ZBX_CONFIG "StartDBSyncers" "${ZBX_STARTDBSYNCERS}"
    update_config_var $ZBX_CONFIG "HistoryCacheSize" "${ZBX_HISTORYCACHESIZE}"
    update_config_var $ZBX_CONFIG "HistoryIndexCacheSize" "${ZBX_HISTORYINDEXCACHESIZE}"

    if [ "$type" == "server" ]; then 
        update_config_var $ZBX_CONFIG "TrendCacheSize" "${ZBX_TRENDCACHESIZE}"
        update_config_var $ZBX_CONFIG "ValueCacheSize" "${ZBX_VALUECACHESIZE}"
    fi

    update_config_var $ZBX_CONFIG "Timeout" "${ZBX_TIMEOUT}"
    update_config_var $ZBX_CONFIG "TrapperTimeout" "${ZBX_TRAPPERIMEOUT}"
    update_config_var $ZBX_CONFIG "UnreachablePeriod" "${ZBX_UNREACHABLEPERIOD}"
    update_config_var $ZBX_CONFIG "UnavailableDelay" "${ZBX_UNAVAILABLEDELAY}"
    update_config_var $ZBX_CONFIG "UnreachableDelay" "${ZBX_UNREACHABLEDELAY}"

    update_config_var $ZBX_CONFIG "AlertScriptsPath" "/usr/lib/zabbix/alertscripts"
    update_config_var $ZBX_CONFIG "ExternalScripts" "/usr/lib/zabbix/externalscripts"

    # Possible few fping locations
    if [ -f "/usr/bin/fping" ]; then
        update_config_var $ZBX_CONFIG "FpingLocation" "/usr/bin/fping"
    else
        update_config_var $ZBX_CONFIG "FpingLocation" "/usr/sbin/fping"
    fi
    if [ -f "/usr/bin/fping6" ]; then
        update_config_var $ZBX_CONFIG "Fping6Location" "/usr/bin/fping6"
    else
        update_config_var $ZBX_CONFIG "Fping6Location" "/usr/sbin/fping6"
    fi

    update_config_var $ZBX_CONFIG "SSHKeyLocation" "$ZABBIX_USER_HOME_DIR/ssh_keys"
    update_config_var $ZBX_CONFIG "LogSlowQueries" "${ZBX_LOGSLOWQUERIES}"

    if [ "$type" == "server" ]; then 
        update_config_var $ZBX_CONFIG "StartProxyPollers" "${ZBX_STARTPROXYPOLLERS}"
        update_config_var $ZBX_CONFIG "ProxyConfigFrequency" "${ZBX_PROXYCONFIGFREQUENCY}"
        update_config_var $ZBX_CONFIG "ProxyDataFrequency" "${ZBX_PROXYDATAFREQUENCY}"
    fi

    update_config_var $ZBX_CONFIG "SSLCertLocation" "$ZABBIX_USER_HOME_DIR/ssl/certs/"
    update_config_var $ZBX_CONFIG "SSLKeyLocation" "$ZABBIX_USER_HOME_DIR/ssl/keys/"
    update_config_var $ZBX_CONFIG "SSLCALocation" "$ZABBIX_USER_HOME_DIR/ssl/ssl_ca/"
    update_config_var $ZBX_CONFIG "LoadModulePath" "$ZABBIX_USER_HOME_DIR/modules/"
    update_config_multiple_var $ZBX_CONFIG "LoadModule" "${ZBX_LOADMODULE}"

    if [ "$type" == "proxy" ]; then
        update_config_var $ZBX_CONFIG "TLSConnect" "${ZBX_TLSCONNECT}"
        update_config_var $ZBX_CONFIG "TLSAccept" "${ZBX_TLSACCEPT}"
    fi
    update_config_var $ZBX_CONFIG "TLSCAFile" "${ZBX_TLSCAFILE}"
    update_config_var $ZBX_CONFIG "TLSCRLFile" "${ZBX_TLSCRLFILE}"

    if [ "$type" == "proxy" ]; then
        update_config_var $ZBX_CONFIG "TLSServerCertIssuer" "${ZBX_TLSSERVERCERTISSUER}"
        update_config_var $ZBX_CONFIG "TLSServerCertSubject" "${ZBX_TLSSERVERCERTSUBJECT}"
    fi

    update_config_var $ZBX_CONFIG "TLSCertFile" "${ZBX_TLSCERTFILE}"
    update_config_var $ZBX_CONFIG "TLSKeyFile" "${ZBX_TLSKEYFILE}"

    if [ "$type" == "proxy" ]; then
        update_config_var $ZBX_CONFIG "TLSPSKIdentity" "${ZBX_TLSPSKIDENTITY}"
        update_config_var $ZBX_CONFIG "TLSPSKFile" "${ZBX_TLSPSKFILE}"
    fi
}


prepare_zbx_web_config() {
    local db_type=$1
    local server_name=""

    echo "** Preparing Zabbix frontend configuration file"

    ZBX_WEB_CONFIG="$ZABBIX_ETC_DIR/web/zabbix.conf.php"

    if [ -f "/usr/share/zabbix/conf/zabbix.conf.php" ]; then
        rm -f "/usr/share/zabbix/conf/zabbix.conf.php"
    fi

    ln -s "$ZBX_WEB_CONFIG" "/usr/share/zabbix/conf/zabbix.conf.php"

    # Different places of PHP configuration file
    if [ -f "/etc/php5/conf.d/99-zabbix.ini" ]; then
        PHP_CONFIG_FILE="/etc/php5/conf.d/99-zabbix.ini"
    elif [ -f "/etc/php5/fpm/conf.d/99-zabbix.ini" ]; then
        PHP_CONFIG_FILE="/etc/php5/fpm/conf.d/99-zabbix.ini"
    elif [ -f "/etc/php5/apache2/conf.d/99-zabbix.ini" ]; then
        PHP_CONFIG_FILE="/etc/php5/apache2/conf.d/99-zabbix.ini"
    elif [ -f "/etc/php/7.0/apache2/conf.d/99-zabbix.ini" ]; then
        PHP_CONFIG_FILE="/etc/php/7.0/apache2/conf.d/99-zabbix.ini"
    elif [ -f "/etc/php/7.0/fpm/conf.d/99-zabbix.ini" ]; then
        PHP_CONFIG_FILE="/etc/php/7.0/fpm/conf.d/99-zabbix.ini"
    fi

    if [ -n "$PHP_CONFIG_FILE" ]; then
        update_config_var "$PHP_CONFIG_FILE" "max_execution_time" "${ZBX_MAXEXECUTIONTIME:-"600"}"
        update_config_var "$PHP_CONFIG_FILE" "memory_limit" "${ZBX_MEMORYLIMIT:-"128M"}" 
        update_config_var "$PHP_CONFIG_FILE" "post_max_size" "${ZBX_POSTMAXSIZE:-"16M"}"
        update_config_var "$PHP_CONFIG_FILE" "upload_max_filesize" "${ZBX_UPLOADMAXFILESIZE:-"2M"}"
        update_config_var "$PHP_CONFIG_FILE" "max_input_time" "${ZBX_MAXINPUTTIME:-"300"}"
        update_config_var "$PHP_CONFIG_FILE" "date.timezone" "${PHP_TZ}"
    else
        echo "**** Zabbix related PHP configuration file not found"
    fi

    # Escaping "/" character in parameter value
    server_name=${ZBX_SERVER_NAME//\//\\/}

    sed -i \
        -e "s/{DB_SERVER_HOST}/${DB_SERVER_HOST}/g" \
        -e "s/{DB_SERVER_PORT}/${DB_SERVER_PORT}/g" \
        -e "s/{DB_SERVER_DBNAME}/${DB_SERVER_DBNAME}/g" \
        -e "s/{DB_SERVER_USER}/${DB_SERVER_ZBX_USER}/g" \
        -e "s/{DB_SERVER_PASS}/${DB_SERVER_ZBX_PASS}/g" \
        -e "s/{ZBX_SERVER_HOST}/${ZBX_SERVER_HOST}/g" \
        -e "s/{ZBX_SERVER_PORT}/${ZBX_SERVER_PORT}/g" \
        -e "s/{ZBX_SERVER_NAME}/$server_name/g" \
    "$ZBX_WEB_CONFIG"

    [ "$db_type" = "postgresql" ] && sed -i "s/MYSQL/POSTGRESQL/g" "$ZBX_WEB_CONFIG"
}

prepare_zbx_agent_config() {
    echo "** Preparing Zabbix agent configuration file"

    ZBX_AGENT_CONFIG=$ZABBIX_ETC_DIR/zabbix_agentd.conf

    ZBX_PASSIVESERVERS=${ZBX_PASSIVESERVERS:-""}
    ZBX_ACTIVESERVERS=${ZBX_ACTIVESERVERS:-""}

    [ -n "$ZBX_PASSIVESERVERS" ] && ZBX_PASSIVESERVERS=","$ZBX_PASSIVESERVERS

    ZBX_PASSIVESERVERS=$ZBX_SERVER_HOST$ZBX_PASSIVESERVERS

    [ -n "$ZBX_ACTIVESERVERS" ] && ZBX_ACTIVESERVERS=","$ZBX_ACTIVESERVERS

    ZBX_ACTIVESERVERS=$ZBX_SERVER_HOST":"$ZBX_SERVER_PORT$ZBX_ACTIVESERVERS

    update_config_var $ZBX_AGENT_CONFIG "PidFile"
    update_config_var $ZBX_AGENT_CONFIG "LogType" "console"
    update_config_var $ZBX_AGENT_CONFIG "LogFile"
    update_config_var $ZBX_AGENT_CONFIG "LogFileSize"
    update_config_var $ZBX_AGENT_CONFIG "DebugLevel" "${ZBX_DEBUGLEVEL}"
    update_config_var $ZBX_AGENT_CONFIG "SourceIP"
    update_config_var $ZBX_AGENT_CONFIG "EnableRemoteCommands" "${ZBX_ENABLEREMOTECOMMANDS}"
    update_config_var $ZBX_AGENT_CONFIG "LogRemoteCommands" "${ZBX_LOGREMOTECOMMANDS}"

    ZBX_PASSIVE_ALLOW=${ZBX_PASSIVE_ALLOW:-"true"}
    if [ "$ZBX_PASSIVE_ALLOW" == "true" ]; then
        echo "** Using '$ZBX_PASSIVESERVERS' servers for passive checks"
        update_config_var $ZBX_AGENT_CONFIG "Server" "${ZBX_PASSIVESERVERS}"
    else
        update_config_var $ZBX_AGENT_CONFIG "Server"
    fi

    update_config_var $ZBX_AGENT_CONFIG "ListenPort"
    update_config_var $ZBX_AGENT_CONFIG "ListenIP" "${ZBX_LISTENIP}"
    update_config_var $ZBX_AGENT_CONFIG "StartAgents" "${ZBX_STARTAGENTS}"

    ZBX_ACTIVE_ALLOW=${ZBX_ACTIVE_ALLOW:-"true"}
    if [ "$ZBX_ACTIVE_ALLOW" == "true" ]; then
        echo "** Using '$ZBX_ACTIVESERVERS' servers for active checks"
        update_config_var $ZBX_AGENT_CONFIG "ServerActive" "${ZBX_ACTIVESERVERS}"
    else
        update_config_var $ZBX_AGENT_CONFIG "ServerActive"
    fi

    update_config_var $ZBX_AGENT_CONFIG "Hostname" "${ZBX_HOSTNAME}"
    update_config_var $ZBX_AGENT_CONFIG "HostnameItem" "${ZBX_HOSTNAMEITEM}"
    update_config_var $ZBX_AGENT_CONFIG "HostMetadata" "${ZBX_METADATA}"
    update_config_var $ZBX_AGENT_CONFIG "HostMetadataItem" "${ZBX_METADATAITEM}"
    update_config_var $ZBX_AGENT_CONFIG "RefreshActiveChecks" "${ZBX_REFRESHACTIVECHECKS}"
    update_config_var $ZBX_AGENT_CONFIG "BufferSend" "${ZBX_BUFFERSEND}"
    update_config_var $ZBX_AGENT_CONFIG "BufferSize" "${ZBX_BUFFERSIZE}"
    update_config_var $ZBX_AGENT_CONFIG "MaxLinesPerSecond" "${ZBX_MAXLINESPERSECOND}"
#    update_config_multiple_var $ZBX_AGENT_CONFIG "Alias" ${ZBX_ALIAS}
    update_config_var $ZBX_AGENT_CONFIG "Timeout" "${ZBX_TIMEOUT}"
    update_config_var $ZBX_AGENT_CONFIG "Include" "/etc/zabbix/zabbix_agentd.d/"
    update_config_var $ZBX_AGENT_CONFIG "UnsafeUserParameters" "${ZBX_UNSAFEUSERPARAMETERS}"
    update_config_var $ZBX_AGENT_CONFIG "LoadModulePath" "$ZABBIX_USER_HOME_DIR/modules/"
    update_config_multiple_var $ZBX_AGENT_CONFIG "LoadModule" "${ZBX_LOADMODULE}"
    update_config_var $ZBX_AGENT_CONFIG "TLSConnect" "${ZBX_TLSCONNECT}"
    update_config_var $ZBX_AGENT_CONFIG "TLSAccept" "${ZBX_TLSACCEPT}"
    update_config_var $ZBX_AGENT_CONFIG "TLSCAFile" "${ZBX_TLSCAFILE}"
    update_config_var $ZBX_AGENT_CONFIG "TLSCRLFile" "${ZBX_TLSCRLFILE}"
    update_config_var $ZBX_AGENT_CONFIG "TLSServerCertIssuer" "${ZBX_TLSSERVERCERTISSUER}"
    update_config_var $ZBX_AGENT_CONFIG "TLSServerCertSubject" "${ZBX_TLSSERVERCERTSUBJECT}"
    update_config_var $ZBX_AGENT_CONFIG "TLSCertFile" "${ZBX_TLSCERTFILE}"
    update_config_var $ZBX_AGENT_CONFIG "TLSKeyFile" "${ZBX_TLSKEYFILE}"
    update_config_var $ZBX_AGENT_CONFIG "TLSPSKIdentity" "${ZBX_TLSPSKIDENTITY}"
    update_config_var $ZBX_AGENT_CONFIG "TLSPSKFile" "${ZBX_TLSPSKFILE}"
}

prepare_java_gateway_config() {
    echo "** Preparing Zabbix Java Gateway log configuration file"

    ZBX_GATEWAY_CONFIG=$ZABBIX_ETC_DIR/zabbix_java_gateway_logback.xml

    if [ -n "${ZBX_DEBUGLEVEL}" ]; then
        echo "Updating $ZBX_GATEWAY_CONFIG 'DebugLevel' parameter: '${ZBX_DEBUGLEVEL}'... updated"
        if [ -f "$ZBX_GATEWAY_CONFIG" ]; then
            sed -i -e "/^.*<root level=/s/=.*/=\"${ZBX_DEBUGLEVEL}\">/" "$ZBX_GATEWAY_CONFIG"
        else
            echo "**** Zabbix Java Gateway log configuration file '$ZBX_GATEWAY_CONFIG' not found"
        fi
    fi
}

prepare_agent() {
    echo "** Preparing Zabbix agent"
    prepare_zbx_agent_config
}

prepare_server() {
    local db_type=$1

    echo "** Preparing Zabbix server"

    check_variables_$db_type "server"
    check_db_connect_$db_type
    create_db_user_$db_type
    create_db_database_$db_type
    create_db_schema_$db_type "server"

    update_zbx_config "server" "$db_type"
}

prepare_proxy() {
    local db_type=$1

    echo "Preparing Zabbix proxy"

    if [ "$db_type" != "sqlite3" ]; then
        check_variables_$db_type "proxy"
        check_db_connect_$db_type
        create_db_user_$db_type
        create_db_database_$db_type
        create_db_schema_$db_type "proxy"
    fi

    update_zbx_config "proxy" $db_type
}

prepare_web() {
    local web_server=$1
    local db_type=$2

    echo "** Preparing Zabbix web-interface"

    check_variables_$db_type
    check_db_connect_$db_type
    prepare_web_server_$web_server
    prepare_zbx_web_config $db_type
}

prepare_java_gateway() {
    echo "** Preparing Zabbix Java Gateway"

    prepare_java_gateway_config
}

#################################################

if [ ! -n "$zbx_type" ]; then
    echo "**** Type of Zabbix component is not specified"
    exit 1
elif [ "$zbx_type" == "dev" ]; then
    echo "** Deploying Zabbix installation from SVN"
else
    if [ ! -n "$zbx_db_type" ]; then
        echo "**** Database type of Zabbix $zbx_type is not specified"
        exit 1
    fi

    if [ -n "$zbx_db_type" ]; then
        if [ -n "$zbx_opt_type" ]; then
            echo "** Deploying Zabbix $zbx_type ($zbx_opt_type) with $zbx_db_type database"
        else
            echo "** Deploying Zabbix $zbx_type with $zbx_db_type database"
        fi
    else
        echo "** Deploying Zabbix $zbx_type"
    fi
fi

prepare_system "$zbx_type" "$zbx_opt_type"

[ "$zbx_type" == "server" ] && prepare_server $zbx_db_type
[ "${ZBX_ADD_SERVER}" == "true" ] && prepare_server ${ZBX_MAIN_DB}

[ "$zbx_type" == "proxy" ] && prepare_proxy $zbx_db_type
[ "${ZBX_ADD_PROXY}" == "true" ] && prepare_proxy ${ZBX_PROXY_DB}

[ "$zbx_type" == "frontend" ] && prepare_web $zbx_opt_type $zbx_db_type
[ "${ZBX_ADD_WEB}" == "true" ] && prepare_web ${ZBX_WEB_SERVER} ${ZBX_MAIN_DB}

[ "$zbx_type" == "agentd" ] && prepare_agent
[ "${ZBX_ADD_AGENT}" == "true" ] && prepare_agent

[ "$zbx_type" == "java-gateway" ] && prepare_java_gateway
[ "${ZBX_ADD_JAVA_GATEWAY}" == "true" ] && prepare_java_gateway

clear_deploy "$zbx_type"

echo "########################################################"

echo "** Executing supervisord"
exec /usr/bin/supervisord -c /etc/supervisor/supervisord.conf

#################################################
