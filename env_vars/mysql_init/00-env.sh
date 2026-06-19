#!/usr/bin/env bash
set -e

read_file_env() {
    local var="$1"
    local file_var="${var}_FILE"

    if [ -n "${!file_var:-}" ]; then
        export "$var"="$(cat "${!file_var}")"
    fi
}

sql_quote() {
    printf "%s" "$1" | sed "s/'/''/g"
}

default_role_sql() {
    local role="$1"
    local user="$2"
    local host="$3"

    if [ "$ZBX_MYSQL_FLAVOR" = "mariadb" ]; then
#        printf "SET DEFAULT ROLE '%s' FOR '%s'@'%s';\n" "$role" "$user" "$host"
        printf "SET DEFAULT ROLE ALL FOR '%s'@'%s';\n" "$user" "$host"
    else
#        printf "SET DEFAULT ROLE '%s' TO '%s'@'%s';\n" "$role" "$user" "$host"
        printf "SET DEFAULT ROLE ALL TO '%s'@'%s';\n" "$user" "$host"
    fi
}

mysql_only_session_variables_admin_sql() {
    if [ "$ZBX_MYSQL_FLAVOR" != "mariadb" ]; then
        printf "GRANT SESSION_VARIABLES_ADMIN ON *.* TO '%s';\n" "$ZBX_PARTITION_DB_ROLE"
    fi
}

identified_by_sql() {
    local password="$1"

    if [ "$ZBX_MYSQL_FLAVOR" = "mariadb" ]; then
        printf "IDENTIFIED BY '%s'" "$password"
    else
        printf "IDENTIFIED WITH caching_sha2_password BY '%s'" "$password"
    fi
}

read_file_env ZBX_SERVER_DB_USER
read_file_env ZBX_SERVER_DB_PASSWORD
read_file_env ZBX_WEB_DB_USER
read_file_env ZBX_WEB_DB_PASSWORD
read_file_env ZBX_PROXY_DB_USER
read_file_env ZBX_PROXY_DB_PASSWORD
read_file_env ZBX_BACKUP_DB_USER
read_file_env ZBX_BACKUP_DB_PASSWORD
read_file_env ZBX_PARTITION_DB_USER
read_file_env ZBX_PARTITION_DB_PASSWORD

: "${ZBX_SERVER_DB_NAME:=zabbix}"
: "${ZBX_PROXY_DB_NAME:=zabbix_proxy}"
: "${ZBX_SERVER_DB_USER_HOST:=%}"
: "${ZBX_WEB_DB_USER_HOST:=%}"
: "${ZBX_PROXY_DB_USER_HOST:=%}"
: "${ZBX_BACKUP_DB_USER_HOST:=%}"
: "${ZBX_PARTITION_DB_USER_HOST:=%}"
: "${ZBX_SERVER_DB_ROLE:=zbx_srv}"
: "${ZBX_WEB_DB_ROLE:=zbx_web}"
: "${ZBX_PROXY_DB_ROLE:=zbx_proxy}"
: "${ZBX_BACKUP_DB_ROLE:=zbx_bckp}"
: "${ZBX_PARTITION_DB_ROLE:=zbx_part}"
: "${ZBX_SERVER_DB_USER:=zabbix_server}"
: "${ZBX_SERVER_DB_PASSWORD:=zabbix}"
: "${ZBX_WEB_DB_USER:=zabbix_web}"
: "${ZBX_WEB_DB_PASSWORD:=zabbix}"
: "${ZBX_PROXY_DB_USER:=zabbix_proxy}"
: "${ZBX_PROXY_DB_PASSWORD:=zabbix}"
: "${ZBX_BACKUP_DB_USER:=zabbix_bckp}"
: "${ZBX_BACKUP_DB_PASSWORD:=zabbix}"
: "${ZBX_PARTITION_DB_USER:=zabbix_part}"
: "${ZBX_PARTITION_DB_PASSWORD:=zabbix}"

ZBX_MYSQL_FLAVOR="$(
    docker_process_sql --skip-column-names --database=mysql \
        -e "SELECT IF(VERSION() LIKE '%MariaDB%', 'mariadb', 'mysql');" |
        tr -d '[:space:]'
)"

ZBX_SERVER_DB_PASSWORD_SQL="$(sql_quote "$ZBX_SERVER_DB_PASSWORD")"
ZBX_WEB_DB_PASSWORD_SQL="$(sql_quote "$ZBX_WEB_DB_PASSWORD")"
ZBX_PROXY_DB_PASSWORD_SQL="$(sql_quote "$ZBX_PROXY_DB_PASSWORD")"
ZBX_BACKUP_DB_PASSWORD_SQL="$(sql_quote "$ZBX_BACKUP_DB_PASSWORD")"
ZBX_PARTITION_DB_PASSWORD_SQL="$(sql_quote "$ZBX_PARTITION_DB_PASSWORD")"
