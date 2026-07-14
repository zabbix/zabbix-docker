# shellcheck shell=bash

entrypoint_hook_info() {
    printf '%s [info]: %s\n' "$(date -u +'%Y-%m-%dT%H:%M:%SZ')" "$*"
}

run_entrypoint_hooks() {
    local directory="${ZABBIX_USER_HOME_DIR:?}/entrypoint.d"
    local hook
    local LC_ALL=C

    [[ -d "$directory" ]] || return 0

    for hook in "$directory"/*; do
        [[ -f "$hook" ]] || continue

        case "$hook" in
            *.sh)
                entrypoint_hook_info "** Running entrypoint hook: $hook"
                /bin/sh "$hook"
                ;;
            *)
                if [[ -x "$hook" ]]; then
                    entrypoint_hook_info "** Running entrypoint hook: $hook"
                    "$hook"
                fi
                ;;
        esac
    done
}
