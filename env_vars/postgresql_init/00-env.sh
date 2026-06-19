#!/usr/bin/env bash
set -e

read_file_env() {
    local var="$1"
    local file_var="${var}_FILE"

    if [ -n "${!file_var:-}" ]; then
        export "$var"="$(cat "${!file_var}")"
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

: "${ZBX_SERVER_DB_NAME:=zabbix}"
: "${ZBX_PROXY_DB_NAME:=zabbix_proxy}"
: "${ZBX_SERVER_DB_SCHEMA:=public}"
: "${ZBX_PROXY_DB_SCHEMA:=public}"
: "${ZBX_SERVER_DB_ROLE:=zbx_srv}"
: "${ZBX_WEB_DB_ROLE:=zbx_web}"
: "${ZBX_PROXY_DB_ROLE:=zbx_proxy}"
: "${ZBX_BACKUP_DB_ROLE:=zbx_bckp}"
: "${ZBX_SERVER_DB_USER:=zabbix}"
: "${ZBX_SERVER_DB_PASSWORD:=zabbix}"
: "${ZBX_WEB_DB_USER:=zabbix_web}"
: "${ZBX_WEB_DB_PASSWORD:=zabbix}"
: "${ZBX_PROXY_DB_USER:=zabbix_proxy}"
: "${ZBX_PROXY_DB_PASSWORD:=zabbix}"
: "${ZBX_BACKUP_DB_USER:=zabbix_bckp}"
: "${ZBX_BACKUP_DB_PASSWORD:=zabbix}"
