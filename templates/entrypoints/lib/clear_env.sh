# shellcheck shell=bash

clear_zbx_env() {
    [[ "${ZBX_CLEAR_ENV:-}" == "false" ]] && return

    local env_var
    while IFS='=' read -r env_var _; do
        case "$env_var" in
            ZABBIX_*) unset "$env_var" ;;
            DB_*) unset "$env_var" ;;
            MYSQL_*) unset "$env_var" ;;
            POSTGRES_*) unset "$env_var" ;;
        esac
    done < <(env)
}
