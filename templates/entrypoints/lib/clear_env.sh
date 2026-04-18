# shellcheck shell=bash

clear_zbx_env() {
    [[ "${ZBX_CLEAR_ENV:-}" == "false" ]] && return

    local env_var

    for env_var in "${!ZBX_@}" "${!DB_@}" "${!MYSQL_@}" "${!POSTGRES_@}"; do
        [[ -n "$env_var" ]] || continue
        unset "$env_var" 2>/dev/null || true
    done
}
