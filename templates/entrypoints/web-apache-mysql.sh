#!/usr/bin/env bash

set -euo pipefail

readonly ENTRYPOINT_LIBS="/usr/lib/docker-entrypoint"
source "${ENTRYPOINT_LIBS}/bootstrap.sh"

source "${ENTRYPOINT_LIBS}/apache.sh"
source "${ENTRYPOINT_LIBS}/mysql.sh"
source "${ENTRYPOINT_LIBS}/php.sh"
source "${ENTRYPOINT_LIBS}/web.sh"

: "${ZBX_SERVER_NAME:=Zabbix docker}"
: "${PHP_TZ:=Europe/Riga}"

: "${DAEMON_USER:=apache}"
: "${DAEMON_GROUP:=apache}"

prepare_runtime_commands() {
    APACHE_BIN="${APACHE_BIN:-/usr/sbin/httpd}"
    PHP_FPM_BIN="${PHP_FPM_BIN:-/usr/sbin/php-fpm}"
    PHP_FPM_CONFIG="${PHP_FPM_CONFIG:-/etc/php-fpm.conf}"

    APACHE_ARGS=(-D FOREGROUND)
    PHP_FPM_ARGS=(--nodaemonize --fpm-config "${PHP_FPM_CONFIG}")
}

start_web_stack() {
    local php_pid=""
    local web_pid=""
    local exit_code=0

    term_handler() {
        [[ -n "$php_pid" ]] && kill "$php_pid" 2>/dev/null || true
        [[ -n "$web_pid" ]] && kill "$web_pid" 2>/dev/null || true
        [[ -n "$php_pid" ]] && wait "$php_pid" 2>/dev/null || true
        [[ -n "$web_pid" ]] && wait "$web_pid" 2>/dev/null || true
    }

    trap term_handler TERM INT

    "${PHP_FPM_BIN}" "${PHP_FPM_ARGS[@]}" &
    php_pid=$!

    "${APACHE_BIN}" "${APACHE_ARGS[@]}" &
    web_pid=$!

    if wait -n "$php_pid" "$web_pid"; then
        exit_code=0
    else
        exit_code=$?
    fi

    term_handler
    return "$exit_code"
}

prepare_service() {
    info "** Preparing Zabbix web-interface (Apache) with MySQL database"

    check_db_variables "zabbix"
    check_db_connect "true"
    prepare_php_config "MYSQL"
    prepare_web_server
    prepare_zbx_config
}

#################################################

if [ $# -eq 0 ]; then
    prepare_service
    prepare_runtime_commands
    start_web_stack
    exit $?
fi

exec "$@"

#################################################
